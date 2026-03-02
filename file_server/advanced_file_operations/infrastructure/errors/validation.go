// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package errors

import (
	"fmt"
	"net/http"
	"strings"
)

// FileValidationError creates a validation error for file-related issues
func NewFileValidationError(filename string, reason string) *AppError {
	return NewValidationError(
		fmt.Sprintf("File validation failed for '%s': %s", filename, reason),
		map[string]any{
			"filename": filename,
			"reason":   reason,
		},
	)
}

// FileSizeError creates an error for file size violations
func NewFileSizeError(filename string, actualSize, maxSize int64) *AppError {
	return NewValidationError(
		fmt.Sprintf("File '%s' size %d bytes exceeds maximum allowed size %d bytes", filename, actualSize, maxSize),
		map[string]any{
			"filename":    filename,
			"actual_size": actualSize,
			"max_size":    maxSize,
		},
	)
}

// FileTypeError creates an error for invalid file types
func NewFileTypeError(filename, contentType string, allowedTypes []string) *AppError {
	return NewValidationError(
		fmt.Sprintf("File '%s' has invalid type '%s'. Allowed types: %s", filename, contentType, strings.Join(allowedTypes, ", ")),
		map[string]any{
			"filename":      filename,
			"content_type":  contentType,
			"allowed_types": allowedTypes,
		},
	)
}

// FileCountError creates an error for file count violations
func NewFileCountError(actualCount, maxCount int) *AppError {
	return NewValidationError(
		fmt.Sprintf("Too many files: %d files provided, maximum %d allowed", actualCount, maxCount),
		map[string]any{
			"actual_count": actualCount,
			"max_count":    maxCount,
		},
	)
}

// PayloadSizeError creates an error for payload size violations
func NewPayloadSizeError(actualSize, maxSize int64) *AppError {
	return NewPayloadTooLargeError(maxSize, actualSize)
}

// InjectionDetectedError creates an error for detected injection attempts
func NewInjectionDetectedError(filename string) *AppError {
	return NewSecurityError(fmt.Sprintf("Potential security threat detected in file '%s'", filename))
}

// ProcessingError wraps processing errors with context
func NewFileProcessingError(filename string, operation string, cause error) *AppError {
	return &AppError{
		Type:       ErrorTypeProcessing,
		Code:       "FILE_PROCESSING_ERROR",
		Message:    fmt.Sprintf("Failed to %s file", operation),
		Details:    map[string]any{"filename": filename},
		HTTPStatus: http.StatusInternalServerError,
		Cause:      cause,
		Safe:       false, // Don't expose internal processing errors or filenames to client
	}
}

// BulkOperationError creates an error for bulk operation failures
func NewBulkOperationError(operation string, failedFiles []string, cause error) *AppError {
	return &AppError{
		Type:    ErrorTypeProcessing,
		Code:    "BULK_OPERATION_FAILED",
		Message: fmt.Sprintf("Bulk %s operation failed for %d files", operation, len(failedFiles)),
		Details: map[string]any{
			"operation":     operation,
			"failed_files":  failedFiles,
			"failure_count": len(failedFiles),
		},
		HTTPStatus: http.StatusInternalServerError,
		Cause:      cause,
		Safe:       true,
	}
}
