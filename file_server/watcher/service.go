package watcher

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/fsnotify/fsnotify"

	"go-file-renamer-wails/file_server/action"
	"go-file-renamer-wails/file_server/config"
	customErrors "go-file-renamer-wails/file_server/errors"
	"go-file-renamer-wails/file_server/processor/options/namer"
	"go-file-renamer-wails/file_server/stats"
)

func NewWatcher(app AppController, cfg config.Config, s *stats.Stats, logger *slog.Logger) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, customErrors.NewProcessingError("Failed to create filesystem watcher", err)
	}

	nameRe, err := regexp.Compile(cfg.NamePattern)
	if err != nil {
		return nil, customErrors.NewValidationError(fmt.Sprintf("Invalid Filename Pattern Regex: %s", err.Error()))
	}

	var selectedNamer namer.Namer
	switch cfg.NamerID {
	case "random":
		selectedNamer = namer.NewRandomNamer(cfg.RandomLength)
	case "datetime":
		dateTimeFormat := cfg.DateTimeFormat
		if dateTimeFormat == "" {
			dateTimeFormat = time.RFC3339
		}
		selectedNamer = namer.NewDateTimeNamer(dateTimeFormat)
	case "template":
		selectedNamer = namer.NewTemplateNamer(cfg.TemplateString)
	case "copy":
		selectedNamer = namer.NewCopyNamer()
	default:
		logger.Warn("Invalid namer. Falling back to random", "namerID", cfg.NamerID)
		selectedNamer = namer.NewRandomNamer(cfg.RandomLength)
	}

	actions := action.GetActionRegistry(cfg)
	selectedAction, ok := actions[cfg.ActionID]
	if !ok {
		logger.Warn("Invalid action. Falling back to none", "actionID", cfg.ActionID)
		selectedAction = actions["none"]
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Watcher{
		app:    app,
		cfg:    cfg,
		fs:     fsWatcher,
		ctx:    ctx,
		cancel: cancel,
		stats:  s,
		nameRe: nameRe,
		namer:  selectedNamer,
		action: selectedAction,
		logger: logger,
	}, nil
}
