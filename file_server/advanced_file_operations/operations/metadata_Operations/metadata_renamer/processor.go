// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package metadata_renamer

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"
	"go-file-renamer-wails/file_server/advanced_file_operations/operations/metadata_Operations/metadata_extraction"
)

// PatternDefinition represents a metadata renaming pattern

type PatternDefinition struct {
	ID string `json:"id"`

	Name string `json:"name"`

	Description string `json:"description"`

	Example string `json:"example"`
}

// Processor implements metadata-based file renaming

type Processor struct {
	service *Service

	logger *slog.Logger
}

// NewProcessor creates a new metadata renamer processor

func NewProcessor(service *Service, logger *slog.Logger) *Processor {
	return &Processor{
		service: service,
		logger:  logger,
	}
}

// Type returns the processor type identifier
func (p *Processor) Type() string {
	return "metadata_rename"
}

// ProcessFile is not supported for metadata-based renaming (requires file content)
func (p *Processor) ProcessFile(ctx context.Context, file common.FileOperation, config map[string]any) (*common.FileOperationResult, error) {
	return nil, errors.NewValidationError(
		"metadata renaming requires file content - use ProcessFileWithContent instead",
		map[string]any{
			"processor": "metadata_renamer",
			"method":    "ProcessFile",
		},
	)
}

// ValidateConfig validates the metadata renamer configuration
func (p *Processor) ValidateConfig(config map[string]any) error {
	pattern, ok := config["pattern"].(string)
	if !ok || pattern == "" {
		return errors.NewValidationError("pattern is required for metadata renaming", nil)
	}
	return nil
}

// ProcessFileWithContent generates a new filename based on file metadata
func (p *Processor) ProcessFileWithContent(ctx context.Context, filename string, content []byte, contentType string, pattern string) (string, error) {
	// Use the metadata extraction processor for consistent extraction
	metadata, err := p.service.metadataExtractionService.ProcessFileWithContent(ctx, filename, content, contentType)
	if err != nil {
		return "", errors.NewFileProcessingError(filename, "extract metadata for renaming", err)
	}

	return p.generateFilename(filename, metadata.Raw, pattern)
}

// ProcessFileWithContentAndPattern generates a new filename based on file metadata and pattern
func (p *Processor) ProcessFileWithContentAndPattern(ctx context.Context, filename string, content []byte, contentType string, pattern string) (string, error) {
	return p.ProcessFileWithContent(ctx, filename, content, contentType, pattern)
}

// GetAvailablePatterns returns all supported metadata renaming patterns
func (p *Processor) GetAvailablePatterns() []PatternDefinition {
	return []PatternDefinition{
		{
			ID:          "exif:datetime",
			Name:        "Date & Time",
			Description: "Rename using EXIF creation date/time",
			Example:     "2024-01-15_14-30-25.jpg",
		},
		{
			ID:          "exif:camera",
			Name:        "Camera Model",
			Description: "Rename using camera make and model",
			Example:     "Canon_EOS_R5.jpg",
		},
		{
			ID:          "exif:location",
			Name:        "GPS Location",
			Description: "Rename using GPS coordinates",
			Example:     "40.7128_-74.0060.jpg",
		},
		{
			ID:          "exif:combined",
			Name:        "Combined Metadata",
			Description: "Rename using camera, date, and ISO",
			Example:     "Canon_2024-01-15_ISO800.jpg",
		},
	}
}

// GetAvailablePatternsForMetadata returns patterns based on what metadata fields are actually present
func (p *Processor) GetAvailablePatternsForMetadata(metadata *metadata_extraction.Metadata) []PatternDefinition {
	if metadata == nil || len(metadata.Raw) == 0 {
		p.logger.Warn("No metadata available, returning default patterns")
		return p.GetAvailablePatterns()
	}

	var patterns []PatternDefinition

	// Check for datetime pattern - look for any datetime-related fields
	if p.hasDateTimeFields(metadata.Raw) {
		patterns = append(patterns, PatternDefinition{
			ID:          "exif:datetime",
			Name:        "Date & Time",
			Description: "Rename using available date/time metadata",
			Example:     "2024-01-15_14-30-25.jpg",
		})
	}

	// Check for camera pattern - look for any camera/manufacturer fields
	if p.hasCameraFields(metadata.Raw) {
		patterns = append(patterns, PatternDefinition{
			ID:          "exif:camera",
			Name:        "Camera/Device Info",
			Description: "Rename using camera or device information",
			Example:     "Canon_EOS_R5.jpg",
		})
	}

	// Check for location pattern - look for any GPS/location fields
	if p.hasLocationFields(metadata.Raw) {
		patterns = append(patterns, PatternDefinition{
			ID:          "exif:location",
			Name:        "GPS Location",
			Description: "Rename using GPS coordinates",
			Example:     "40.7128_-74.0060.jpg",
		})
	}

	// Check for combined pattern - requires at least datetime + camera or datetime + location
	hasDateTime := p.hasDateTimeFields(metadata.Raw)
	hasCamera := p.hasCameraFields(metadata.Raw)
	hasLocation := p.hasLocationFields(metadata.Raw)

	if hasDateTime && (hasCamera || hasLocation) {
		patterns = append(patterns, PatternDefinition{
			ID:          "exif:combined",
			Name:        "Combined Metadata",
			Description: "Rename using multiple metadata fields",
			Example:     "Canon_2024-01-15_ISO800.jpg",
		})
	}

	// If no patterns were found, return default patterns
	if len(patterns) == 0 {
		p.logger.Warn("No patterns could be generated from metadata, returning defaults")
		return p.GetAvailablePatterns()
	}

	p.logger.Debug("Generated patterns from metadata", "count", len(patterns), "hasDateTime", hasDateTime, "hasCamera", hasCamera, "hasLocation", hasLocation)
	return patterns
}

// hasDateTimeFields checks if any datetime-related fields are available in the metadata
func (p *Processor) hasDateTimeFields(metadata map[string]any) bool {
	datetimeFields := []string{
		"DateTimeOriginal", "CreateDate", "ModifyDate", "DateTime",
		"DateTimeDigitized", "DateCreated", "DateModified", "CreationDate",
		"FileCreateDate", "FileModifyDate", "FileAccessDate", // File system dates
	}

	for _, field := range datetimeFields {
		if _, exists := metadata[field]; exists {
			return true
		}
	}
	return false
}

// hasCameraFields checks if any camera-related fields are available in the metadata
func (p *Processor) hasCameraFields(metadata map[string]any) bool {
	cameraFields := []string{
		"Make", "Model", "CameraMake", "CameraModel", "LensMake", "LensModel",
		"CameraManufacturer", "CameraModelName", "LensManufacturer", "LensModelName",
		"DeviceManufacturer", "DeviceModel", // Device/Profile fields
	}

	for _, field := range cameraFields {
		if _, exists := metadata[field]; exists {
			return true
		}
	}
	return false
}

// hasLocationFields checks if any GPS/location-related fields are available in the metadata
func (p *Processor) hasLocationFields(metadata map[string]any) bool {
	locationFields := []string{
		"GPSLatitude", "GPSLongitude", "GPSLatitudeRef", "GPSLongitudeRef",
		"GPSPosition", "GPSAltitude", "GPSAltitudeRef", "GPSLocation",
		"Latitude", "Longitude", "Location", "GPS",
	}

	for _, field := range locationFields {
		if _, exists := metadata[field]; exists {
			return true
		}
	}
	return false
}

// generateFilename creates a new filename based on metadata and pattern
func (p *Processor) generateFilename(original string, metadata map[string]any, pattern string) (string, error) {
	ext := filepath.Ext(original)

	switch pattern {
	case "exif:datetime":
		return p.generateDateTimeFilename(metadata, ext)
	case "exif:camera":
		return p.generateCameraFilename(metadata, ext)
	case "exif:location":
		return p.generateLocationFilename(metadata, ext)
	case "exif:combined":
		return p.generateCombinedFilename(metadata, ext)
	default:
		return "", errors.NewValidationError("unknown metadata pattern", pattern)
	}
}

// generateDateTimeFilename creates filename from EXIF date/time
func (p *Processor) generateDateTimeFilename(metadata map[string]any, ext string) (string, error) {
	// Try multiple date fields
	dateFields := []string{"DateTimeOriginal", "CreateDate", "ModifyDate", "DateTime"}

	for _, field := range dateFields {
		if val, ok := metadata[field]; ok {
			dateStr := fmt.Sprintf("%v", val)
			// Parse and format: "2024:01:15 14:30:25" -> "2024-01-15_14-30-25"
			cleaned := strings.ReplaceAll(dateStr, ":", "-")
			cleaned = strings.ReplaceAll(cleaned, " ", "_")
			return cleaned + ext, nil
		}
	}

	// Fallback to current timestamp
	return time.Now().Format("2006-01-02_15-04-05") + ext, nil
}

// generateCameraFilename creates filename from camera info
func (p *Processor) generateCameraFilename(metadata map[string]any, ext string) (string, error) {
	make := p.getMetadataString(metadata, "Make")
	model := p.getMetadataString(metadata, "Model")

	if make == "" && model == "" {
		return "Unknown_Camera" + ext, nil
	}

	// Clean and combine: "Canon EOS R5" -> "Canon_EOS_R5"
	name := strings.TrimSpace(make + " " + model)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "-")

	return name + ext, nil
}

// generateLocationFilename creates filename from GPS coordinates
func (p *Processor) generateLocationFilename(metadata map[string]any, ext string) (string, error) {
	lat := p.getMetadataString(metadata, "GPSLatitude")
	lon := p.getMetadataString(metadata, "GPSLongitude")

	if lat == "" || lon == "" {
		return "No_Location" + ext, nil
	}

	// Format: "40.7128_-74.0060"
	name := fmt.Sprintf("%s_%s", lat, lon)
	name = strings.ReplaceAll(name, " ", "")

	return name + ext, nil
}

// generateCombinedFilename creates comprehensive filename
func (p *Processor) generateCombinedFilename(metadata map[string]any, ext string) (string, error) {
	parts := []string{}

	// Add camera
	if make := p.getMetadataString(metadata, "Make"); make != "" {
		parts = append(parts, strings.ReplaceAll(make, " ", "_"))
	}

	// Add date
	if date := p.getMetadataString(metadata, "DateTimeOriginal"); date != "" {
		datePart := strings.Split(date, " ")[0]
		datePart = strings.ReplaceAll(datePart, ":", "-")
		parts = append(parts, datePart)
	}

	// Add ISO
	if iso := p.getMetadataString(metadata, "ISO"); iso != "" {
		parts = append(parts, "ISO"+iso)
	}

	if len(parts) == 0 {
		return "metadata" + ext, nil
	}

	return strings.Join(parts, "_") + ext, nil
}

// getMetadataString safely extracts a string value from metadata
func (p *Processor) getMetadataString(metadata map[string]any, key string) string {
	if val, ok := metadata[key]; ok {
		return strings.TrimSpace(fmt.Sprintf("%v", val))
	}
	return ""
}
