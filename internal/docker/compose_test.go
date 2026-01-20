package docker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tobias/muxbee/internal/config"
)

func TestDockerComposeYAMLEmbedded(t *testing.T) {
	assert.NotEmpty(t, DockerComposeYAML)
	assert.Contains(t, string(DockerComposeYAML), "name: muxbee")
	assert.Contains(t, string(DockerComposeYAML), "postgres:")
	assert.Contains(t, string(DockerComposeYAML), "synapse:")
	assert.Contains(t, string(DockerComposeYAML), "element:")
	assert.Contains(t, string(DockerComposeYAML), "mautrix-whatsapp:")
}

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))

	cfg := &config.Config{
		Postgres: config.PostgresConfig{
			Password: "testpass123",
		},
	}

	compose := New(cfg)
	assert.NotNil(t, compose)
	assert.Contains(t, compose.configDir, "muxbee")
	assert.Contains(t, compose.dataDir, "muxbee")

	// Check environment variables
	hasConfigDir := false
	hasDataDir := false
	hasPostgresPass := false
	for _, env := range compose.env {
		if env == "CONFIG_DIR="+compose.configDir {
			hasConfigDir = true
		}
		if env == "DATA_DIR="+compose.dataDir {
			hasDataDir = true
		}
		if env == "POSTGRES_PASSWORD=testpass123" {
			hasPostgresPass = true
		}
	}
	assert.True(t, hasConfigDir)
	assert.True(t, hasDataDir)
	assert.True(t, hasPostgresPass)
}

func TestWriteComposeFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))

	cfg := &config.Config{
		Postgres: config.PostgresConfig{
			Password: "testpass",
		},
	}

	compose := New(cfg)
	err := compose.WriteComposeFile()
	require.NoError(t, err)

	// Check that file was written
	composePath := filepath.Join(config.ConfigDir(), "docker-compose.yml")
	assert.FileExists(t, composePath)

	// Verify contents
	content, err := os.ReadFile(composePath)
	require.NoError(t, err)
	assert.Equal(t, DockerComposeYAML, content)
}

func TestGetProfiles(t *testing.T) {
	elementDisabled := false

	tests := []struct {
		name     string
		cfg      *config.Config
		expected []string
	}{
		{
			name: "default (element enabled)",
			cfg: &config.Config{
				EnabledBridges: []string{},
				HTTPS:          config.HTTPSConfig{Enabled: false},
			},
			expected: []string{"element"},
		},
		{
			name: "element disabled",
			cfg: &config.Config{
				EnabledBridges: []string{},
				ElementEnabled: &elementDisabled,
				HTTPS:          config.HTTPSConfig{Enabled: false},
			},
			expected: []string{},
		},
		{
			name: "with bridges",
			cfg: &config.Config{
				EnabledBridges: []string{"whatsapp", "telegram"},
				HTTPS:          config.HTTPSConfig{Enabled: false},
			},
			expected: []string{"whatsapp", "telegram", "element"},
		},
		{
			name: "with https",
			cfg: &config.Config{
				EnabledBridges: []string{},
				HTTPS:          config.HTTPSConfig{Enabled: true},
			},
			expected: []string{"element", "https"},
		},
		{
			name: "with bridges and https",
			cfg: &config.Config{
				EnabledBridges: []string{"signal"},
				HTTPS:          config.HTTPSConfig{Enabled: true},
			},
			expected: []string{"signal", "element", "https"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profiles := GetProfiles(tt.cfg)
			assert.Equal(t, tt.expected, profiles)
		})
	}
}

func TestParseServiceName(t *testing.T) {
	tests := []struct {
		containerName string
		expected      string
	}{
		{"muxbee-postgres-1", "postgres"},
		{"muxbee-synapse-1", "synapse"},
		{"muxbee-mautrix-whatsapp-1", "mautrix-whatsapp"},
		{"muxbee-element-1", "element"},
		{"simple", "simple"},
	}

	for _, tt := range tests {
		t.Run(tt.containerName, func(t *testing.T) {
			result := ParseServiceName(tt.containerName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestComposePath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))

	cfg := &config.Config{
		Postgres: config.PostgresConfig{Password: "test"},
	}

	compose := New(cfg)
	path := compose.composePath()
	assert.Contains(t, path, "docker-compose.yml")
	assert.Contains(t, path, "muxbee")
}

func TestBuildCommand(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))

	cfg := &config.Config{
		Postgres: config.PostgresConfig{Password: "testpass"},
	}

	compose := New(cfg)
	cmd := compose.buildCommand("up", "-d")

	// Check that the command is docker
	assert.Equal(t, "docker", cmd.Path[len(cmd.Path)-6:])

	// Check args include compose, -f, and the actual args
	args := cmd.Args[1:] // skip program name
	assert.Contains(t, args, "compose")
	assert.Contains(t, args, "-f")
	assert.Contains(t, args, "up")
	assert.Contains(t, args, "-d")
}

func TestBuildCommand_Environment(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))

	cfg := &config.Config{
		Postgres: config.PostgresConfig{Password: "secretpass"},
	}

	compose := New(cfg)
	cmd := compose.buildCommand("ps")

	// Check environment includes our custom vars
	envMap := make(map[string]bool)
	for _, e := range cmd.Env {
		envMap[e] = true
	}

	assert.True(t, envMap["POSTGRES_PASSWORD=secretpass"])
}

func TestBuildCommand_MultipleArgs(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))

	cfg := &config.Config{
		Postgres: config.PostgresConfig{Password: "test"},
	}

	compose := New(cfg)
	cmd := compose.buildCommand("--profile", "whatsapp", "--profile", "signal", "up", "-d")

	args := cmd.Args
	// Find indices of profiles
	hasProfile1 := false
	hasProfile2 := false
	for i, arg := range args {
		if arg == "--profile" && i+1 < len(args) {
			if args[i+1] == "whatsapp" {
				hasProfile1 = true
			}
			if args[i+1] == "signal" {
				hasProfile2 = true
			}
		}
	}
	assert.True(t, hasProfile1, "should have whatsapp profile")
	assert.True(t, hasProfile2, "should have signal profile")
}

func TestIsServiceRunning_NotRunning(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))

	cfg := &config.Config{
		Postgres: config.PostgresConfig{Password: "test"},
	}

	compose := New(cfg)
	// Without Docker running, IsServiceRunning should return false
	result := compose.IsServiceRunning("nonexistent-service")
	assert.False(t, result)
}

func TestIsRunning_NoServices(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))

	cfg := &config.Config{
		Postgres: config.PostgresConfig{Password: "test"},
	}

	compose := New(cfg)
	// Without Docker running, IsRunning should return false
	result := compose.IsRunning()
	assert.False(t, result)
}
