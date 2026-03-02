// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package renamers

import (
	"context"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"
)

func preserveOriginalFromConfig(config map[string]interface{}) bool {
	if preserve, ok := config["preserveOriginalName"].(bool); ok {
		return preserve
	}

	if opts, ok := config["options"].(map[string]interface{}); ok {
		if preserve, ok := opts["preserveOriginalName"].(bool); ok {
			return preserve
		}
	}

	return false
}

// Processor implements the FileProcessor interface for file renaming
type Processor struct{}

// NewProcessor creates a new renamer processor
func NewProcessor() *Processor {
	return &Processor{}
}

// Type returns the processor type identifier
func (p *Processor) Type() string {
	return "rename"
}

// ProcessFile processes a single file by renaming it according to the configuration
func (p *Processor) ProcessFile(ctx context.Context, file common.FileOperation, config map[string]interface{}) (*common.FileOperationResult, error) {
	pattern, ok := config["pattern"].(string)
	if !ok {
		pattern = "default"
	}

	preserveOriginal := preserveOriginalFromConfig(config)

	newName, err := GenerateName(file.Source, pattern, Options{PreserveOriginalName: preserveOriginal})
	if err != nil {
		return nil, err
	}

	result := &common.FileOperationResult{
		Original: file,
		New: common.FileOperation{
			Source: newName,
		},
		Status: "completed",
	}

	return result, nil
}

// ValidateConfig validates the renamer configuration
func (p *Processor) ValidateConfig(config map[string]interface{}) error {
	if pattern, ok := config["pattern"].(string); ok {
		// Basic validation - could be extended
		if pattern == "" {
			return errors.NewValidationError("pattern cannot be empty", nil)
		}
	}
	return nil
}
