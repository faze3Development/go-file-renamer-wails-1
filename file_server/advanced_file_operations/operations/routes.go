// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package operations

import (
	"log/slog"

	"github.com/go-chi/chi/v5"

	"go-file-renamer-wails/file_server/advanced_file_operations/operations/file_operations/bulk_file_processing"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/monitoring_and_statistics"
)

// RegisterRoutes initializes all operational handlers and registers their routes.
func RegisterRoutes(r chi.Router, orchestrator *Orchestrator, logger *slog.Logger) {
	// Job status and monitoring endpoints
	processingHandler := monitoring_and_statistics.NewHandler(orchestrator.JobStatusService, orchestrator, logger)
	mux, ok := r.(*chi.Mux)
	if !ok {
		mux = chi.NewRouter()
		r.Mount("/", mux)
	}
	processingHandler.RegisterRoutes(mux)

	// Bulk file processing endpoints
	bulkProcessingHandler := bulk_file_processing.NewHandler(orchestrator.BulkFileProcessingService, logger)
	bulkProcessingHandler.RegisterRoutes(mux)

	// Feature-specific endpoints
	metadataRemovalHandler := orchestrator.MetadataRemovalHandler
	metadataRemovalHandler.RegisterRoutes(mux)

	// Metadata renamer endpoints
	metadataRenamerHandler := orchestrator.MetadataRenamerHandler
	metadataRenamerHandler.RegisterRoutes(mux)

	optimizerHandler := orchestrator.OptimizerHandler
	optimizerHandler.RegisterRoutes(mux)

	// Public renamer endpoints
	renamerHandler := orchestrator.RenamerHandler
	renamerHandler.RegisterRoutes(mux)

	// Metadata writer endpoints
	orchestrator.MetadataWriterHandler.RegisterRoutes(mux)
}
