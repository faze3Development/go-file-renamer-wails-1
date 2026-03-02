package metadata_renamer

import (
	"log/slog"

	"go-file-renamer-wails/file_server/advanced_file_operations/operations/metadata_Operations/metadata_extraction"
)

type Service struct {
	metadataExtractionService *metadata_extraction.Service
	logger                    *slog.Logger
}

func NewService(metadataExtractionService *metadata_extraction.Service, logger *slog.Logger) *Service {
	return &Service{
		metadataExtractionService: metadataExtractionService,
		logger:                    logger,
	}
}
