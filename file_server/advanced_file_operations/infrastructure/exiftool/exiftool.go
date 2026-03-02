// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package exiftool

import (
	"log/slog"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"
)

// Config holds configuration options for the ExifTool service.
// All options correspond to ExifTool command-line parameters.
type Config struct {
	// Buffer settings for reading ExifTool output
	Buffer     []byte
	BufferSize int

	// Character encoding settings
	Charset string

	// API options (can be multiple)
	ApiValues []string

	// Processing options
	NoPrintConversion        bool
	ExtractEmbedded          bool
	ExtractAllBinaryMetadata bool

	// Format options
	DateFormat  string
	CoordFormat string

	// Output options
	PrintGroupNames string

	// File handling options
	BackupOriginal           bool
	ClearFieldsBeforeWriting bool

	// Binary path
	ExiftoolBinaryPath string
}

// DefaultConfig returns a configuration with sensible defaults for production use.
func DefaultConfig() *Config {
	return &Config{
		Buffer:     make([]byte, 128*1024), // 128KB buffer
		BufferSize: 64 * 1024,              // 64KB max size
		Charset:    "filename=utf8",
		ApiValues:  []string{},
		// NoPrintConversion: false (use human-readable output by default)
		// ExtractEmbedded: false (don't extract embedded metadata by default)
		// ExtractAllBinaryMetadata: false (don't extract binary metadata by default)
		DateFormat:  "%Y:%m:%d %H:%M:%S",
		CoordFormat: "%+f",
		// PrintGroupNames: "" (don't print group names by default)
		BackupOriginal:           false, // Overwrite original files
		ClearFieldsBeforeWriting: false, // Don't clear fields before writing
		// ExiftoolBinaryPath: "" (use PATH)
	}
}

// Service defines the interface for interacting with ExifTool.
// This allows for mocking and dependency injection.
type Service interface {
	// RemoveAllMetadata removes all EXIF and other metadata from a file's content.
	RemoveAllMetadata(file *common.File) error

	// ExtractMetadata reads all metadata from a file's content and returns it as a map.
	ExtractMetadata(file *common.File) (map[string]any, error)

	// WriteMetadata writes specific metadata fields to a file's content.
	// The metadata map contains field names as keys and values to write.
	// Example: {"Author": "John Doe", "Copyright": "2024", "Keywords": "travel,sunset"}
	WriteMetadata(file *common.File, metadata map[string]any) error

	// ListAvailableFields returns all available metadata field names that ExifTool can read.
	// This is useful for dynamic pattern discovery and validation.
	ListAvailableFields() ([]string, error)

	// Close gracefully shuts down the ExifTool service and releases resources.
	Close() error
}

// NewService creates a new ExifTool service with the provided configuration.
// The service uses a single ExifTool process with the stay_open feature for optimal performance.
// Thread-safe: The underlying go-exiftool library uses mutex locking for concurrent access.
func NewService(config *Config, logger *slog.Logger) (Service, error) {
	if config == nil {
		config = DefaultConfig()
	}
	if logger == nil {
		return nil, errors.NewSystemError("logger cannot be nil", nil)
	}

	return newServiceImpl(config, logger)
}
