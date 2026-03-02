// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package renamers

import (
	"log/slog"
	"net/http"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Handler manages HTTP requests for the renamer feature.
type Handler struct {
	logger *slog.Logger
}

// NewHandler creates a new renamer handler.
func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{logger: logger}
}

// RegisterRoutes registers the renamer routes with the router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Method("GET", "/patterns", h.wrap(h.PatternsHandler))
	r.Method("GET", "/namers", h.wrap(h.NamersHandler))
}

func (h *Handler) wrap(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		errors.WrapHandler(handler, h.logger, requestID)(w, r)
	}
}

// PatternsResponse represents the response for the PatternsHandler
type PatternsResponse struct {
	Patterns []string `json:"patterns"`
}

// NamersResponse represents the response for the NamersHandler
type NamersResponse struct {
	Namers []string `json:"namers"`
}

// PatternsHandler returns available renaming patterns
func (h *Handler) PatternsHandler(w http.ResponseWriter, r *http.Request) error {
	patterns := []string{"default", "timestamp", "random", "custom-date", "regex-replace"}
	return errors.SafeJSONResponse(w, PatternsResponse{Patterns: patterns})
}

// NamersHandler returns available namers
func (h *Handler) NamersHandler(w http.ResponseWriter, r *http.Request) error {
	namers := []string{"basic", "timestamp", "hash"}
	return errors.SafeJSONResponse(w, NamersResponse{Namers: namers})
}
