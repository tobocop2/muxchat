package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tobocop2/muxbee/internal/config"
	"github.com/tobocop2/muxbee/internal/docker"
	"github.com/tobocop2/muxbee/internal/generator"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize muxbee configuration",
	Long: `Initialize muxbee by creating configuration files and setting up
the Matrix homeserver environment.

This command creates:
  - Configuration directory structure
  - Synapse homeserver configuration
  - Element Web configuration
  - Docker Compose file`,
	RunE: runInit,
}

var (
	initServerName string
	initHTTPS      bool
	initDomain     string
	initEmail      string
	initForce      bool
	initNoElement  bool
)

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVar(&initServerName, "server-name", "localhost", "Matrix server name")
	initCmd.Flags().BoolVar(&initHTTPS, "https", false, "Enable HTTPS with Caddy")
	initCmd.Flags().StringVar(&initDomain, "domain", "", "Domain name for HTTPS")
	initCmd.Flags().StringVar(&initEmail, "email", "", "Email for Let's Encrypt certificates")
	initCmd.Flags().BoolVar(&initForce, "force", false, "Overwrite existing configuration")
	initCmd.Flags().BoolVar(&initNoElement, "no-element", false, "Don't run Element Web (use your own Matrix client)")
}

func runInit(cmd *cobra.Command, args []string) error {
	if config.Exists() && !initForce {
		return fmt.Errorf("configuration already exists, use --force to overwrite")
	}

	if err := docker.DockerAvailable(); err != nil {
		return fmt.Errorf("Docker is not available: %w\nPlease install Docker and ensure it is running", err)
	}

	connectivityMode := "local"
	if initHTTPS {
		connectivityMode = "public"
		if initDomain == "" {
			return fmt.Errorf("--domain is required when using --https")
		}
		if initEmail == "" {
			return fmt.Errorf("--email is required when using --https")
		}
		initServerName = initDomain
	}

	fmt.Println("Initializing muxbee...")

	cfg, err := config.NewDefaultConfig()
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	cfg.ServerName = initServerName
	cfg.ConnectivityMode = connectivityMode
	cfg.HTTPS.Enabled = initHTTPS
	cfg.HTTPS.Domain = initDomain
	cfg.HTTPS.Email = initEmail

	if initNoElement {
		elementEnabled := false
		cfg.ElementEnabled = &elementEnabled
	}

	if cfg.EnsureAvailablePorts() {
		fmt.Println("Note: Default ports were in use, using alternative ports.")
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	gen := generator.New()
	if err := gen.GenerateAll(cfg); err != nil {
		return fmt.Errorf("failed to generate configs: %w", err)
	}

	compose := docker.New(cfg)
	if err := compose.WriteComposeFile(); err != nil {
		return fmt.Errorf("failed to write docker-compose.yml: %w", err)
	}

	fmt.Println()
	fmt.Println("Configuration created successfully!")
	fmt.Println()
	fmt.Printf("  Config directory: %s\n", config.ConfigDir())
	fmt.Printf("  Data directory:   %s\n", config.DataDir())
	fmt.Println()
	fmt.Printf("  Admin username:   %s\n", cfg.Admin.Username)
	fmt.Printf("  Admin password:   %s\n", cfg.Admin.Password)
	fmt.Println()

	if cfg.IsElementEnabled() {
		fmt.Printf("  Element URL:      %s\n", cfg.ElementURL())
	} else {
		fmt.Println("  Element:          disabled (use your own Matrix client)")
	}
	fmt.Printf("  Synapse URL:      %s\n", cfg.PublicBaseURL())
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Enable bridges:  muxbee bridge enable whatsapp")
	fmt.Println("  2. Start services:  muxbee up")
	if cfg.IsElementEnabled() {
		fmt.Println("  3. Open Element:    muxbee open")
	} else {
		fmt.Printf("  3. Connect your Matrix client to %s\n", cfg.PublicBaseURL())
	}

	return nil
}
