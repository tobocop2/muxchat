package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tobias/muxbee/internal/config"
)

func setupTestEnv(t *testing.T) string {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))
	return tmpDir
}

func TestGenerateSynapse(t *testing.T) {
	setupTestEnv(t)

	gen := New()
	data := SynapseData{
		ServerName:         "test.example.com",
		PublicBaseURL:      "https://test.example.com",
		HTTPS:              true,
		Postgres:           config.PostgresConfig{User: "synapse", Password: "testpass", Database: "synapse"},
		RegistrationSecret: "secret123",
		Bridges:            []string{"whatsapp", "telegram"},
	}

	err := gen.GenerateSynapse(data)
	require.NoError(t, err)

	// Verify file was created
	configDir := config.ConfigDir()
	content, err := os.ReadFile(filepath.Join(configDir, "synapse", "homeserver.yaml"))
	require.NoError(t, err)

	assert.Contains(t, string(content), "server_name: \"test.example.com\"")
	assert.Contains(t, string(content), "public_baseurl: \"https://test.example.com\"")
	assert.Contains(t, string(content), "x_forwarded: true")
	assert.Contains(t, string(content), "password: testpass")
	assert.Contains(t, string(content), "/bridges/whatsapp/registration.yaml")
	assert.Contains(t, string(content), "/bridges/telegram/registration.yaml")
}

func TestGenerateSynapseLogConfig(t *testing.T) {
	setupTestEnv(t)

	gen := New()
	err := gen.GenerateSynapseLogConfig()
	require.NoError(t, err)

	configDir := config.ConfigDir()
	content, err := os.ReadFile(filepath.Join(configDir, "synapse", "log.config"))
	require.NoError(t, err)

	assert.Contains(t, string(content), "version: 1")
	assert.Contains(t, string(content), "handlers:")
}

func TestGenerateElement(t *testing.T) {
	setupTestEnv(t)

	gen := New()
	data := ElementData{
		HomeserverURL: "http://localhost:8008",
		ServerName:    "localhost",
	}

	err := gen.GenerateElement(data)
	require.NoError(t, err)

	configDir := config.ConfigDir()
	content, err := os.ReadFile(filepath.Join(configDir, "element", "config.json"))
	require.NoError(t, err)

	assert.Contains(t, string(content), "http://localhost:8008")
	assert.Contains(t, string(content), "\"server_name\": \"localhost\"")
	assert.Contains(t, string(content), "\"brand\": \"Mux Chat\"")
}

func TestGenerateCaddy(t *testing.T) {
	setupTestEnv(t)

	gen := New()
	data := CaddyData{
		Domain: "chat.example.com",
		Email:  "admin@example.com",
	}

	err := gen.GenerateCaddy(data)
	require.NoError(t, err)

	configDir := config.ConfigDir()
	content, err := os.ReadFile(filepath.Join(configDir, "caddy", "Caddyfile"))
	require.NoError(t, err)

	assert.Contains(t, string(content), "chat.example.com")
	assert.Contains(t, string(content), "reverse_proxy /_matrix/* synapse:8008")
	assert.Contains(t, string(content), "tls admin@example.com")
}

func TestGenerateBridgeConfig(t *testing.T) {
	setupTestEnv(t)

	gen := New()
	data := BridgeConfigData{
		ServerName: "localhost",
		Port:       29318,
		AdminUser:  "admin",
	}

	bridgeNames := []string{"whatsapp", "telegram", "signal", "gmessages", "discord", "slack", "meta", "twitter", "bluesky", "linkedin", "googlechat", "gvoice", "irc"}

	for _, name := range bridgeNames {
		t.Run(name, func(t *testing.T) {
			err := gen.GenerateBridgeConfig(name, data)
			require.NoError(t, err)

			// Bridge configs are written to data directory
			dataDir := config.DataDir()
			content, err := os.ReadFile(filepath.Join(dataDir, "bridges", name, "config.yaml"))
			require.NoError(t, err)

			assert.Contains(t, string(content), "domain: localhost")
			assert.Contains(t, string(content), "@admin:localhost")
		})
	}
}

func TestGenerateBridgeRegistration(t *testing.T) {
	setupTestEnv(t)

	gen := New()
	data := BridgeRegistrationData{
		Name:            "whatsapp",
		Port:            29318,
		ASToken:         "as_token_123",
		HSToken:         "hs_token_456",
		BotUsername:     "whatsappbot",
		NamespacePrefix: "whatsapp_",
		ServerName:      "localhost",
	}

	err := gen.GenerateBridgeRegistration(data)
	require.NoError(t, err)

	configDir := config.ConfigDir()
	content, err := os.ReadFile(filepath.Join(configDir, "bridges", "whatsapp", "registration.yaml"))
	require.NoError(t, err)

	assert.Contains(t, string(content), "id: whatsapp")
	assert.Contains(t, string(content), "url: http://mautrix-whatsapp:29318")
	assert.Contains(t, string(content), "as_token: as_token_123")
	assert.Contains(t, string(content), "hs_token: hs_token_456")
	assert.Contains(t, string(content), "sender_localpart: whatsappbot")
	assert.Contains(t, string(content), "@whatsapp_.*:localhost")
}

func TestGenerateAll(t *testing.T) {
	setupTestEnv(t)

	cfg := &config.Config{
		ServerName:       "test.local",
		ConnectivityMode: "local",
		Postgres: config.PostgresConfig{
			User:     "synapse",
			Password: "dbpass123",
			Database: "synapse",
		},
		Admin: config.AdminConfig{
			Username: "admin",
			Password: "adminpass",
		},
		HTTPS: config.HTTPSConfig{
			Enabled: false,
		},
		EnabledBridges: []string{"whatsapp"},
	}

	gen := New()
	err := gen.GenerateAll(cfg)
	require.NoError(t, err)

	configDir := config.ConfigDir()
	dataDir := config.DataDir()

	// Check Synapse files
	assert.FileExists(t, filepath.Join(configDir, "synapse", "homeserver.yaml"))
	assert.FileExists(t, filepath.Join(configDir, "synapse", "log.config"))

	// Check Element file
	assert.FileExists(t, filepath.Join(configDir, "element", "config.json"))

	// Check bridge config file (in data dir)
	assert.FileExists(t, filepath.Join(dataDir, "bridges", "whatsapp", "config.yaml"))
	// Check bridge registration file (in config dir for Synapse)
	assert.FileExists(t, filepath.Join(configDir, "bridges", "whatsapp", "registration.yaml"))

	// Check data directories
	assert.DirExists(t, filepath.Join(dataDir, "bridges", "whatsapp"))

	// Caddy should not be generated when HTTPS is disabled
	_, err = os.Stat(filepath.Join(configDir, "caddy", "Caddyfile"))
	assert.True(t, os.IsNotExist(err))
}

func TestGenerateAllWithHTTPS(t *testing.T) {
	setupTestEnv(t)

	cfg := &config.Config{
		ServerName:       "chat.example.com",
		ConnectivityMode: "public",
		Postgres: config.PostgresConfig{
			User:     "synapse",
			Password: "dbpass123",
			Database: "synapse",
		},
		Admin: config.AdminConfig{
			Username: "admin",
			Password: "adminpass",
		},
		HTTPS: config.HTTPSConfig{
			Enabled: true,
			Domain:  "chat.example.com",
			Email:   "admin@example.com",
		},
		EnabledBridges: []string{},
	}

	gen := New()
	err := gen.GenerateAll(cfg)
	require.NoError(t, err)

	configDir := config.ConfigDir()

	// Check Caddy file exists
	assert.FileExists(t, filepath.Join(configDir, "caddy", "Caddyfile"))

	content, err := os.ReadFile(filepath.Join(configDir, "caddy", "Caddyfile"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "chat.example.com")
}

func TestTemplatesEmbedded(t *testing.T) {
	// Verify all required templates are embedded
	entries, err := templateFS.ReadDir("templates")
	require.NoError(t, err)

	dirs := make([]string, 0)
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}

	assert.Contains(t, dirs, "synapse")
	assert.Contains(t, dirs, "element")
	assert.Contains(t, dirs, "caddy")
	assert.Contains(t, dirs, "bridges")

	// Check synapse templates
	synapseFiles, _ := templateFS.ReadDir("templates/synapse")
	synapseNames := make([]string, len(synapseFiles))
	for i, f := range synapseFiles {
		synapseNames[i] = f.Name()
	}
	assert.Contains(t, synapseNames, "homeserver.yaml.tmpl")
	assert.Contains(t, synapseNames, "log.config.tmpl")

	// Check bridge templates
	bridgeFiles, _ := templateFS.ReadDir("templates/bridges")
	bridgeNames := make([]string, len(bridgeFiles))
	for i, f := range bridgeFiles {
		bridgeNames[i] = f.Name()
	}
	assert.Contains(t, bridgeNames, "registration.yaml.tmpl")
	assert.Contains(t, bridgeNames, "whatsapp.yaml.tmpl")
	assert.Contains(t, bridgeNames, "telegram.yaml.tmpl")
	assert.Contains(t, bridgeNames, "signal.yaml.tmpl")
	assert.Contains(t, bridgeNames, "gmessages.yaml.tmpl")
	assert.Contains(t, bridgeNames, "discord.yaml.tmpl")
	assert.Contains(t, bridgeNames, "slack.yaml.tmpl")
	assert.Contains(t, bridgeNames, "meta.yaml.tmpl")
	assert.Contains(t, bridgeNames, "twitter.yaml.tmpl")
	assert.Contains(t, bridgeNames, "bluesky.yaml.tmpl")
	assert.Contains(t, bridgeNames, "linkedin.yaml.tmpl")
	assert.Contains(t, bridgeNames, "googlechat.yaml.tmpl")
	assert.Contains(t, bridgeNames, "gvoice.yaml.tmpl")
	assert.Contains(t, bridgeNames, "irc.yaml.tmpl")
}

func TestSynapseNoBridges(t *testing.T) {
	setupTestEnv(t)

	gen := New()
	data := SynapseData{
		ServerName:         "localhost",
		PublicBaseURL:      "http://localhost:8008",
		HTTPS:              false,
		Postgres:           config.PostgresConfig{User: "synapse", Password: "pass", Database: "synapse"},
		RegistrationSecret: "secret",
		Bridges:            []string{},
	}

	err := gen.GenerateSynapse(data)
	require.NoError(t, err)

	configDir := config.ConfigDir()
	content, err := os.ReadFile(filepath.Join(configDir, "synapse", "homeserver.yaml"))
	require.NoError(t, err)

	// Should have empty array for app_service_config_files
	assert.True(t, strings.Contains(string(content), "[]") || !strings.Contains(string(content), "/bridges/"))
}
