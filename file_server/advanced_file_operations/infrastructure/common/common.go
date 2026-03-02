// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package common

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// File size limits
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB

	// Common file size limits (int64 for consistent usage throughout codebase)
	MaxFileSize1MB   = int64(1 * MB)
	MaxFileSize2MB   = int64(2 * MB)
	MaxFileSize10MB  = int64(10 * MB)
	MaxFileSize50MB  = int64(50 * MB)
	MaxFileSize100MB = int64(100 * MB)

	// Payload size limits (int64 for consistent usage throughout codebase)
	MaxPayload10MB = int64(10 * MB)
	MaxPayload32MB = int64(32 * MB)
	MaxPayload50MB = int64(50 * MB)
)

// GetMaxFileSizeFromEnv reads MAX_FILE_SIZE from environment and parses it
// Format: "10MB", "50MB", "100MB"
// Returns the size in bytes as int64, or defaultSize if not set/invalid
func GetMaxFileSizeFromEnv(defaultSize int64) int64 {
	sizeStr := os.Getenv("MAX_FILE_SIZE")
	if sizeStr == "" {
		return defaultSize
	}

	// Parse size string (e.g., "10MB", "50MB")
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	var multiplier int64 = 1
	if strings.HasSuffix(sizeStr, "GB") {
		multiplier = GB
		sizeStr = strings.TrimSuffix(sizeStr, "GB")
	} else if strings.HasSuffix(sizeStr, "MB") {
		multiplier = MB
		sizeStr = strings.TrimSuffix(sizeStr, "MB")
	} else if strings.HasSuffix(sizeStr, "KB") {
		multiplier = KB
		sizeStr = strings.TrimSuffix(sizeStr, "KB")
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil || size <= 0 {
		return defaultSize
	}

	return size * multiplier
}

// Configuration parsing helper functions
func GetBoolFromConfig(config map[string]any, key string, defaultValue bool) bool {
	if val, ok := config[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func GetStringFromConfig(config map[string]any, key string, defaultValue string) string {
	if val, ok := config[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return defaultValue
}

func GetInt64FromConfig(config map[string]any, key string, defaultValue int64) int64 {
	if val, ok := config[key]; ok {
		if i, ok := val.(int64); ok {
			return i
		}
		if i, ok := val.(int); ok {
			return int64(i)
		}
		if f, ok := val.(float64); ok {
			return int64(f)
		}
	}
	return defaultValue
}

// ParseMaxFileSizeFromString parses a maxFileSize string value (e.g., "52428800") into int64
func ParseMaxFileSizeFromString(value string, defaultValue int64) int64 {
	if value == "" {
		return defaultValue
	}
	if parsed, err := strconv.ParseInt(value, 10, 64); err == nil && parsed > 0 {
		return parsed
	}
	return defaultValue
}

// File represents a file to be processed (contains actual file data)
type File struct {
	Filename    string `json:"filename"`
	Content     []byte `json:"content"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
}

// FileInfo represents file information for storage operations
type FileInfo struct {
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	ContentType string    `json:"contentType"`
	UploadedAt  time.Time `json:"uploadedAt"`
}

// FileOperation represents a file to be processed
type FileOperation struct {
	Source string `json:"source"`
}

// FileOperationResult represents the result of a file operation
type FileOperationResult struct {
	Original FileOperation `json:"original"`
	New      FileOperation `json:"new"`
	Status   string        `json:"status"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `json:"page" query:"page"`
	PageSize int `json:"pageSize" query:"pageSize"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Total       int  `json:"total"`
	Page        int  `json:"page"`
	PageSize    int  `json:"pageSize"`
	HasNext     bool `json:"hasNext"`
	HasPrevious bool `json:"hasPrevious"`
}

// Common operation types
const (
	OperationRename         = "rename"
	OperationMetadataRemove = "metadata_removal"
)

// Common operation statuses
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
)
