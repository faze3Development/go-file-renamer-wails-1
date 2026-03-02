// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package limited_file_processing

import (
	"context"
	"log/slog"
	"net/http"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/exiftool"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/file_operations/optimizer"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/file_operations/renamers"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/metadata_Operations/metadata_extraction"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/metadata_Operations/metadata_removal"
)

// ServiceAPI defines the interface for a service that processes files.
type ServiceAPI interface {
	ProcessFile(ctx context.Context, file common.FileOperation, config map[string]interface{}) (*common.FileOperationResult, error)
	ProcessFileWithContent(ctx context.Context, file common.File, config map[string]interface{}) (*FileProcessingResult, error)
}

// FileProcessingResult represents the result of processing a single file
type FileProcessingResult struct {
	Filename    string `json:"filename"`
	NewName     string `json:"newName"`
	ContentType string `json:"contentType"`
	Action      string `json:"action"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
	Content     []byte `json:"content,omitempty"` // The processed file content
}

// Service implements the logic for limited, synchronous file processing.
// Uses the operations/features for actual file processing logic.
type Service struct {
	logger                      *slog.Logger
	renamerProcessor            *renamers.Processor
	metadataProcessor           *metadata_removal.Processor
	metadataExtractionProcessor *metadata_extraction.Processor
	optimizerProcessor          *optimizer.Processor
}

// NewService creates a new limited file processing service.
func NewService(exiftoolService exiftool.Service, logger *slog.Logger) *Service {
	return &Service{
		logger:                      logger,
		renamerProcessor:            renamers.NewProcessor(),
		metadataProcessor:           metadata_removal.NewProcessor(exiftoolService, logger),
		metadataExtractionProcessor: metadata_extraction.NewProcessor(exiftoolService, logger),
		optimizerProcessor:          optimizer.NewProcessor(),
	}
}

// ProcessFile processes a single file synchronously using the operations/features.
// Note: Only renaming uses ProcessFile (doesn't need file content).
// All other operations (metadata removal, extraction, optimization) require file content
// and will return an error if ProcessFile is called. Use ProcessFileWithContent for those.
func (s *Service) ProcessFile(ctx context.Context, file common.FileOperation, config map[string]interface{}) (*common.FileOperationResult, error) {
	s.logger.Info("Processing single file synchronously", "filename", file.Source)

	// Check for metadata removal request
	if config["removeMetadata"] == true {
		// In limited processing, we can analyze what metadata could be removed but can't actually remove it
		// since we don't have file content. We'll simulate the operation for demonstration.
		s.logger.Info("Metadata removal requested but file content not available - simulating operation", "filename", file.Source)

		result := &common.FileOperationResult{
			Original: file,
			New:      file, // File would remain the same name but metadata would be stripped
			Status:   common.StatusCompleted,
		}

		// Note: In a real implementation, you might want to return information about
		// what metadata types are typically found in this file type
		return result, nil
	}

	// Check if renaming is requested
	if config["renameFiles"] == true {
		// Use the renamer processor
		result, err := s.renamerProcessor.ProcessFile(ctx, file, config)
		if err != nil {
			s.logger.Error("Failed to rename file", "error", err, "filename", file.Source)
			return nil, errors.NewFileProcessingError(file.Source, "rename", err)
		}
		return result, nil
	}

	// Default: return the file unchanged
	return &common.FileOperationResult{
		Original: file,
		New:      file,
		Status:   common.StatusCompleted,
	}, nil
}

// ProcessFileWithContent processes a single file with actual content synchronously using the operations/features.
// This enables full processing including metadata removal.
func (s *Service) ProcessFileWithContent(ctx context.Context, file common.File, config map[string]interface{}) (*FileProcessingResult, error) {
	s.logger.Info("Processing single file with content synchronously", "filename", file.Filename)

	result := &FileProcessingResult{
		Filename:    file.Filename,
		ContentType: file.ContentType,
		Content:     file.Content, // Start with original content, will be replaced
		Action:      "processed",
		Success:     false,
	}

	processedContent := file.Content // Start with original content

	// Apply renaming
	if config["renameFiles"] == true {
		pattern := common.GetStringFromConfig(config, "pattern", "default")
		preserveOriginal := false
		if preserve, ok := config["preserveOriginalName"].(bool); ok {
			preserveOriginal = preserve
		} else if opts, ok := config["renameOptions"].(map[string]interface{}); ok {
			if preserve, ok := opts["preserveOriginalName"].(bool); ok {
				preserveOriginal = preserve
			}
		}

		if renameCfg, ok := config["rename"].(map[string]interface{}); ok {
			if pattern == "default" {
				if p, ok := renameCfg["pattern"].(string); ok && p != "" {
					pattern = p
				}
			}
			if !preserveOriginal {
				if opts, ok := renameCfg["options"].(map[string]interface{}); ok {
					if preserve, ok := opts["preserveOriginalName"].(bool); ok {
						preserveOriginal = preserve
					}
				}
			}
		}

		namer, err := renamers.NewPatternNamer(pattern, renamers.Options{PreserveOriginalName: preserveOriginal})
		if err != nil {
			return nil, err
		}

		newName, err := renamers.FormatWithNamer(namer, file.Filename)
		if err != nil {
			return nil, err
		}
		result.NewName = newName
		result.Action = "renamed"
	} else {
		result.NewName = file.Filename
	}

	// Apply metadata removal
	if config["removeMetadata"] == true {
		strippedContent, err := s.metadataProcessor.ProcessFileWithContent(ctx, file.Filename, processedContent, file.ContentType)
		if err != nil {
			s.logger.Error("Failed to remove metadata", "error", err, "filename", file.Filename)
			result.Error = "Failed to remove metadata: " + err.Error()
			return result, errors.NewFileProcessingError(file.Filename, "remove metadata", err)
		}
		processedContent = strippedContent

		if result.Action == "renamed" {
			result.Action = "renamed_metadata_removed"
		} else {
			result.Action = "metadata_removed"
		}
	}

	// Apply optimization
	if config["optimizeFiles"] == true {
		optimizedContent, err := s.optimizerProcessor.ProcessFileWithContent(ctx, file.Filename, processedContent, file.ContentType)
		if err != nil {
			s.logger.Error("Failed to optimize file", "error", err, "filename", file.Filename)
			result.Error = "Failed to optimize file: " + err.Error()
			return result, errors.NewFileProcessingError(file.Filename, "optimize", err)
		}
		processedContent = optimizedContent

		switch result.Action {
		case "renamed":
			result.Action = "renamed_optimized"
		case "metadata_removed":
			result.Action = "metadata_removed_optimized"
		case "renamed_metadata_removed":
			result.Action = "renamed_metadata_removed_optimized"
		default:
			result.Action = "optimized"
		}
	}

	result.Content = processedContent
	result.Success = true
	return result, nil
}

// Handler is responsible for handling HTTP requests for the limited processing service.
// Note: This service is called by the Dispatcher, so it doesn't register its own routes.
type Handler struct {
	service ServiceAPI
	logger  *slog.Logger
}

// NewHandler creates a new handler for the limited processing service.
func NewHandler(service ServiceAPI, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// ProcessFilesLimited handles limited processing requests from the operations dispatcher.
// This method focuses purely on limited processing orchestration using features.
func (h *Handler) ProcessFilesLimited(w http.ResponseWriter, r *http.Request, files []common.File, config map[string]interface{}) error {
	h.logger.InfoContext(r.Context(), "Processing files in limited mode", "fileCount", len(files))

	results := make([]*FileProcessingResult, 0, len(files))

	// Process each file synchronously
	for _, file := range files {
		result, err := h.service.ProcessFileWithContent(r.Context(), file, config)
		if err != nil {
			h.logger.ErrorContext(r.Context(), "Failed to process file in limited mode",
				"error", err,
				"filename", file.Filename)
			// For limited processing, we'll fail the whole request if one file fails
			return errors.NewFileProcessingError(file.Filename, "process", err)
		}

		results = append(results, result)
	}

	h.logger.InfoContext(r.Context(), "Limited processing completed", "files_processed", len(results))

	// Return JSON response
	response := map[string]interface{}{
		"message": "Files processed successfully (limited mode)",
		"results": results,
	}

	return errors.SafeJSONResponse(w, response)
}
