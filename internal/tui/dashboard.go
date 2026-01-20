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

var dashboardSpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// DashboardModel handles the dashboard screen
type DashboardModel struct {
	services    []docker.ServiceStatus
	isLoading   bool
	loadingOp   string
	loadingStep string // Current step within a multi-step operation
	spinnerIdx  int
	lastUpdated time.Time

	// For multi-step operations like update
	updateChan chan updateStepMsg
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
		m.loadingStep = ""
		m.services = msg.services
		m.lastUpdated = time.Now()
		return m, nil

	case dashboardSpinnerMsg:
		if m.isLoading {
			m.spinnerIdx = (m.spinnerIdx + 1) % len(dashboardSpinnerFrames)
			return m, m.spinnerTick()
		}
		return m, nil

	case updateStepMsg:
		if msg.done {
			m.isLoading = false
			m.loadingOp = ""
			m.loadingStep = ""
			m.services = msg.services
			m.lastUpdated = time.Now()
			m.updateChan = nil
			return m, nil
		}
		// Only update step if a new one was provided
		if msg.step != "" {
			m.loadingStep = msg.step
		}
		return m, m.checkUpdateProgress()

	case tea.KeyMsg:
		switch msg.String() {
		case "s":
			// Block starting new operations while one is in progress
			if m.isLoading {
				return m, nil
			}
			if compose != nil && cfg != nil {
				m.isLoading = true
				m.loadingOp = "Starting"
				m.spinnerIdx = 0
				return m, tea.Batch(m.startServicesCmd(cfg, compose), m.spinnerTick())
			}
		case "x":
			if m.isLoading {
				return m, nil
			}
			if compose != nil && cfg != nil {
				m.isLoading = true
				m.loadingOp = "Stopping"
				m.spinnerIdx = 0
				return m, tea.Batch(m.stopServicesCmd(cfg, compose), m.spinnerTick())
			}
		case "r":
			if m.isLoading {
				return m, nil
			}
			if compose != nil && cfg != nil {
				m.isLoading = true
				m.loadingOp = "Restarting"
				m.spinnerIdx = 0
				return m, tea.Batch(m.restartServicesCmd(cfg, compose), m.spinnerTick())
			}
		case "o":
			// Open browser works even during loading
			if cfg != nil && cfg.IsElementEnabled() {
				openBrowser(cfg.ElementURL())
			}
		case "e":
			if m.isLoading {
				return m, nil
			}
			if cfg != nil && compose != nil {
				m.isLoading = true
				m.loadingOp = "Toggling Element"
				m.spinnerIdx = 0
				return m, tea.Batch(m.toggleElementCmd(cfg, compose), m.spinnerTick())
			}
		case "u":
			if m.isLoading {
				return m, nil
			}
			if cfg != nil && compose != nil {
				m.isLoading = true
				m.loadingOp = "Updating"
				m.loadingStep = "Pulling images"
				m.spinnerIdx = 0
				m.updateChan = make(chan updateStepMsg, 1)
				go m.updateServicesBackground(cfg, compose)
				return m, tea.Batch(m.spinnerTick(), m.checkUpdateProgress())
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
		spinner := dashboardSpinnerFrames[m.spinnerIdx]
		if m.loadingStep != "" {
			s += SubtitleStyle.Render(spinner+" "+m.loadingOp+": "+m.loadingStep+"...") + "\n"
		} else {
			s += SubtitleStyle.Render(spinner+" "+m.loadingOp+"...") + "\n"
		}
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
			version := ""
			if svc.Version != "" {
				version = " " + VersionStyle.Render(svc.Version)
			}
			s += "  " + name + version + " " + status + "\n"
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

type dashboardSpinnerMsg struct{}

type updateStepMsg struct {
	step     string // Current step description
	done     bool   // Whether the whole update is done
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

// spinnerTick returns a command that ticks the spinner
func (m *DashboardModel) spinnerTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return dashboardSpinnerMsg{}
	})
}

// checkUpdateProgress polls for update progress messages
func (m *DashboardModel) checkUpdateProgress() tea.Cmd {
	ch := m.updateChan
	return func() tea.Msg {
		if ch == nil {
			return nil
		}
		select {
		case msg := <-ch:
			return msg
		case <-time.After(100 * time.Millisecond):
			// Keep polling - return a msg that will trigger another check
			return updateStepMsg{step: "", done: false}
		}
	}
}

// updateServicesBackground runs the update in the background, sending progress
func (m *DashboardModel) updateServicesBackground(cfg *config.Config, compose *docker.Compose) {
	profiles := docker.GetProfiles(cfg)

	// Step 1: Pull images
	m.updateChan <- updateStepMsg{step: "Pulling images", done: false}
	compose.PullQuiet(profiles)

	// Step 2: Stop services
	m.updateChan <- updateStepMsg{step: "Stopping services", done: false}
	compose.DownQuiet(profiles)

	// Step 3: Start services
	m.updateChan <- updateStepMsg{step: "Starting services", done: false}
	compose.UpForceRecreateQuiet(profiles)

	// Done
	services, _ := compose.Status()
	m.updateChan <- updateStepMsg{step: "", done: true, services: services}
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
