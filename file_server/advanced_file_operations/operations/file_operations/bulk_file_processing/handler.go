// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package bulk_file_processing

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Handler handles bulk file processing HTTP requests
type Handler struct {
	service ServiceAPI
	logger  *slog.Logger
}

// NewHandler creates a new bulk file processing handler
func NewHandler(service ServiceAPI, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers bulk processing routes
func (h *Handler) RegisterRoutes(r *chi.Mux) {
	h.logger.Info("Registering bulk processing routes")
	r.Route("/bulk", func(r chi.Router) {
		r.Method("POST", "/process", h.wrap(h.ProcessFiles))
	})
}

func (h *Handler) wrap(handler func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		errors.WrapHandler(handler, h.logger, requestID)(w, r)
	}
}

// ProcessFiles handles bulk file processing requests
func (h *Handler) ProcessFiles(w http.ResponseWriter, r *http.Request) error {
	h.logger.InfoContext(r.Context(), "Bulk processing request received")

	// Parse multipart form
	maxPayloadSize := common.GetMaxFileSizeFromEnv(common.MaxPayload32MB)
	err := r.ParseMultipartForm(maxPayloadSize)
	if err != nil {
		h.logger.WarnContext(r.Context(), "Failed to parse multipart form", "error", err)
		return errors.NewValidationError("Failed to parse multipart form", err)
	}

	// Get files from form
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		return errors.NewValidationError("No files provided", nil)
	}

	// Get processing options using centralized logic
	config := buildProcessingOptionsFromForm(r)
	renameOptions := renameOptionsFromInterface(config["renameOptions"])
	options := ProcessingOptions{
		RenameFiles:    common.GetBoolFromConfig(config, "renameFiles", false),
		RemoveMetadata: common.GetBoolFromConfig(config, "removeMetadata", false),
		OptimizeFiles:  common.GetBoolFromConfig(config, "optimizeFiles", false),
		Pattern:        common.GetStringFromConfig(config, "pattern", "default"),
		Namer:          common.GetStringFromConfig(config, "namer", "basic"),
		Rename:         renameOptions,
		MaxFileSize:    common.GetInt64FromConfig(config, "maxFileSize", common.MaxFileSize50MB), // 50MB default, configurable from frontend
		AllowedTypes:   []string{"image/jpeg", "image/png", "image/gif", "application/pdf"},
	}

	// Convert multipart files to FileUpload structs
	var commonFiles []common.File
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			h.logger.WarnContext(r.Context(), "Failed to open file", "error", err, "filename", fileHeader.Filename)
			continue
		}

		// Read file content
		fileContent, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			h.logger.WarnContext(r.Context(), "Failed to read file content", "error", err, "filename", fileHeader.Filename)
			continue
		}

		commonFile := common.File{
			Filename:    fileHeader.Filename,
			Content:     fileContent,
			ContentType: fileHeader.Header.Get("Content-Type"),
			Size:        fileHeader.Size,
		}

		commonFiles = append(commonFiles, commonFile)
	}

	if len(commonFiles) == 0 {
		return errors.NewValidationError("No valid files to process", nil)
	}

	// Process files
	result, err := h.service.ProcessBulkFiles(r.Context(), "anonymous", commonFiles, options)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Bulk processing failed", "error", err)
		return err
	}

	// Prepare response
	var responseResults []map[string]any
	for _, fileResult := range result.Results {
		responseResult := map[string]any{
			"success": fileResult.Success,
			"action":  fileResult.Action,
		}

		responseResult["filename"] = fileResult.Filename
		if fileResult.NewName != "" {
			responseResult["newName"] = fileResult.NewName
		}
		if fileResult.ContentType != "" {
			responseResult["contentType"] = fileResult.ContentType
		}

		if fileResult.Success {
			// Get processed data from cache and base64 encode it
			processedData, exists := h.service.GetProcessedData(result.JobID, fileResult.Filename)
			if !exists {
				responseResult["error"] = "Processed data not available"
				continue
			}
			encodedContent := base64.StdEncoding.EncodeToString(processedData)
			responseResult["data"] = map[string]any{
				"newName":        fileResult.NewName,
				"encodedContent": encodedContent,
				"contentType":    fileResult.ContentType,
			}
		} else {
			responseResult["error"] = fileResult.Error
		}

		responseResults = append(responseResults, responseResult)
	}

	response := ProcessFilesResponse{
		Results: responseResults,
		JobID:   result.JobID,
	}

	h.logger.InfoContext(r.Context(), "Bulk processing completed successfully",
		"jobId", result.JobID,
		"totalFiles", result.TotalFiles,
		"successCount", result.SuccessCount,
		"failureCount", result.FailureCount,
		"duration", result.Duration)

	return errors.SafeJSONResponse(w, response)
}

// ProcessFilesResponse represents the response for the ProcessFiles handler
type ProcessFilesResponse struct {
	Results     []map[string]any `json:"results"`
	JobID       string           `json:"jobId"`
	DownloadURL string           `json:"downloadUrl,omitempty"`
}

func buildProcessingOptionsFromForm(r *http.Request) map[string]interface{} {
	config := make(map[string]interface{})

	// Parse multipart form if not already parsed
	if r.MultipartForm == nil {
		maxPayloadSize := common.GetMaxFileSizeFromEnv(common.MaxPayload32MB)
		r.ParseMultipartForm(maxPayloadSize)
	}

	// Boolean processing options
	if r.FormValue("renameFiles") == "true" {
		config["renameFiles"] = true
	}
	if r.FormValue("removeMetadata") == "true" {
		config["removeMetadata"] = true
	}
	if r.FormValue("optimizeFiles") == "true" {
		config["optimizeFiles"] = true
	}

	// String configuration options with sanitization
	if pattern := r.FormValue("pattern"); pattern != "" {
		config["pattern"] = pattern
	}

	// Numeric configuration options
	if maxFileSize := r.FormValue("maxFileSize"); maxFileSize != "" {
		config["maxFileSize"] = common.ParseMaxFileSizeFromString(maxFileSize, common.MaxFileSize50MB)
	}

	if rawConfig := r.FormValue("config"); rawConfig != "" {
		if ops, err := parseOperationsConfig(rawConfig, true); err == nil {
			applyOperationsConfig(config, ops)
		}
	}

	if rawOptions := r.FormValue("options"); rawOptions != "" {
		if ops, err := parseOperationsConfig(rawOptions, false); err == nil {
			applyOperationsConfig(config, ops)
		}
	}

	return config
}

type operationsConfigPayload struct {
	Rename struct {
		Enabled bool                   `json:"enabled"`
		Pattern string                 `json:"pattern"`
		Namer   string                 `json:"namer"`
		Options RenameOperationOptions `json:"options"`
	} `json:"rename"`
	MetadataRemoval struct {
		Enabled bool `json:"enabled"`
	} `json:"metadataRemoval"`
	Optimizer struct {
		Enabled bool `json:"enabled"`
	} `json:"optimizer"`
}

func parseOperationsConfig(raw string, wrapped bool) (*operationsConfigPayload, error) {
	if wrapped {
		var payload struct {
			Operations operationsConfigPayload `json:"operations"`
		}
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			return nil, err
		}
		return &payload.Operations, nil
	}

	var ops operationsConfigPayload
	if err := json.Unmarshal([]byte(raw), &ops); err != nil {
		return nil, err
	}
	return &ops, nil
}

func applyOperationsConfig(config map[string]interface{}, ops *operationsConfigPayload) {
	if ops == nil {
		return
	}

	if ops.Rename.Enabled {
		config["renameFiles"] = true
	}
	if ops.Rename.Pattern != "" {
		config["pattern"] = ops.Rename.Pattern
	}
	if ops.Rename.Namer != "" {
		config["namer"] = ops.Rename.Namer
	}
	config["renameOptions"] = normalizeRenameOptions(ops.Rename.Options)

	if ops.MetadataRemoval.Enabled {
		config["removeMetadata"] = true
	}
	if ops.Optimizer.Enabled {
		config["optimizeFiles"] = true
	}
}
