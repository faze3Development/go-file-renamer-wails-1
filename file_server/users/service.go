package users

import (
	"encoding/json"
	"fmt"
	"go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/security"
	"go-file-renamer-wails/file_server/config"
	"os"
	"path/filepath"
)

const profileFileName = "profiles.json"

// --- Profile Management ---

// SaveProfile saves a given configuration under a specific name.
func SaveProfile(name string, cfg config.Config) error {
	profiles, err := LoadProfiles()
	if err != nil {
		return err
	}

	profiles[name] = cfg

	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return err
	}

	path, err := getProfilesPath()
	if err != nil {
		return err
	}

	// Use 0600 for permissions to restrict access to the current user.
	return os.WriteFile(path, data, 0600)
}

// LoadProfiles reads all saved profiles from the JSON file.
func LoadProfiles() (map[string]config.Config, error) {
	path, err := getProfilesPath()
	if err != nil {
		return nil, err
	}

	// Validate the path to prevent directory traversal
	validatedPath, err := security.ValidateFilePath(path, "")
	if err != nil {
		return nil, fmt.Errorf("invalid profiles path: %w", err)
	}

	profiles := make(map[string]config.Config)

	data, err := os.ReadFile(validatedPath) // #nosec G304 - path validated above
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, which is fine. Return an empty map.
			return profiles, nil
		}
		return nil, fmt.Errorf("failed to read profiles file: %w", err)
	}

	// If the file is empty, return an empty map.
	if len(data) == 0 {
		return profiles, nil
	}

	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, fmt.Errorf("failed to parse profiles JSON: %w", err)
	}
	return profiles, nil
}

// GetProfile retrieves a single profile by its name.
func GetProfile(name string) (config.Config, error) {
	profiles, err := LoadProfiles()
	if err != nil {
		return config.Config{}, err
	}

	profile, ok := profiles[name]
	if !ok {
		return config.Config{}, fmt.Errorf("profile '%s' not found", name)
	}

	return profile, nil
}

// DeleteProfile removes a profile by its name.
func DeleteProfile(name string) error {
	profiles, err := LoadProfiles()
	if err != nil {
		return err
	}

	if _, ok := profiles[name]; !ok {
		return fmt.Errorf("profile '%s' not found, cannot delete", name)
	}

	delete(profiles, name)

	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return err
	}

	path, err := getProfilesPath()
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// getProfilesPath determines the path for the configuration file in the user's standard config location.
func getProfilesPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not find user config directory: %w", err)
	}

	appConfigDir := filepath.Join(configDir, "go-file-renamer")
	if err := os.MkdirAll(appConfigDir, 0750); err != nil {
		return "", fmt.Errorf("could not create app config directory: %w", err)
	}

	return filepath.Join(appConfigDir, profileFileName), nil
}
