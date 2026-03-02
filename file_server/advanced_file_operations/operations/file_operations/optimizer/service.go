// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package optimizer

import (
	"context"
	"log/slog"
)

// Service orchestrates optimization operations around the processor.
type Service struct {
	processor *Processor
	logger    *slog.Logger
}

// NewService creates a new optimizer service instance.
func NewService(logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}

	return &Service{
		processor: NewProcessor(),
		logger:    logger,
	}
}

// ProcessFileWithContent runs optimization on the provided file payload.
func (s *Service) ProcessFileWithContent(ctx context.Context, filename string, content []byte, contentType string) ([]byte, error) {
	return s.processor.ProcessFileWithContent(ctx, filename, content, contentType)
}
