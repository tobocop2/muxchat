package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tobocop2/muxbee/internal/config"
	"github.com/tobocop2/muxbee/internal/docker"
)

var logsCmd = &cobra.Command{
	Use:   "logs [service]",
	Short: "View logs from muxbee services",
	Long: `View logs from muxbee services.

If no service is specified, logs from all services are shown.

Examples:
  muxbee logs              # All services
  muxbee logs synapse      # Synapse only
  muxbee logs -f           # Follow all logs
  muxbee logs -f synapse   # Follow Synapse logs`,
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
		return fmt.Errorf("failed to load config: %w\nRun 'muxbee init' first", err)
	}

	compose := docker.New(cfg)

	service := ""
	if len(args) > 0 {
		service = args[0]
	}

	return compose.Logs(service, logsFollow, logsTail)
}
