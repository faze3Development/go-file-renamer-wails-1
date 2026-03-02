package metadata_extraction

import (
	"context"
	"log/slog"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/exiftool"
)

type Service struct {
	exiftoolService exiftool.Service
	processor       *Processor
	logger          *slog.Logger
}

func NewService(exiftoolService exiftool.Service, logger *slog.Logger) *Service {
	return &Service{
		exiftoolService: exiftoolService,
		processor:       NewProcessor(exiftoolService, logger),
		logger:          logger,
	}
}

// ProcessFileWithContent extracts metadata for the provided file payload.
func (s *Service) ProcessFileWithContent(ctx context.Context, filename string, content []byte, contentType string) (*Metadata, error) {
	return s.processor.ProcessFileWithContent(ctx, filename, content, contentType)
}

// Processor exposes the underlying processor for callers needing advanced access.
func (s *Service) Processor() *Processor {
	return s.processor
}
