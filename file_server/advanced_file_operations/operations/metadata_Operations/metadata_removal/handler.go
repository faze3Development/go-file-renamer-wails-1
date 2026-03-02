// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package metadata_removal

import (
	"encoding/base64"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/security"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/metadata_Operations/metadata_removal/pdf"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Handler manages HTTP requests for the metadata removal feature.
type Handler struct {
	logger  *slog.Logger
	service *Service
}

// NewHandler creates a new metadata removal handler.
func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

func (h *Handler) RegisterRoutes(r *chi.Mux) {
	r.Method("POST", "/metadata/images", h.wrap(h.ImagesHandler))
	r.Method("POST", "/metadata/pdf", h.wrap(h.PDFHandler))
}

func (h *Handler) wrap(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		errors.WrapHandler(handler, h.logger, requestID)(w, r)
	}
}

// ImagesHandler handles image metadata removal requests.
func (h *Handler) ImagesHandler(w http.ResponseWriter, r *http.Request) error {
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
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			h.logger.Warn("Failed to close file", "error", err)
		}
	}(file)

	// Read file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		h.logger.Error("Failed to read file content", "error", err)
		return errors.NewSystemError("Failed to read file content", err)
	}

	// Process the image metadata removal using ExifTool
	fileObj := &common.File{
		Filename:    header.Filename,
		Content:     fileContent,
		ContentType: header.Header.Get("Content-Type"),
		Size:        header.Size,
	}

	err = h.service.exiftoolService.RemoveAllMetadata(fileObj)
	if err != nil {
		h.logger.Error("Image metadata strip failed", "error", err)
		return errors.NewProcessingError("Image metadata strip failed", err)
	}

	stripped := fileObj.Content

	h.logger.Info("Successfully processed image metadata removal",
		"file_size", header.Size)

	// Base64 encode the processed file content for JSON transport
	encodedContent := base64.StdEncoding.EncodeToString(stripped)

	// Return JSON response with file information and processed file data
	response := map[string]interface{}{
		"success":        true,
		"originalName":   header.Filename,
		"action":         "metadata_removed",
		"encodedContent": encodedContent,
		"contentType":    header.Header.Get("Content-Type"),
	}

	return errors.SafeJSONResponse(w, response)
}

// PDFHandler handles PDF metadata removal requests.
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
		return errors.NewSystemError("Failed to open validated file", err)
	}
	defer file.Close()

	// Process the PDF metadata removal
	stripped, err := pdf.StripPDFMetadata(file, h.logger)
	if err != nil {
		h.logger.Error("PDF metadata strip failed", "error", err)
		return errors.NewProcessingError("PDF metadata strip failed", err)
	}

	h.logger.Info("Successfully processed PDF metadata removal",
		"file_size", header.Size)

	// Base64 encode the processed file content for JSON transport
	encodedContent := base64.StdEncoding.EncodeToString(stripped)

	// Return JSON response with file information and processed file data
	response := map[string]interface{}{
		"success":        true,
		"originalName":   header.Filename,
		"action":         "pdf_metadata_removed",
		"encodedContent": encodedContent,
		"contentType":    "application/pdf",
	}

	return errors.SafeJSONResponse(w, response)
}
