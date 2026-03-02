// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package logger

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type LogLevel int

type Logger interface {
	Info(format string, args ...any)
	Warn(format string, args ...any)
	err(format string, args ...any)
	Debug(format string, args ...any)
}

const (
	ErrorLevel LogLevel = iota
	WarnLevel
	InfoLevel
	DebugLevel
)

func (lv LogLevel) String() string {
	switch lv {
	case ErrorLevel:
		return "ERROR"
	case WarnLevel:
		return "WARN"
	case InfoLevel:
		return "INFO"
	case DebugLevel:
		return "DEBUG"
	default:
		return "INFO"
	}
}

// Severity returns the lowercase representation used by the shared log schema.
func (lv LogLevel) Severity() string {
	switch lv {
	case ErrorLevel:
		return "error"
	case WarnLevel:
		return "warn"
	case InfoLevel:
		return "info"
	case DebugLevel:
		return "debug"
	default:
		return "info"
	}
}

// StructuredLogEntry is the canonical payload emitted to the frontend log store.
type StructuredLogEntry struct {
	ID        string
	Timestamp string
	Severity  string
	Message   string
	Context   string
	Details   any
	Source    string
	Metadata  map[string]any
}

// Emit forwards a structured log payload to the frontend using the shared schema.
func Emit(ctx context.Context, entry StructuredLogEntry) {
	if ctx == nil {
		return
	}

	timestamp := entry.Timestamp
	if timestamp == "" {
		timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}

	severity := entry.Severity
	if severity == "" {
		severity = InfoLevel.Severity()
	}

	id := entry.ID
	if id == "" {
		id = generateLogID()
	}

	payload := map[string]any{
		"id":        id,
		"timestamp": timestamp,
		"severity":  severity,
		"message":   entry.Message,
	}

	if entry.Context != "" {
		payload["context"] = entry.Context
	}
	if entry.Source != "" {
		payload["source"] = entry.Source
	}
	if entry.Details != nil {
		payload["details"] = sanitizeValue(entry.Details)
	}
	if metadata := sanitizeMetadata(entry.Metadata); len(metadata) > 0 {
		payload["metadata"] = metadata
	}

	wailsRuntime.EventsEmit(ctx, "log:entry", payload)
}

// LogToFrontend is retained for backward compatibility; prefer Emit for new code.
func LogToFrontend(ctx context.Context, lv LogLevel, format string, args ...any) {
	Emit(ctx, StructuredLogEntry{
		Severity: lv.Severity(),
		Message:  fmt.Sprintf(format, args...),
		Source:   "backend",
	})
}

func generateLogID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err == nil {
		return fmt.Sprintf("log-%x", buf)
	}
	return fmt.Sprintf("log-%d", time.Now().UTC().UnixNano())
}

func sanitizeMetadata(metadata map[string]any) map[string]any {
	if len(metadata) == 0 {
		return map[string]any{}
	}

	sanitized := make(map[string]any, len(metadata))
	for k, v := range metadata {
		if k == "" {
			continue
		}
		sanitized[k] = sanitizeValue(v)
	}
	return sanitized
}

func sanitizeValue(value any) any {
	switch v := value.(type) {
	case nil,
		string,
		bool,
		int,
		int8,
		int16,
		int32,
		int64,
		uint,
		uint8,
		uint16,
		uint32,
		uint64,
		float32,
		float64,
		map[string]any,
		[]any:
		return v
	case time.Time:
		return v.UTC().Format(time.RFC3339Nano)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// KeyValuesToMap converts a variadic key/value sequence into sanitized metadata.
func KeyValuesToMap(values ...any) map[string]any {
	length := len(values)
	if length < 2 {
		return map[string]any{}
	}

	// Ensure we only process complete key/value pairs.
	if length%2 != 0 {
		length--
	}

	metadata := make(map[string]any, length/2)
	for i := 0; i < length; i += 2 {
		key, ok := values[i].(string)
		if !ok || key == "" {
			continue
		}
		metadata[key] = sanitizeValue(values[i+1])
	}
	return metadata
}

// KeyValuesToAttrs mirrors slog's variadic logging helpers so existing log output is preserved.
func KeyValuesToAttrs(values ...any) []slog.Attr {
	length := len(values)
	if length < 2 {
		return nil
	}
	if length%2 != 0 {
		length--
	}

	attrs := make([]slog.Attr, 0, length/2)
	for i := 0; i < length; i += 2 {
		key, ok := values[i].(string)
		if !ok || key == "" {
			continue
		}
		attrs = append(attrs, slog.Any(key, values[i+1]))
	}
	return attrs
}

// ToSlogLevel maps our LogLevel enumeration to slog.Level.
func ToSlogLevel(level LogLevel) slog.Level {
	switch level {
	case DebugLevel:
		return slog.LevelDebug
	case WarnLevel:
		return slog.LevelWarn
	case ErrorLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// New creates a new structured logger for the application
func New(err string) *slog.Logger {
	logLevelStr := err
	var logLevel slog.Level
	switch strings.ToLower(logLevelStr) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	case "fatal":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	logOpts := &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   a.Key,
					Value: slog.StringValue(time.Now().UTC().Format(time.RFC3339Nano)),
				}
			}

			return a
		},
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, logOpts))
	logger = logger.With(
		"service", "cloud-file-processor",
		"version", getVersion(),
	)

	return logger
}

// getVersion returns the current application version
func getVersion() string {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		if info, ok := debug.ReadBuildInfo(); ok {
			version = info.Main.Version
		}
		if version == "" {
			version = "dev"
		}
	}

	return version
}

func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value("logger").(*slog.Logger); ok {
		return logger
	}
	return nil
}
