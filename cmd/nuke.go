package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tobias/muxchat/internal/config"
	"github.com/tobias/muxchat/internal/docker"
)

var nukeCmd = &cobra.Command{
	Use:   "nuke",
	Short: "Remove all muxchat data and configuration",
	Long: `Completely remove all muxchat data and configuration.

This will:
  - Stop all running services
  - Remove Docker volumes
  - Delete configuration directory
  - Delete data directory

THIS ACTION IS IRREVERSIBLE. All your messages, media, and settings will be lost.`,
	RunE: runNuke,
}

var nukeYes bool

func init() {
	rootCmd.AddCommand(nukeCmd)
	nukeCmd.Flags().BoolVarP(&nukeYes, "yes", "y", false, "Skip confirmation")
}

func runNuke(cmd *cobra.Command, args []string) error {
	configDir := config.ConfigDir()
	dataDir := config.DataDir()

	fmt.Println("WARNING: This will permanently delete ALL muxchat data!")
	fmt.Println()
	fmt.Printf("  Config: %s\n", configDir)
	fmt.Printf("  Data:   %s\n", dataDir)
	fmt.Println()

	if !nukeYes {
		fmt.Print("Type 'yes' to confirm: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	fmt.Println()

	if config.Exists() {
		cfg, err := config.Load()
		if err == nil {
			compose := docker.New(cfg)
			profiles := docker.GetProfiles(cfg)

			compose.Down(profiles)
			compose.DownVolumes()
		}
	}

	fmt.Printf("Removing %s...\n", configDir)
	if err := os.RemoveAll(configDir); err != nil {
		fmt.Printf("Warning: failed to remove config dir: %v\n", err)
	}

	fmt.Printf("Removing %s...\n", dataDir)
	if err := os.RemoveAll(dataDir); err != nil {
		fmt.Printf("Warning: failed to remove data dir: %v\n", err)
	}

	fmt.Println()
	fmt.Println("Mautrix Chat has been completely removed.")
	fmt.Println("Run 'muxchat init' to start fresh.")

	return nil
}
