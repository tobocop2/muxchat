package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tobias/muxbee/internal/config"
	"github.com/tobias/muxbee/internal/docker"
	"github.com/tobias/muxbee/internal/generator"
	"github.com/tobias/muxbee/internal/matrix"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start muxbee services",
	Long: `Start all muxbee services including the Matrix homeserver,
Element Web client, and any enabled bridges.`,
	RunE: runUp,
}

var upPull bool

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().BoolVar(&upPull, "pull", false, "Pull latest images before starting")
}

func runUp(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nRun 'muxbee init' first", err)
	}

	gen := generator.New()
	if err := gen.GenerateAll(cfg); err != nil {
		return fmt.Errorf("failed to generate configs: %w", err)
	}

	compose := docker.New(cfg)
	if err := compose.WriteComposeFile(); err != nil {
		return fmt.Errorf("failed to write docker-compose.yml: %w", err)
	}

	profiles := docker.GetProfiles(cfg)

	if upPull {
		if err := compose.Pull(profiles); err != nil {
			return fmt.Errorf("failed to pull images: %w", err)
		}
	}

	if err := compose.Up(profiles); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}
	fmt.Println()

	if cfg.HTTPS.Enabled {
		fmt.Printf("  Element Web: https://%s\n", cfg.HTTPS.Domain)
		fmt.Printf("  Matrix API:  https://%s/_matrix/\n", cfg.HTTPS.Domain)
	} else {
		fmt.Printf("  Element Web: %s\n", cfg.ElementURL())
		fmt.Printf("  Matrix API:  %s\n", cfg.PublicBaseURL())
	}
	fmt.Println()

	if err := setupAdminUser(cfg); err != nil {
		fmt.Printf("Note: Could not create admin user: %v\n", err)
	}

	if len(cfg.EnabledBridges) > 0 {
		fmt.Println("Waiting for bridges to start...")
		compose.WaitForBridges(cfg.EnabledBridges, 30)

		fmt.Println("Setting up bridge bot conversations...")
		if err := matrix.SetupBotsForUser(cfg); err != nil {
			fmt.Printf("Note: Bot setup incomplete: %v\n", err)
		}
		fmt.Println()
	}

	fmt.Println("Run 'muxbee open' to launch Element in your browser.")
	fmt.Println("Run 'muxbee status' to check service status.")

	return nil
}

func setupAdminUser(cfg *config.Config) error {
	markerFile := filepath.Join(config.DataDir(), ".admin_setup_done")
	if fileExists(markerFile) {
		return nil
	}

	fmt.Println("Setting up admin user...")
	time.Sleep(3 * time.Second) // Wait for Synapse to be ready

	cmd := exec.Command("docker", "exec", "muxbee-synapse-1",
		"register_new_matrix_user",
		"-u", cfg.Admin.Username,
		"-p", cfg.Admin.Password,
		"-a",
		"-c", "/data/homeserver.yaml",
		"http://localhost:8008",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outStr := string(output)
		if !strings.Contains(outStr, "already taken") && !strings.Contains(outStr, "already exists") {
			return fmt.Errorf("failed to create admin user: %s", outStr)
		}
	}

	fmt.Printf("  Admin user: %s\n", cfg.Admin.Username)
	fmt.Printf("  Password:   %s\n", cfg.Admin.Password)
	fmt.Println()

	_ = writeMarkerFile(markerFile)
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func writeMarkerFile(path string) error {
	return os.WriteFile(path, []byte("done"), 0644)
}
