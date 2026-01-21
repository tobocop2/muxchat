package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tobocop2/muxbee/internal/config"
	"github.com/tobocop2/muxbee/internal/docker"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update all services to latest versions",
	Long: `Pull the latest Docker images for all services and restart them.
This is useful when new bridge versions are released.`,
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	compose := docker.New(cfg)
	profiles := docker.GetProfiles(cfg)

	// Show what will be updated
	services := []string{"synapse", "postgres"}
	if cfg.IsElementEnabled() {
		services = append(services, "element")
	}
	if cfg.HTTPS.Enabled {
		services = append(services, "caddy")
	}
	for _, bridge := range cfg.EnabledBridges {
		services = append(services, "mautrix-"+bridge)
	}

	fmt.Println("Updating services:", strings.Join(services, ", "))
	fmt.Println()

	// Pull images
	fmt.Println("==> Pulling latest images...")
	if err := compose.Pull(profiles); err != nil {
		return fmt.Errorf("failed to pull images: %w", err)
	}
	fmt.Println()

	// Stop services
	fmt.Println("==> Stopping services...")
	if err := compose.Down(profiles); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}
	fmt.Println()

	// Start with force recreate to ensure new images are used
	fmt.Println("==> Starting services with updated images...")
	if err := compose.UpForceRecreate(profiles); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}
	fmt.Println()

	fmt.Println("Update complete!")
	return nil
}
