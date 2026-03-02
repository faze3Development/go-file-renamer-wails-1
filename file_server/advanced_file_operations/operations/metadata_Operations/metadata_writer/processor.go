// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package metadata_writer

import (
	"context"
	"fmt"
	"log/slog"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/exiftool"
)

// Processor implements metadata writing/modification
type Processor struct {
	exiftoolService exiftool.Service
	logger          *slog.Logger
}

// NewProcessor creates a new metadata writer processor
func NewProcessor(exiftoolService exiftool.Service, logger *slog.Logger) *Processor {
	return &Processor{
		exiftoolService: exiftoolService,
		logger:          logger,
	}
}

// Type returns the processor type identifier
func (p *Processor) Type() string {
	return "metadata_write"
}

// ProcessFile is not supported for metadata writing (requires file content)
func (p *Processor) ProcessFile(ctx context.Context, file common.FileOperation, config map[string]any) (*common.FileOperationResult, error) {
	return nil, errors.NewValidationError(
		"metadata writing requires file content - use ProcessFileWithContent instead",
		map[string]any{
			"processor": "metadata_writer",
			"method":    "ProcessFile",
		},
	)
}

// ValidateConfig validates the metadata writer configuration
func (p *Processor) ValidateConfig(config map[string]any) error {
	metadata, ok := config["metadata"].(map[string]any)
	if !ok || len(metadata) == 0 {
		p.logger.Info("Validation failed: no metadata fields provided",
			"metadata_type", fmt.Sprintf("%T", config["metadata"]),
			"metadata_value", config["metadata"],
			"metadata_length", len(metadata))
		return errors.NewValidationError(
			"at least one metadata field must be provided for writing",
			map[string]any{
				"received_metadata": metadata,
			},
		)
	}
	return nil
}

// ProcessFileWithContent writes metadata to a file
func (p *Processor) ProcessFileWithContent(ctx context.Context, filename string, content []byte, contentType string, metadata map[string]any) ([]byte, error) {
	// Create file object
	file := &common.File{
		Filename:    filename,
		Content:     content,
		ContentType: contentType,
		Size:        int64(len(content)),
	}

	// Write metadata using ExifTool service
	err := p.exiftoolService.WriteMetadata(file, metadata)
	if err != nil {
		return nil, errors.NewFileProcessingError(filename, "write metadata", err)
	}

	p.logger.Info("Successfully wrote metadata to file",
		"filename", filename,
		"field_count", len(metadata))

	return file.Content, nil
}
