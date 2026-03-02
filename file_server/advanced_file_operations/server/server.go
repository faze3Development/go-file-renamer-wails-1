// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/unrolled/secure"
)

// Server holds the dependencies for the HTTP server.
type Server struct {
	processingService *operations.Orchestrator
	logger            *slog.Logger
	router            *chi.Mux
	httpServer        *http.Server
}

// NewServer creates and configures a new server instance.
func NewServer(processingService *operations.Orchestrator, logger *slog.Logger) *Server {
	s := &Server{
		processingService: processingService,
		logger:            logger,
		router:            chi.NewRouter(),
	}

	// Setup middleware
	s.setupMiddleware(s.router)

	s.httpServer = &http.Server{
		Handler:           s.router,
		ReadHeaderTimeout: 10 * time.Second, // Prevent Slowloris attacks
	}

	return s
}

// GetRouter returns the server's router.
func (s *Server) GetRouter() *chi.Mux {
	return s.router
}

// Start begins listening for HTTP requests.
func (s *Server) Start(addr string) error {
	s.logger.Info("🚀 Starting Cloud File Bulk Renamer API server", "address", addr)
	s.httpServer.Addr = addr
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
}

// setupMiddleware configures global middleware for the router.
func (s *Server) setupMiddleware(router *chi.Mux) {
	// Get allowed origins from environment variable, split by comma
	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsEnv == "" {
		// Provide a sensible default for development if the env var is not set.
		allowedOriginsEnv = "http://localhost:3000,http://localhost:3001,http://localhost:5173"
	}
	allowedOrigins := strings.Split(allowedOriginsEnv, ",")

	// Get the environment once at startup for the security middleware.
	env := os.Getenv("ENV")

	// --- Global Middleware ---
	router.Use(cors.Handler(cors.Options{
		// Use the dynamically loaded origins.
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Requested-With", "Origin"},
		ExposedHeaders:   []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300, // Cache preflight for 5 minutes
	}))

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(errors.PanicRecovery(s.logger, ""))
	router.Use(middleware.Timeout(30 * time.Second))
	router.Use(requestSizeLimiter(common.GetMaxFileSizeFromEnv(common.MaxPayload32MB)))
	router.Use(loggerMiddleware(s.logger))
	router.Use(securityMiddleware(env))

	// --- Public Routes ---
	router.Get("/health", healthCheckHandler)
	router.Options("/health", healthCheckHandler) // Handle CORS preflight requests
}

// loggerMiddleware injects the logger into the request context.
func loggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), common.LoggerKey, logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// securityMiddleware returns a middleware that sets security-related headers.
// It uses the "unrolled/secure" library for a robust and standard implementation.
func securityMiddleware(env string) func(http.Handler) http.Handler {
	isProduction := env == "production"
	secureMiddleware := secure.New(secure.Options{
		// Enable HSTS only in production.
		STSSeconds:           31536000,
		STSIncludeSubdomains: true,
		STSPreload:           true,
		IsDevelopment:        !isProduction,

		// Other standard security headers.
		ContentTypeNosniff: true,
		BrowserXssFilter:   true,
		FrameDeny:          true,
	})

	// The returned middleware function.
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			secureMiddleware.HandlerFuncWithNext(w, r, next.ServeHTTP)
		})
	}
}

// requestSizeLimiter middleware limits the size of incoming requests
func requestSizeLimiter(maxSize int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxSize {
				errors.HTTPErrorHandler(w, errors.NewPayloadTooLargeError(maxSize, r.ContentLength), nil, "")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// healthCheckHandler provides a simple health check response.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	errors.SafeJSONResponse(w, map[string]string{"status": "ok"})
}
