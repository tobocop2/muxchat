package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDir(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	t.Setenv("XDG_CONFIG_HOME", "/tmp/test-config")
	assert.Equal(t, "/tmp/test-config/muxbee", ConfigDir())

	// Test without XDG_CONFIG_HOME
	t.Setenv("XDG_CONFIG_HOME", "")
	home, _ := os.UserHomeDir()
	assert.Equal(t, filepath.Join(home, ".config", "muxbee"), ConfigDir())
}

func TestDataDir(t *testing.T) {
	// Test with XDG_DATA_HOME set
	t.Setenv("XDG_DATA_HOME", "/tmp/test-data")
	assert.Equal(t, "/tmp/test-data/muxbee", DataDir())

	// Test without XDG_DATA_HOME
	t.Setenv("XDG_DATA_HOME", "")
	home, _ := os.UserHomeDir()
	assert.Equal(t, filepath.Join(home, ".local", "share", "muxbee"), DataDir())
}

func TestGeneratePassword(t *testing.T) {
	pass1, err := GeneratePassword(16)
	require.NoError(t, err)
	assert.Len(t, pass1, 16)

	pass2, err := GeneratePassword(16)
	require.NoError(t, err)
	assert.Len(t, pass2, 16)

	// Passwords should be different
	assert.NotEqual(t, pass1, pass2)
}

func TestNewDefaultConfig(t *testing.T) {
	cfg, err := NewDefaultConfig()
	require.NoError(t, err)

	assert.Equal(t, "localhost", cfg.ServerName)
	assert.Equal(t, "local", cfg.ConnectivityMode)
	assert.Equal(t, "synapse", cfg.Postgres.User)
	assert.Equal(t, "synapse", cfg.Postgres.Database)
	assert.Len(t, cfg.Postgres.Password, 32)
	assert.Equal(t, "admin", cfg.Admin.Username)
	assert.Len(t, cfg.Admin.Password, 16)
	assert.False(t, cfg.HTTPS.Enabled)
	assert.Empty(t, cfg.EnabledBridges)
}

func TestConfigSaveAndLoad(t *testing.T) {
	// Use temp directory
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg, err := NewDefaultConfig()
	require.NoError(t, err)

	cfg.ServerName = "test.example.com"
	cfg.EnabledBridges = []string{"whatsapp", "telegram"}

	err = cfg.Save()
	require.NoError(t, err)

	loaded, err := Load()
	require.NoError(t, err)

	assert.Equal(t, cfg.ServerName, loaded.ServerName)
	assert.Equal(t, cfg.ConnectivityMode, loaded.ConnectivityMode)
	assert.Equal(t, cfg.Postgres.Password, loaded.Postgres.Password)
	assert.Equal(t, cfg.Admin.Password, loaded.Admin.Password)
	assert.Equal(t, cfg.EnabledBridges, loaded.EnabledBridges)
}

func TestBridgeOperations(t *testing.T) {
	cfg := &Config{
		EnabledBridges: []string{"whatsapp"},
	}

	assert.True(t, cfg.IsBridgeEnabled("whatsapp"))
	assert.False(t, cfg.IsBridgeEnabled("telegram"))

	cfg.EnableBridge("telegram")
	assert.True(t, cfg.IsBridgeEnabled("telegram"))
	assert.Equal(t, []string{"whatsapp", "telegram"}, cfg.EnabledBridges)

	// Don't add duplicates
	cfg.EnableBridge("telegram")
	assert.Equal(t, []string{"whatsapp", "telegram"}, cfg.EnabledBridges)

	cfg.DisableBridge("whatsapp")
	assert.False(t, cfg.IsBridgeEnabled("whatsapp"))
	assert.Equal(t, []string{"telegram"}, cfg.EnabledBridges)
}

func TestPublicBaseURL(t *testing.T) {
	cfg := &Config{
		ServerName: "localhost",
		HTTPS: HTTPSConfig{
			Enabled: false,
		},
	}
	assert.Equal(t, "http://localhost:8008", cfg.PublicBaseURL())

	cfg.HTTPS.Enabled = true
	cfg.HTTPS.Domain = "chat.example.com"
	assert.Equal(t, "https://chat.example.com", cfg.PublicBaseURL())
}

func TestElementURL(t *testing.T) {
	cfg := &Config{
		ServerName: "localhost",
		HTTPS: HTTPSConfig{
			Enabled: false,
		},
	}
	assert.Equal(t, "http://localhost:8080", cfg.ElementURL())

	cfg.HTTPS.Enabled = true
	cfg.HTTPS.Domain = "chat.example.com"
	assert.Equal(t, "https://chat.example.com", cfg.ElementURL())
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	assert.False(t, Exists())

	cfg, err := NewDefaultConfig()
	require.NoError(t, err)
	err = cfg.Save()
	require.NoError(t, err)

	assert.True(t, Exists())
}

func TestEnsureDirs(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))

	err := EnsureDirs()
	require.NoError(t, err)

	// Check that directories were created
	configDir := ConfigDir()
	dataDir := DataDir()

	assert.DirExists(t, filepath.Join(configDir, "synapse"))
	assert.DirExists(t, filepath.Join(configDir, "element"))
	assert.DirExists(t, filepath.Join(configDir, "caddy"))
	assert.DirExists(t, filepath.Join(configDir, "bridges"))
	assert.DirExists(t, filepath.Join(dataDir, "synapse"))
	assert.DirExists(t, filepath.Join(dataDir, "postgres"))
	assert.DirExists(t, filepath.Join(dataDir, "caddy"))
	assert.DirExists(t, filepath.Join(dataDir, "bridges"))
}

func TestGetOrCreateBridgeTokens(t *testing.T) {
	cfg := &Config{}

	// First call should create new tokens
	tokens1, err := cfg.GetOrCreateBridgeTokens("whatsapp")
	require.NoError(t, err)
	assert.Len(t, tokens1.ASToken, 64)
	assert.Len(t, tokens1.HSToken, 64)

	// Second call should return same tokens
	tokens2, err := cfg.GetOrCreateBridgeTokens("whatsapp")
	require.NoError(t, err)
	assert.Equal(t, tokens1.ASToken, tokens2.ASToken)
	assert.Equal(t, tokens1.HSToken, tokens2.HSToken)

	// Different bridge should get different tokens
	tokens3, err := cfg.GetOrCreateBridgeTokens("telegram")
	require.NoError(t, err)
	assert.NotEqual(t, tokens1.ASToken, tokens3.ASToken)
}

func TestGetOrCreateDoublePuppetTokens(t *testing.T) {
	cfg := &Config{}

	// First call should create new tokens
	tokens1, err := cfg.GetOrCreateDoublePuppetTokens()
	require.NoError(t, err)
	assert.Len(t, tokens1.ASToken, 64)
	assert.Len(t, tokens1.HSToken, 64)

	// Second call should return same tokens
	tokens2, err := cfg.GetOrCreateDoublePuppetTokens()
	require.NoError(t, err)
	assert.Equal(t, tokens1.ASToken, tokens2.ASToken)
	assert.Equal(t, tokens1.HSToken, tokens2.HSToken)
}

func TestBridgeTokensPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg, err := NewDefaultConfig()
	require.NoError(t, err)

	// Create bridge tokens
	tokens, err := cfg.GetOrCreateBridgeTokens("whatsapp")
	require.NoError(t, err)

	// Save and reload
	err = cfg.Save()
	require.NoError(t, err)

	loaded, err := Load()
	require.NoError(t, err)

	// Tokens should persist
	loadedTokens, err := loaded.GetOrCreateBridgeTokens("whatsapp")
	require.NoError(t, err)
	assert.Equal(t, tokens.ASToken, loadedTokens.ASToken)
	assert.Equal(t, tokens.HSToken, loadedTokens.HSToken)
}

func TestSynapsePort_Default(t *testing.T) {
	cfg := &Config{}
	assert.Equal(t, 8008, cfg.SynapsePort())
}

func TestSynapsePort_Custom(t *testing.T) {
	cfg := &Config{
		Ports: PortsConfig{Synapse: 9008},
	}
	assert.Equal(t, 9008, cfg.SynapsePort())
}

func TestElementPort_Default(t *testing.T) {
	cfg := &Config{}
	assert.Equal(t, 8080, cfg.ElementPort())
}

func TestElementPort_Custom(t *testing.T) {
	cfg := &Config{
		Ports: PortsConfig{Element: 9080},
	}
	assert.Equal(t, 9080, cfg.ElementPort())
}

func TestIsElementEnabled_Default(t *testing.T) {
	cfg := &Config{}
	assert.True(t, cfg.IsElementEnabled())
}

func TestIsElementEnabled_ExplicitTrue(t *testing.T) {
	enabled := true
	cfg := &Config{ElementEnabled: &enabled}
	assert.True(t, cfg.IsElementEnabled())
}

func TestIsElementEnabled_ExplicitFalse(t *testing.T) {
	disabled := false
	cfg := &Config{ElementEnabled: &disabled}
	assert.False(t, cfg.IsElementEnabled())
}

func TestPublicBaseURL_WithCustomPort(t *testing.T) {
	cfg := &Config{
		ServerName: "myserver",
		Ports:      PortsConfig{Synapse: 9008},
	}
	assert.Equal(t, "http://myserver:9008", cfg.PublicBaseURL())
}

func TestElementURL_WithCustomPort(t *testing.T) {
	cfg := &Config{
		ServerName: "myserver",
		Ports:      PortsConfig{Element: 9080},
	}
	assert.Equal(t, "http://myserver:9080", cfg.ElementURL())
}

func TestLoadNonexistentConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	_, err := Load()
	assert.Error(t, err)
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create directory and write invalid YAML
	err := os.MkdirAll(filepath.Join(tmpDir, "muxbee"), 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "muxbee", "settings.yaml"), []byte("invalid: [yaml: content"), 0600)
	require.NoError(t, err)

	_, err = Load()
	assert.Error(t, err)
}

func TestDisableBridge_NotEnabled(t *testing.T) {
	cfg := &Config{
		EnabledBridges: []string{"whatsapp"},
	}

	// Disabling a bridge that's not enabled should be a no-op
	cfg.DisableBridge("telegram")
	assert.Equal(t, []string{"whatsapp"}, cfg.EnabledBridges)
}

func TestDisableBridge_Empty(t *testing.T) {
	cfg := &Config{
		EnabledBridges: []string{},
	}

	cfg.DisableBridge("whatsapp")
	assert.Empty(t, cfg.EnabledBridges)
}

func TestBridgeTokens_PartialTokens(t *testing.T) {
	cfg := &Config{
		BridgeTokens: map[string]BridgeTokens{
			"whatsapp": {ASToken: "partial", HSToken: ""},
		},
	}

	// Should regenerate when tokens are incomplete
	tokens, err := cfg.GetOrCreateBridgeTokens("whatsapp")
	require.NoError(t, err)
	assert.Len(t, tokens.ASToken, 64)
	assert.Len(t, tokens.HSToken, 64)
}

func TestDoublePuppetTokens_PartialTokens(t *testing.T) {
	cfg := &Config{
		DoublePuppetTokens: &BridgeTokens{ASToken: "partial", HSToken: ""},
	}

	// Should regenerate when tokens are incomplete
	tokens, err := cfg.GetOrCreateDoublePuppetTokens()
	require.NoError(t, err)
	assert.Len(t, tokens.ASToken, 64)
	assert.Len(t, tokens.HSToken, 64)
}

func TestGeneratePassword_DifferentLengths(t *testing.T) {
	tests := []int{8, 16, 32, 64}
	for _, length := range tests {
		t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
			pass, err := GeneratePassword(length)
			require.NoError(t, err)
			assert.Len(t, pass, length)
		})
	}
}

func TestIsBridgeEnabled_EmptyList(t *testing.T) {
	cfg := &Config{EnabledBridges: []string{}}
	assert.False(t, cfg.IsBridgeEnabled("whatsapp"))
}

func TestIsBridgeEnabled_NilList(t *testing.T) {
	cfg := &Config{}
	assert.False(t, cfg.IsBridgeEnabled("whatsapp"))
}
