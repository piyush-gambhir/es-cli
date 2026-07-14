package config

import (
	"os"
	"path/filepath"
)

const appName = "es-cli"

// ConfigDir returns the configuration directory, respecting XDG_CONFIG_HOME.
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, appName)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", appName)
	}
	return filepath.Join(home, ".config", appName)
}

// ConfigFilePath returns the full path to the config file.
func ConfigFilePath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}
