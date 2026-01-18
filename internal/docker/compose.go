package docker

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/tobias/muxchat/internal/config"
)

//go:embed docker-compose.yml
var DockerComposeYAML []byte

// Compose provides Docker Compose operations
type Compose struct {
	configDir string
	dataDir   string
	env       []string
}

// ServiceStatus represents the status of a Docker service
type ServiceStatus struct {
	Name    string
	State   string
	Health  string
	Running bool
}

// New creates a new Compose instance
func New(cfg *config.Config) *Compose {
	configDir := config.ConfigDir()
	dataDir := config.DataDir()

	return &Compose{
		configDir: configDir,
		dataDir:   dataDir,
		env: []string{
			"CONFIG_DIR=" + configDir,
			"DATA_DIR=" + dataDir,
			"POSTGRES_PASSWORD=" + cfg.Postgres.Password,
			fmt.Sprintf("SYNAPSE_PORT=%d", cfg.SynapsePort()),
			fmt.Sprintf("ELEMENT_PORT=%d", cfg.ElementPort()),
		},
	}
}

// WriteComposeFile writes the docker-compose.yml to the config directory
func (c *Compose) WriteComposeFile() error {
	if err := config.EnsureDirs(); err != nil {
		return err
	}
	path := filepath.Join(c.configDir, "docker-compose.yml")
	return os.WriteFile(path, DockerComposeYAML, 0644)
}

// composePath returns the path to docker-compose.yml
func (c *Compose) composePath() string {
	return filepath.Join(c.configDir, "docker-compose.yml")
}

// buildCommand creates an exec.Cmd for docker compose
func (c *Compose) buildCommand(args ...string) *exec.Cmd {
	fullArgs := append([]string{"compose", "-f", c.composePath()}, args...)
	cmd := exec.Command("docker", fullArgs...)
	cmd.Env = append(os.Environ(), c.env...)
	return cmd
}

// Up starts services with specified profiles
func (c *Compose) Up(profiles []string) error {
	args := []string{}
	for _, p := range profiles {
		args = append(args, "--profile", p)
	}
	args = append(args, "up", "-d")

	cmd := c.buildCommand(args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// UpQuiet starts services without output (for background/TUI use)
func (c *Compose) UpQuiet(profiles []string) error {
	args := []string{}
	for _, p := range profiles {
		args = append(args, "--profile", p)
	}
	args = append(args, "up", "-d", "--quiet-pull")

	cmd := c.buildCommand(args...)
	return cmd.Run()
}

// Down stops all services
func (c *Compose) Down(profiles []string) error {
	args := []string{}
	for _, p := range profiles {
		args = append(args, "--profile", p)
	}
	args = append(args, "down", "--remove-orphans")

	cmd := c.buildCommand(args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// DownQuiet stops services without output
func (c *Compose) DownQuiet(profiles []string) error {
	args := []string{}
	for _, p := range profiles {
		args = append(args, "--profile", p)
	}
	args = append(args, "down", "--remove-orphans")

	cmd := c.buildCommand(args...)
	return cmd.Run()
}

// DownVolumes stops all services and removes volumes
func (c *Compose) DownVolumes() error {
	cmd := c.buildCommand("down", "-v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Status returns the status of all services
func (c *Compose) Status() ([]ServiceStatus, error) {
	cmd := c.buildCommand("ps", "--format", "json", "-a")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return nil, nil
	}

	var statuses []ServiceStatus
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var svc struct {
			Name   string `json:"Name"`
			State  string `json:"State"`
			Health string `json:"Health"`
		}
		if err := json.Unmarshal([]byte(line), &svc); err != nil {
			continue
		}

		statuses = append(statuses, ServiceStatus{
			Name:    svc.Name,
			State:   svc.State,
			Health:  svc.Health,
			Running: svc.State == "running",
		})
	}

	return statuses, nil
}

// IsRunning checks if services are currently running
func (c *Compose) IsRunning() bool {
	statuses, err := c.Status()
	if err != nil {
		return false
	}
	for _, s := range statuses {
		if s.Running {
			return true
		}
	}
	return false
}

// IsServiceRunning checks if a specific service is running
func (c *Compose) IsServiceRunning(serviceName string) bool {
	statuses, err := c.Status()
	if err != nil {
		return false
	}
	for _, s := range statuses {
		if strings.Contains(s.Name, serviceName) && s.Running {
			return true
		}
	}
	return false
}

// Logs streams logs for a specific service or all services
func (c *Compose) Logs(service string, follow bool, tail string) error {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	if tail != "" {
		args = append(args, "--tail", tail)
	}
	if service != "" {
		args = append(args, service)
	}

	cmd := c.buildCommand(args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// LogsReader returns a reader for streaming logs
func (c *Compose) LogsReader(service string, follow bool, tail string) (io.ReadCloser, error) {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	if tail != "" {
		args = append(args, "--tail", tail)
	}
	if service != "" {
		args = append(args, service)
	}

	cmd := c.buildCommand(args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return stdout, nil
}

// Pull pulls images for specified profiles
func (c *Compose) Pull(profiles []string) error {
	args := []string{}
	for _, p := range profiles {
		args = append(args, "--profile", p)
	}
	args = append(args, "pull")

	cmd := c.buildCommand(args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// PullQuiet pulls images without output (for background/TUI use)
func (c *Compose) PullQuiet(profiles []string) error {
	args := []string{}
	for _, p := range profiles {
		args = append(args, "--profile", p)
	}
	args = append(args, "pull", "-q")

	cmd := c.buildCommand(args...)
	return cmd.Run()
}

// Restart restarts a specific service
func (c *Compose) Restart(service string) error {
	cmd := c.buildCommand("restart", service)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RestartQuiet restarts a service without output
func (c *Compose) RestartQuiet(service string) error {
	cmd := c.buildCommand("restart", service)
	return cmd.Run()
}

// StopService stops and removes a specific service container
func (c *Compose) StopService(service string) error {
	cmd := c.buildCommand("rm", "-f", "-s", service)
	return cmd.Run()
}

// GetProfiles returns the profiles to enable based on config
func GetProfiles(cfg *config.Config) []string {
	profiles := make([]string, len(cfg.EnabledBridges))
	copy(profiles, cfg.EnabledBridges)

	if cfg.IsElementEnabled() {
		profiles = append(profiles, "element")
	}

	if cfg.HTTPS.Enabled {
		profiles = append(profiles, "https")
	}

	return profiles
}

// DockerAvailable checks if Docker is available and running
func DockerAvailable() error {
	cmd := exec.Command("docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// WaitForBridges waits for all specified bridges to be running
func (c *Compose) WaitForBridges(bridges []string, timeout int) error {
	if len(bridges) == 0 {
		return nil
	}

	expected := make(map[string]bool)
	for _, b := range bridges {
		expected["mautrix-"+b] = false
	}

	for i := 0; i < timeout; i++ {
		statuses, err := c.Status()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		allRunning := true
		for _, s := range statuses {
			svcName := ParseServiceName(s.Name)
			if _, ok := expected[svcName]; ok {
				if s.Running {
					expected[svcName] = true
				} else {
					allRunning = false
				}
			}
		}

		if allRunning {
			allFound := true
			for _, running := range expected {
				if !running {
					allFound = false
					break
				}
			}
			if allFound {
				return nil
			}
		}

		time.Sleep(time.Second)
	}

	return nil
}

// ParseServiceName extracts the service name from a container name
func ParseServiceName(containerName string) string {
	const prefix = "muxchat-"
	if strings.HasPrefix(containerName, prefix) {
		remainder := strings.TrimPrefix(containerName, prefix)
		if idx := strings.LastIndex(remainder, "-"); idx > 0 {
			return remainder[:idx]
		}
		return remainder
	}
	parts := strings.Split(containerName, "-")
	if len(parts) >= 2 {
		return strings.Join(parts[1:len(parts)-1], "-")
	}
	return containerName
}
