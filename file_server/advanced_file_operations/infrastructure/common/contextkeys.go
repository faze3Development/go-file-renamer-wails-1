// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package common

// Use a private type for context keys to avoid collisions.
type contextKey string

const (
	// LoggerKey is the key for the slog.Logger in the context.
	LoggerKey = contextKey("logger")
)
