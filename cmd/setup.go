package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tobias/muxbee/internal/config"
	"github.com/tobias/muxbee/internal/matrix"
)

var setupBotsCmd = &cobra.Command{
	Use:   "setup-bots",
	Short: "Create DM rooms with all enabled bridge bots",
	Long: `Creates direct message rooms with all your enabled bridge bots
and sends welcome messages with login instructions.

Run this after enabling bridges to get started quickly without
having to manually find and message each bot.`,
	RunE: runSetupBots,
}

func init() {
	rootCmd.AddCommand(setupBotsCmd)
}

func runSetupBots(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nRun 'muxbee init' first", err)
	}

	if len(cfg.EnabledBridges) == 0 {
		fmt.Println("No bridges enabled.")
		fmt.Println("Enable bridges with: muxbee bridge enable <name>")
		return nil
	}

	fmt.Println("Setting up bridge bot conversations...")
	fmt.Println()

	// Give a moment for any recently started services
	time.Sleep(2 * time.Second)

	if err := matrix.SetupBotsForUser(cfg); err != nil {
		return fmt.Errorf("failed to setup bots: %w", err)
	}

	fmt.Println()
	fmt.Println("Done! Check Element for your bridge bot conversations.")
	fmt.Println("Each bot has instructions for linking your account.")

	return nil
}
