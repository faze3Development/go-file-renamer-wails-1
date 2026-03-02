package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"go-file-renamer-wails/file_server/action"
	"go-file-renamer-wails/file_server/advanced_file_operations"
	"go-file-renamer-wails/file_server/config"
	"go-file-renamer-wails/file_server/logger"
	"go-file-renamer-wails/file_server/processor/options/namer"
	"go-file-renamer-wails/file_server/processor/options/patterns"
	"go-file-renamer-wails/file_server/stats"
	"go-file-renamer-wails/file_server/users"
	"go-file-renamer-wails/file_server/watcher"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

// --- App Struct ---

type App struct {
	ctx                    context.Context
	cfg                    config.Config
	watcher                *watcher.Watcher
	stats                  *stats.Stats
	mu                     sync.Mutex
	wg                     sync.WaitGroup
	logger                 *slog.Logger
	advancedFileOperations *advanced_file_operations.AdvancedFileOperations
}

func NewApp() *App {
	appLogger := logger.New("development")
	advFileOps, err := advanced_file_operations.NewAdvancedFileOperations(appLogger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialise advanced file operations: %v\n", err)
		os.Exit(1)
	}

	return &App{
		stats:                  &stats.Stats{}, // Initialize stats
		logger:                 appLogger,
		advancedFileOperations: advFileOps,
	}
}

// --- Wails Lifecycle Hooks ---

func (a *App) Startup(ctx context.Context) {
	ctx = context.WithValue(ctx, "logger", a.logger)
	a.ctx = ctx
}

func (a *App) Shutdown(_ context.Context) {
	a.StopWatching()
}

// --- AppController Interface Implementation ---

func (a *App) Logf(level logger.LogLevel, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)

	if baseLogger := logger.FromContext(a.ctx); baseLogger != nil {
		baseLogger.Log(a.ctx, logger.ToSlogLevel(level), msg)
	}

	logger.Emit(a.ctx, logger.StructuredLogEntry{
		Severity: level.Severity(),
		Message:  msg,
		Source:   "backend.app",
	})
}

func (a *App) GetContext() context.Context {
	return a.ctx

}

// --- Frontend Callable Methods ---

func (a *App) StartWatching(cfg config.Config) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.watcher != nil {
		return fmt.Errorf("watcher is already running")
	}

	a.cfg = cfg

	w, err := watcher.NewWatcher(a, a.cfg, a.stats, logger.FromContext(a.ctx))
	if err != nil {
		a.Logf(logger.ErrorLevel, "Failed to create watcher: %v", err)
		return err
	}
	a.watcher = w

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		runtime.EventsEmit(a.ctx, "watcherStarted")
		a.Logf(logger.InfoLevel, "Watcher started.")
		if err := a.watcher.Start(); err != nil {
			a.Logf(logger.ErrorLevel, "Watcher failed: %v", err)
		}
		a.Logf(logger.InfoLevel, "Watcher has stopped.")

		a.mu.Lock()
		// Emit final stats once (removed duplicate).
		stats.EmitStats(a.ctx, a.stats)
		a.watcher = nil
		a.mu.Unlock()

		runtime.EventsEmit(a.ctx, "watcherStopped")
	}()

	return nil
}

func (a *App) StopWatching() {
	a.mu.Lock()
	if a.watcher != nil {
		// Get the watcher instance and unlock before waiting.
		// This prevents a deadlock.
		watcherToStop := a.watcher
		a.mu.Unlock()

		watcherToStop.Stop()
		a.wg.Wait() // Now we can safely wait for the goroutine to finish.
	} else {
		a.mu.Unlock() // Unlock if there was no watcher to stop.
	}
}

func (a *App) SelectDirectory() (string, error) {
	selected, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Directory",
	})
	if err != nil {
		return "", err
	}
	if selected != "" {
		info, err := os.Stat(selected)
		if err == nil && !info.IsDir() {
			return filepath.Dir(selected), nil
		}
	}
	return selected, nil
}

func (a *App) SelectActionDirectory() (string, error) {
	return a.SelectDirectory()
}

func (a *App) SaveProfile(name string, cfg config.Config) error {
	return users.SaveProfile(name, cfg)
}

func (a *App) LoadProfiles() (map[string]config.Config, error) {
	return users.LoadProfiles()
}

func (a *App) DeleteProfile(name string) error {
	return users.DeleteProfile(name)
}

func (a *App) GetNamerInfo() []namer.Info {
	return namer.GetNamerInfo()
}

func (a *App) GetActionInfo() []action.Info {
	return action.GetActionInfo()
}

func (a *App) GetPatternInfo() []patterns.PatternInfo {
	return patterns.GetPatternInfo()
}

// --- Main ---
func main() {
	app := NewApp()
	err := wails.Run(&options.App{
		Title:            "File Renamer Pro",
		Width:            1400,
		Height:           900,
		MinWidth:         1400,
		MinHeight:        900,
		MaxWidth:         1400,
		MaxHeight:        900,
		DisableResize:    false,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 10, G: 10, B: 10, A: 1},
		OnStartup:        app.Startup,
		OnShutdown:       app.Shutdown,
		Bind: []interface{}{
			app,
			app.advancedFileOperations,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
