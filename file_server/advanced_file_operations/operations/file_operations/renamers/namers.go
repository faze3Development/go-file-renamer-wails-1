// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package renamers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"
)

const maxSequentialBaseLength = 120

var (
	unsafeSequentialChars = regexp.MustCompile(`[^a-zA-Z0-9-_ ]+`)
	sequentialWhitespace  = regexp.MustCompile(`\s+`)
)

// Options controls runtime behaviour of renamers derived from pattern strings.
type Options struct {
	PreserveOriginalName bool
}

func sanitizeSequentialBase(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	sanitized := unsafeSequentialChars.ReplaceAllString(trimmed, " ")
	sanitized = sequentialWhitespace.ReplaceAllString(sanitized, " ")
	sanitized = strings.TrimSpace(sanitized)

	if len(sanitized) > maxSequentialBaseLength {
		sanitized = sanitized[:maxSequentialBaseLength]
	}

	// Convert whitespace to hyphen for filesystem safety and consistency with frontend previews.
	sanitized = strings.ReplaceAll(sanitized, " ", "-")

	return sanitized
}

// SequentialNamer generates sequential filenames with optional padding.
type SequentialNamer struct {
	base             string
	next             int
	padLength        int
	keepExtension    bool
	preserveOriginal bool
	mu               sync.Mutex
}

// ShouldKeepExtension indicates whether the original extension should be retained.
func (n *SequentialNamer) ShouldKeepExtension() bool {
	return n.keepExtension
}

// GenerateName returns the next sequential filename.
func (n *SequentialNamer) GenerateName(originalName, _ string) (string, error) {
	n.mu.Lock()
	current := n.next
	n.next++
	n.mu.Unlock()

	base := n.base
	if base == "" && n.preserveOriginal {
		base = sanitizeSequentialBase(originalName)
	}
	if base == "" {
		base = "file"
	}

	number := strconv.Itoa(current)
	if n.padLength > 0 {
		number = fmt.Sprintf("%0*d", n.padLength, current)
	}

	return fmt.Sprintf("%s_%s", base, number), nil
}

// Namer defines the interface for a file namer.
type Namer interface {
	GenerateName(originalName, ext string) (string, error)
}

// TimestampNamer renames a file using a timestamp.
type TimestampNamer struct{}

func (n *TimestampNamer) GenerateName(originalName, ext string) (string, error) {
	return fmt.Sprintf("%s_%d%s", originalName, time.Now().Unix(), ext), nil
}

// RandomNamer renames a file using a random string.
type RandomNamer struct{}

func (n *RandomNamer) GenerateName(originalName, ext string) (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", errors.NewSystemError("failed to generate random string", err)
	}
	randomID := hex.EncodeToString(bytes)
	return fmt.Sprintf("%s_%s%s", originalName, randomID, ext), nil
}

// DateNamer renames a file using a date.
type DateNamer struct {
	Layout string
	Date   string
}

func (n *DateNamer) GenerateName(originalName, ext string) (string, error) {
	if n.Date != "" {
		return fmt.Sprintf("%s_%s%s", originalName, n.Date, ext), nil
	}
	dateStr := time.Now().Format(n.Layout)
	return fmt.Sprintf("%s_%s%s", originalName, dateStr, ext), nil
}

// RegexNamer renames a file using a regex.
type RegexNamer struct {
	Find    *regexp.Regexp
	Replace string
}

func (n *RegexNamer) GenerateName(originalName, ext string) (string, error) {
	result := n.Find.ReplaceAllString(originalName, n.Replace)
	return result + ext, nil
}

// CombinedNamer combines multiple namers.
type CombinedNamer struct {
	Namers []Namer
}

func (n *CombinedNamer) GenerateName(originalName, ext string) (string, error) {
	result := originalName
	keepExtension := true
	for _, namer := range n.Namers {
		if controller, ok := namer.(interface{ ShouldKeepExtension() bool }); ok {
			if !controller.ShouldKeepExtension() {
				keepExtension = false
			}
		}

		var err error
		result, err = namer.GenerateName(result, "")
		if err != nil {
			return "", err
		}
	}
	if keepExtension {
		return result + ext, nil
	}
	return result, nil
}

// NewNamerFromString creates a namer from a pattern string.
func NewNamerFromString(pattern string, opts Options) (Namer, error) {
	if strings.Contains(pattern, "|") {
		parts := strings.Split(pattern, "|")
		var namers []Namer
		for _, part := range parts {
			namer, err := NewNamerFromString(part, opts)
			if err != nil {
				return nil, err
			}
			if namer != nil {
				namers = append(namers, namer)
			}
		}
		if len(namers) == 0 {
			return nil, nil
		}
		return &CombinedNamer{Namers: namers}, nil
	}

	if strings.HasPrefix(pattern, "date:") {
		parts := strings.Split(strings.TrimPrefix(pattern, "date:"), ":")
		layout := "2006-01-02"
		var date string
		if len(parts) > 0 && parts[0] != "" {
			layout = parts[0]
		}
		if len(parts) > 1 && parts[1] != "" {
			date = parts[1]
		}
		return &DateNamer{Layout: layout, Date: date}, nil
	}

	if strings.HasPrefix(pattern, "seq:") {
		parts := strings.Split(strings.TrimPrefix(pattern, "seq:"), ":")
		base := ""
		if len(parts) > 0 {
			base = sanitizeSequentialBase(parts[0])
		}

		start := 1
		if len(parts) > 1 {
			if parsed, err := strconv.Atoi(parts[1]); err == nil {
				start = parsed
			}
		}
		if start < 0 {
			start = 0
		}

		pad := 2
		if len(parts) > 2 {
			if parsed, err := strconv.Atoi(parts[2]); err == nil {
				pad = parsed
			}
		}
		if pad < 0 {
			pad = 0
		}
		if pad > 6 {
			pad = 6
		}

		keepExtension := true
		if len(parts) > 3 {
			keepExtension = parts[3] != "0"
		}

		return &SequentialNamer{
			base:             base,
			next:             start,
			padLength:        pad,
			keepExtension:    keepExtension,
			preserveOriginal: opts.PreserveOriginalName,
		}, nil
	}

	if strings.HasPrefix(pattern, "regex:") {
		parts := strings.SplitN(strings.TrimPrefix(pattern, "regex:"), ":", 2)
		if len(parts) == 2 {
			re, err := regexp.Compile(parts[0])
			if err != nil {
				return nil, errors.NewValidationError("invalid regex pattern", err)
			}
			return &RegexNamer{Find: re, Replace: parts[1]}, nil
		}
		return nil, errors.NewValidationError("invalid regex pattern format", nil)
	}

	switch pattern {
	case "timestamp":
		return &TimestampNamer{}, nil
	case "random":
		return &RandomNamer{}, nil
	case "default":
		return nil, nil // No-op
	default:
		return nil, errors.NewValidationError("unknown pattern", pattern)
	}
}

func NewPatternNamer(pattern string, opts Options) (Namer, error) {
	if pattern == "" || pattern == "default" {
		return nil, nil
	}
	return NewNamerFromString(pattern, opts)
}

func GenerateName(original, pattern string, opts Options) (string, error) {
	ext := filepath.Ext(original)
	name := strings.TrimSuffix(original, ext)

	namer, err := NewNamerFromString(pattern, opts)
	if err != nil {
		return "", err
	}

	if namer == nil { // default case
		return original, nil
	}

	return namer.GenerateName(name, ext)
}

func FormatWithNamer(namer Namer, original string) (string, error) {
	if namer == nil {
		return original, nil
	}

	ext := filepath.Ext(original)
	name := strings.TrimSuffix(original, ext)
	return namer.GenerateName(name, ext)
}
