package metadata_removal

import (
	"log/slog"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/exiftool"
)

type Service struct {
	exiftoolService exiftool.Service
	logger          *slog.Logger
}

func NewService(exiftoolService exiftool.Service, logger *slog.Logger) *Service {
	return &Service{
		exiftoolService: exiftoolService,
		logger:          logger,
	}
}
