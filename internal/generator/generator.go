package generator

import (
	"embed"
	"os"
	"path/filepath"
	"text/template"

	"github.com/tobias/muxchat/internal/bridges"
	"github.com/tobias/muxchat/internal/config"
)

//go:embed templates/*
var templateFS embed.FS

// SynapseData contains data for Synapse homeserver template
type SynapseData struct {
	ServerName         string
	PublicBaseURL      string
	HTTPS              bool
	Postgres           config.PostgresConfig
	RegistrationSecret string
	DoublePuppetSecret string
	Bridges            []string
}

// ElementData contains data for Element Web config template
type ElementData struct {
	HomeserverURL string
	ServerName    string
}

// CaddyData contains data for Caddy reverse proxy template
type CaddyData struct {
	Domain string
	Email  string
}

// BridgeConfigData contains data for bridge config templates
type BridgeConfigData struct {
	ServerName         string
	Port               int
	AdminUser          string
	ASToken            string
	HSToken            string
	BotUsername        string
	TelegramAPIID      string // Only used for telegram bridge
	TelegramAPIHash    string // Only used for telegram bridge
	DoublePuppetSecret string // Shared secret for double puppeting
}

// BridgeRegistrationData contains data for bridge registration template
type BridgeRegistrationData struct {
	Name            string
	Port            int
	ASToken         string
	HSToken         string
	BotUsername     string
	NamespacePrefix string
	ServerName      string
}

// DoublePuppetRegistrationData contains data for the doublepuppet appservice registration
type DoublePuppetRegistrationData struct {
	ASToken    string
	HSToken    string
	ServerName string
}

// Generator handles template rendering
type Generator struct {
	configDir string
	dataDir   string
}

// New creates a new Generator
func New() *Generator {
	return &Generator{
		configDir: config.ConfigDir(),
		dataDir:   config.DataDir(),
	}
}

// GenerateSynapse generates the Synapse homeserver configuration
func (g *Generator) GenerateSynapse(data SynapseData) error {
	tmpl, err := template.ParseFS(templateFS, "templates/synapse/homeserver.yaml.tmpl")
	if err != nil {
		return err
	}

	dir := filepath.Join(g.configDir, "synapse")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(dir, "homeserver.yaml"))
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

// GenerateSynapseLogConfig generates the Synapse logging configuration
func (g *Generator) GenerateSynapseLogConfig() error {
	tmpl, err := template.ParseFS(templateFS, "templates/synapse/log.config.tmpl")
	if err != nil {
		return err
	}

	dir := filepath.Join(g.configDir, "synapse")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(dir, "log.config"))
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, nil)
}

// GenerateElement generates the Element Web configuration
func (g *Generator) GenerateElement(data ElementData) error {
	tmpl, err := template.ParseFS(templateFS, "templates/element/config.json.tmpl")
	if err != nil {
		return err
	}

	dir := filepath.Join(g.configDir, "element")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(dir, "config.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

// GenerateCaddy generates the Caddy reverse proxy configuration
func (g *Generator) GenerateCaddy(data CaddyData) error {
	tmpl, err := template.ParseFS(templateFS, "templates/caddy/Caddyfile.tmpl")
	if err != nil {
		return err
	}

	dir := filepath.Join(g.configDir, "caddy")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(dir, "Caddyfile"))
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

// GenerateBridgeConfig generates configuration for a specific bridge
// Writes to data directory since bridges expect writable /data with config.yaml inside
func (g *Generator) GenerateBridgeConfig(bridgeName string, data BridgeConfigData) error {
	templatePath := "templates/bridges/" + bridgeName + ".yaml.tmpl"
	tmpl, err := template.ParseFS(templateFS, templatePath)
	if err != nil {
		return err
	}

	// Write to data directory, not config directory
	dir := filepath.Join(g.dataDir, "bridges", bridgeName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(dir, "config.yaml"))
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

// GenerateDoublePuppetRegistration generates the doublepuppet appservice registration for Synapse
func (g *Generator) GenerateDoublePuppetRegistration(data DoublePuppetRegistrationData) error {
	tmpl, err := template.ParseFS(templateFS, "templates/synapse/doublepuppet-registration.yaml.tmpl")
	if err != nil {
		return err
	}

	// Write to synapse data dir (mounted at /data in container)
	dir := filepath.Join(g.dataDir, "synapse")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(dir, "doublepuppet-registration.yaml"))
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

// GenerateBridgeRegistration generates the appservice registration for a bridge
// Writes to both CONFIG_DIR (for Synapse) and DATA_DIR (for the bridge to use)
func (g *Generator) GenerateBridgeRegistration(data BridgeRegistrationData) error {
	tmpl, err := template.ParseFS(templateFS, "templates/bridges/registration.yaml.tmpl")
	if err != nil {
		return err
	}

	// Write to config dir for Synapse
	configBridgeDir := filepath.Join(g.configDir, "bridges", data.Name)
	if err := os.MkdirAll(configBridgeDir, 0755); err != nil {
		return err
	}

	f1, err := os.Create(filepath.Join(configBridgeDir, "registration.yaml"))
	if err != nil {
		return err
	}
	if err := tmpl.Execute(f1, data); err != nil {
		f1.Close()
		return err
	}
	f1.Close()

	// Also write to data dir for the bridge to find
	dataBridgeDir := filepath.Join(g.dataDir, "bridges", data.Name)
	if err := os.MkdirAll(dataBridgeDir, 0755); err != nil {
		return err
	}

	f2, err := os.Create(filepath.Join(dataBridgeDir, "registration.yaml"))
	if err != nil {
		return err
	}
	defer f2.Close()

	// Re-parse template for second write
	tmpl2, err := template.ParseFS(templateFS, "templates/bridges/registration.yaml.tmpl")
	if err != nil {
		return err
	}
	return tmpl2.Execute(f2, data)
}

// GenerateAll generates all configuration files based on the given config
func (g *Generator) GenerateAll(cfg *config.Config) error {
	// Ensure directories exist
	if err := config.EnsureDirs(); err != nil {
		return err
	}

	// Generate registration secret
	regSecret, err := config.GeneratePassword(64)
	if err != nil {
		return err
	}

	// Get or create double puppet appservice tokens (persisted in config)
	var oldDoublePuppetASToken string
	if cfg.DoublePuppetTokens != nil {
		oldDoublePuppetASToken = cfg.DoublePuppetTokens.ASToken
	}
	doublePuppetTokens, err := cfg.GetOrCreateDoublePuppetTokens()
	if err != nil {
		return err
	}
	doublePuppetTokensChanged := oldDoublePuppetASToken != doublePuppetTokens.ASToken

	// Generate doublepuppet appservice registration (for Synapse)
	doublePuppetRegData := DoublePuppetRegistrationData{
		ASToken:    doublePuppetTokens.ASToken,
		HSToken:    doublePuppetTokens.HSToken,
		ServerName: cfg.ServerName,
	}
	if err := g.GenerateDoublePuppetRegistration(doublePuppetRegData); err != nil {
		return err
	}

	// Format the double puppet secret for bridges: "as_token:TOKEN"
	doublePuppetSecret := "as_token:" + doublePuppetTokens.ASToken

	// Generate Synapse config
	synapseData := SynapseData{
		ServerName:         cfg.ServerName,
		PublicBaseURL:      cfg.PublicBaseURL(),
		HTTPS:              cfg.HTTPS.Enabled,
		Postgres:           cfg.Postgres,
		RegistrationSecret: regSecret,
		DoublePuppetSecret: doublePuppetSecret,
		Bridges:            cfg.EnabledBridges,
	}
	if err := g.GenerateSynapse(synapseData); err != nil {
		return err
	}

	// Generate Synapse log config
	if err := g.GenerateSynapseLogConfig(); err != nil {
		return err
	}

	// Generate Element config
	elementData := ElementData{
		HomeserverURL: cfg.PublicBaseURL(),
		ServerName:    cfg.ServerName,
	}
	if err := g.GenerateElement(elementData); err != nil {
		return err
	}

	// Generate Caddy config if HTTPS is enabled
	if cfg.HTTPS.Enabled {
		caddyData := CaddyData{
			Domain: cfg.HTTPS.Domain,
			Email:  cfg.HTTPS.Email,
		}
		if err := g.GenerateCaddy(caddyData); err != nil {
			return err
		}
	}

	// Track if we need to save config (new tokens were generated)
	configChanged := false

	// Generate bridge configs
	for _, bridgeName := range cfg.EnabledBridges {
		bridge := bridges.Get(bridgeName)
		if bridge == nil {
			continue
		}

		// Get existing tokens or generate new ones (persisted in config)
		var oldASToken string
		if cfg.BridgeTokens != nil {
			oldASToken = cfg.BridgeTokens[bridgeName].ASToken
		}
		tokens, err := cfg.GetOrCreateBridgeTokens(bridgeName)
		if err != nil {
			return err
		}
		if oldASToken != tokens.ASToken {
			configChanged = true
		}

		// Generate bridge config (with tokens)
		bridgeConfigData := BridgeConfigData{
			ServerName:         cfg.ServerName,
			Port:               bridge.Port,
			AdminUser:          cfg.Admin.Username,
			ASToken:            tokens.ASToken,
			HSToken:            tokens.HSToken,
			BotUsername:        bridge.BotUsername(),
			DoublePuppetSecret: doublePuppetSecret,
		}

		// Add telegram-specific credentials if available
		if bridgeName == "telegram" && cfg.Telegram != nil {
			bridgeConfigData.TelegramAPIID = cfg.Telegram.APIID
			bridgeConfigData.TelegramAPIHash = cfg.Telegram.APIHash
		}

		if err := g.GenerateBridgeConfig(bridgeName, bridgeConfigData); err != nil {
			return err
		}

		// Generate registration (with same tokens)
		regData := BridgeRegistrationData{
			Name:            bridgeName,
			Port:            bridge.Port,
			ASToken:         tokens.ASToken,
			HSToken:         tokens.HSToken,
			BotUsername:     bridge.BotUsername(),
			NamespacePrefix: bridge.NamespacePrefix(),
			ServerName:      cfg.ServerName,
		}
		if err := g.GenerateBridgeRegistration(regData); err != nil {
			return err
		}

		// Ensure bridge data directory exists
		bridgeDataDir := filepath.Join(g.dataDir, "bridges", bridgeName)
		if err := os.MkdirAll(bridgeDataDir, 0755); err != nil {
			return err
		}
	}

	// Save config if tokens were generated
	if configChanged || doublePuppetTokensChanged {
		if err := cfg.Save(); err != nil {
			return err
		}
	}

	return nil
}
