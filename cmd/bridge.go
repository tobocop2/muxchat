package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tobocop2/muxbee/internal/bridges"
	"github.com/tobocop2/muxbee/internal/config"
	"github.com/tobocop2/muxbee/internal/docker"
	"github.com/tobocop2/muxbee/internal/generator"
)

var bridgeCmd = &cobra.Command{
	Use:   "bridge",
	Short: "Manage messaging bridges",
	Long:  `Manage messaging bridges for WhatsApp, Telegram, Signal, and more.`,
}

var bridgeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available bridges",
	Long:  `List all available messaging bridges and their status.`,
	RunE:  runBridgeList,
}

var bridgeEnableCmd = &cobra.Command{
	Use:   "enable <bridge>",
	Short: "Enable a messaging bridge",
	Long: `Enable a messaging bridge.

Run 'muxbee bridge list' to see available bridges.`,
	Args: cobra.ExactArgs(1),
	RunE: runBridgeEnable,
}

var bridgeDisableCmd = &cobra.Command{
	Use:   "disable <bridge>",
	Short: "Disable a messaging bridge",
	Long:  `Disable an enabled messaging bridge.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runBridgeDisable,
}

var bridgeLoginCmd = &cobra.Command{
	Use:   "login <bridge>",
	Short: "Show login instructions for a bridge",
	Long:  `Display login instructions for a specific bridge.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runBridgeLogin,
}

func init() {
	rootCmd.AddCommand(bridgeCmd)
	bridgeCmd.AddCommand(bridgeListCmd)
	bridgeCmd.AddCommand(bridgeEnableCmd)
	bridgeCmd.AddCommand(bridgeDisableCmd)
	bridgeCmd.AddCommand(bridgeLoginCmd)
}

func runBridgeList(cmd *cobra.Command, args []string) error {
	var enabledBridges []string
	if cfg, err := config.Load(); err == nil {
		enabledBridges = cfg.EnabledBridges
	}

	isEnabled := func(name string) bool {
		for _, b := range enabledBridges {
			if b == name {
				return true
			}
		}
		return false
	}

	fmt.Println("Available Bridges:")
	fmt.Println()

	for _, b := range bridges.List() {
		status := ""
		if isEnabled(b.Name) {
			status = " [enabled]"
		}
		fmt.Printf("  %-12s %s%s\n", b.Name, b.Description, status)
	}
	fmt.Println()

	return nil
}

func runBridgeEnable(cmd *cobra.Command, args []string) error {
	bridgeName := args[0]

	bridge := bridges.Get(bridgeName)
	if bridge == nil {
		return fmt.Errorf("unknown bridge: %s\nRun 'muxbee bridge list' to see available bridges", bridgeName)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nRun 'muxbee init' first", err)
	}

	if cfg.IsBridgeEnabled(bridgeName) {
		fmt.Printf("Bridge '%s' is already enabled.\n", bridgeName)
		return nil
	}

	if bridge.RequiresAPICredentials {
		if err := promptForAPICredentials(cfg, bridgeName); err != nil {
			return err
		}
	}

	cfg.EnableBridge(bridgeName)

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	gen := generator.New()
	if err := gen.GenerateAll(cfg); err != nil {
		return fmt.Errorf("failed to generate configs: %w", err)
	}

	fmt.Printf("Bridge '%s' enabled.\n", bridgeName)

	compose := docker.New(cfg)
	if compose.IsServiceRunning("synapse") {
		fmt.Println("Services are running. Applying changes...")

		// Restart Synapse first so it loads the new bridge registration
		fmt.Println("  Restarting Synapse to load bridge registration...")
		if err := compose.RestartQuiet("synapse"); err != nil {
			return fmt.Errorf("failed to restart Synapse: %w", err)
		}

		time.Sleep(3 * time.Second) // Wait for Synapse to be ready

		fmt.Printf("  Starting %s bridge...\n", bridgeName)
		profiles := docker.GetProfiles(cfg)
		if err := compose.UpQuiet(profiles); err != nil {
			return fmt.Errorf("failed to start bridge: %w", err)
		}

		fmt.Println("Done!")
	} else {
		fmt.Println()
		fmt.Println("Run 'muxbee up' to start services.")
	}

	fmt.Printf("Run 'muxbee bridge login %s' for login instructions.\n", bridgeName)

	return nil
}

func promptForAPICredentials(cfg *config.Config, bridgeName string) error {
	reader := bufio.NewReader(os.Stdin)

	switch bridgeName {
	case "telegram":
		if cfg.Telegram != nil && cfg.Telegram.APIID != "" && cfg.Telegram.APIHash != "" {
			fmt.Println("Telegram API credentials already configured.")
			fmt.Print("Update them? (y/N): ")
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				return nil
			}
		}

		fmt.Println()
		fmt.Println("Telegram requires API credentials from https://my.telegram.org")
		fmt.Println()
		fmt.Println("To get your credentials:")
		fmt.Println("  1. Go to https://my.telegram.org")
		fmt.Println("  2. Log in with your phone number")
		fmt.Println("  3. Go to 'API development tools'")
		fmt.Println("  4. Create a new application")
		fmt.Println("  5. Copy the api_id and api_hash")
		fmt.Println()

		fmt.Print("Enter API ID: ")
		apiID, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		apiID = strings.TrimSpace(apiID)
		if apiID == "" {
			return fmt.Errorf("API ID is required")
		}

		fmt.Print("Enter API Hash: ")
		apiHash, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		apiHash = strings.TrimSpace(apiHash)
		if apiHash == "" {
			return fmt.Errorf("API Hash is required")
		}

		cfg.Telegram = &config.TelegramConfig{
			APIID:   apiID,
			APIHash: apiHash,
		}
		fmt.Println()
		fmt.Println("Telegram credentials saved.")
	}

	return nil
}

func runBridgeDisable(cmd *cobra.Command, args []string) error {
	bridgeName := args[0]

	if !bridges.Exists(bridgeName) {
		return fmt.Errorf("unknown bridge: %s", bridgeName)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nRun 'muxbee init' first", err)
	}

	if !cfg.IsBridgeEnabled(bridgeName) {
		fmt.Printf("Bridge '%s' is not enabled.\n", bridgeName)
		return nil
	}

	cfg.DisableBridge(bridgeName)

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	gen := generator.New()
	if err := gen.GenerateAll(cfg); err != nil {
		return fmt.Errorf("failed to generate configs: %w", err)
	}

	fmt.Printf("Bridge '%s' disabled.\n", bridgeName)
	fmt.Println("Run 'muxbee up' to apply changes.")

	return nil
}

func runBridgeLogin(cmd *cobra.Command, args []string) error {
	bridgeName := args[0]

	bridge := bridges.Get(bridgeName)
	if bridge == nil {
		return fmt.Errorf("unknown bridge: %s", bridgeName)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nRun 'muxbee init' first", err)
	}

	if !cfg.IsBridgeEnabled(bridgeName) {
		return fmt.Errorf("bridge '%s' is not enabled\nRun 'muxbee bridge enable %s' first", bridgeName, bridgeName)
	}

	fmt.Printf("Login instructions for %s:\n", bridge.Description)
	fmt.Println()

	instructions := strings.ReplaceAll(bridge.LoginInstructions, "SERVER", cfg.ServerName)
	fmt.Println(instructions)

	return nil
}
