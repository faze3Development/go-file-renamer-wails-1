// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package metadata_renamer

import (
	"io"
	"log/slog"
	"net/http"
	"strings"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Handler manages HTTP requests for metadata-based renaming
type Handler struct {
	processor *Processor
	logger    *slog.Logger
	service   *Service
}

// NewHandler creates a new metadata renamer handler
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		processor: NewProcessor(service, logger),
		logger:    logger,
		service:   service,
	}
}

// RegisterRoutes registers the metadata renamer routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Method("GET", "/metadata-rename/patterns", h.wrap(h.PatternsHandler))
	r.Method("POST", "/metadata-rename/process", h.wrap(h.ProcessHandler))
}

func (h *Handler) wrap(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		errors.WrapHandler(handler, h.logger, requestID)(w, r)
	}
}

// PatternsHandler returns available metadata-based renaming patterns for a specific file
func (h *Handler) PatternsHandler(w http.ResponseWriter, r *http.Request) error {
	// Parse multipart form to get the file
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		h.logger.Warn("Failed to parse multipart form", "error", err)
		return errors.NewValidationError("Failed to parse form data", err)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.logger.Warn("No file provided for pattern discovery", "error", err)
		return errors.NewValidationError("File is required to determine available patterns", err)
	}
	defer file.Close()

	// Read file content
	fileData, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error("Failed to read uploaded file", "error", err)
		return errors.NewFileProcessingError("Failed to read file", "", err)
	}

	// Use the full metadata extraction processor for comprehensive analysis
	extractedMetadata, err := h.processor.service.metadataExtractionService.ProcessFileWithContent(r.Context(), header.Filename, fileData, header.Header.Get("Content-Type"))
	if err != nil {
		h.logger.Warn("Failed to extract metadata for pattern discovery", "error", err)
		// Fall back to default patterns if extraction fails
		patterns := h.processor.GetAvailablePatterns()
		return errors.SafeJSONResponse(w, map[string]any{
			"patterns": patterns,
			"warning":  "Could not analyze file metadata, showing default patterns",
		})
	}

	// Generate patterns based on the structured metadata
	patterns := h.processor.GetAvailablePatternsForMetadata(extractedMetadata)

	return errors.SafeJSONResponse(w, map[string]any{
		"patterns":        patterns,
		"availableFields": h.getAvailableFieldNames(extractedMetadata.Raw),
		"metadata":        extractedMetadata, // Include full metadata for debugging/transparency
	})
}

// getAvailableFieldNames extracts field names from metadata, filtering out binary data and internal fields
func (h *Handler) getAvailableFieldNames(metadata map[string]any) []string {
	var fields []string

	for key, value := range metadata {
		// Skip binary data fields (they contain "(Binary data" in the string representation)
		if strValue, ok := value.(string); ok {
			if strings.Contains(strValue, "(Binary data") {
				continue
			}
		}

		// Skip obviously internal/system fields
		switch key {
		case "SourceFile", "Directory", "FileName", "FilePermissions", "ExifToolVersion":
			continue
		}

		fields = append(fields, key)
	}

	return fields
}

// ProcessHandler handles metadata-based file renaming requests
func (h *Handler) ProcessHandler(w http.ResponseWriter, r *http.Request) error {
	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		h.logger.Warn("Failed to parse multipart form", "error", err)
		return errors.NewValidationError("Failed to parse form data", err)
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		h.logger.Warn("No file provided", "error", err)
		return errors.NewValidationError("File is required", err)
	}
	defer file.Close()

	// Get pattern from form
	pattern := r.FormValue("pattern")
	if pattern == "" {
		pattern = "exif:datetime" // default pattern
	}

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error("Failed to read file content", "error", err)
		return errors.NewSystemError("Failed to read file", err)
	}

	// Process the file with metadata-based renaming
	newFilename, err := h.processor.ProcessFileWithContentAndPattern(r.Context(), header.Filename, fileContent, header.Header.Get("Content-Type"), pattern)
	if err != nil {
		h.logger.Error("Metadata renaming failed", "error", err, "filename", header.Filename)
		return errors.NewProcessingError("Metadata renaming failed", err)
	}

	h.logger.Info("Successfully processed metadata renaming",
		"original_filename", header.Filename,
		"new_filename", newFilename,
		"pattern", pattern)

	// Return response with new filename
	response := map[string]interface{}{
		"success":      true,
		"originalName": header.Filename,
		"newName":      newFilename,
		"action":       "metadata_renamed",
		"pattern":      pattern,
		"contentType":  header.Header.Get("Content-Type"),
		"originalSize": header.Size,
	}

	return errors.SafeJSONResponse(w, response)
}
