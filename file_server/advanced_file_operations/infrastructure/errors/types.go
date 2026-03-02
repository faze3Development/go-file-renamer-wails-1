// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package errors

import (
	"fmt"
	"net/http"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeValidation      ErrorType = "validation"
	ErrorTypeProcessing      ErrorType = "processing"
	ErrorTypeSecurity        ErrorType = "security"
	ErrorTypeSystem          ErrorType = "system"
	ErrorTypeNotFound        ErrorType = "not_found"
	ErrorTypeConflict        ErrorType = "conflict"
	ErrorTypePayloadTooLarge ErrorType = "payload_too_large"
	ErrorTypeRateLimit       ErrorType = "rate_limit"
	ErrorTypeUnknown         ErrorType = "unknown"
)

// AppError represents a structured application error
type AppError struct {
	Type       ErrorType   `json:"type"`
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	HTTPStatus int         `json:"-"`
	Cause      error       `json:"-"`
	Safe       bool        `json:"-"` // Whether it's safe to expose details to client
	Unknown    string      `json:"unknown,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying cause
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Common error constructors

// NewValidationError creates a validation error
func NewValidationError(message string, details interface{}) *AppError {
	return &AppError{
		Type:       ErrorTypeValidation,
		Code:       "VALIDATION_ERROR",
		Message:    message,
		Details:    details,
		HTTPStatus: http.StatusBadRequest,
		Safe:       true,
	}
}

// NewProcessingError creates a processing error
func NewProcessingError(message string, cause error) *AppError {
	return &AppError{
		Type:       ErrorTypeProcessing,
		Code:       "PROCESSING_ERROR",
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
		Cause:      cause,
		Safe:       false,
	}
}

// NewSecurityError creates a security error
func NewSecurityError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeSecurity,
		Code:       "SECURITY_VIOLATION",
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
		Safe:       false, // Don't expose security details
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Type:       ErrorTypeNotFound,
		Code:       "RESOURCE_NOT_FOUND",
		Message:    fmt.Sprintf("%s not found", resource),
		HTTPStatus: http.StatusNotFound,
		Safe:       true,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeConflict,
		Code:       "RESOURCE_CONFLICT",
		Message:    message,
		HTTPStatus: http.StatusConflict,
		Safe:       true,
	}
}

// NewPayloadTooLargeError creates a payload too large error
func NewPayloadTooLargeError(maxSize, actualSize int64) *AppError {
	return &AppError{
		Type:       ErrorTypePayloadTooLarge,
		Code:       "PAYLOAD_TOO_LARGE",
		Message:    fmt.Sprintf("Payload size %d bytes exceeds maximum allowed size %d bytes", actualSize, maxSize),
		HTTPStatus: http.StatusRequestEntityTooLarge,
		Safe:       true,
	}
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeRateLimit,
		Code:       "RATE_LIMIT_EXCEEDED",
		Message:    message,
		HTTPStatus: http.StatusTooManyRequests,
		Safe:       true,
	}
}

// NewSystemError creates a system/internal error
func NewSystemError(message string, cause error) *AppError {
	return &AppError{
		Type:       ErrorTypeSystem,
		Code:       "SYSTEM_ERROR",
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
		Cause:      cause,
		Safe:       false,
	}
}

func NewUnknownError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeUnknown,
		Code:       "Unknown",
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
		Cause:      NewSystemError(message, nil),
		Safe:       false,
	}
}
