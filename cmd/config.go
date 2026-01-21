package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tobocop2/muxbee/internal/config"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage muxbee configuration",
	Long:  `View and manage muxbee configuration.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current muxbee configuration.`,
	RunE:  runConfigShow,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration paths",
	Long:  `Display the paths to muxbee configuration and data directories.`,
	RunE:  runConfigPath,
}

var showSecrets bool

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configPathCmd)

	configShowCmd.Flags().BoolVar(&showSecrets, "show-secrets", false, "Show passwords and tokens (default: masked)")
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nRun 'muxbee init' first", err)
	}

	// Mask passwords for display unless --show-secrets is set
	displayCfg := *cfg
	if !showSecrets {
		displayCfg.Postgres.Password = "********"
		displayCfg.Admin.Password = "********"
	}

	data, err := yaml.Marshal(displayCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Println("Current configuration:")
	fmt.Println()
	fmt.Println(string(data))

	if !showSecrets {
		fmt.Println("(passwords masked, use --show-secrets to reveal)")
	}

	return nil
}

func runConfigPath(cmd *cobra.Command, args []string) error {
	fmt.Printf("Config directory: %s\n", config.ConfigDir())
	fmt.Printf("Data directory:   %s\n", config.DataDir())
	fmt.Printf("Settings file:    %s\n", config.SettingsPath())

	return nil
}
