package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/tobias/muxchat/internal/config"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open Element Web in your browser",
	Long:  `Open the Element Web client in your default browser.`,
	RunE:  runOpen,
}

func init() {
	rootCmd.AddCommand(openCmd)
}

func runOpen(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nRun 'muxchat init' first", err)
	}

	url := cfg.ElementURL()

	fmt.Printf("Opening %s in your browser...\n", url)

	if err := openBrowser(url); err != nil {
		return fmt.Errorf("failed to open browser: %w\nManually navigate to: %s", err, url)
	}

	return nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
