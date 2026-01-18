package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tobias/muxchat/internal/config"
	"github.com/tobias/muxchat/internal/docker"
)

var logsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "View logs from muxchat services",
	Long: `View logs from muxchat services.

If no service is specified, logs from all services are shown.

Examples:
  muxchat logs              # All services
  muxchat logs synapse      # Synapse only
  muxchat logs -f           # Follow all logs
  muxchat logs -f synapse   # Follow Synapse logs`,
	RunE: runLogs,
}

var (
	logsFollow bool
	logsTail   string
)

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().StringVarP(&logsTail, "tail", "n", "100", "Number of lines to show")
}

func runLogs(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nRun 'muxchat init' first", err)
	}

	compose := docker.New(cfg)

	service := ""
	if len(args) > 0 {
		service = args[0]
	}

	return compose.Logs(service, logsFollow, logsTail)
}
