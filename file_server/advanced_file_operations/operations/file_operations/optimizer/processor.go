// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package optimizer

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"
	"sync"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"

	"github.com/disintegration/imaging"
)

// Processor implements the FileProcessor interface for file optimization
type Processor struct{}

// NewProcessor creates a new optimizer processor
func NewProcessor() *Processor {
	return &Processor{}
}

// Type returns the processor type identifier
func (p *Processor) Type() string {
	return "optimizer"
}

// ProcessFile processes a single file by optimizing it
func (p *Processor) ProcessFile(ctx context.Context, file common.FileOperation, config map[string]interface{}) (*common.FileOperationResult, error) {
	// Since we don't have file content in this interface, we simulate the operation
	// and return information about what optimization would accomplish.
	// The actual processing should use ProcessFileWithContent method.

	result := &common.FileOperationResult{
		Original: file,
		New:      file, // Same filename, but optimized
		Status:   common.StatusCompleted,
	}

	// To-Do: actually process the file
	// For now, we return success to indicate the operation is supported

	return result, nil
}

// ValidateConfig validates the optimizer configuration
func (p *Processor) ValidateConfig(config map[string]any) error {
	// Optimizer doesn't need special configuration currently
	return nil
}

// GetSupportedTypes returns the file types supported for optimization
func (p *Processor) GetSupportedTypes() []string {
	return []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"application/pdf",
		"text/plain",
		"text/markdown",
		"application/json",
		"application/octet-stream", // Generic fallback
	}
}

// ProcessFileWithContent processes a file with content for optimization
func (p *Processor) ProcessFileWithContent(ctx context.Context, filename string, content []byte, contentType string) ([]byte, error) {
	if len(content) == 0 {
		return nil, errors.NewValidationError("cannot optimize empty file", nil)
	}

	// Check for reasonable file size limits (100MB for optimization)
	maxFileSize := common.MaxFileSize100MB
	if int64(len(content)) > maxFileSize {
		return nil, errors.NewPayloadTooLargeError(maxFileSize, int64(len(content)))
	}

	reader := bytes.NewReader(content)

	// Determine file type and apply appropriate optimization
	if contentType == "application/pdf" || (len(content) > 4 && string(content[:4]) == "%%PDF") {
		// For PDF files, apply optimization
		optimizedContent, err := optimizePDFFile(reader, filename)
		if err != nil {
			return nil, errors.NewFileProcessingError(filename, "optimize PDF", err)
		}
		return optimizedContent, nil
	} else if strings.HasPrefix(contentType, "image/") {
		// For image files, apply image-specific optimizations
		optimizedContent, err := optimizeImageFile(reader, filename, contentType)
		if err != nil {
			return nil, errors.NewFileProcessingError(filename, "optimize image", err)
		}
		return optimizedContent, nil
	} else if strings.HasPrefix(contentType, "text/") || contentType == "application/json" {
		// For text-based files, apply text optimizations
		optimizedContent, err := optimizeTextFile(reader, filename, contentType)
		if err != nil {
			return nil, errors.NewFileProcessingError(filename, "optimize text", err)
		}
		return optimizedContent, nil
	} else {
		// For other files, validate but return as-is with size check
		if len(content) < 1024 {
			return nil, errors.NewValidationError(fmt.Sprintf("file '%s' too small for meaningful optimization: %d bytes", filename, len(content)), nil)
		}
		return content, nil // Return original content for unsupported file types
	}
}

// ProcessFilesWithContent processes multiple files concurrently.
type FileProcessingResult struct {
	Filename string
	Content  []byte
	Err      error
}

func (p *Processor) ProcessFilesWithContent(ctx context.Context, files []common.File) <-chan FileProcessingResult {
	results := make(chan FileProcessingResult, len(files))
	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func(file common.File) {
			defer wg.Done()
			optimizedContent, err := p.ProcessFileWithContent(ctx, file.Filename, file.Content, file.ContentType)
			results <- FileProcessingResult{
				Filename: file.Filename,
				Content:  optimizedContent,
				Err:      err,
			}
		}(file)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

// optimizePDFFile optimizes PDF files
// Currently implements basic PDF validation and returns original content.
// For production-grade optimization, consider integrating:
// - github.com/pdfcpu/pdfcpu for comprehensive PDF optimization and compression
// - External tools: Ghostscript for advanced PDF processing, QPDF for restructuring
func optimizePDFFile(reader *bytes.Reader, filename string) ([]byte, error) {
	// Reset reader to beginning
	reader.Seek(0, 0)

	// Read all content
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.NewSystemError(fmt.Sprintf("failed to read PDF content for %s", filename), err)
	}

	// Basic PDF validation - check PDF header
	if len(content) < 4 || string(content[:4]) != "%%PDF" {
		return nil, errors.NewValidationError(fmt.Sprintf("invalid PDF file %s: missing PDF header", filename), nil)
	}

	// Basic size validation for PDFs
	if len(content) < 1024 {
		return nil, errors.NewValidationError(fmt.Sprintf("PDF file %s too small: %d bytes", filename, len(content)), nil)
	}

	// Basic PDF optimization: Remove trailing whitespace and normalize line endings
	// This provides minimal size reduction without external dependencies
	optimized := bytes.TrimSpace(content)
	optimized = bytes.ReplaceAll(optimized, []byte("\r\n"), []byte("\n"))

	// Note: For significant size reduction, use pdfcpu:
	// import "github.com/pdfcpu/pdfcpu/pkg/api"
	// err := api.OptimizeFile(inputPath, outputPath, nil)

	return optimized, nil
}

// optimizeImageFile optimizes image files using the imaging library
// Implements advanced optimization: smart resizing for large images, quality control, and format-specific encoding
// Uses github.com/disintegration/imaging for high-quality resize with Lanczos filter
func optimizeImageFile(reader *bytes.Reader, filename string, contentType string) ([]byte, error) {
	// Reset reader to beginning
	reader.Seek(0, 0)

	// Read all content
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.NewSystemError(fmt.Sprintf("failed to read image content for %s", filename), err)
	}

	// Basic image validation
	if len(content) < 10 {
		return nil, errors.NewValidationError(fmt.Sprintf("image file %s too small: %d bytes", filename, len(content)), nil)
	}

	// Validate image headers based on content type
	switch contentType {
	case "image/jpeg":
		// JPEG files start with FF D8
		if len(content) < 2 || content[0] != 0xFF || content[1] != 0xD8 {
			return nil, errors.NewValidationError(fmt.Sprintf("invalid JPEG file %s: incorrect header", filename), nil)
		}
	case "image/png":
		// PNG files start with 89 50 4E 47
		if len(content) < 8 || string(content[:8]) != "\x89PNG\r\n\x1a\n" {
			return nil, errors.NewValidationError(fmt.Sprintf("invalid PNG file %s: incorrect header", filename), nil)
		}
	case "image/gif":
		// GIF files start with GIF87a or GIF89a
		if len(content) < 6 || (string(content[:6]) != "GIF87a" && string(content[:6]) != "GIF89a") {
			return nil, errors.NewValidationError(fmt.Sprintf("invalid GIF file %s: incorrect header", filename), nil)
		}
	case "image/webp":
		// WebP files have RIFF header
		if len(content) < 12 || string(content[:4]) != "RIFF" || string(content[8:12]) != "WEBP" {
			return nil, errors.NewValidationError(fmt.Sprintf("invalid WebP file %s: incorrect header", filename), nil)
		}
	}

	// Decode the image
	reader.Seek(0, 0)
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, errors.NewSystemError(fmt.Sprintf("failed to decode image %s: %v", filename, err), err)
	}

	// Get image dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Smart resizing for very large images (>4000px on any dimension)
	// This provides significant size reduction while maintaining good visual quality
	const maxDimension = 4000
	if width > maxDimension || height > maxDimension {
		// Calculate new dimensions maintaining aspect ratio
		if width > height {
			height = int(float64(height) * float64(maxDimension) / float64(width))
			width = maxDimension
		} else {
			width = int(float64(width) * float64(maxDimension) / float64(height))
			height = maxDimension
		}

		// Resize using Lanczos filter for high quality
		img = imaging.Resize(img, width, height, imaging.Lanczos)
	}

	// Re-encode with optimization based on format
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		// JPEG: Use quality 85 for good balance between size and quality
		opts := &jpeg.Options{Quality: 85}
		if err := jpeg.Encode(&buf, img, opts); err != nil {
			return nil, errors.NewSystemError(fmt.Sprintf("failed to encode JPEG %s", filename), err)
		}
	case "png":
		// PNG: Use best compression for maximum size reduction
		encoder := &png.Encoder{CompressionLevel: png.BestCompression}
		if err := encoder.Encode(&buf, img); err != nil {
			return nil, errors.NewSystemError(fmt.Sprintf("failed to encode PNG %s", filename), err)
		}
	case "gif":
		// GIF: Re-encode with default options
		if err := gif.Encode(&buf, img, nil); err != nil {
			return nil, errors.NewSystemError(fmt.Sprintf("failed to encode GIF %s", filename), err)
		}
	default:
		// For unsupported formats (webp, etc.), return original
		// Note: For webp optimization, use github.com/chai2010/webp or external tools
		return content, nil
	}

	optimizedContent := buf.Bytes()

	// Only return optimized version if it's actually smaller
	if len(optimizedContent) < len(content) {
		return optimizedContent, nil
	}

	// If optimization didn't reduce size, return original
	return content, nil
}

// optimizeTextFile optimizes text-based files
func optimizeTextFile(reader *bytes.Reader, filename string, contentType string) ([]byte, error) {
	// Reset reader to beginning
	reader.Seek(0, 0)

	// Read all content
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.NewSystemError(fmt.Sprintf("failed to read text content for %s (%s)", filename, contentType), err)
	}

	// Basic size validation
	if len(content) < 1 {
		return nil, errors.NewValidationError(fmt.Sprintf("text file %s is empty", filename), nil)
	}

	// Convert to string for text processing
	textContent := string(content)

	// Normalize line endings (convert Windows \r\n to Unix \n)
	textContent = strings.ReplaceAll(textContent, "\r\n", "\n")

	// Split into lines for processing
	lines := strings.Split(textContent, "\n")
	optimizedLines := make([]string, 0, len(lines))

	// Process each line
	for _, line := range lines {
		// Trim trailing whitespace
		line = strings.TrimRight(line, " \t")

		// Skip empty lines at the end (but preserve empty lines within content)
		if line == "" && len(optimizedLines) == 0 {
			continue // Skip leading empty lines
		}

		optimizedLines = append(optimizedLines, line)
	}

	// Remove trailing empty lines
	for len(optimizedLines) > 0 && optimizedLines[len(optimizedLines)-1] == "" {
		optimizedLines = optimizedLines[:len(optimizedLines)-1]
	}

	// Join back with normalized line endings
	optimizedContent := strings.Join(optimizedLines, "\n")

	// Additional optimizations based on content type
	// switch contentType {
	// case "application/json":
	// 	// Basic JSON validation - should start with { or [
	// 	optimizedContent = strings.TrimSpace(optimizedContent)
	// 	if len(optimizedContent) == 0 {
	// 		return nil, errors.NewValidationError("JSON file is empty after optimization", nil)
	// 	}
	// 	if optimizedContent[0] != '{' && optimizedContent[0] != '[' {
	// 		return nil, errors.NewValidationError("invalid JSON file: must start with '{' or '['", nil)
	// 	}
	// }

	return []byte(optimizedContent), nil
}
