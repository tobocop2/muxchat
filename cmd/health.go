package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/tobocop2/muxbee/internal/config"
	"github.com/tobocop2/muxbee/internal/docker"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check health of muxbee services",
	Long: `Perform health checks on all muxbee services.

This checks:
  - Docker availability
  - Container status
  - Synapse API health
  - Element Web availability`,
	RunE: runHealth,
}

func init() {
	rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, args []string) error {
	fmt.Println("Mautrix Chat Health Check")
	fmt.Println("====================")
	fmt.Println()

	allHealthy := true

	fmt.Print("Docker:       ")
	if err := docker.DockerAvailable(); err != nil {
		fmt.Println("FAIL - Docker not available")
		allHealthy = false
	} else {
		fmt.Println("OK")
	}

	fmt.Print("Config:       ")
	if !config.Exists() {
		fmt.Println("FAIL - No configuration found")
		allHealthy = false
		return nil
	}
	fmt.Println("OK")

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Config Load:  FAIL - %v\n", err)
		return nil
	}

	compose := docker.New(cfg)

	fmt.Print("Services:     ")
	statuses, err := compose.Status()
	if err != nil {
		fmt.Printf("FAIL - %v\n", err)
		allHealthy = false
	} else if len(statuses) == 0 {
		fmt.Println("NOT RUNNING")
		allHealthy = false
	} else {
		runningCount := 0
		for _, s := range statuses {
			if s.Running {
				runningCount++
			}
		}
		if runningCount == len(statuses) {
			fmt.Printf("OK (%d/%d running)\n", runningCount, len(statuses))
		} else {
			fmt.Printf("DEGRADED (%d/%d running)\n", runningCount, len(statuses))
			allHealthy = false
		}
	}

	fmt.Print("Synapse API:  ")
	synapseHealthURL := fmt.Sprintf("http://localhost:%d/health", cfg.SynapsePort())
	if checkHTTP(synapseHealthURL, 5*time.Second) {
		fmt.Println("OK")
	} else {
		fmt.Println("FAIL")
		allHealthy = false
	}

	if cfg.IsElementEnabled() {
		fmt.Print("Element Web:  ")
		if checkHTTP(cfg.ElementURL(), 5*time.Second) {
			fmt.Println("OK")
		} else {
			fmt.Println("FAIL")
			allHealthy = false
		}
	}

	fmt.Println()

	if allHealthy {
		fmt.Println("All systems operational.")
	} else {
		fmt.Println("Some checks failed. Run 'muxbee status' for details.")
	}

	return nil
}

func checkHTTP(url string, timeout time.Duration) bool {
	client := http.Client{Timeout: timeout}

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 400
}
