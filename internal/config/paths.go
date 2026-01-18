package config

import (
	"os"
	"path/filepath"
)

// ConfigDir returns the configuration directory path following XDG spec
func ConfigDir() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "muxchat")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "muxchat")
}

// DataDir returns the data directory path following XDG spec
func DataDir() string {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, "muxchat")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "muxchat")
}

// SettingsPath returns the path to the settings.yaml file
func SettingsPath() string {
	return filepath.Join(ConfigDir(), "settings.yaml")
}

// DockerComposePath returns the path to the docker-compose.yml file
func DockerComposePath() string {
	return filepath.Join(ConfigDir(), "docker-compose.yml")
}

// EnsureDirs creates the config and data directories if they don't exist
func EnsureDirs() error {
	configDir := ConfigDir()
	dataDir := DataDir()

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	// Create subdirectories
	subdirs := []string{
		filepath.Join(configDir, "synapse"),
		filepath.Join(configDir, "element"),
		filepath.Join(configDir, "caddy"),
		filepath.Join(configDir, "bridges"),
		filepath.Join(dataDir, "synapse"),
		filepath.Join(dataDir, "postgres"),
		filepath.Join(dataDir, "caddy"),
		filepath.Join(dataDir, "bridges"),
	}

	for _, dir := range subdirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}
