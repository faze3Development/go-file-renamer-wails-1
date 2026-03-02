// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package metadata_extraction

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/exiftool"
)

// Processor implements metadata extraction logic backed by ExifTool.
type Processor struct {
	exifService exiftool.Service
	logger      *slog.Logger
}

// NewProcessor creates a new metadata extraction processor
func NewProcessor(exifService exiftool.Service, logger *slog.Logger) *Processor {
	return &Processor{
		logger:      logger,
		exifService: exifService,
	}
}

// Type returns the processor type identifier
func (p *Processor) Type() string {
	return "metadata_extraction"
}

// ProcessFile is part of the FileProcessor interface but is not used for metadata extraction.
// Metadata extraction always requires file content, so use ProcessFileWithContent instead.
// This method exists only for interface compliance and will return an error if called.
func (p *Processor) ProcessFile(ctx context.Context, file common.FileOperation, config map[string]any) (*common.FileOperationResult, error) {
	return nil, errors.NewValidationError(
		"metadata extraction requires file content - use ProcessFileWithContent instead",
		map[string]interface{}{
			"processor": "metadata_extraction",
			"method":    "ProcessFile",
			"note":      "This method is not supported for metadata operations",
		},
	)
}

// ValidateConfig validates the metadata extraction configuration
func (p *Processor) ValidateConfig(config map[string]any) error {
	// Metadata extraction doesn't need special configuration
	return nil
}

// GetSupportedTypes returns the file types supported for metadata extraction
// ExifTool supports hundreds of file formats, so this is a subset of commonly supported types
func (p *Processor) GetSupportedTypes() []string {
	return []string{
		// Images
		"image/jpeg", "image/png", "image/gif", "image/webp", "image/tiff",
		"image/bmp", "image/heic", "image/heif", "image/raw", "image/cr2",
		"image/nef", "image/arw", "image/dng",

		// Videos
		"video/mp4", "video/avi", "video/mov", "video/mkv", "video/webm",
		"video/flv", "video/wmv", "video/m4v",

		// Audio
		"audio/mp3", "audio/wav", "audio/flac", "audio/m4a", "audio/aac",
		"audio/ogg", "audio/wma",

		// Documents
		"application/pdf", "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint", "application/vnd.openxmlformats-officedocument.presentationml.presentation",

		// Archives
		"application/zip", "application/x-rar-compressed", "application/x-7z-compressed",

		// Other common formats
		"text/plain", "application/json", "application/xml",
	}
}

// ProcessFileWithContent processes a file with content for metadata extraction using ExifTool
func (p *Processor) ProcessFileWithContent(ctx context.Context, filename string, content []byte, contentType string) (*Metadata, error) {
	// Create a common.File struct to pass to the exiftool service.
	file := &common.File{
		Filename:    filename,
		Content:     content,
		ContentType: contentType,
		Size:        int64(len(content)),
	}

	// Extract metadata using ExifTool
	rawFields, err := p.exifService.ExtractMetadata(file)
	if err != nil {
		// The service already provides a well-structured error.
		return nil, err
	}

	metadata := &Metadata{
		Filename:    filename,
		Size:        file.Size,
		ContentType: contentType,
	}

	// Process the extracted metadata
	metadata.HasMetadata = len(rawFields) > 0

	// Store all raw metadata
	metadata.Raw = rawFields

	// Try to categorize metadata by type
	p.categorizeMetadata(metadata, rawFields)

	p.logger.Info("Successfully extracted metadata", "filename", filename, "field_count", len(rawFields))

	return metadata, nil
}

// categorizeMetadata attempts to organize raw ExifTool metadata into structured categories
func (p *Processor) categorizeMetadata(metadata *Metadata, fields map[string]any) {
	// Check content type to determine categorization approach
	contentType := strings.ToLower(metadata.ContentType)

	// Image files
	if strings.HasPrefix(contentType, "image/") {
		metadata.Image = p.extractImageMetadataFromFields(fields)
	}

	// PDF files
	if contentType == "application/pdf" || strings.Contains(contentType, "pdf") {
		metadata.PDF = p.extractPDFMetadataFromFields(fields)
	}

	// Other file types - we can add more categorization logic here as needed
}

// extractImageMetadataFromFields extracts image metadata from ExifTool fields
func (p *Processor) extractImageMetadataFromFields(fields map[string]any) *ImageMetadata {
	meta := &ImageMetadata{
		HasExif:  false,
		ExifData: make(map[string]any),
	}

	// Check for EXIF data presence
	if len(fields) > 0 {
		meta.HasExif = true
	}

	// Extract dimensions
	meta.Width = getInt(fields, "ImageWidth")
	meta.Height = getInt(fields, "ImageHeight")

	// Camera information
	meta.CameraMake = getString(fields, "Make")
	meta.CameraModel = getString(fields, "Model")
	meta.LensModel = getString(fields, "LensModel")

	// Camera settings
	meta.ISO = getString(fields, "ISO")
	meta.Aperture = getString(fields, "FNumber", "Aperture")
	meta.ShutterSpeed = getString(fields, "ExposureTime", "ShutterSpeed")
	meta.FocalLength = getString(fields, "FocalLength")

	// Orientation
	meta.Orientation = getString(fields, "Orientation")

	// Software
	meta.Software = getString(fields, "Software")

	// Date and time
	meta.DateTime = getString(fields, "DateTimeOriginal", "CreateDate")

	// Color space
	meta.ColorSpace = getString(fields, "ColorSpace")

	// GPS location data
	lat := getFloat(fields, "GPSLatitude")
	lon := getFloat(fields, "GPSLongitude")
	if lat != 0 || lon != 0 {
		meta.Location = &LocationMetadata{
			Latitude:  lat,
			Longitude: lon,
			Altitude:  getFloat(fields, "GPSAltitude"),
		}
	}

	// Store all EXIF data for advanced users
	meta.ExifData = fields

	return meta
}

// extractPDFMetadataFromFields extracts PDF metadata from ExifTool fields
func (p *Processor) extractPDFMetadataFromFields(fields map[string]any) *PDFMetadata {
	meta := &PDFMetadata{}

	meta.Title = getString(fields, "Title")
	meta.Author = getString(fields, "Author")
	meta.Subject = getString(fields, "Subject")
	meta.Creator = getString(fields, "Creator")
	meta.Producer = getString(fields, "Producer")
	meta.Keywords = getString(fields, "Keywords")
	meta.PageCount = getInt(fields, "PageCount")
	meta.PDFVersion = getString(fields, "PDFVersion")
	meta.CreationDate = getString(fields, "CreateDate")
	meta.ModDate = getString(fields, "ModifyDate")

	return meta
}

// --- Private Helper Functions ---

// getString safely extracts a string from the map, checking multiple keys as fallbacks.
func getString(fields map[string]any, keys ...string) string {
	for _, key := range keys {
		if val, ok := fields[key]; ok {
			return fmt.Sprintf("%v", val)
		}
	}
	return ""
}

// getInt safely extracts an integer from the map.
func getInt(fields map[string]any, key string) int {
	if val, ok := fields[key]; ok {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int64:
			return int(v)
		case int:
			return v
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return 0
}

// getFloat safely extracts a float64 from the map.
func getFloat(fields map[string]any, key string) float64 {
	if val, ok := fields[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int64:
			return float64(v)
		case int:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return 0
}
