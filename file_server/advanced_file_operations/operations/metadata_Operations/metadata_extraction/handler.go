// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package metadata_extraction

import (
	"io"
	"log/slog"
	"net/http"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/security"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Handler manages HTTP requests for the metadata extraction feature.
type Handler struct {
	logger    *slog.Logger
	processor *Processor
}

// NewHandler creates a new metadata extraction handler.
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		logger:    logger,
		processor: service.Processor(),
	}
}

func (h *Handler) RegisterRoutes(r *chi.Mux) {
	r.Method("POST", "/extract", h.wrap(h.ExtractMetadataHandler))
}

func (h *Handler) wrap(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		errors.WrapHandler(handler, h.logger, requestID)(w, r)
	}
}

// ExtractMetadataHandler handles metadata extraction requests.
func (h *Handler) ExtractMetadataHandler(w http.ResponseWriter, r *http.Request) error {
	config := security.DefaultFileConfig()

	// Validate file
	header, err := security.ValidateFile(r, config)
	if err != nil {
		h.logger.Warn("File validation failed", "error", err)
		return errors.NewValidationError("File validation failed", err)
	}

	file, err := header.Open()
	if err != nil {
		h.logger.Error("Failed to open validated file", "error", err)
		return errors.NewSystemError("Failed to open validated file", err)
	}
	defer file.Close()

	// Read file content fully; single Read can truncate multipart streams
	content, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error("Failed to read file content", "error", err)
		return errors.NewSystemError("Failed to read file content", err)
	}

	// Extract metadata
	metadata, err := h.processor.ProcessFileWithContent(r.Context(), header.Filename, content, header.Header.Get("Content-Type"))
	if err != nil {
		h.logger.Error("Metadata extraction failed", "error", err)
		return errors.NewProcessingError("Metadata extraction failed", err)
	}

	h.logger.Info("Successfully extracted metadata",
		"file_size", header.Size,
		"filename", header.Filename,
		"has_metadata", metadata.HasMetadata)

	// Return JSON response with extracted metadata
	response := map[string]any{
		"success":  true,
		"filename": header.Filename,
		"metadata": metadata,
	}

	return errors.SafeJSONResponse(w, response)
}
