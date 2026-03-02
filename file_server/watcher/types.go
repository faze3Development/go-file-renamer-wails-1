package watcher

import (
	"context"
	"log/slog"

	"go-file-renamer-wails/file_server/action"
	"go-file-renamer-wails/file_server/config"
	"go-file-renamer-wails/file_server/processor/options/namer"
	"go-file-renamer-wails/file_server/stats"
	"regexp"

	"github.com/fsnotify/fsnotify"
)

// --- Structs & Interfaces ---

type AppController interface {
	GetContext() context.Context
}
type Watcher struct {
	ctx    context.Context
	app    AppController
	cfg    config.Config
	fs     *fsnotify.Watcher
	cancel context.CancelFunc
	stats  *stats.Stats
	nameRe *regexp.Regexp
	namer  namer.Namer
	action action.Action
	logger *slog.Logger
}
