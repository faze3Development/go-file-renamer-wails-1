// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package errors

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/mrz1836/go-sanitize"
)

// ErrorResponse represents the JSON structure sent to clients
type ErrorResponse struct {
	Success   bool      `json:"success"`
	Error     *AppError `json:"error"`
	RequestID string    `json:"request_id,omitempty"`
}

// HTTPErrorHandler handles HTTP error responses with structured JSON
func HTTPErrorHandler(w http.ResponseWriter, err error, logger *slog.Logger, requestID string) {
	// Convert to AppError if it's not already
	var appErr *AppError
	if e, ok := err.(*AppError); ok {
		appErr = e
	} else {
		// Wrap unknown errors as system errors
		appErr = NewSystemError("An unexpected error occurred", err)
	}

	// Log the full error details (including stack trace for system errors)
	logError(appErr, logger, requestID)

	// Create safe response for client
	response := ErrorResponse{
		Success:   false,
		Error:     createSafeErrorResponse(appErr),
		RequestID: requestID,
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(appErr.HTTPStatus)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(true) // Prevent XSS in error messages
	if encodeErr := encoder.Encode(response); encodeErr != nil {
		// If JSON encoding fails, send plain text fallback
		logger.Error("Failed to encode error response", "error", encodeErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// createSafeErrorResponse creates a client-safe version of the error
func createSafeErrorResponse(err *AppError) *AppError {
	if err.Safe {
		// Return the error as-is for safe errors
		return &AppError{
			Type:    err.Type,
			Code:    err.Code,
			Message: err.Message,
			Details: err.Details,
		}
	}

	// For unsafe errors, return a generic message
	return &AppError{
		Type:    err.Type,
		Code:    err.Code,
		Message: "An error occurred while processing your request",
	}
}

// logError logs the error with appropriate detail level
func logError(err *AppError, logger *slog.Logger, requestID string) {
	logAttrs := []slog.Attr{
		slog.String("error_type", string(err.Type)),
		slog.String("error_code", err.Code),
		slog.String("request_id", requestID),
		slog.Int("http_status", err.HTTPStatus),
	}

	// Add cause if present
	if err.Cause != nil {
		logAttrs = append(logAttrs, slog.String("cause", err.Cause.Error()))
	}

	// Add stack trace for system errors
	if err.Type == ErrorTypeSystem {
		logAttrs = append(logAttrs, slog.String("stack_trace", string(debug.Stack())))
	}

	// Log based on error type
	args := make([]any, len(logAttrs))
	for i, attr := range logAttrs {
		args[i] = attr
	}

	switch err.Type {
	case ErrorTypeSystem:
		logger.Error(err.Message, args...)
	case ErrorTypeSecurity:
		logger.Warn(err.Message, args...)
	default:
		logger.Info(err.Message, args...)
	}
}

// PanicRecovery middleware recovers from panics and returns structured errors
func PanicRecovery(logger *slog.Logger, requestID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Panic recovered", "panic", err, "stack", string(debug.Stack()), "request_id", requestID)
					appErr := NewSystemError("Internal server error occurred", nil)
					HTTPErrorHandler(w, appErr, logger, requestID)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// SafeJSONResponse safely encodes JSON responses with HTML escaping and security headers
func SafeJSONResponse(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// Use json.HTMLEscape to prevent XSS in JSON responses
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(true) // Escape HTML characters in JSON strings

	return encoder.Encode(data)
}

// SanitizeJSONStrings recursively sanitizes string values in JSON structures
func SanitizeJSONStrings(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		sanitized := make(map[string]interface{})
		for key, value := range v {
			sanitized[key] = SanitizeJSONStrings(value)
		}
		return sanitized
	case []interface{}:
		sanitized := make([]interface{}, len(v))
		for i, item := range v {
			sanitized[i] = SanitizeJSONStrings(item)
		}
		return sanitized
	case string:
		return SanitizeInput(v)
	default:
		return v
	}
}

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

// WrapHandler wraps an HTTP handler with error handling
func WrapHandler(handler func(w http.ResponseWriter, r *http.Request) error, logger *slog.Logger, requestID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			HTTPErrorHandler(w, err, logger, requestID)
		}
	}
}
