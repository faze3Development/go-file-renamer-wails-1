// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.
package security

import (
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"

	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/time/rate"
)

// IPRateLimiter holds rate limiters for different IPs
type IPRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	cleanup  time.Duration
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
		cleanup:  time.Minute * 10, // Clean up old limiters every 10 minutes
	}

	// Start cleanup goroutine
	go limiter.cleanupRoutine()

	return limiter
}

// getLimiter gets or creates a rate limiter for an IP
func (i *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(i.rate, i.burst)
		i.limiters[ip] = limiter
	}

	return limiter
}

// Allow checks if a request from the given IP is allowed
func (i *IPRateLimiter) Allow(ip string) bool {
	limiter := i.getLimiter(ip)
	return limiter.Allow()
}

// cleanupRoutine periodically removes old limiters
func (i *IPRateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(i.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		i.mu.Lock()
		for ip, limiter := range i.limiters {
			// Remove limiters that haven't been used recently
			if limiter.Tokens() == float64(i.burst) {
				delete(i.limiters, ip)
			}
		}
		i.mu.Unlock()
	}
}

// RateLimitingMiddleware creates rate limiting middleware
func RateLimitingMiddleware(limiter *IPRateLimiter, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			ip := getClientIP(r)

			if !limiter.Allow(ip) {
				requestID := middleware.GetReqID(r.Context())
				errors.HTTPErrorHandler(w, errors.NewRateLimitError("Too many requests"), logger, requestID)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the real client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-real-ip header
	if xri := r.Header.Get("X-real-ip"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

// Helper function to create rate limiters for different scenarios
func CreateFileUploadRateLimiter() *IPRateLimiter {
	// More restrictive for file uploads
	rateLimit, _ := strconv.ParseFloat(getEnv("UPLOAD_RATE_LIMIT_RPS", "0.083"), 64) // ~0.083 rps = ~5 rpm
	burst, _ := strconv.Atoi(getEnv("UPLOAD_RATE_LIMIT_BURST", "5"))
	return NewIPRateLimiter(rate.Limit(rateLimit), burst)
}

func CreateAPIRateLimiter() *IPRateLimiter {
	// Standard API rate limiting
	rateLimit, _ := strconv.ParseFloat(getEnv("API_RATE_LIMIT_RPS", "1"), 64) // 1 rps = 60 rpm
	burst, _ := strconv.Atoi(getEnv("API_RATE_LIMIT_BURST", "60"))
	return NewIPRateLimiter(rate.Limit(rateLimit), burst)
}

// getEnv gets an environment variable with a fallback default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
