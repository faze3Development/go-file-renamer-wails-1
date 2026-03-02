// Copyright (c) 2025 FAZE3 DEVELOPMENT LLC
// All rights reserved.

package pdf

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/errors"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/exiftool"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/security"
)

// StripPDFMetadata removes metadata from a PDF file using ExifTool.
func StripPDFMetadata(pdf io.Reader, logger *slog.Logger) ([]byte, error) {
	// Read the input PDF
	inputData, err := io.ReadAll(pdf)
	if err != nil {
		logger.Error("failed to read PDF", "error", err)
		return nil, errors.NewSystemError("failed to read PDF", err)
	}

	// Create temporary input file
	inputFile, err := os.CreateTemp("", "pdf-input-*.pdf")
	if err != nil {
		logger.Error("failed to create temp input file", "error", err)
		return nil, errors.NewSystemError("failed to create temp file", err)
	}
	defer os.Remove(inputFile.Name())
	defer inputFile.Close()

	// Write PDF data to temp file
	if _, err := inputFile.Write(inputData); err != nil {
		logger.Error("failed to write to temp file", "error", err)
		return nil, errors.NewSystemError("failed to write temp file", err)
	}
	inputFile.Close()

	// Create temporary output file
	outputFile, err := os.CreateTemp("", "pdf-output-*.pdf")
	if err != nil {
		logger.Error("failed to create temp output file", "error", err)
		return nil, errors.NewSystemError("failed to create temp file", err)
	}
	defer os.Remove(outputFile.Name())
	defer outputFile.Close()
	outputFile.Close() // Close it so ExifTool can write to it

	// Validate temp file paths before using in command
	validatedInputPath, err := security.ValidateFilePath(inputFile.Name(), "")
	if err != nil {
		logger.Error("invalid input temp file path", "path", inputFile.Name(), "error", err)
		return nil, errors.NewSystemError("invalid temp file path for exiftool input", err)
	}
	validatedOutputPath, err := security.ValidateFilePath(outputFile.Name(), "")
	if err != nil {
		logger.Error("invalid output temp file path", "path", outputFile.Name(), "error", err)
		return nil, errors.NewSystemError("invalid temp file path for exiftool output", err)
	}

	// Run ExifTool to remove all metadata
	// exiftool -all= -o output.pdf input.pdf
	cmd := exec.Command("exiftool", "-all=", "-o", validatedOutputPath, validatedInputPath) // #nosec G204 - command is hardcoded, paths validated

	// Capture stderr for logging
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		logger.Error("ExifTool failed to remove PDF metadata",
			"error", err,
			"stderr", stderr.String())
		return nil, errors.NewProcessingError("failed to remove PDF metadata with ExifTool", err)
	}

	// Read the cleaned PDF
	cleanData, err := os.ReadFile(outputFile.Name())
	if err != nil {
		logger.Error("failed to read cleaned PDF", "error", err)
		return nil, errors.NewSystemError("failed to read cleaned PDF", err)
	}

	logger.Info("successfully removed PDF metadata",
		"original_size", len(inputData),
		"cleaned_size", len(cleanData))

	return cleanData, nil
}

// PDFMetadata represents extracted PDF metadata
type PDFMetadata struct {
	Title        string `json:"title,omitempty"`
	Author       string `json:"author,omitempty"`
	Subject      string `json:"subject,omitempty"`
	Creator      string `json:"creator,omitempty"`
	Producer     string `json:"producer,omitempty"`
	Keywords     string `json:"keywords,omitempty"`
	CreationDate string `json:"creationDate,omitempty"`
	ModDate      string `json:"modDate,omitempty"`
	PageCount    int    `json:"pageCount,omitempty"`
	PDFVersion   string `json:"pdfVersion,omitempty"`
}

// ExtractPDFMetadata extracts metadata from a PDF file using ExifTool
func ExtractPDFMetadata(pdf io.Reader, exifService exiftool.Service, logger *slog.Logger) (*PDFMetadata, error) {
	inputData, err := io.ReadAll(pdf)
	if err != nil {
		return nil, errors.NewSystemError("failed to read PDF for metadata extraction", err)
	}

	// Create a temporary file for ExifTool to process
	tempFile, err := os.CreateTemp("", "pdf-metadata-*")
	if err != nil {
		logger.Warn("Failed to create temp file for PDF metadata extraction", "error", err)
		return &PDFMetadata{}, nil
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write content to temp file
	if _, err := tempFile.Write(inputData); err != nil {
		logger.Warn("Failed to write PDF content to temp file", "error", err)
		return &PDFMetadata{}, nil
	}
	tempFile.Close()

	// Use ExifTool to extract metadata
	// Create a temporary file for the exiftool service
	tempFileForExif := &common.File{
		Filename: tempFile.Name(),
		Content:  inputData,
		Size:     int64(len(inputData)),
	}

	metadata, err := exifService.ExtractMetadata(tempFileForExif)
	if err != nil {
		logger.Warn("Failed to extract PDF metadata", "error", err)
		return &PDFMetadata{}, nil
	}

	if len(metadata) == 0 {
		logger.Warn("No metadata extracted from PDF by ExifTool")
		return &PDFMetadata{}, nil
	}

	// Extract PDF-specific metadata from ExifTool fields
	pdfMetadata := &PDFMetadata{}

	if title, ok := metadata["Title"]; ok {
		pdfMetadata.Title = fmt.Sprintf("%v", title)
	}

	if author, ok := metadata["Author"]; ok {
		pdfMetadata.Author = fmt.Sprintf("%v", author)
	}

	if subject, ok := metadata["Subject"]; ok {
		pdfMetadata.Subject = fmt.Sprintf("%v", subject)
	}

	if creator, ok := metadata["Creator"]; ok {
		pdfMetadata.Creator = fmt.Sprintf("%v", creator)
	}

	if producer, ok := metadata["Producer"]; ok {
		pdfMetadata.Producer = fmt.Sprintf("%v", producer)
	}

	if keywords, ok := metadata["Keywords"]; ok {
		pdfMetadata.Keywords = fmt.Sprintf("%v", keywords)
	}

	if pageCount, ok := metadata["PageCount"]; ok {
		if pc, ok := pageCount.(float64); ok {
			pdfMetadata.PageCount = int(pc)
		}
	}

	if pdfVersion, ok := metadata["PDFVersion"]; ok {
		pdfMetadata.PDFVersion = fmt.Sprintf("%v", pdfVersion)
	}

	if creationDate, ok := metadata["CreateDate"]; ok {
		pdfMetadata.CreationDate = fmt.Sprintf("%v", creationDate)
	}

	if modDate, ok := metadata["ModifyDate"]; ok {
		pdfMetadata.ModDate = fmt.Sprintf("%v", modDate)
	}

	return pdfMetadata, nil
}
