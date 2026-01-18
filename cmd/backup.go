package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/tobias/muxchat/internal/config"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup muxchat data and configuration",
	Long: `Create a backup of all muxchat data and configuration.

The backup includes:
  - Configuration files (synapse, element, bridges)
  - Data directories (database, media, bridge state)
  - Settings file`,
	RunE: runBackup,
}

var backupOutput string

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().StringVarP(&backupOutput, "output", "o", "", "Output file path (default: muxchat-backup-TIMESTAMP.tar.gz)")
}

func runBackup(cmd *cobra.Command, args []string) error {
	if !config.Exists() {
		return fmt.Errorf("no configuration found\nRun 'muxchat init' first")
	}

	if backupOutput == "" {
		timestamp := time.Now().Format("20060102-150405")
		backupOutput = fmt.Sprintf("muxchat-backup-%s.tar.gz", timestamp)
	}

	configDir := config.ConfigDir()
	dataDir := config.DataDir()

	fmt.Printf("Creating backup...\n")
	fmt.Printf("  Config: %s\n", configDir)
	fmt.Printf("  Data:   %s\n", dataDir)

	f, err := os.Create(backupOutput)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	if err := addDirToTar(tw, configDir, "config"); err != nil {
		return fmt.Errorf("failed to backup config: %w", err)
	}

	if err := addDirToTar(tw, dataDir, "data"); err != nil {
		return fmt.Errorf("failed to backup data: %w", err)
	}

	fmt.Printf("\nBackup created: %s\n", backupOutput)
	return nil
}

func addDirToTar(tw *tar.Writer, srcDir, prefix string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.Join(prefix, relPath)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}

		return nil
	})
}
