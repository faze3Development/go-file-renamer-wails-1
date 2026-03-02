// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package operations

import (
	"fmt"
	"log/slog"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/exiftool"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/file_operations/bulk_file_processing"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/file_operations/optimizer"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/file_operations/renamers"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/metadata_Operations/metadata_extraction"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/metadata_Operations/metadata_removal"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/metadata_Operations/metadata_renamer"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/metadata_Operations/metadata_writer"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/monitoring_and_statistics"
)

// Orchestrator holding and providing access to all

// operational services. This simplifies dependency injection into the API layer.

type Orchestrator struct {
	JobStatusService          monitoring_and_statistics.ServiceAPI // Job status/stats
	BulkFileProcessingService bulk_file_processing.ServiceAPI      // Async, large jobs
	ExifToolService           exiftool.Service                     // ExifTool service for metadata operations

	// Feature services
	MetadataExtractionService *metadata_extraction.Service
	MetadataRemovalService    *metadata_removal.Service
	MetadataRenamerService    *metadata_renamer.Service
	OptimizerService          *optimizer.Service

	// Feature handlers
	MetadataRemovalHandler *metadata_removal.Handler
	MetadataRenamerHandler *metadata_renamer.Handler
	MetadataWriterHandler  *metadata_writer.Handler
	OptimizerHandler       *optimizer.Handler
	RenamerHandler         *renamers.Handler
}

// NewOrchestrator creates and initializes a new Orchestrator.

func NewOrchestrator(exiftoolService exiftool.Service, logger *slog.Logger) *Orchestrator {

	jobStatusService := monitoring_and_statistics.NewService()

	bulkProcessingService := bulk_file_processing.NewService(jobStatusService, exiftoolService, logger)

	// Initialize feature services
	metadataExtractionService := metadata_extraction.NewService(exiftoolService, logger)
	metadataRemovalService := metadata_removal.NewService(exiftoolService, logger)
	metadataRenamerService := metadata_renamer.NewService(metadataExtractionService, logger)
	optimizerService := optimizer.NewService(logger)

	// Initialize feature handlers
	metadataRemovalHandler := metadata_removal.NewHandler(metadataRemovalService, logger)
	metadataRenamerHandler := metadata_renamer.NewHandler(metadataRenamerService, logger)
	metadataWriterHandler := metadata_writer.NewHandler(exiftoolService, logger)
	optimizerHandler := optimizer.NewHandler(optimizerService, logger)
	renamerHandler := renamers.NewHandler(logger)

	// Create orchestrator instance for job lookup
	orchestrator := &Orchestrator{
		JobStatusService:          jobStatusService,
		BulkFileProcessingService: bulkProcessingService,
		ExifToolService:           exiftoolService,

		// Feature services
		MetadataExtractionService: metadataExtractionService,
		MetadataRemovalService:    metadataRemovalService,
		MetadataRenamerService:    metadataRenamerService,
		OptimizerService:          optimizerService,

		// Feature handlers
		MetadataRemovalHandler: metadataRemovalHandler,
		MetadataRenamerHandler: metadataRenamerHandler,
		MetadataWriterHandler:  metadataWriterHandler,
		OptimizerHandler:       optimizerHandler,
		RenamerHandler:         renamerHandler,
	}

	return orchestrator

}

// GetExifToolService returns the ExifTool service instance
func (o *Orchestrator) GetExifToolService() exiftool.Service {
	return o.ExifToolService
}

// GetJob looks up a job across all job services (bulk processing, monitoring, etc.)
func (o *Orchestrator) GetJob(jobID string) (any, error) {
	// First try bulk processing service
	if bulkJob, err := o.BulkFileProcessingService.GetJob(jobID); err == nil && bulkJob != nil {
		return bulkJob, nil
	}

	// Could add other job services here in the future
	// For now, just return nil if not found in bulk service
	return nil, fmt.Errorf("job not found")
}
