// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package metadata_removal

import (
	"bytes"
	"context"
	"log/slog"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/exiftool"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/metadata_Operations/metadata_removal/pdf"
)

// Processor implements the FileProcessor interface for metadata removal.
// Supports JPEG, PNG, and PDF metadata stripping using ExifTool:
// - Images: Comprehensive metadata removal via ExifTool
// - PDFs: Metadata optimization via pdfcpu
type Processor struct {
	exifService exiftool.Service
	logger      *slog.Logger
}

// NewProcessor creates a new metadata removal processor
func NewProcessor(exifService exiftool.Service, logger *slog.Logger) *Processor {
	return &Processor{exifService: exifService, logger: logger}
}

// Type returns the processor type identifier
func (p *Processor) Type() string {
	return "metadata_removal"
}

// ProcessFile is part of the FileProcessor interface but is not used for metadata removal.
// Metadata removal always requires file content, so use ProcessFileWithContent instead.
// This method exists only for interface compliance and will return an error if called.
func (p *Processor) ProcessFile(ctx context.Context, file common.FileOperation, config map[string]interface{}) (*common.FileOperationResult, error) {
	return nil, errors.NewValidationError(
		"metadata removal requires file content - use ProcessFileWithContent instead",
		map[string]interface{}{
			"processor": "metadata_removal",
			"method":    "ProcessFile",
			"note":      "This method is not supported for metadata operations",
		},
	)
}

// ValidateConfig validates the metadata removal configuration
func (p *Processor) ValidateConfig(config map[string]interface{}) error {
	// Metadata removal doesn't need special configuration
	return nil
}

// GetSupportedTypes returns the file types supported for metadata removal
func (p *Processor) GetSupportedTypes() []string {
	return []string{
		"image/jpeg",
		"image/png",
		"application/pdf",
	}
}

// ProcessFileWithContent processes a file with content for metadata removal
func (p *Processor) ProcessFileWithContent(ctx context.Context, filename string, content []byte, contentType string) ([]byte, error) {
	reader := bytes.NewReader(content)

	// Determine file type and use appropriate stripping method
	if contentType == "application/pdf" || (len(content) > 4 && string(content[:4]) == "%%PDF") {
		// Handle PDF files
		processedContent, err := pdf.StripPDFMetadata(reader, p.logger)
		if err != nil {
			return nil, errors.NewFileProcessingError(filename, "strip PDF metadata", err)
		}
		return processedContent, nil
	} else {
		// Handle image files using ExifTool
		file := &common.File{
			Filename:    filename,
			Content:     content,
			ContentType: contentType,
			Size:        int64(len(content)),
		}

		err := p.exifService.RemoveAllMetadata(file)
		if err != nil {
			return nil, errors.NewFileProcessingError(filename, "strip image metadata", err)
		}
		return file.Content, nil
	}
}
