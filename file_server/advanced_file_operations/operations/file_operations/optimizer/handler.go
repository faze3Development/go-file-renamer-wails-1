// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package optimizer

import (
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/security"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Handler manages HTTP requests for the optimizer feature.
type Handler struct {
	logger  *slog.Logger
	service *Service
}

// NewHandler creates a new optimizer handler.
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	if service == nil {
		service = NewService(logger)
	}
	return &Handler{
		logger:  logger,
		service: service,
	}
}

func (h *Handler) RegisterRoutes(r *chi.Mux) {
	r.Method("GET", "/optimizer/options", h.wrap(h.OptionsHandler))
	r.Method("POST", "/optimizer/images", h.wrap(h.ImagesHandler))
	r.Method("POST", "/optimizer/pdf", h.wrap(h.PDFHandler))
	r.Method("POST", "/optimizer/files", h.wrap(h.GenericHandler))
}

func (h *Handler) wrap(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		errors.WrapHandler(handler, h.logger, requestID)(w, r)
	}
}

// OptionsHandler returns available optimization options
func (h *Handler) OptionsHandler(w http.ResponseWriter, r *http.Request) error {
	options := []map[string]any{
		{
			"type":              "image",
			"name":              "Image Optimization",
			"description":       "Optimize image files (JPEG, PNG, WebP) for size and quality",
			"supported_formats": []string{"jpg", "jpeg", "png", "webp"},
			"parameters": map[string]any{
				"quality": map[string]any{
					"type":        "integer",
					"min":         1,
					"max":         100,
					"default":     85,
					"description": "Image quality (1-100, higher = better quality, larger file)",
				},
				"resize": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"width": map[string]any{
							"type":        "integer",
							"min":         1,
							"description": "Target width in pixels",
						},
						"height": map[string]any{
							"type":        "integer",
							"min":         1,
							"description": "Target height in pixels",
						},
						"maintain_aspect": map[string]any{
							"type":        "boolean",
							"default":     true,
							"description": "Maintain aspect ratio when resizing",
						},
					},
				},
			},
		},
		{
			"type":              "pdf",
			"name":              "PDF Optimization",
			"description":       "Optimize PDF files for size and performance",
			"supported_formats": []string{"pdf"},
			"parameters": map[string]any{
				"compression": map[string]any{
					"type":        "string",
					"options":     []string{"low", "medium", "high"},
					"default":     "medium",
					"description": "Compression level for PDF optimization",
				},
				"remove_metadata": map[string]any{
					"type":        "boolean",
					"default":     false,
					"description": "Remove metadata from PDF files",
				},
			},
		},
		{
			"type":              "generic",
			"name":              "Generic File Optimization",
			"description":       "General file compression and optimization",
			"supported_formats": []string{"txt", "json", "xml", "csv"},
			"parameters": map[string]any{
				"compression_level": map[string]any{
					"type":        "integer",
					"min":         1,
					"max":         9,
					"default":     6,
					"description": "Compression level (1-9, higher = better compression, slower)",
				},
			},
		},
	}

	return errors.SafeJSONResponse(w, map[string]any{
		"options": options,
	})
}

// ImagesHandler handles image optimization requests.
func (h *Handler) ImagesHandler(w http.ResponseWriter, r *http.Request) error {
	config := security.DefaultFileConfig()

	// Validate file
	header, err := security.ValidateFile(r, config)
	if err != nil {
		h.logger.Warn("File validation failed", "error", err)
		return errors.NewValidationError("File validation failed", err)
	}

	// Additional validation for image files
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		h.logger.Warn("Invalid content type for image optimization", "content_type", contentType)
		return errors.NewValidationError("Invalid file type for image optimization", nil)
	}

	file, err := header.Open()
	if err != nil {
		h.logger.Error("Failed to open validated file", "error", err, "filename", header.Filename)
		return errors.NewSystemError("failed to open file", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			h.logger.Warn("Failed to close file", "error", closeErr, "filename", header.Filename)
		}
	}()

	// Read file content with error handling
	fileContent, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error("Failed to read file content", "error", err, "filename", header.Filename)
		return errors.NewSystemError("failed to read file", err)
	}

	// Validate file content size matches header
	if int64(len(fileContent)) != header.Size {
		h.logger.Warn("File size mismatch", "header_size", header.Size, "actual_size", len(fileContent), "filename", header.Filename)
		return errors.NewValidationError("File size validation failed", nil)
	}

	// Process the file optimization
	optimized, err := h.service.ProcessFileWithContent(r.Context(), header.Filename, fileContent, contentType)
	if err != nil {
		h.logger.Error("File optimization failed", "error", err, "filename", header.Filename, "content_type", contentType)
		return errors.NewProcessingError("optimization failed", err)
	}

	// Validate optimization results
	if len(optimized) == 0 {
		h.logger.Error("Optimization produced empty result", "filename", header.Filename)
		return errors.NewSystemError("optimization produced invalid result", nil)
	}

	h.logger.Info("Successfully processed image optimization",
		"filename", header.Filename,
		"original_size", header.Size,
		"optimized_size", len(optimized),
		"content_type", contentType)

	// Base64 encode the processed file content for JSON transport
	encodedContent := base64.StdEncoding.EncodeToString(optimized)

	// Calculate size reduction percentage
	sizeReduction := float64(header.Size-int64(len(optimized))) / float64(header.Size) * 100

	// Return JSON response with file information and processed file data
	response := map[string]any{
		"success":          true,
		"originalName":     header.Filename,
		"action":           "optimized",
		"encodedContent":   encodedContent,
		"contentType":      contentType,
		"originalSize":     header.Size,
		"optimizedSize":    len(optimized),
		"sizeReduction":    sizeReduction,
		"optimizationType": "image",
	}

	return errors.SafeJSONResponse(w, response)
}

// PDFHandler handles PDF optimization requests.
func (h *Handler) PDFHandler(w http.ResponseWriter, r *http.Request) error {
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
		return errors.NewSystemError("failed to open file", err)
	}
	defer file.Close()

	// Read file content with error handling
	fileContent, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error("Failed to read file content", "error", err, "filename", header.Filename)
		return errors.NewSystemError("failed to read file", err)
	}

	// Process the PDF optimization
	optimized, err := h.service.ProcessFileWithContent(r.Context(), header.Filename, fileContent, header.Header.Get("Content-Type"))
	if err != nil {
		h.logger.Error("PDF optimization failed", "error", err)
		return errors.NewProcessingError("PDF optimization failed", err)
	}

	h.logger.Info("Successfully processed PDF optimization",
		"original_size", header.Size,
		"optimized_size", len(optimized))

	// Base64 encode the processed file content for JSON transport
	encodedContent := base64.StdEncoding.EncodeToString(optimized)

	// Return JSON response with file information and processed file data
	response := map[string]any{
		"success":        true,
		"originalName":   header.Filename,
		"action":         "pdf_optimized",
		"encodedContent": encodedContent,
		"contentType":    "application/pdf",
		"originalSize":   header.Size,
		"optimizedSize":  len(optimized),
	}

	return errors.SafeJSONResponse(w, response)
}

// GenericHandler handles generic file optimization requests.
func (h *Handler) GenericHandler(w http.ResponseWriter, r *http.Request) error {
	// Create custom config for optimizer that includes text files
	baseConfig := security.DefaultFileConfig()
	config := &security.FileValidationConfig{
		MaxSizeBytes: baseConfig.MaxSizeBytes,
		AllowedTypes: append(baseConfig.AllowedTypes, "text/plain", "text/markdown", "application/json"),
		BlockedTypes: baseConfig.BlockedTypes,
		AllowedExts:  append(baseConfig.AllowedExts, ".txt", ".md", ".json"),
		BlockedExts:  baseConfig.BlockedExts,
	}

	// Validate file
	header, err := security.ValidateFile(r, config)
	if err != nil {
		h.logger.Warn("File validation failed", "error", err)
		return errors.NewValidationError("File validation failed", err)
	}

	// Additional validation for generic files
	contentType := header.Header.Get("Content-Type")

	file, err := header.Open()
	if err != nil {
		h.logger.Error("Failed to open validated file", "error", err, "filename", header.Filename)
		return errors.NewSystemError("failed to open file", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			h.logger.Warn("Failed to close file", "error", closeErr, "filename", header.Filename)
		}
	}()

	// Read file content with error handling
	fileContent, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error("Failed to read file content", "error", err, "filename", header.Filename)
		return errors.NewSystemError("failed to read file", err)
	}

	// Process the file optimization
	optimized, err := h.service.ProcessFileWithContent(r.Context(), header.Filename, fileContent, contentType)
	if err != nil {
		h.logger.Error("File optimization failed", "error", err)
		return errors.NewProcessingError("File optimization failed", err)
	}

	h.logger.Info("Successfully processed file optimization",
		"original_size", header.Size,
		"optimized_size", len(optimized))

	// Base64 encode the processed file content for JSON transport
	encodedContent := base64.StdEncoding.EncodeToString(optimized)

	// Return JSON response with file information and processed file data
	response := map[string]any{
		"success":        true,
		"originalName":   header.Filename,
		"action":         "optimized",
		"encodedContent": encodedContent,
		"contentType":    header.Header.Get("Content-Type"),
		"originalSize":   header.Size,
		"optimizedSize":  len(optimized),
	}

	return errors.SafeJSONResponse(w, response)
}
