package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tobias/muxchat/internal/config"
)

var restoreCmd = &cobra.Command{
	Use:   "restore <backup-file>",
	Short: "Restore muxchat from a backup",
	Long: `Restore muxchat data and configuration from a backup file.

This will overwrite existing configuration and data.
Services should be stopped before restoring.`,
	Args: cobra.ExactArgs(1),
	RunE: runRestore,
}

var restoreForce bool

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().BoolVarP(&restoreForce, "force", "f", false, "Overwrite existing data without confirmation")
}

func runRestore(cmd *cobra.Command, args []string) error {
	backupFile := args[0]

	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupFile)
	}

	configDir := config.ConfigDir()
	dataDir := config.DataDir()

	configExists := dirHasContents(configDir)
	dataExists := dirHasContents(dataDir)

	if (configExists || dataExists) && !restoreForce {
		fmt.Println("Warning: Existing data will be overwritten!")
		fmt.Printf("  Config: %s\n", configDir)
		fmt.Printf("  Data:   %s\n", dataDir)
		fmt.Println()
		fmt.Println("Use --force to proceed, or run 'muxchat down' first.")
		return nil
	}

	fmt.Printf("Restoring from %s...\n", backupFile)

	f, err := os.Open(backupFile)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed to read gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		var destDir string
		var relPath string

		if strings.HasPrefix(header.Name, "config/") {
			destDir = configDir
			relPath = strings.TrimPrefix(header.Name, "config/")
		} else if strings.HasPrefix(header.Name, "data/") {
			destDir = dataDir
			relPath = strings.TrimPrefix(header.Name, "data/")
		} else {
			continue
		}

		destPath := filepath.Join(destDir, relPath)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(destPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}

			outFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
			outFile.Close()
		}
	}

	fmt.Println("Restore complete!")
	fmt.Println()
	fmt.Println("Run 'muxchat up' to start services.")

	return nil
}

func dirHasContents(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	return len(entries) > 0
}
