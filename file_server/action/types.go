package action

import (
	"context"
	"log/slog"
)

// ActionInfo provides metadata about an action for the UI.
type Info struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// FieldLabel describes the input needed from the user, e.g., "Destination Folder"
	FieldLabel string `json:"fieldLabel"`
}

// Action defines the interface for different post-rename strategies.
type Action interface {
	Execute(ctx context.Context, filePath string, logger *slog.Logger) error
	Info() Info
}

// 3. CopyAction
type CopyAction struct {
	destinationPath string
}

// 2. MoveAction
type MoveAction struct {
	destinationPath string
}

// 1. NoneAction (Default)
type NoneAction struct{}

// 4. AdvancedOperationsAction integrates advanced pipelines.
type AdvancedOperationsAction struct{}
