package action

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/security"
)

// --- Action Implementations ---
func (a *NoneAction) Execute(context.Context, string, *slog.Logger) error {
	return nil // Do nothing
}

func (a *NoneAction) Info() Info {
	return Info{
		ID:          "none",
		Name:        "None",
		Description: "Do not perform any action after renaming.",
		FieldLabel:  "", // No input needed
	}
}

func (a *MoveAction) Execute(ctx context.Context, filePath string, logger *slog.Logger) error {
	if a.destinationPath == "" {
		return fmt.Errorf("move action destination path is not set")
	}
	fileName := filepath.Base(filePath)
	newPath := filepath.Join(a.destinationPath, fileName)

	logger.Info("MOVING: %s -> %s", filePath, newPath)
	// Ensure the destination directory exists
	if err := os.MkdirAll(a.destinationPath, 0750); err != nil {
		return fmt.Errorf("could not create move destination directory: %w", err)
	}

	return os.Rename(filePath, newPath)
}

func (a *MoveAction) Info() Info {
	return Info{
		ID:          "move",
		Name:        "Move to Folder",
		Description: "Moves the renamed file to a specified folder.",
		FieldLabel:  "Destination Folder",
	}
}

func (a *CopyAction) Execute(ctx context.Context, filePath string, logger *slog.Logger) error {
	if a.destinationPath == "" {
		return fmt.Errorf("copy action destination path is not set")
	}
	fileName := filepath.Base(filePath)
	newPath := filepath.Join(a.destinationPath, fileName)

	logger.Info("COPYING: %s -> %s", filePath, newPath)

	// Ensure the destination directory exists
	if err := os.MkdirAll(a.destinationPath, 0750); err != nil {
		return fmt.Errorf("could not create copy destination directory: %w", err)
	}

	return copyFile(ctx, filePath, newPath, logger)
}

func (a *CopyAction) Info() Info {
	return Info{
		ID:          "copy",
		Name:        "Copy to Folder",
		Description: "Copies the renamed file to a specified folder.",
		FieldLabel:  "Destination Folder",
	}
}

func (a *AdvancedOperationsAction) Execute(ctx context.Context, filePath string, logger *slog.Logger) error {
	if logger != nil {
		logger.Info("Advanced file operations action selected", "file", filePath)
	}
	return nil
}

func (a *AdvancedOperationsAction) Info() Info {
	return Info{
		ID:          "advanced_operations",
		Name:        "Advanced File Operations",
		Description: "Use the advanced processing pipeline for metadata workflows and optimizations.",
	}
}

// copyFile is a helper to copy a file from src to dst with context cancellation support.
// It respects context cancellation during the I/O operation, allowing for timeout and cancellation.
func copyFile(ctx context.Context, src, dst string, logger *slog.Logger) error {
	// Validate paths to prevent directory traversal attacks
	validatedSrc, err := security.ValidateFilePath(src, "")
	if err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}
	validatedDst, err := security.ValidateFilePath(dst, "")
	if err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	// Check if context is already canceled before starting
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context canceled before copy: %w", err)
	}

	in, err := os.Open(validatedSrc) // #nosec G304 - path validated above
	if err != nil {
		return err
	}
	defer func(in *os.File) {
		err := in.Close()
		if err != nil {
			// Log the error, but we can't do much else. The primary operation might have succeeded.
			logger.Warn("Failed to close source file %s: %v", validatedSrc, err)
		}
	}(in)

	fi, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(validatedDst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fi.Mode()) // #nosec G304 - path validated above
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			logger.Warn("Failed to close destination file %s: %v", validatedDst, err)
		}
	}(out)

	// Use a buffer for copying with periodic cancellation checks
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("copy operation canceled: %w", ctx.Err())
		default:
		}

		// Read from source
		nr, er := in.Read(buf)
		if nr > 0 {
			// Write to destination
			nw, ew := out.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = fmt.Errorf("invalid write result")
				}
			}
			if ew != nil {
				return ew
			}
			if nr != nw {
				return io.ErrShortWrite
			}
		}
		if er != nil {
			if er != io.EOF {
				return er
			}
			break
		}
	}

	return nil
}
