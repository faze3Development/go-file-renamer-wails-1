// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package security

import (
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"

	"github.com/go-playground/validator/v10"
	"github.com/mrz1836/go-sanitize"
)

// SanitizeInput sanitizes user input using go-sanitize library
func SanitizeInput(input string) string {
	if input == "" {
		return input
	}
	// Use go-sanitize XSS function to remove attack vectors
	input = sanitize.XSS(input)
	// Remove HTML tags for additional safety
	input = sanitize.HTML(input)
	// Trim whitespace and limit length
	input = strings.TrimSpace(input)
	if len(input) > 1000 {
		input = input[:1000]
	}
	return input
}

// SanitizeFilename sanitizes filenames using go-sanitize PathName function
func SanitizeFilename(filename string) string {
	if filename == "" {
		return "unnamed_file"
	}
	// Use go-sanitize PathName - removes dangerous path characters
	filename = sanitize.PathName(filename)
	// Ensure reasonable length
	if len(filename) > 255 {
		filename = filename[:255]
	}
	// Ensure non-empty filename
	if filename == "" {
		filename = "unnamed_file"
	}
	return filename
}

// ValidateFileType validates file content type against allowed types
func ValidateFileType(contentType string) bool {
	allowedTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"application/pdf",
	}

	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

// CheckForInjection detects potential injection attacks in input
func CheckForInjection(input string) bool {
	// Use go-sanitize XSS function to detect injection patterns
	sanitized := sanitize.XSS(input)
	return sanitized != input // If input changed, injection detected
}

// ValidateFileCount checks if file count is within limits
func ValidateFileCount(fileCount int, maxFiles int) error {
	if fileCount > maxFiles {
		return fmt.Errorf("too many files: maximum %d files allowed", maxFiles)
	}
	return nil
}

// ValidateIndividualFileSize checks individual file size
func ValidateIndividualFileSize(size int64, maxSize int64) error {
	if size > maxSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", size, maxSize)
	}
	return nil
}

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// RequestValidator provides centralized validation for file processing requests
type RequestValidator struct {
	logger *slog.Logger
}

// NewRequestValidator creates a new request validator
func NewRequestValidator(logger *slog.Logger) *RequestValidator {
	return &RequestValidator{logger: logger}
}

// ValidateAndParseMultipartRequest validates and parses a multipart form request
func (v *RequestValidator) ValidateAndParseMultipartRequest(r *http.Request, maxPayloadSize int64) ([]common.File, map[string]interface{}, error) {
	// Parse multipart form
	err := r.ParseMultipartForm(maxPayloadSize)
	if err != nil {
		v.logger.Warn("Failed to parse multipart form", "error", err)
		return nil, nil, errors.NewValidationError("Failed to parse form data", map[string]interface{}{
			"reason": "multipart form parsing failed",
			"error":  err.Error(),
		})
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		return nil, nil, errors.NewValidationError("No files provided", map[string]interface{}{
			"reason": "empty file list",
		})
	}

	// Validate and convert files
	commonFiles, err := v.validateAndConvertFiles(files, common.MaxFileSize10MB)
	if err != nil {
		return nil, nil, err
	}

	if len(commonFiles) == 0 {
		return nil, nil, errors.NewValidationError("No valid files to process", map[string]interface{}{
			"reason": "all files failed validation",
		})
	}

	// Extract processing configuration
	config := v.extractProcessingConfig(r)

	return commonFiles, config, nil
}

// validateAndConvertFiles validates and converts multipart files to common.File structs
func (v *RequestValidator) validateAndConvertFiles(files []*multipart.FileHeader, maxFileSize int64) ([]common.File, error) {
	var commonFiles []common.File

	for _, fileHeader := range files {
		// Sanitize filename
		sanitizedFilename := SanitizeFilename(fileHeader.Filename)

		// Check for injection in filename
		if CheckForInjection(fileHeader.Filename) {
			v.logger.Warn("Potential injection detected in filename", "original", fileHeader.Filename)
			continue // Skip potentially dangerous files
		}

		file, err := fileHeader.Open()
		if err != nil {
			v.logger.Warn("Failed to open file", "error", err, "filename", sanitizedFilename)
			continue
		}

		// Validate individual file size
		if err := ValidateIndividualFileSize(fileHeader.Size, maxFileSize); err != nil {
			v.logger.Warn("File size validation failed", "error", err, "filename", sanitizedFilename)
			file.Close()
			continue
		}

		// Read file content
		fileContent, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			v.logger.Warn("Failed to read file content", "error", err, "filename", sanitizedFilename)
			continue
		}

		// Validate file type
		contentType := fileHeader.Header.Get("Content-Type")
		if !ValidateFileType(contentType) {
			v.logger.Warn("Invalid file type", "contentType", contentType, "filename", sanitizedFilename)
			continue
		}

		commonFile := common.File{
			Filename:    sanitizedFilename,
			Content:     fileContent,
			ContentType: contentType,
			Size:        fileHeader.Size,
		}

		commonFiles = append(commonFiles, commonFile)
	}

	return commonFiles, nil
}

// extractProcessingConfig extracts processing configuration from form data
func (v *RequestValidator) extractProcessingConfig(r *http.Request) map[string]interface{} {
	config := make(map[string]interface{})

	// Boolean processing options
	if r.FormValue("renameFiles") == "true" {
		config["renameFiles"] = true
	}
	if r.FormValue("removeMetadata") == "true" {
		config["removeMetadata"] = true
	}
	if r.FormValue("optimizeFiles") == "true" {
		config["optimizeFiles"] = true
	}

	// String configuration options with sanitization
	if pattern := r.FormValue("pattern"); pattern != "" {
		config["pattern"] = SanitizeInput(pattern)
	}

	// Numeric configuration options
	if maxFileSize := r.FormValue("maxFileSize"); maxFileSize != "" {
		config["maxFileSize"] = common.ParseMaxFileSizeFromString(maxFileSize, common.MaxFileSize50MB)
	}

	return config
}

// ValidateJobRequest validates job-specific parameters
func (v *RequestValidator) ValidateJobRequest(fileCount int) error {
	// Validate file count based on authentication status
	maxFiles := 100 // Authenticated limit
	return ValidateFileCount(fileCount, maxFiles)
}

// FileValidationConfig holds configuration for file validation
type FileValidationConfig struct {
	MaxSizeBytes int64
	AllowedTypes []string
	BlockedTypes []string
	AllowedExts  []string
	BlockedExts  []string
}

// DefaultFileConfig returns secure defaults for file validation
func DefaultFileConfig() *FileValidationConfig {
	return &FileValidationConfig{
		MaxSizeBytes: common.MaxFileSize50MB, // 50MB default
		AllowedTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
			"application/pdf",
		},
		BlockedTypes: []string{
			"application/x-executable",
			"application/x-msdownload",
			"application/x-script",
		},
		AllowedExts: []string{
			".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf",
		},
		BlockedExts: []string{
			".exe", ".bat", ".cmd", ".com", ".pif", ".scr", ".vbs", ".js", ".jar",
		},
	}
}

// ValidateFile performs comprehensive file validation
func ValidateFile(r *http.Request, config *FileValidationConfig) (*multipart.FileHeader, error) {
	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, errors.NewValidationError("failed to read file", err)
	}

	// Validate file size
	if header.Size > config.MaxSizeBytes {
		file.Close()
		return nil, errors.NewPayloadTooLargeError(config.MaxSizeBytes, header.Size)
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !contains(config.AllowedExts, ext) {
		file.Close()
		return nil, errors.NewValidationError("file extension is not allowed", ext)
	}

	if contains(config.BlockedExts, ext) {
		file.Close()
		return nil, errors.NewValidationError("file extension is blocked for security reasons", ext)
	}

	// Validate MIME type
	contentType := header.Header.Get("Content-Type")
	if contentType != "" {
		if !contains(config.AllowedTypes, contentType) {
			file.Close()
			return nil, errors.NewValidationError("content type is not allowed", contentType)
		}

		if contains(config.BlockedTypes, contentType) {
			file.Close()
			return nil, errors.NewValidationError("content type is blocked for security reasons", contentType)
		}
	}

	// Additional security check: read first few bytes to detect file type
	if err := validateFileContent(file, header); err != nil {
		file.Close()
		return nil, err
	}

	return header, nil
}

// validateFileContent performs additional content-based validation
func validateFileContent(file multipart.File, _ *multipart.FileHeader) error {
	// Reset file position
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return errors.NewSystemError("failed to seek file", err)
		}
	}

	// Read first 512 bytes for magic number detection
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return errors.NewSystemError("failed to read file header", err)
	}

	// Check for executable signatures (basic check)
	if isExecutableContent(buffer[:n]) {
		return errors.NewValidationError("file appears to be an executable, which is not allowed", nil)
	}

	// Reset file position for actual processing
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return errors.NewSystemError("failed to reset file position", err)
		}
	}

	return nil
}

// isExecutableContent performs basic detection of executable content
func isExecutableContent(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	// Check for common executable signatures
	signatures := [][]byte{
		{0x4D, 0x5A},             // PE (Windows executable)
		{0x7F, 0x45, 0x4C, 0x46}, // ELF (Linux executable)
		{0xFE, 0xED, 0xFA},       // Mach-O (macOS executable)
		{0xCA, 0xFE, 0xBA, 0xBE}, // Java class file
		{0x23, 0x21},             // Shebang (script files)
	}

	for _, sig := range signatures {
		if len(data) >= len(sig) && compareBytes(data[:len(sig)], sig) {
			return true
		}
	}

	return false
}

// compareBytes compares two byte slices
func compareBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ValidateStruct validates a struct using the validator library
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// SanitizeString sanitizes a string input
func SanitizeString(input string) string {
	// Remove null bytes and control characters
	input = strings.Map(func(r rune) rune {
		if r < 32 && r != 9 && r != 10 && r != 13 { // Keep tab, LF, CR
			return -1
		}
		return r
	}, input)

	// Trim whitespace
	return strings.TrimSpace(input)
}

// ValidateProfileName validates and sanitizes profile names
func ValidateProfileName(name string) (string, error) {
	if name == "" {
		return "", errors.NewValidationError("profile name cannot be empty", nil)
	}

	if len(name) > 100 {
		return "", errors.NewValidationError("profile name too long (max 100 characters)", nil)
	}

	// Check for dangerous characters
	dangerous := []string{"<", ">", "&", "\"", "'", "/", "\\", ".."}
	for _, char := range dangerous {
		if strings.Contains(name, char) {
			return "", errors.NewValidationError("profile name contains invalid character", char)
		}
	}

	return SanitizeString(name), nil
}

// ValidatePath prevents directory traversal attacks by ensuring paths are safe
// Returns the validated path or an error if the path contains traversal attempts
func ValidatePath(path string) (string, error) {
	if path == "" {
		return "", errors.NewValidationError("path cannot be empty", nil)
	}

	// Clean the path to resolve any .. or . components
	cleanPath := filepath.Clean(path)

	// Check for directory traversal attempts
	if strings.Contains(cleanPath, "..") {
		return "", errors.NewValidationError("path contains directory traversal attempt", path)
	}

	// Ensure path doesn't start with dangerous patterns
	if strings.HasPrefix(cleanPath, "/") || strings.HasPrefix(cleanPath, "\\") {
		return "", errors.NewValidationError("absolute paths not allowed", path)
	}

	// Additional check: ensure the path doesn't contain null bytes or other dangerous characters
	if strings.Contains(cleanPath, "\x00") {
		return "", errors.NewValidationError("path contains null bytes", path)
	}

	return cleanPath, nil
}

// ValidateFilePath validates a file path for security and returns sanitized path
func ValidateFilePath(path string, baseDir string) (string, error) {
	// First validate the path is safe
	safePath, err := ValidatePath(path)
	if err != nil {
		return "", err
	}

	// If baseDir is provided, ensure the path is within the base directory
	if baseDir != "" {
		baseDir = filepath.Clean(baseDir)
		// Join with base directory and clean again
		fullPath := filepath.Join(baseDir, safePath)
		fullPath = filepath.Clean(fullPath)

		// Ensure the resulting path is still within the base directory
		relPath, err := filepath.Rel(baseDir, fullPath)
		if err != nil || strings.HasPrefix(relPath, "..") {
			return "", errors.NewValidationError("path escapes base directory", path)
		}
		return fullPath, nil
	}

	return safePath, nil
}
