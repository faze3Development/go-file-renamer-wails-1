// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package advanced_file_operations

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/exiftool"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/file_operations/bulk_file_processing"
)

// AdvancedFileOperations centralises access to the orchestration layer used by the
// Wails frontend. The struct is bound in main.go so any exported methods on this
// type become available to the UI layer.
type AdvancedFileOperations struct {
	ExiftoolService   exiftool.Service
	ProcessingService *operations.Orchestrator
	Logger            *slog.Logger
}

// FileProcessor defines the interface implemented by feature processors such as
// renamers, metadata removal, optimisation, etc.
type FileProcessor interface {
	Type() string
	ProcessFile(ctx context.Context, file common.FileOperation, config map[string]interface{}) (*common.FileOperationResult, error)
	ValidateConfig(config map[string]interface{}) error
}

// Exported type aliases keep backend and frontend option contracts in sync without
// duplicating struct definitions.
type (
	SequentialNamingOptions = bulk_file_processing.SequentialNamingOptions
	RenameOperationOptions  = bulk_file_processing.RenameOperationOptions
)

var defaultAllowedContentTypes = []string{
	"image/jpeg",
	"image/png",
	"image/gif",
	"application/pdf",
}

// BulkProcessingFile captures the file data supplied by the frontend for bulk
// processing. Content arrives as base64 so we can marshal over Wails bindings.
type BulkProcessingFile struct {
	Filename      string `json:"filename"`
	ContentBase64 string `json:"contentBase64"`
	ContentType   string `json:"contentType,omitempty"`
	Size          int64  `json:"size"`
}

// BulkProcessingOptions mirrors the orchestrator configuration. Fields are
// intentionally aligned with ProcessingOptions to keep conversions straightforward.
type BulkProcessingOptions struct {
	RenameFiles    bool                   `json:"renameFiles"`
	RemoveMetadata bool                   `json:"removeMetadata"`
	OptimizeFiles  bool                   `json:"optimizeFiles"`
	CompressFiles  bool                   `json:"compressFiles"`
	Pattern        string                 `json:"pattern,omitempty"`
	Namer          string                 `json:"namer,omitempty"`
	Rename         RenameOperationOptions `json:"renameOptions"`
	AllowedTypes   []string               `json:"allowedTypes,omitempty"`
	MaxFileSize    int64                  `json:"maxFileSize,omitempty"`
}

// BulkProcessingRequest is the payload received from the frontend.
type BulkProcessingRequest struct {
	UserID  string                `json:"userId,omitempty"`
	Files   []BulkProcessingFile  `json:"files"`
	Options BulkProcessingOptions `json:"options"`
}

// BulkProcessingResultFile represents per-file processing results sent back to
// the UI. Processed file data is returned as base64 so the frontend can generate
// downloads without touching disk.
type BulkProcessingResultFile struct {
	Filename      string `json:"filename"`
	NewName       string `json:"newName,omitempty"`
	Success       bool   `json:"success"`
	Error         string `json:"error,omitempty"`
	Action        string `json:"action,omitempty"`
	ContentType   string `json:"contentType,omitempty"`
	ContentBase64 string `json:"contentBase64,omitempty"`
}

// BulkProcessingResponse summarises the outcome of a bulk processing request.
type BulkProcessingResponse struct {
	JobID        string                     `json:"jobId"`
	TotalFiles   int                        `json:"totalFiles"`
	SuccessCount int                        `json:"successCount"`
	FailureCount int                        `json:"failureCount"`
	Duration     int64                      `json:"durationMs"`
	Results      []BulkProcessingResultFile `json:"results"`
}

// NewAdvancedFileOperations wires ExifTool and the processing orchestrator using
// the shared logger. It is invoked during application start and returns the
// instance bound to the frontend.
func NewAdvancedFileOperations(logger *slog.Logger) (*AdvancedFileOperations, error) {
	exiftoolConfig := exiftool.DefaultConfig()
	exiftoolConfig.ExiftoolBinaryPath = filepath.Join("file_server", "advanced_file_operations", "exiftool-13.39_64", "exiftool.exe")

	exiftoolService, err := exiftool.NewService(exiftoolConfig, logger)
	if err != nil {
		logger.Error("failed to initialise ExifTool service", "error", err)
		return nil, err
	}

	processingService := operations.NewOrchestrator(exiftoolService, logger)

	return &AdvancedFileOperations{
		ExiftoolService:   exiftoolService,
		ProcessingService: processingService,
		Logger:            logger,
	}, nil
}

func (o BulkProcessingOptions) toProcessingOptions() bulk_file_processing.ProcessingOptions {
	allowedTypes := o.AllowedTypes
	if len(allowedTypes) == 0 {
		allowedTypes = defaultAllowedContentTypes
	}

	maxFileSize := o.MaxFileSize
	if maxFileSize == 0 {
		maxFileSize = common.MaxFileSize50MB
	}

	return bulk_file_processing.ProcessingOptions{
		RenameFiles:    o.RenameFiles,
		RemoveMetadata: o.RemoveMetadata,
		CompressFiles:  o.CompressFiles,
		OptimizeFiles:  o.OptimizeFiles,
		Pattern:        o.Pattern,
		Namer:          o.Namer,
		Rename:         bulk_file_processing.RenameOperationOptions(o.Rename),
		AllowedTypes:   allowedTypes,
		MaxFileSize:    maxFileSize,
	}
}

// ProcessBulkFiles handles bulk file processing requests coming from the frontend.
// Files are supplied as base64 payloads which are decoded and streamed through the
// orchestrator. Processed outputs are returned as base64 strings so the UI can
// offer direct downloads without temporary disk writes.
func (a *AdvancedFileOperations) ProcessBulkFiles(req BulkProcessingRequest) (*BulkProcessingResponse, error) {
	if len(req.Files) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	userID := strings.TrimSpace(req.UserID)
	if userID == "" {
		userID = "desktop-user"
	}

	files := make([]common.File, 0, len(req.Files))
	for _, file := range req.Files {
		if strings.TrimSpace(file.Filename) == "" {
			return nil, fmt.Errorf("file missing filename")
		}

		data, err := base64.StdEncoding.DecodeString(file.ContentBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode file %s: %w", file.Filename, err)
		}

		size := file.Size
		if size == 0 {
			size = int64(len(data))
		}

		contentType := strings.TrimSpace(file.ContentType)
		if contentType == "" {
			contentType = http.DetectContentType(data)
		}

		files = append(files, common.File{
			Filename:    file.Filename,
			Content:     data,
			ContentType: contentType,
			Size:        size,
		})
	}

	options := req.Options.toProcessingOptions()

	result, err := a.ProcessingService.BulkFileProcessingService.ProcessBulkFiles(context.Background(), userID, files, options)
	if err != nil {
		return nil, fmt.Errorf("bulk file processing failed: %w", err)
	}

	response := &BulkProcessingResponse{
		JobID:        result.JobID,
		TotalFiles:   result.TotalFiles,
		SuccessCount: result.SuccessCount,
		FailureCount: result.FailureCount,
		Duration:     result.Duration,
		Results:      make([]BulkProcessingResultFile, 0, len(result.Results)),
	}

	for _, fileResult := range result.Results {
		item := BulkProcessingResultFile{
			Filename:    fileResult.Filename,
			NewName:     fileResult.NewName,
			Success:     fileResult.Success,
			Error:       fileResult.Error,
			Action:      fileResult.Action,
			ContentType: fileResult.ContentType,
		}

		if fileResult.Success {
			if data, ok := a.ProcessingService.BulkFileProcessingService.GetProcessedData(result.JobID, fileResult.Filename); ok && len(data) > 0 {
				item.ContentBase64 = base64.StdEncoding.EncodeToString(data)
			}
		}

		response.Results = append(response.Results, item)
	}

	return response, nil
}

// GetBulkProcessingJob returns a snapshot of a previously submitted bulk
// processing job, allowing the frontend to poll for updates.
func (a *AdvancedFileOperations) GetBulkProcessingJob(jobID string) (*bulk_file_processing.ProcessingJob, error) {
	if strings.TrimSpace(jobID) == "" {
		return nil, fmt.Errorf("jobID is required")
	}

	return a.ProcessingService.BulkFileProcessingService.GetJob(jobID)
}
