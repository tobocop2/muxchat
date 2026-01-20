package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tobias/muxbee/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive TUI",
	Long:  `Launch the interactive terminal user interface for managing muxbee.`,
	RunE:  runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTUI(cmd *cobra.Command, args []string) error {
	return tui.Run()
}
