// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package exiftool

import (
	stdErrors "errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
	apperrors "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/security"

	"github.com/barasher/go-exiftool"
)

type serviceImpl struct {
	mu     sync.RWMutex
	et     *exiftool.Exiftool
	config *Config
	logger *slog.Logger
	opts   []func(*exiftool.Exiftool) error
}

var errNoMetadataResults = stdErrors.New("exiftool returned no metadata results")

// newServiceImpl creates a new ExifTool service instance with the provided configuration.
// This is the internal constructor called by NewService.
func newServiceImpl(config *Config, lg *slog.Logger) (Service, error) {
	svc := &serviceImpl{
		config: config,
		logger: lg,
		opts:   buildOptions(config),
	}

	if err := svc.startExiftool(); err != nil {
		// use the service logger (svc.logger) rather than the package value
		svc.logger.Error("Failed to initialize ExifTool", "error", err)
		return nil, apperrors.NewSystemError("failed to initialize ExifTool service", err)
	}

	svc.logger.Info("ExifTool service initialized successfully")
	return svc, nil
}

func buildOptions(config *Config) []func(*exiftool.Exiftool) error {
	var opts []func(*exiftool.Exiftool) error

	if config.Buffer != nil && config.BufferSize > 0 {
		opts = append(opts, exiftool.Buffer(config.Buffer, config.BufferSize))
	}

	if config.Charset != "" {
		opts = append(opts, exiftool.Charset(config.Charset))
	}

	for _, apiValue := range config.ApiValues {
		opts = append(opts, exiftool.Api(apiValue))
	}

	if config.NoPrintConversion {
		opts = append(opts, exiftool.NoPrintConversion())
	}
	if config.ExtractEmbedded {
		opts = append(opts, exiftool.ExtractEmbedded())
	}
	if config.ExtractAllBinaryMetadata {
		opts = append(opts, exiftool.ExtractAllBinaryMetadata())
	}

	if config.DateFormat != "" {
		opts = append(opts, exiftool.DateFormant(config.DateFormat))
	}
	if config.CoordFormat != "" {
		opts = append(opts, exiftool.CoordFormant(config.CoordFormat))
	}

	if config.PrintGroupNames != "" {
		opts = append(opts, exiftool.PrintGroupNames(config.PrintGroupNames))
	}

	if config.BackupOriginal {
		opts = append(opts, exiftool.BackupOriginal())
	}
	if config.ClearFieldsBeforeWriting {
		opts = append(opts, exiftool.ClearFieldsBeforeWriting())
	}

	if config.ExiftoolBinaryPath != "" {
		opts = append(opts, exiftool.SetExiftoolBinaryPath(config.ExiftoolBinaryPath))
	}

	return opts
}

func (s *serviceImpl) startExiftool() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.et != nil {
		return nil
	}

	et, err := exiftool.NewExiftool(s.opts...)
	if err != nil {
		return err
	}

	s.et = et
	return nil
}

func (s *serviceImpl) ensureExiftool() error {
	s.mu.RLock()
	if s.et != nil {
		s.mu.RUnlock()
		return nil
	}
	s.mu.RUnlock()

	if err := s.startExiftool(); err != nil {
		return err
	}

	s.logger.Info("ExifTool process started")
	return nil
}

func (s *serviceImpl) getExiftool() *exiftool.Exiftool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.et
}

func (s *serviceImpl) restartExiftool(operation string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.et != nil {
		if err := s.et.Close(); err != nil {
			s.logger.Warn("Error closing ExifTool during restart", "operation", operation, "error", err)
		}
		s.et = nil
	}

	et, err := exiftool.NewExiftool(s.opts...)
	if err != nil {
		return err
	}

	s.et = et
	s.logger.Info("ExifTool process restarted", "operation", operation)
	return nil
}

func (s *serviceImpl) withExiftoolRetry(operation string, fn func() error) error {
	if err := s.ensureExiftool(); err != nil {
		return err
	}

	if err := fn(); err != nil {
		if !s.shouldRestart(err) {
			return err
		}

		s.logger.Warn("ExifTool failure detected, attempting restart", "operation", operation, "error", err)

		if restartErr := s.restartExiftool(operation); restartErr != nil {
			s.logger.Error("Failed to restart ExifTool", "operation", operation, "error", restartErr)
			return restartErr
		}

		if retryErr := fn(); retryErr != nil {
			return retryErr
		}
	}

	return nil
}

func (s *serviceImpl) shouldRestart(err error) bool {
	if err == nil || err == errNoMetadataResults {
		return false
	}

	if _, ok := err.(*exec.ExitError); ok {
		return true
	}

	if stdErrors.Is(err, io.EOF) {
		return true
	}

	if stdErrors.Is(err, syscall.EPIPE) || stdErrors.Is(err, syscall.ECONNRESET) {
		return true
	}

	var pathErr *os.PathError
	if stdErrors.As(err, &pathErr) {
		if stdErrors.Is(pathErr.Err, syscall.EPIPE) {
			return true
		}
	}

	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "broken pipe") || strings.Contains(msg, "pipe has been ended") {
		return true
	}
	if strings.Contains(msg, "exiftool") && strings.Contains(msg, "exit status") {
		return true
	}
	if strings.Contains(msg, "no such process") {
		return true
	}

	return false
}

// RemoveAllMetadata removes all metadata from a file's content by writing it to a
// temporary file, processing it, and reading it back.
func (s *serviceImpl) RemoveAllMetadata(file *common.File) error {
	tmpPath, cleanup, err := s.createTempFile(file.Content, "exif-remove-")
	if err != nil {
		return err
	}
	defer cleanup()

	operationErr := s.withExiftoolRetry("remove metadata", func() error {
		fm := exiftool.EmptyFileMetadata()
		fm.File = tmpPath
		fm.Clear("all")

		fileMetadatas := []exiftool.FileMetadata{fm}
		tool := s.getExiftool()
		if tool == nil {
			return stdErrors.New("exiftool process is not available")
		}

		tool.WriteMetadata(fileMetadatas)
		return fileMetadatas[0].Err
	})
	if operationErr != nil {
		s.logger.Error("Failed to remove metadata", "filename", file.Filename, "error", operationErr)
		return apperrors.NewFileProcessingError(file.Filename, "remove metadata with exiftool", operationErr)
	}

	// Validate temp path before reading
	validatedTmpPath, err := security.ValidateFilePath(tmpPath, "")
	if err != nil {
		s.logger.Error("Invalid temp file path", "path", tmpPath, "error", err)
		return apperrors.NewSystemError("invalid temp file path for reading processed content", err)
	}

	processedContent, err := os.ReadFile(validatedTmpPath) // #nosec G304 - path validated above
	if err != nil {
		s.logger.Error("Failed to read processed file content", "filename", file.Filename, "error", err)
		return apperrors.NewSystemError("failed to read processed file content after metadata removal", err)
	}

	file.Content = processedContent
	file.Size = int64(len(processedContent))

	return nil
}

// ExtractMetadata reads all metadata from a file's content by writing it to a
// temporary file and then processing it.
func (s *serviceImpl) ExtractMetadata(file *common.File) (map[string]any, error) {
	tmpPath, cleanup, err := s.createTempFile(file.Content, "exif-extract-")
	if err != nil {
		return nil, err
	}
	defer cleanup()

	var results []exiftool.FileMetadata
	operationErr := s.withExiftoolRetry("extract metadata", func() error {
		tool := s.getExiftool()
		if tool == nil {
			return stdErrors.New("exiftool process is not available")
		}

		results = tool.ExtractMetadata(tmpPath)
		if len(results) == 0 {
			return errNoMetadataResults
		}

		return results[0].Err
	})

	if operationErr != nil {
		if stdErrors.Is(operationErr, errNoMetadataResults) {
			s.logger.Error("No results returned from metadata extraction", "filename", file.Filename)
			return nil, apperrors.NewFileProcessingError(file.Filename, "extract metadata with exiftool: no results returned", nil)
		}

		s.logger.Error("Failed to extract metadata", "filename", file.Filename, "error", operationErr)
		return nil, apperrors.NewFileProcessingError(file.Filename, "extract metadata with exiftool", operationErr)
	}

	return results[0].Fields, nil
}

// WriteMetadata writes specific metadata fields to a file's content.
// The file content is updated in-place with the new metadata.
func (s *serviceImpl) WriteMetadata(file *common.File, metadata map[string]any) error {
	tmpPath, cleanup, err := s.createTempFile(file.Content, "exif-write-")
	if err != nil {
		return err
	}
	defer cleanup()

	operationErr := s.withExiftoolRetry("write metadata", func() error {
		fm := exiftool.EmptyFileMetadata()
		fm.File = tmpPath

		for key, value := range metadata {
			switch v := value.(type) {
			case string:
				fm.SetString(key, v)
			case int:
				fm.SetInt(key, int64(v))
			case int64:
				fm.SetInt(key, v)
			case float64:
				fm.SetFloat(key, v)
			case []string:
				fm.SetStrings(key, v)
			default:
				fm.SetString(key, fmt.Sprintf("%v", v))
			}
		}

		fileMetadatas := []exiftool.FileMetadata{fm}
		tool := s.getExiftool()
		if tool == nil {
			return stdErrors.New("exiftool process is not available")
		}

		tool.WriteMetadata(fileMetadatas)
		return fileMetadatas[0].Err
	})

	if operationErr != nil {
		s.logger.Error("Failed to write metadata", "filename", file.Filename, "error", operationErr)
		return apperrors.NewFileProcessingError(file.Filename, "write metadata with exiftool", operationErr)
	}

	// Validate temp path before reading
	validatedTmpPath, err := security.ValidateFilePath(tmpPath, "")
	if err != nil {
		s.logger.Error("Invalid temp file path", "path", tmpPath, "error", err)
		return apperrors.NewSystemError("invalid temp file path for reading processed content", err)
	}

	processedContent, err := os.ReadFile(validatedTmpPath) // #nosec G304 - path validated above
	if err != nil {
		s.logger.Error("Failed to read processed file content", "filename", file.Filename, "error", err)
		return apperrors.NewSystemError("failed to read processed file content after metadata write", err)
	}

	file.Content = processedContent
	file.Size = int64(len(processedContent))

	s.logger.Info("Successfully wrote metadata", "filename", file.Filename, "field_count", len(metadata))
	return nil
}

// createTempFile creates a temporary directory, writes content to a file in that directory,
// and returns the file path and a cleanup function.
// The cleanup function should be called with defer to ensure proper cleanup.
func (s *serviceImpl) createTempFile(content []byte, prefix string) (path string, cleanup func(), err error) {
	// Create temporary directory (no trailing whitespace to avoid invalid paths)
	tmpDir, err := os.MkdirTemp("", prefix+"*")
	if err != nil {
		return "", nil, apperrors.NewSystemError("failed to create temporary directory", err)
	}

	// Define cleanup function that removes the entire directory
	cleanup = func() {
		_ = os.RemoveAll(tmpDir)
	}

	// Create file in temp directory
	tmpFilePath := filepath.Join(tmpDir, "file")
	if err := os.WriteFile(tmpFilePath, content, 0600); err != nil {
		cleanup()
		return "", nil, apperrors.NewSystemError("failed to write to temporary file", err)
	}

	return tmpFilePath, cleanup, nil
}

// ListAvailableFields returns all available metadata field names that ExifTool can read.
// This executes 'exiftool -list' and parses the output to return field names.
func (s *serviceImpl) ListAvailableFields() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.et == nil {
		return nil, apperrors.NewSystemError("ExifTool service not initialized", nil)
	}

	// Validate binary path to prevent command injection
	if s.config.ExiftoolBinaryPath == "" {
		return nil, apperrors.NewSystemError("ExifTool binary path not configured", nil)
	}
	validatedBinaryPath, err := security.ValidatePath(s.config.ExiftoolBinaryPath)
	if err != nil {
		s.logger.Error("Invalid ExifTool binary path", "path", s.config.ExiftoolBinaryPath, "error", err)
		return nil, apperrors.NewSystemError("invalid ExifTool binary path", err)
	}

	// Execute exiftool -list command
	cmd := exec.Command(validatedBinaryPath, "-list") // #nosec G204 - binary path validated above
	output, err := cmd.Output()
	if err != nil {
		s.logger.Error("Failed to execute exiftool -list", "error", err)
		return nil, apperrors.NewSystemError("Failed to list available ExifTool fields", err)
	}

	// Parse the output - each line contains field names separated by spaces
	lines := strings.Split(string(output), "\n")
	var fields []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split by spaces and add each field
		lineFields := strings.Fields(line)
		for _, field := range lineFields {
			field = strings.TrimSpace(field)
			if field != "" {
				fields = append(fields, field)
			}
		}
	}

	s.logger.Debug("Listed available ExifTool fields", "count", len(fields))
	return fields, nil
}

// Close gracefully shuts down the ExifTool service and releases resources.
func (s *serviceImpl) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.et == nil {
		return nil
	}

	s.logger.Info("Closing ExifTool service")
	if err := s.et.Close(); err != nil {
		s.logger.Warn("Error closing ExifTool service, proceeding with shutdown", "error", err)
		s.et = nil
		return nil
	}

	s.et = nil
	s.logger.Info("ExifTool service closed successfully")
	return nil
}
