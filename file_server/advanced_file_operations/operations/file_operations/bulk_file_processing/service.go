// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package bulk_file_processing

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/exiftool"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/file_operations/optimizer"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/file_operations/renamers"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/metadata_Operations/metadata_extraction"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/metadata_Operations/metadata_removal"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/monitoring_and_statistics"
)

// processedDataCache temporarily stores processed file data in memory
// Key: jobID + ":" + filename, Value: processed data
var processedDataCache = make(map[string][]byte)
var cacheMutex = sync.RWMutex{}

// ServiceAPI defines the interface for the bulk file processing service.
type ServiceAPI interface {
	ProcessBulkFiles(ctx context.Context, userID string, files []common.File, options ProcessingOptions) (*BulkProcessingResult, error)
	GetProcessedData(jobID, filename string) ([]byte, bool)
	GetJob(jobID string) (*ProcessingJob, error)
}

// Service manages bulk file processing operations
type Service struct {
	jobStatusService            monitoring_and_statistics.ServiceAPI
	logger                      *slog.Logger
	renamerProcessor            *renamers.Processor
	metadataProcessor           *metadata_removal.Processor
	metadataExtractionProcessor *metadata_extraction.Processor
	optimizerProcessor          *optimizer.Processor
	jobs                        map[string]*ProcessingJob
	jobsMutex                   sync.RWMutex
}

// NewService creates a new bulk file processing service
func NewService(jobStatusService monitoring_and_statistics.ServiceAPI, exiftoolService exiftool.Service, logger *slog.Logger) *Service {
	return &Service{
		jobStatusService:            jobStatusService,
		logger:                      logger,
		renamerProcessor:            renamers.NewProcessor(),
		metadataProcessor:           metadata_removal.NewProcessor(exiftoolService, logger),
		metadataExtractionProcessor: metadata_extraction.NewProcessor(exiftoolService, logger),
		optimizerProcessor:          optimizer.NewProcessor(),
		jobs:                        make(map[string]*ProcessingJob),
	}
}

// FileUploadMetadata represents file metadata for storage (without content)
type FileUploadMetadata struct {
	Filename    string `firestore:"filename"`
	ContentType string `firestore:"contentType"`
	Size        int64  `firestore:"size"`
}

// SequentialNamingOptions configures sequential renaming behaviour.
type SequentialNamingOptions struct {
	Enabled       bool   `json:"enabled" firestore:"enabled"`
	BaseName      string `json:"baseName" firestore:"baseName"`
	StartIndex    int    `json:"startIndex" firestore:"startIndex"`
	PadLength     int    `json:"padLength" firestore:"padLength"`
	KeepExtension bool   `json:"keepExtension" firestore:"keepExtension"`
}

// RenameOperationOptions captures rename feature toggles.
type RenameOperationOptions struct {
	PreserveOriginalName bool                    `json:"preserveOriginalName" firestore:"preserveOriginalName"`
	AddTimestamp         bool                    `json:"addTimestamp" firestore:"addTimestamp"`
	AddRandomID          bool                    `json:"addRandomId" firestore:"addRandomId"`
	AddCustomDate        bool                    `json:"addCustomDate" firestore:"addCustomDate"`
	CustomDate           string                  `json:"customDate,omitempty" firestore:"customDate,omitempty"`
	UseRegexReplace      bool                    `json:"useRegexReplace" firestore:"useRegexReplace"`
	RegexFind            string                  `json:"regexFind,omitempty" firestore:"regexFind,omitempty"`
	RegexReplace         string                  `json:"regexReplace,omitempty" firestore:"regexReplace,omitempty"`
	Sequential           SequentialNamingOptions `json:"sequentialNaming" firestore:"sequentialNaming"`
}

// ProcessingOptions defines the processing configuration
type ProcessingOptions struct {
	RenameFiles    bool                   `json:"renameFiles" firestore:"renameFiles"`
	RemoveMetadata bool                   `json:"removeMetadata" firestore:"removeMetadata"`
	CompressFiles  bool                   `json:"compressFiles" firestore:"compressFiles"` // 🆕 New compression option
	OptimizeFiles  bool                   `json:"optimizeFiles" firestore:"optimizeFiles"`
	Pattern        string                 `json:"pattern" firestore:"pattern"`
	Namer          string                 `json:"namer" firestore:"namer"`
	Rename         RenameOperationOptions `json:"renameOptions" firestore:"renameOptions"`
	MaxFileSize    int64                  `json:"maxFileSize" firestore:"maxFileSize"`
	AllowedTypes   []string               `json:"allowedTypes" firestore:"allowedTypes"`
}

// ProcessingJob represents a bulk processing job
type ProcessingJob struct {
	ID          string                 `json:"id" firestore:"id"`
	UserID      string                 `json:"userId" firestore:"userId"`
	Status      string                 `json:"status" firestore:"status"`
	Files       []FileUploadMetadata   `json:"files" firestore:"files"`
	Options     ProcessingOptions      `json:"options" firestore:"options"`
	Results     []FileProcessingResult `json:"results" firestore:"results"`
	CreatedAt   time.Time              `json:"createdAt" firestore:"createdAt"`
	StartedAt   *time.Time             `json:"startedAt,omitempty" firestore:"startedAt,omitempty"`
	CompletedAt *time.Time             `json:"completedAt,omitempty" firestore:"completedAt,omitempty"`
	Duration    *int64                 `json:"durationMs,omitempty" firestore:"durationMs,omitempty"`
	Error       string                 `json:"-" firestore:"-"` // Internal error, not for JSON response
}

// FileProcessingResult represents the result of processing a single file
type FileProcessingResult struct {
	Filename    string `json:"filename" firestore:"filename"`
	NewName     string `json:"newName,omitempty" firestore:"newName,omitempty"`
	Success     bool   `json:"success" firestore:"success"`
	Action      string `json:"action" firestore:"action"`
	Error       string `json:"error,omitempty" firestore:"error,omitempty"`
	DataID      string `json:"-" firestore:"dataId"` // Reference to stored file data
	ContentType string `json:"contentType,omitempty" firestore:"contentType,omitempty"`
}

// BulkProcessingResult represents the overall result of bulk processing
type BulkProcessingResult struct {
	JobID        string                 `json:"jobId"`
	TotalFiles   int                    `json:"totalFiles"`
	SuccessCount int                    `json:"successCount"`
	FailureCount int                    `json:"failureCount"`
	Results      []FileProcessingResult `json:"results"`
	Duration     int64                  `json:"durationMs"`
}

// Request represents a bulk file processing request
type Request struct {
	Files  []common.FileOperation `json:"files"`
	Config map[string]interface{} `json:"config"`
}

// ProcessBulkFiles processes multiple files in bulk
func (s *Service) ProcessBulkFiles(ctx context.Context, userID string, files []common.File, options ProcessingOptions) (*BulkProcessingResult, error) {
	// Validate files
	if len(files) == 0 {
		return nil, errors.NewValidationError("no files provided", nil)
	}

	// Check file limits
	if len(files) > 100 {
		return nil, errors.NewValidationError("maximum 100 files allowed per batch", nil)
	}

	// Convert files to metadata for storage
	filesMetadata := make([]FileUploadMetadata, 0, len(files))
	for _, file := range files {
		filesMetadata = append(filesMetadata, FileUploadMetadata{
			Filename:    file.Filename,
			ContentType: file.ContentType,
			Size:        file.Size,
		})
	}

	// Create processing job
	job := &ProcessingJob{
		ID:        generateJobID(),
		UserID:    userID,
		Status:    "processing",
		Files:     filesMetadata,
		Options:   options,
		Results:   make([]FileProcessingResult, 0, len(files)),
		CreatedAt: time.Now(),
	}

	now := time.Now()
	job.StartedAt = &now

	s.jobsMutex.Lock()
	s.jobs[job.ID] = job
	s.jobsMutex.Unlock()

	var renameNamer renamers.Namer
	var err error
	if options.RenameFiles {
		renameNamer, err = renamers.NewPatternNamer(options.Pattern, renamers.Options{
			PreserveOriginalName: options.Rename.PreserveOriginalName,
		})
		if err != nil {
			return nil, errors.NewValidationError("invalid rename pattern", err)
		}
	}

	// Process files concurrently
	var wg sync.WaitGroup
	resultsChan := make(chan FileProcessingResult, len(files))

	for _, file := range files {
		wg.Add(1)
		go func(f common.File) {
			defer wg.Done()
			resultsChan <- s.processFile(job.ID, f, options, renameNamer)
		}(file)
	}

	wg.Wait()
	close(resultsChan)

	results := make([]FileProcessingResult, 0, len(files))
	successCount := 0
	failureCount := 0
	for result := range resultsChan {
		results = append(results, result)
		if result.Success {
			successCount++
		} else {
			failureCount++
		}
	}

	// Update job
	completedAt := time.Now()
	job.CompletedAt = &completedAt
	duration := completedAt.Sub(now).Milliseconds()

	s.jobsMutex.Lock()
	job.Results = results
	job.Status = "completed"
	job.Duration = &duration
	s.jobsMutex.Unlock()

	// Log processing completion
	s.logger.InfoContext(ctx, "Bulk processing completed",
		"jobId", job.ID,
		"userId", userID,
		"totalFiles", len(files),
		"successCount", successCount,
		"failureCount", failureCount,
		"duration", duration)

	return &BulkProcessingResult{
		JobID:        job.ID,
		TotalFiles:   len(files),
		SuccessCount: successCount,
		FailureCount: failureCount,
		Results:      results,
		Duration:     duration,
	}, nil
}

// GetJob returns a snapshot of the specified processing job, if it exists.
func (s *Service) GetJob(jobID string) (*ProcessingJob, error) {
	s.jobsMutex.RLock()
	defer s.jobsMutex.RUnlock()
	job, ok := s.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("job %s not found", jobID)
	}

	// Return a shallow copy to prevent external mutation of internal state.
	copied := *job
	copied.Files = append([]FileUploadMetadata(nil), job.Files...)
	copied.Results = append([]FileProcessingResult(nil), job.Results...)
	return &copied, nil
}

// processFile processes a single file
func (s *Service) processFile(jobID string, file common.File, options ProcessingOptions, namer renamers.Namer) FileProcessingResult {
	result := FileProcessingResult{
		Filename:    file.Filename,
		ContentType: file.ContentType,
	}
	processedContent := file.Content // Start with original content
	result.Action = "processed"

	// Apply renaming for paying users
	if options.RenameFiles {
		newName, err := renamers.FormatWithNamer(namer, file.Filename)
		if err != nil {
			s.logger.Warn("Failed to generate new name", "error", err, "filename", file.Filename)
			result.Success = false
			result.Error = "Failed to generate new name: " + err.Error()
			return result
		}
		result.NewName = newName
		result.Action = "renamed"
	} else {
		result.NewName = file.Filename
	}

	// Apply metadata removal for paying users
	if options.RemoveMetadata {
		strippedContent, err := s.metadataProcessor.ProcessFileWithContent(context.Background(), file.Filename, processedContent, file.ContentType)
		if err != nil {
			s.logger.Warn("Failed to remove metadata", "error", err, "filename", file.Filename)
			result.Success = false
			result.Error = "Failed to remove metadata: " + err.Error()
			return result
		}
		processedContent = strippedContent

		if result.Action == "renamed" {
			result.Action = "renamed_metadata_removed"
		} else {
			result.Action = "metadata_removed"
		}
	}

	// Apply optimization for paying users
	if options.OptimizeFiles {
		optimizedContent, err := s.optimizerProcessor.ProcessFileWithContent(context.Background(), file.Filename, processedContent, file.ContentType)
		if err != nil {
			s.logger.Warn("Failed to optimize file", "error", err, "filename", file.Filename)
			result.Success = false
			result.Error = "Failed to optimize file: " + err.Error()
			return result
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

	// 🆕 Apply file compression - just add this block!
	if options.CompressFiles {
		compressedContent, err := compressFile(processedContent, file.ContentType)
		if err != nil {
			s.logger.Warn("Failed to compress file", "error", err, "filename", file.Filename)
			result.Success = false
			result.Error = "Failed to compress file: " + err.Error()
			return result
		}
		processedContent = compressedContent

		// Update action string
		if result.Action != "" && result.Action != "processed" {
			result.Action += "_compressed"
		} else {
			result.Action = "compressed"
		}
	}

	// Generate cache key and store processed content
	cacheKey := fmt.Sprintf("%s:%s", jobID, file.Filename)
	cacheMutex.Lock()
	processedDataCache[cacheKey] = processedContent
	cacheMutex.Unlock()

	result.Success = true
	return result
}

// compressFile applies compression to file content based on type
func compressFile(content []byte, contentType string) ([]byte, error) {
	// For now, just return the original content (placeholder implementation)
	// In a real implementation, you could use gzip, brotli, or format-specific compression
	// For images: use image optimization libraries
	// For text: use gzip compression
	// For other formats: appropriate compression algorithms

	// Placeholder: simulate some compression by removing redundant whitespace for text files
	if strings.HasPrefix(contentType, "text/") {
		// Simple text compression - remove extra whitespace
		compressed := strings.Fields(strings.TrimSpace(string(content)))
		return []byte(strings.Join(compressed, " ")), nil
	}

	// For other file types, return as-is for now
	return content, nil
}

// generateJobID generates a unique job ID
func generateJobID() string {
	return fmt.Sprintf("bulk_%d", time.Now().UnixNano())
}

// GetProcessedData retrieves processed file data from cache
func (s *Service) GetProcessedData(jobID, filename string) ([]byte, bool) {
	cacheKey := fmt.Sprintf("%s:%s", jobID, filename)
	cacheMutex.RLock()
	data, exists := processedDataCache[cacheKey]
	cacheMutex.RUnlock()
	return data, exists
}
