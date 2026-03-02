package watcher

import (
	"context"
	"errors" // Standard library errors package
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/security"
	customErrors "go-file-renamer-wails/file_server/errors" // Custom errors package
	"go-file-renamer-wails/file_server/logger"
	"go-file-renamer-wails/file_server/stats"

	"github.com/fsnotify/fsnotify"
)

const numWorkers = 4      // Number of concurrent file processing workers
const jobQueueSize = 1024 // Buffer size for the job queue
const watcherSource = "watcher"

func (w *Watcher) Start() error {
	defer func() {
		if err := w.fs.Close(); err != nil && err.Error() != "file already closed" {
			w.logDebug("Error closing watcher filesystem handle", "error", err)
		}
	}()

	// --- Worker Pool Setup ---
	jobQueue := make(chan string, jobQueueSize)
	var workerWg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		workerWg.Add(1)
		go func(workerID int) {
			defer workerWg.Done()
			w.logDebug("Worker started", "workerID", workerID)
			w.fileProcessor(jobQueue)
		}(i)
	}

	for _, root := range w.cfg.WatchPaths {
		if err := w.addWatchDir(root); err != nil {
			w.logWarn("Failed to watch", "root", root, "error", err)
			atomic.AddUint64(&w.stats.Errors, 1)
		}
	}

	if !w.cfg.NoInitialScan {
		w.logInfo("Performing initial scan...")
		for _, d := range w.cfg.WatchPaths {
			if err := w.scanDir(d, jobQueue); err != nil {
				w.logWarn("Scan error", "directory", d, "error", err)
				atomic.AddUint64(&w.stats.Errors, 1)
			}
		}
		w.logInfo("Initial scan complete.")
	}

	statsTicker := time.NewTicker(1 * time.Second)
	defer statsTicker.Stop()

	go func() {
		for {
			select {
			case <-w.ctx.Done():
				return
			case <-statsTicker.C:
				stats.EmitStats(w.app.GetContext(), w.stats)
			}
		}
	}()

	for {
		select {
		case <-w.ctx.Done():
			// Context is canceled, close the job queue to signal workers to stop.
			close(jobQueue)
			// Wait for all workers to finish their current job and exit.
			w.logDebug("Waiting for workers to finish...")
			workerWg.Wait()
			w.logDebug("All workers have finished.")
			return nil
		case ev, ok := <-w.fs.Events:
			if !ok {
				return nil
			}
			w.handleEvent(ev, jobQueue)
		case err, ok := <-w.fs.Errors:
			if !ok {
				return nil
			}
			w.logWarn("Watcher error", "error", err)
			atomic.AddUint64(&w.stats.Errors, 1)
		}
	}
}

func (w *Watcher) Stop() {
	w.cancel()
}

func (w *Watcher) handleEvent(ev fsnotify.Event, jobQueue chan<- string) {
	if ev.Op&fsnotify.Create == fsnotify.Create {
		if fi, err := os.Lstat(ev.Name); err == nil && fi.IsDir() {
			_ = w.addWatchDir(ev.Name)
			return
		}
	}

	if ev.Op&(fsnotify.Create|fsnotify.Rename|fsnotify.Write) != 0 {
		if fi, err := os.Lstat(ev.Name); err == nil && fi.Mode().IsRegular() {
			atomic.AddUint64(&w.stats.Scanned, 1)
			// Send the file path to the job queue instead of spawning a goroutine
			select {
			case jobQueue <- ev.Name:
				w.logDebug("Queued for processing", "file", ev.Name)
			default:
				w.logWarn("Job queue is full. Discarding event", "file", ev.Name)
			}
		}
	}
}

func (w *Watcher) addWatchDir(dir string) error {
	if err := w.fs.Add(dir); err != nil {
		return err
	}
	w.logInfo("Watching", "directory", dir)
	if !w.cfg.Recursive {
		return nil
	}
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			w.logWarn("Traversal issue while adding watch", "path", path, "error", err)
			atomic.AddUint64(&w.stats.Errors, 1)
			return nil
		}
		if d.IsDir() && path != dir {
			if err := w.fs.Add(path); err == nil {
				w.logInfo("Watching", "path", path)
			} else {
				w.logWarn("Failed to watch", "path", path, "error", err)
				atomic.AddUint64(&w.stats.Errors, 1)
			}
		}
		return nil
	})
}

func (w *Watcher) scanDir(root string, jobQueue chan<- string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			w.logWarn("Traversal issue", "path", path, "error", err)
			atomic.AddUint64(&w.stats.Errors, 1)
			return nil
		}
		if !d.IsDir() {
			atomic.AddUint64(&w.stats.Scanned, 1)
			// Send the file path to the job queue
			select {
			case jobQueue <- path:
			default:
				w.logWarn("Job queue is full during initial scan. Discarding", "file", path)
			}
		}
		if !w.cfg.Recursive && d.IsDir() && path != root {
			return fs.SkipDir
		}
		return nil
	})
}

// fileProcessor is the function run by each worker.
// It reads file paths from the jobQueue and processes them.
func (w *Watcher) fileProcessor(jobQueue <-chan string) {
	for path := range jobQueue {
		w.logDebug("Processing", "file", path)
		_ = w.handleFileWithRetry(path)
	}
}

func (w *Watcher) handleFileWithRetry(path string) error {
	retries := w.cfg.Retries
	if retries < 1 {
		retries = 1
	}

	for attempt := 0; attempt < retries; attempt++ {
		err := w.handleFile(path)
		// Success on the first try or a subsequent one
		if err == nil {
			return nil
		}

		// --- Smart Retry Logic ---

		// Case 1: File disappeared during processing. This is not an error we should retry.
		if errors.Is(err, os.ErrNotExist) {
			w.logDebug("File disappeared, ignoring", "file", path)
			return nil
		}

		// Case 2: Check for specific, non-retriable processing errors.
		var processingErr *customErrors.ProcessingError
		if errors.As(err, &processingErr) {
			w.logWarn("Permanent error processing. No retry.", "file", path, "error", err)
			atomic.AddUint64(&w.stats.Errors, 1)
			return err
		}

		// Case 3: For other errors (likely transient I/O issues like file locks), perform a retry.
		backoff := time.Duration(100*(1<<attempt)) * time.Millisecond // Exponential backoff
		w.logDebug("Retry due to transient error", "attempt", attempt+1, "retries", retries, "file", path, "error", err, "next_attempt_in", backoff)
		time.Sleep(backoff)
	}

	lastErr := fmt.Errorf("failed to process %s after %d retries", path, retries)
	w.logWarn(lastErr.Error())
	atomic.AddUint64(&w.stats.Errors, 1)
	return lastErr
}

func (w *Watcher) handleFile(path string) error {
	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)

	if !w.nameRe.MatchString(base) {
		w.logDebug("SKIP (pattern mismatch)", "file", path)
		atomic.AddUint64(&w.stats.Skipped, 1)
		return nil
	}

	if err := w.waitForStableFile(path); err != nil {
		return customErrors.NewProcessingError(fmt.Sprintf("File is not stable: %s", path), err) // Using customErrors
	}

	if ext == "" {
		if imgExt, ok := w.detectImageExtension(path); ok {
			ext = imgExt
		}
	}

	newPath, err := w.generateUniqueName(dir, ext, base)
	if err != nil {
		return err
	}

	if w.cfg.DryRun {
		w.logInfo("DRY-RUN", "old_path", path, "new_path", newPath)
		atomic.AddUint64(&w.stats.Renamed, 1)
		return nil
	}

	if err := os.Rename(path, newPath); err != nil {
		if isCrossDevice(err) {
			w.logDebug("Cross-device rename detected, using copy+remove", "old_path", path, "new_path", newPath)
			if copyErr := w.copyFile(path, newPath); copyErr != nil { // Using customErrors
				return customErrors.NewProcessingError(fmt.Sprintf("Copy fallback failed for %s", path), copyErr)
			}
			if removeErr := os.Remove(path); removeErr != nil { // Using customErrors
				return customErrors.NewProcessingError(fmt.Sprintf("Cleanup of original file failed for %s", path), removeErr)
			}
		} else {
			return err
		}
	}

	w.logInfo("RENAMED", "old_path", path, "new_path", newPath)
	atomic.AddUint64(&w.stats.Renamed, 1)

	if err := w.action.Execute(w.ctx, newPath, w.logger); err != nil {
		w.logError("Post-rename action failed", "file", newPath, "error", err)
		atomic.AddUint64(&w.stats.Errors, 1)
	}

	return nil
}

func (w *Watcher) generateUniqueName(dir, ext, originalBaseName string) (string, error) {
	for i := uint64(0); i < 100; i++ {
		baseName, err := w.namer.GenerateName(originalBaseName, i)
		if err != nil {
			return "", err
		}
		candidate := filepath.Join(dir, baseName+ext) // Using standard library errors.Is
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		}
	}
	return "", customErrors.NewProcessingError("Failed to generate a unique name after many attempts", nil) // Using customErrors
}

func (w *Watcher) waitForStableFile(path string) error {
	settle := time.Duration(w.cfg.Settle) * time.Millisecond
	timeout := time.Duration(w.cfg.SettleTimeout) * time.Second

	if settle <= 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var lastSize int64 = -1
	var lastModTime time.Time

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out after %v waiting for %s to be stable", timeout, path)
		default:
			info, err := os.Stat(path)
			if err != nil {
				return err
			}

			if lastSize == -1 {
				lastSize = info.Size()
				lastModTime = info.ModTime()
			} else {
				if info.Size() == lastSize && info.ModTime().Equal(lastModTime) {
					if time.Since(lastModTime) >= settle {
						w.logDebug("File is stable", "file", path)
						return nil
					}
				} else {
					lastSize = info.Size()
					lastModTime = info.ModTime()
				}
			}
			time.Sleep(settle / 3)
		}
	}
}

func (w *Watcher) detectImageExtension(path string) (string, bool) {
	// Validate path to prevent directory traversal attacks
	validatedPath, err := security.ValidateFilePath(path, "")
	if err != nil {
		w.logWarn("Invalid path for image extension detection", "path", path, "error", err)
		return "", false
	}

	f, err := os.Open(validatedPath) // #nosec G304 - path validated above
	if err != nil {
		return "", false
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			w.logWarn("Failed to close file after type detection", "file", path, "error", err)
		}
	}(f)

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		w.logWarn("Could not read file for type detection", "file", path, "error", err)
		return "", false
	}
	buf = buf[:n]

	contentType := http.DetectContentType(buf)
	w.logDebug("Detected content type", "content_type", contentType, "file", path)

	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/gif":
		return ".gif", true
	case "image/webp":
		return ".webp", true
	case "image/bmp":
		return ".bmp", true
	case "image/tiff":
		return ".tiff", true
	default:
		return "", false
	}
}

func (w *Watcher) copyFile(src, dst string) error {
	// Validate paths to prevent directory traversal attacks
	validatedSrc, err := security.ValidateFilePath(src, "")
	if err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}
	validatedDst, err := security.ValidateFilePath(dst, "")
	if err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	in, err := os.Open(validatedSrc) // #nosec G304 - path validated above
	if err != nil {
		return err
	}
	defer func() {
		_ = in.Close()
	}()

	fi, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(validatedDst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fi.Mode()) // #nosec G304 - path validated above
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	_, err = io.Copy(out, in)
	return err
}

func (w *Watcher) log(level logger.LogLevel, message string, kv ...any) {
	if w == nil {
		return
	}

	if w.logger != nil {
		w.logger.LogAttrs(context.Background(), logger.ToSlogLevel(level), message, logger.KeyValuesToAttrs(kv...)...)
	}

	if w.app == nil {
		return
	}

	ctx := w.app.GetContext()
	if ctx == nil {
		return
	}

	metadata := logger.KeyValuesToMap(kv...)
	entry := logger.StructuredLogEntry{
		Severity: level.Severity(),
		Message:  message,
		Source:   watcherSource,
	}
	if len(metadata) > 0 {
		entry.Metadata = metadata
	}
	logger.Emit(ctx, entry)
}

func (w *Watcher) logDebug(message string, kv ...any) {
	w.log(logger.DebugLevel, message, kv...)
}

func (w *Watcher) logInfo(message string, kv ...any) {
	w.log(logger.InfoLevel, message, kv...)
}

func (w *Watcher) logWarn(message string, kv ...any) {
	w.log(logger.WarnLevel, message, kv...)
}

func (w *Watcher) logError(message string, kv ...any) {
	w.log(logger.ErrorLevel, message, kv...)
}

func isCrossDevice(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, syscall.EXDEV) { // Using standard library errors.Is
		return true
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "cross-device") || strings.Contains(s, "invalid cross-device link")
}
