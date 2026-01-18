package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tobias/muxchat/internal/tui"
)

// Version is set at build time via ldflags
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "muxchat",
	Short: "Self-hosted Matrix server with messaging bridges",
	Long: `muxchat is a self-hosted Matrix server with messaging bridges.

Run without arguments to launch the TUI, or use subcommands for CLI access.`,
	Version: Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run()
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate("muxchat version {{.Version}}\n")
}
