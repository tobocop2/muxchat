package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the muxbee settings
type Config struct {
	ServerName         string                  `yaml:"server_name"`
	ConnectivityMode   string                  `yaml:"connectivity_mode"` // local, private, public
	ElementEnabled     *bool                   `yaml:"element_enabled,omitempty"` // nil = true (default)
	Ports              PortsConfig             `yaml:"ports,omitempty"`
	Postgres           PostgresConfig          `yaml:"postgres"`
	Admin              AdminConfig             `yaml:"admin"`
	HTTPS              HTTPSConfig             `yaml:"https"`
	EnabledBridges     []string                `yaml:"enabled_bridges"`
	BridgeTokens       map[string]BridgeTokens `yaml:"bridge_tokens,omitempty"`
	Telegram           *TelegramConfig         `yaml:"telegram,omitempty"`
	DoublePuppetTokens *BridgeTokens           `yaml:"double_puppet_tokens,omitempty"`
}

// PortsConfig holds the ports for services
type PortsConfig struct {
	Synapse int `yaml:"synapse,omitempty"` // Default: 8008
	Element int `yaml:"element,omitempty"` // Default: 8080
}

// SynapsePort returns the Synapse port (default 8008)
func (c *Config) SynapsePort() int {
	if c.Ports.Synapse != 0 {
		return c.Ports.Synapse
	}
	return 8008
}

// ElementPort returns the Element port (default 8080)
func (c *Config) ElementPort() int {
	if c.Ports.Element != 0 {
		return c.Ports.Element
	}
	return 8080
}

// IsPortAvailable checks if a port is available for use
func IsPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// FindAvailablePort finds an available port starting from the given port
func FindAvailablePort(startPort int) int {
	for port := startPort; port < startPort+100; port++ {
		if IsPortAvailable(port) {
			return port
		}
	}
	return startPort // fallback
}

// EnsureAvailablePorts checks default ports and assigns available ones if needed
func (c *Config) EnsureAvailablePorts() bool {
	changed := false

	if c.Ports.Synapse == 0 {
		if !IsPortAvailable(8008) {
			c.Ports.Synapse = FindAvailablePort(8008)
			changed = true
		}
	}

	if c.Ports.Element == 0 {
		defaultElementPort := 8080
		if c.Ports.Synapse == 8080 {
			defaultElementPort = 8081
		}
		if !IsPortAvailable(defaultElementPort) {
			c.Ports.Element = FindAvailablePort(defaultElementPort)
			changed = true
		}
	}

	return changed
}

// IsElementEnabled returns whether Element Web should run (defaults to true)
func (c *Config) IsElementEnabled() bool {
	return c.ElementEnabled == nil || *c.ElementEnabled
}

// TelegramConfig holds Telegram API credentials
type TelegramConfig struct {
	APIID   string `yaml:"api_id"`
	APIHash string `yaml:"api_hash"`
}

// BridgeTokens holds the appservice tokens for a bridge
type BridgeTokens struct {
	ASToken string `yaml:"as_token"`
	HSToken string `yaml:"hs_token"`
}

// PostgresConfig holds database connection settings
type PostgresConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

// AdminConfig holds admin user credentials
type AdminConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// HTTPSConfig holds HTTPS/TLS settings
type HTTPSConfig struct {
	Enabled bool   `yaml:"enabled"`
	Domain  string `yaml:"domain,omitempty"`
	Email   string `yaml:"email,omitempty"`
}

// GeneratePassword creates a cryptographically secure random password
func GeneratePassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// NewDefaultConfig creates a new config with default values and generated passwords
func NewDefaultConfig() (*Config, error) {
	postgresPass, err := GeneratePassword(32)
	if err != nil {
		return nil, err
	}

	adminPass, err := GeneratePassword(16)
	if err != nil {
		return nil, err
	}

	return &Config{
		ServerName:       "localhost",
		ConnectivityMode: "local",
		Postgres: PostgresConfig{
			User:     "synapse",
			Password: postgresPass,
			Database: "synapse",
		},
		Admin: AdminConfig{
			Username: "admin",
			Password: adminPass,
		},
		HTTPS: HTTPSConfig{
			Enabled: false,
		},
		EnabledBridges: []string{},
	}, nil
}

// Load reads the config from the settings file
func Load() (*Config, error) {
	path := SettingsPath()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save writes the config to the settings file
func (c *Config) Save() error {
	if err := EnsureDirs(); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(SettingsPath(), data, 0600)
}

// Exists checks if a config file already exists
func Exists() bool {
	_, err := os.Stat(SettingsPath())
	return err == nil
}

// IsBridgeEnabled checks if a specific bridge is enabled
func (c *Config) IsBridgeEnabled(name string) bool {
	for _, b := range c.EnabledBridges {
		if b == name {
			return true
		}
	}
	return false
}

// EnableBridge adds a bridge to the enabled list
func (c *Config) EnableBridge(name string) {
	if !c.IsBridgeEnabled(name) {
		c.EnabledBridges = append(c.EnabledBridges, name)
	}
}

// DisableBridge removes a bridge from the enabled list
func (c *Config) DisableBridge(name string) {
	bridges := make([]string, 0, len(c.EnabledBridges))
	for _, b := range c.EnabledBridges {
		if b != name {
			bridges = append(bridges, b)
		}
	}
	c.EnabledBridges = bridges
}

// GetOrCreateBridgeTokens returns existing tokens for a bridge or creates new ones
func (c *Config) GetOrCreateBridgeTokens(bridgeName string) (BridgeTokens, error) {
	if c.BridgeTokens == nil {
		c.BridgeTokens = make(map[string]BridgeTokens)
	}

	if tokens, exists := c.BridgeTokens[bridgeName]; exists && tokens.ASToken != "" && tokens.HSToken != "" {
		return tokens, nil
	}

	asToken, err := GeneratePassword(64)
	if err != nil {
		return BridgeTokens{}, err
	}
	hsToken, err := GeneratePassword(64)
	if err != nil {
		return BridgeTokens{}, err
	}

	tokens := BridgeTokens{
		ASToken: asToken,
		HSToken: hsToken,
	}
	c.BridgeTokens[bridgeName] = tokens
	return tokens, nil
}

// GetOrCreateDoublePuppetTokens returns existing tokens or creates new ones for the doublepuppet appservice
func (c *Config) GetOrCreateDoublePuppetTokens() (BridgeTokens, error) {
	if c.DoublePuppetTokens != nil && c.DoublePuppetTokens.ASToken != "" && c.DoublePuppetTokens.HSToken != "" {
		return *c.DoublePuppetTokens, nil
	}

	asToken, err := GeneratePassword(64)
	if err != nil {
		return BridgeTokens{}, err
	}
	hsToken, err := GeneratePassword(64)
	if err != nil {
		return BridgeTokens{}, err
	}

	tokens := BridgeTokens{
		ASToken: asToken,
		HSToken: hsToken,
	}
	c.DoublePuppetTokens = &tokens
	return tokens, nil
}

// PublicBaseURL returns the public URL for the homeserver
func (c *Config) PublicBaseURL() string {
	if c.HTTPS.Enabled && c.HTTPS.Domain != "" {
		return "https://" + c.HTTPS.Domain
	}
	return fmt.Sprintf("http://%s:%d", c.ServerName, c.SynapsePort())
}

// ElementURL returns the URL to access Element Web
func (c *Config) ElementURL() string {
	if c.HTTPS.Enabled && c.HTTPS.Domain != "" {
		return "https://" + c.HTTPS.Domain
	}
	return fmt.Sprintf("http://%s:%d", c.ServerName, c.ElementPort())
}
