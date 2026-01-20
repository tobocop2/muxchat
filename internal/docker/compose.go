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
	"sync"
	"time"

	"github.com/tobias/muxbee/internal/config"
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
	Version string
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

// UpForceRecreate starts services and forces container recreation (for updates)
func (c *Compose) UpForceRecreate(profiles []string) error {
	args := []string{}
	for _, p := range profiles {
		args = append(args, "--profile", p)
	}
	args = append(args, "up", "-d", "--force-recreate")

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

// UpForceRecreateQuiet starts services with force recreate, without output
func (c *Compose) UpForceRecreateQuiet(profiles []string) error {
	args := []string{}
	for _, p := range profiles {
		args = append(args, "--profile", p)
	}
	args = append(args, "up", "-d", "--force-recreate", "--quiet-pull")

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

	// First pass: collect all services
	type svcInfo struct {
		Name    string `json:"Name"`
		State   string `json:"State"`
		Health  string `json:"Health"`
		Image   string `json:"Image"`
		Service string `json:"Service"`
	}
	var services []svcInfo

	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var svc svcInfo
		if err := json.Unmarshal([]byte(line), &svc); err != nil {
			continue
		}
		services = append(services, svc)
	}

	// Fetch versions in parallel for running containers
	versions := make([]string, len(services))
	var wg sync.WaitGroup
	for i, svc := range services {
		if svc.State != "running" {
			continue
		}
		wg.Add(1)
		go func(idx int, s svcInfo) {
			defer wg.Done()
			versions[idx] = getContainerVersion(s.Name, s.Service, s.Image)
		}(i, svc)
	}
	wg.Wait()

	// Build result
	var statuses []ServiceStatus
	for i, svc := range services {
		statuses = append(statuses, ServiceStatus{
			Name:    svc.Name,
			State:   svc.State,
			Health:  svc.Health,
			Running: svc.State == "running",
			Version: versions[i],
		})
	}

	return statuses, nil
}

// getContainerVersion attempts to get the version of a container
func getContainerVersion(containerName, service, image string) string {
	// First try the OCI version label
	cmd := exec.Command("docker", "inspect", containerName,
		"--format", `{{index .Config.Labels "org.opencontainers.image.version"}}`)
	out, err := cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(out))
		if version != "" && version != "<no value>" {
			return cleanVersion(version)
		}
	}

	// For images with explicit version tags (e.g., postgres:17), extract from image
	if idx := strings.LastIndex(image, ":"); idx != -1 {
		tag := image[idx+1:]
		if tag != "latest" && tag != "" {
			return tag
		}
	}

	// For mautrix bridges, try multiple methods
	if strings.HasPrefix(service, "mautrix-") {
		// Method 1: Try running the Go binary with --version
		binary := "/usr/bin/" + service
		cmd = exec.Command("docker", "exec", containerName, binary, "--version")
		out, err = cmd.Output()
		if err == nil {
			// Parse "mautrix-discord 0.7.5+dev.11b1ea5a (Nov 25 2025...)"
			// or "mautrix-whatsapp v26.01+dev.4d9366c2 (built at...)"
			version := strings.TrimSpace(string(out))
			parts := strings.Fields(version)
			if len(parts) >= 2 {
				v := parts[1]
				// Strip +dev.xxx suffix
				if idx := strings.Index(v, "+"); idx != -1 {
					v = v[:idx]
				}
				return cleanVersion(v)
			}
		}

		// Method 2: Try pip show for Python bridges
		pipPkg := strings.Replace(service, "-", "_", 1) // mautrix-telegram -> mautrix_telegram
		cmd = exec.Command("docker", "exec", containerName, "pip", "show", pipPkg)
		out, err = cmd.Output()
		if err == nil {
			// Parse "Version: 0.5.2+dev.xxx"
			for _, line := range strings.Split(string(out), "\n") {
				if strings.HasPrefix(line, "Version:") {
					v := strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
					// Strip +dev.xxx suffix
					if idx := strings.Index(v, "+"); idx != -1 {
						v = v[:idx]
					}
					return cleanVersion(v)
				}
			}
		}
	}

	return ""
}

// cleanVersion removes common prefixes like "v" from version strings
func cleanVersion(version string) string {
	version = strings.TrimPrefix(version, "v")
	return version
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
	const prefix = "muxbee-"
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
