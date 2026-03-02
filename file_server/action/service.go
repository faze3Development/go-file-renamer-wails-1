package action

import (
	"go-file-renamer-wails/file_server/config"
)

// GetActionRegistry creates and configures all available actions.
func GetActionRegistry(cfg config.Config) map[string]Action {
	// The destination path for move/copy is read from the generic ActionConfig map.
	// This allows for flexible configuration from the frontend.
	destinationPath := ""
	if cfg.ActionConfig != nil {
		if value, ok := cfg.ActionConfig["destination"]; ok && value != "" {
			destinationPath = value
		} else if value, ok := cfg.ActionConfig["destinationPath"]; ok {
			destinationPath = value
		}
	}

	return map[string]Action{
		"none": &NoneAction{},
		"move": &MoveAction{
			destinationPath: destinationPath,
		},
		"copy": &CopyAction{
			destinationPath: destinationPath,
		},
		"advanced_operations": &AdvancedOperationsAction{},
	}
}

// GetActionInfo returns metadata for all available actions.
func GetActionInfo() []Info {
	return []Info{
		(&NoneAction{}).Info(),
		(&MoveAction{}).Info(),
		(&CopyAction{}).Info(),
		(&AdvancedOperationsAction{}).Info(),
	}
}
