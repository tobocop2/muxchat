package tui

import (
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tobias/muxchat/internal/config"
	"github.com/tobias/muxchat/internal/docker"
)

// DashboardModel handles the dashboard screen
type DashboardModel struct {
	services    []docker.ServiceStatus
	isLoading   bool
	loadingOp   string
	lastUpdated time.Time
}

// NewDashboardModel creates a new dashboard model
func NewDashboardModel() *DashboardModel {
	return &DashboardModel{
		services: []docker.ServiceStatus{},
	}
}

// Init initializes the dashboard
func (m *DashboardModel) Init() tea.Cmd {
	return nil
}

// Update handles dashboard events
func (m *DashboardModel) Update(msg tea.Msg, cfg *config.Config, compose *docker.Compose) (*DashboardModel, tea.Cmd) {
	switch msg := msg.(type) {
	case servicesUpdatedMsg:
		m.isLoading = false
		m.loadingOp = ""
		m.services = msg.services
		m.lastUpdated = time.Now()
		return m, nil
	case tea.KeyMsg:
		if m.isLoading {
			return m, nil
		}
		switch msg.String() {
		case "s":
			if compose != nil && cfg != nil {
				m.isLoading = true
				m.loadingOp = "Starting"
				return m, m.startServicesCmd(cfg, compose)
			}
		case "x":
			if compose != nil && cfg != nil {
				m.isLoading = true
				m.loadingOp = "Stopping"
				return m, m.stopServicesCmd(cfg, compose)
			}
		case "r":
			if compose != nil && cfg != nil {
				m.isLoading = true
				m.loadingOp = "Restarting"
				return m, m.restartServicesCmd(cfg, compose)
			}
		case "o":
			if cfg != nil && cfg.IsElementEnabled() {
				openBrowser(cfg.ElementURL())
			}
		case "e":
			if cfg != nil && compose != nil {
				m.isLoading = true
				m.loadingOp = "Toggling Element"
				return m, m.toggleElementCmd(cfg, compose)
			}
		}
	}
	return m, nil
}

// View renders the dashboard
func (m *DashboardModel) View(cfg *config.Config) string {
	if cfg == nil {
		return ErrorStyle.Render("Config not loaded")
	}

	var s string

	// Header with server info
	s += TitleStyle.Render("muxchat") + "  "
	s += SubtitleStyle.Render(cfg.ServerName+" · "+cfg.ConnectivityMode) + "\n"

	// Status line
	if m.isLoading {
		s += SubtitleStyle.Render(m.loadingOp+"...") + "\n"
	} else if !m.lastUpdated.IsZero() {
		s += SubtitleStyle.Render("updated "+m.lastUpdated.Format("15:04:05")) + "\n"
	}
	s += "\n"

	// Services
	s += TitleStyle.Render("Services") + "\n"
	if len(m.services) == 0 {
		s += "  " + SubtitleStyle.Render("none running") + "\n"
	} else {
		for _, svc := range m.services {
			name := docker.ParseServiceName(svc.Name)
			status := RenderStatus(svc.Running, svc.Health)
			s += "  " + name + " " + status + "\n"
		}
	}
	s += "\n"

	// Bridges
	s += TitleStyle.Render("Bridges") + "\n"
	if len(cfg.EnabledBridges) == 0 {
		s += "  " + SubtitleStyle.Render("none") + "\n"
	} else {
		s += "  " + strings.Join(cfg.EnabledBridges, ", ") + "\n"
	}
	s += "\n"

	// Element status
	s += TitleStyle.Render("Element") + "  "
	if cfg.IsElementEnabled() {
		s += SuccessStyle.Render("enabled") + "\n"
	} else {
		s += SubtitleStyle.Render("disabled") + "\n"
	}
	s += "\n"

	// URLs
	s += TitleStyle.Render("URLs") + "\n"
	if cfg.IsElementEnabled() {
		s += "  Element Web:  " + cfg.ElementURL() + "\n"
	}
	s += "  Matrix API:   " + cfg.PublicBaseURL() + "\n"

	return s
}

type servicesUpdatedMsg struct {
	services []docker.ServiceStatus
}

func (m *DashboardModel) startServicesCmd(cfg *config.Config, compose *docker.Compose) tea.Cmd {
	return func() tea.Msg {
		profiles := docker.GetProfiles(cfg)
		compose.UpQuiet(profiles)
		services, _ := compose.Status()
		return servicesUpdatedMsg{services: services}
	}
}

func (m *DashboardModel) stopServicesCmd(cfg *config.Config, compose *docker.Compose) tea.Cmd {
	return func() tea.Msg {
		profiles := docker.GetProfiles(cfg)
		compose.DownQuiet(profiles)
		services, _ := compose.Status()
		return servicesUpdatedMsg{services: services}
	}
}

func (m *DashboardModel) restartServicesCmd(cfg *config.Config, compose *docker.Compose) tea.Cmd {
	return func() tea.Msg {
		profiles := docker.GetProfiles(cfg)
		compose.DownQuiet(profiles)
		compose.UpQuiet(profiles)
		services, _ := compose.Status()
		return servicesUpdatedMsg{services: services}
	}
}

func (m *DashboardModel) toggleElementCmd(cfg *config.Config, compose *docker.Compose) tea.Cmd {
	return func() tea.Msg {
		enabled := !cfg.IsElementEnabled()
		cfg.ElementEnabled = &enabled
		cfg.Save()

		profiles := docker.GetProfiles(cfg)
		compose.DownQuiet(profiles)
		compose.UpQuiet(profiles)
		services, _ := compose.Status()
		return servicesUpdatedMsg{services: services}
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	cmd.Start()
}
