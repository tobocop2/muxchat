package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tobias/muxbee/internal/config"
	"github.com/tobias/muxbee/internal/docker"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop muxbee services",
	Long:  `Stop all running muxbee services.`,
	RunE:  runDown,
}

func init() {
	rootCmd.AddCommand(downCmd)
}

func runDown(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nRun 'muxbee init' first", err)
	}

	compose := docker.New(cfg)
	profiles := docker.GetProfiles(cfg)

	if err := compose.Down(profiles); err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}

	return nil
}
