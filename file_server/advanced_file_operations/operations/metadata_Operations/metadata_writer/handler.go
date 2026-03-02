// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package metadata_writer

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/exiftool"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Handler manages HTTP requests for metadata writing
type Handler struct {
	processor *Processor
	logger    *slog.Logger
}

// NewHandler creates a new metadata writer handler
func NewHandler(exiftoolService exiftool.Service, logger *slog.Logger) *Handler {
	return &Handler{
		processor: NewProcessor(exiftoolService, logger),
		logger:    logger,
	}
}

// RegisterRoutes registers the metadata writer routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Method("GET", "/write/fields", h.wrap(h.FieldsHandler))
	r.Method("POST", "/write/process", h.wrap(h.ProcessHandler))
}

func (h *Handler) wrap(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		errors.WrapHandler(handler, h.logger, requestID)(w, r)
	}
}

// FieldsHandler returns commonly used writable metadata fields
func (h *Handler) FieldsHandler(w http.ResponseWriter, r *http.Request) error {
	fields := []map[string]any{
		{"name": "Author", "type": "string", "description": "Creator of the content", "category": "general", "writable": true},
		{"name": "Copyright", "type": "string", "description": "Copyright notice", "category": "general", "writable": true},
		{"name": "Keywords", "type": "string", "description": "Comma-separated keywords", "category": "general", "writable": true},
		{"name": "Title", "type": "string", "description": "Title of the content", "category": "general", "writable": true},
		{"name": "Subject", "type": "string", "description": "Subject or description", "category": "general", "writable": true},
		{"name": "Comment", "type": "string", "description": "User comment", "category": "general", "writable": true},
		{"name": "Artist", "type": "string", "description": "Artist name", "category": "general", "writable": true},
		{"name": "Rating", "type": "int", "description": "Rating (0-5)", "category": "general", "writable": true},
		{"name": "ImageWidth", "type": "int", "description": "Image width in pixels", "category": "image", "writable": false},
		{"name": "ImageHeight", "type": "int", "description": "Image height in pixels", "category": "image", "writable": false},
		{"name": "ISO", "type": "int", "description": "ISO sensitivity", "category": "exif", "writable": true},
		{"name": "FocalLength", "type": "string", "description": "Focal length", "category": "exif", "writable": true},
		{"name": "ApertureValue", "type": "string", "description": "Aperture value", "category": "exif", "writable": true},
		{"name": "ShutterSpeedValue", "type": "string", "description": "Shutter speed", "category": "exif", "writable": true},
	}

	return errors.SafeJSONResponse(w, map[string]any{
		"fields": fields,
		"filtering": map[string]any{
			"description": "Backend automatically filters metadata fields",
			"skipped_fields": []string{
				"filename", "size", "contentType", "hasMetadata",
				"image", "raw", "Directory*", "FileAccessDate*",
				"FileCreateDate*", "FileModifyDate*", "FileName*",
				"FilePermissions*", "FileSize*", "FileType*",
				"SourceFile*", "ExifToolVersion*", "*Binary data*",
			},
			"note": "Any field not in the skipped list will be accepted if it has a non-empty value",
		},
	})
}

// ProcessHandler handles metadata writing requests
func (h *Handler) ProcessHandler(w http.ResponseWriter, r *http.Request) error {
	// Parse multipart form
	maxPayloadSize := common.GetMaxFileSizeFromEnv(common.MaxPayload10MB)
	err := r.ParseMultipartForm(maxPayloadSize)
	if err != nil {
		return errors.NewValidationError("Failed to parse multipart form", err)
	}

	// Get file from form
	file, fileHeader, err := r.FormFile("files")
	if err != nil {
		return errors.NewValidationError("No file provided", err)
	}
	defer file.Close()

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return errors.NewFileProcessingError(fileHeader.Filename, "read file content", err)
	}

	// Get metadata from form
	metadataJSON := r.FormValue("metadata")
	h.logger.Info("Received metadata JSON", "json", metadataJSON, "length", len(metadataJSON))
	if metadataJSON == "" {
		return errors.NewValidationError("metadata field is required", nil)
	}

	var metadata map[string]any
	if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
		return errors.NewValidationError("Invalid metadata JSON format", err)
	}

	// Allow any non-empty string metadata fields (more flexible than predefined list)
	filteredMetadata := make(map[string]any)
	skippedFields := make([]string, 0)
	emptyFields := make([]string, 0)

	h.logger.Info("Processing metadata fields", "original_keys", getKeys(metadata))
	for key, value := range metadata {
		// Skip internal/file fields that shouldn't be written back
		if shouldSkipField(key) {
			h.logger.Info("Skipping field", "key", key)
			skippedFields = append(skippedFields, key)
			continue
		}
		// Accept any non-empty value
		if value != nil && value != "" {
			// Convert dotted notation to ExifTool tag names
			exifTag := convertToExifTag(key)
			h.logger.Info("Accepting field", "key", key, "exifTag", exifTag, "value", value)
			filteredMetadata[exifTag] = value
		} else {
			h.logger.Info("Rejecting field due to empty value", "key", key, "value", value)
			emptyFields = append(emptyFields, key)
		}
	}
	h.logger.Info("Filtered metadata",
		"filtered_keys", getKeys(filteredMetadata),
		"filtered_count", len(filteredMetadata),
		"skipped_fields", skippedFields,
		"empty_fields", emptyFields)

	// Validate filtered metadata
	if err := h.processor.ValidateConfig(map[string]any{"metadata": filteredMetadata}); err != nil {
		return err
	}

	// Process file
	processedContent, err := h.processor.ProcessFileWithContent(
		r.Context(),
		fileHeader.Filename,
		fileContent,
		fileHeader.Header.Get("Content-Type"),
		filteredMetadata,
	)
	if err != nil {
		return err
	}

	// Return processed file
	w.Header().Set("Content-Type", fileHeader.Header.Get("Content-Type"))
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileHeader.Filename+"\"")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(processedContent)))

	_, err = w.Write(processedContent)
	if err != nil {
		return errors.NewSystemError("Failed to write response", err)
	}

	return nil
}

// shouldSkipField determines if a metadata field should be skipped during writing
func shouldSkipField(key string) bool {
	// Skip file metadata that shouldn't be written back
	skipFields := []string{
		"filename", "size", "contentType", "hasMetadata",
	}

	for _, skip := range skipFields {
		if strings.Contains(key, skip) {
			return true
		}
	}

	// Skip complex nested objects (but allow dotted field names like "image.height")
	if key == "image" || key == "raw" {
		return true
	}

	// Skip fields that start with common internal prefixes
	if strings.HasPrefix(key, "Directory") ||
		strings.HasPrefix(key, "FileAccessDate") ||
		strings.HasPrefix(key, "FileCreateDate") ||
		strings.HasPrefix(key, "FileModifyDate") ||
		strings.HasPrefix(key, "FileName") ||
		strings.HasPrefix(key, "FilePermissions") ||
		strings.HasPrefix(key, "FileSize") ||
		strings.HasPrefix(key, "FileType") ||
		strings.HasPrefix(key, "SourceFile") ||
		strings.HasPrefix(key, "ExifToolVersion") {
		return true
	}

	// Skip fields that look like they contain complex data
	if strings.Contains(key, "Binary data") ||
		strings.Contains(key, "(Binary data") {
		return true
	}

	return false
}

// convertToExifTag converts dotted field notation to ExifTool tag names
func convertToExifTag(key string) string {
	// Handle common dotted notation conversions
	switch key {
	case "image.width":
		return "ImageWidth"
	case "image.height":
		return "ImageHeight"
	case "image.bitsPerSample":
		return "BitsPerSample"
	case "exif.iso":
		return "ISO"
	case "exif.focalLength":
		return "FocalLength"
	case "exif.aperture":
		return "ApertureValue"
	case "exif.shutterSpeed":
		return "ShutterSpeedValue"
	case "raw.bitsPerSample":
		return "BitsPerSample"
	}

	// Convert dotted notation to CamelCase
	if strings.Contains(key, ".") {
		parts := strings.Split(key, ".")
		for i, part := range parts {
			if len(part) > 0 {
				parts[i] = strings.ToUpper(string(part[0])) + part[1:]
			}
		}
		return strings.Join(parts, "")
	}

	// Return as-is if no conversion needed
	return key
}

// getKeys returns the keys of a map for logging
func getKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
