package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tobias/muxchat/internal/config"
	"github.com/tobias/muxchat/internal/docker"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of muxchat services",
	Long:  `Display the current status of all muxchat services.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nRun 'muxchat init' first", err)
	}

	compose := docker.New(cfg)
	statuses, err := compose.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if len(statuses) == 0 {
		fmt.Println("No services running.")
		fmt.Println("Run 'muxchat up' to start services.")
		return nil
	}

	fmt.Println("Service Status:")
	fmt.Println()

	for _, s := range statuses {
		status := "stopped"
		if s.Running {
			status = "running"
			if s.Health != "" && s.Health != "healthy" {
				status = fmt.Sprintf("running (%s)", s.Health)
			}
		}

		serviceName := docker.ParseServiceName(s.Name)
		fmt.Printf("  %-20s %s\n", serviceName, status)
	}

	fmt.Println()

	if compose.IsRunning() {
		if cfg.IsElementEnabled() {
			fmt.Printf("Element Web: %s\n", cfg.ElementURL())
		}
		fmt.Printf("Matrix API:  %s\n", cfg.PublicBaseURL())
	}

	return nil
}
