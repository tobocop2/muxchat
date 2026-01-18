package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tobias/muxchat/internal/config"
	"github.com/tobias/muxchat/internal/docker"
)

const statusRefreshInterval = 2 * time.Second

// Screen represents different TUI screens
type Screen int

const (
	ScreenDashboard Screen = iota
	ScreenBridges
	ScreenLogs
	ScreenSettings
	ScreenWizard
)

// Model represents the TUI application state
type Model struct {
	screen       Screen
	width        int
	height       int
	config       *config.Config
	compose      *docker.Compose
	services     []docker.ServiceStatus
	err          error
	configExists bool
	quitting     bool
	autoStarting bool

	// Sub-models for each screen
	dashboard *DashboardModel
	bridges   *BridgesModel
	logs      *LogsModel
	settings  *SettingsModel
	wizard    *WizardModel
}

// New creates a new TUI model
func New() Model {
	m := Model{
		screen:   ScreenDashboard,
		services: []docker.ServiceStatus{},
	}

	// Check if config exists
	m.configExists = config.Exists()

	// Load config if it exists
	if m.configExists {
		cfg, err := config.Load()
		if err != nil {
			m.err = err
		} else {
			m.config = cfg
			m.compose = docker.New(cfg)
			m.autoStarting = true
		}
	}

	// Initialize sub-models
	m.dashboard = NewDashboardModel()
	m.bridges = NewBridgesModel()
	m.logs = NewLogsModel()
	m.settings = NewSettingsModel(m.config)
	m.wizard = NewWizardModel()

	// Set dashboard to loading state if we'll auto-start
	if m.autoStarting {
		m.dashboard.isLoading = true
		m.dashboard.loadingOp = "Starting"
	}

	return m
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	// If no config exists, start the wizard
	if !m.configExists {
		m.screen = ScreenWizard
		return m.wizard.Init()
	}

	// Otherwise, start on dashboard and auto-start services
	// Also start the ticker for real-time updates
	return tea.Batch(
		m.dashboard.Init(),
		m.autoStartCmd(),
		tickCmd(),
	)
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Global key handlers
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q":
			if m.screen == ScreenSettings {
				// Return to dashboard from settings
				m.screen = ScreenDashboard
				return m, m.fetchStatusCmd()
			}
			if m.screen != ScreenWizard {
				m.quitting = true
				return m, tea.Quit
			}
		case "esc":
			if m.screen == ScreenSettings && !m.settings.editing {
				// Return to dashboard from settings
				m.screen = ScreenDashboard
				return m, m.fetchStatusCmd()
			}
		case "d":
			if m.screen != ScreenWizard && m.configExists {
				m.stopLogStreaming()
				m.screen = ScreenDashboard
				return m, m.fetchStatusCmd()
			}
		case "b":
			if m.screen != ScreenWizard && m.screen != ScreenSettings && m.configExists {
				m.stopLogStreaming()
				m.screen = ScreenBridges
				return m, nil
			}
		case "l":
			if m.screen != ScreenWizard && m.screen != ScreenSettings && m.configExists {
				m.screen = ScreenLogs
				return m, m.logs.StartStreaming(m.compose)
			}
		case "c":
			if m.screen != ScreenWizard && m.screen != ScreenSettings && m.configExists {
				m.stopLogStreaming()
				// Reinitialize settings with current config
				m.settings = NewSettingsModel(m.config)
				m.screen = ScreenSettings
				return m, nil
			}
		}

	case statusMsg:
		m.services = msg.services
		m.err = msg.err
		m.dashboard.services = msg.services
		m.dashboard.lastUpdated = time.Now()
		return m, nil

	case autoStartMsg:
		m.autoStarting = false
		m.services = msg.services
		m.err = msg.err
		m.dashboard.services = msg.services
		m.dashboard.isLoading = false
		m.dashboard.lastUpdated = time.Now()
		return m, tickCmd()

	case tickMsg:
		if m.screen == ScreenDashboard && !m.dashboard.isLoading && m.compose != nil {
			return m, tea.Batch(m.fetchStatusCmd(), tickCmd())
		}
		return m, tickCmd()

	case configCreatedMsg:
		// Wizard completed, reload config and switch to dashboard
		m.configExists = true
		cfg, err := config.Load()
		if err != nil {
			m.err = err
		} else {
			m.config = cfg
			m.compose = docker.New(cfg)
			m.settings = NewSettingsModel(cfg)
		}
		m.screen = ScreenDashboard
		return m, tea.Batch(m.fetchStatusCmd(), tickCmd())

	case settingsSavedMsg:
		// Reload config after settings saved
		if msg.err == nil {
			cfg, err := config.Load()
			if err == nil {
				m.config = cfg
				m.compose = docker.New(cfg)
			}
		}
	}

	// Delegate to active screen
	var cmd tea.Cmd
	switch m.screen {
	case ScreenDashboard:
		m.dashboard, cmd = m.dashboard.Update(msg, m.config, m.compose)
	case ScreenBridges:
		m.bridges, cmd = m.bridges.Update(msg, m.config)
	case ScreenLogs:
		m.logs, cmd = m.logs.Update(msg, m.compose)
	case ScreenSettings:
		m.settings, cmd = m.settings.Update(msg, m.config, m.compose)
	case ScreenWizard:
		m.wizard, cmd = m.wizard.Update(msg)
	}

	return m, cmd
}

func (m *Model) stopLogStreaming() {
	if m.screen == ScreenLogs && m.logs.streaming {
		m.logs.streaming = false
		if m.logs.reader != nil {
			m.logs.reader.Close()
			m.logs.reader = nil
		}
	}
}

// View implements tea.Model
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var content string

	switch m.screen {
	case ScreenDashboard:
		content = m.dashboard.View(m.config)
	case ScreenBridges:
		content = m.bridges.View(m.config)
	case ScreenLogs:
		content = m.logs.View()
	case ScreenSettings:
		content = m.settings.View()
	case ScreenWizard:
		content = m.wizard.View()
	}

	// Add navigation tabs if not in wizard
	if m.screen != ScreenWizard && m.configExists {
		tabs := m.renderTabs()
		help := m.renderHelp()
		return tabs + "\n\n" + content + "\n\n" + help
	}

	return content
}

func (m Model) renderTabs() string {
	tabs := []string{"Dashboard", "Bridges", "Logs", "Settings"}
	screenToTab := map[Screen]int{
		ScreenDashboard: 0,
		ScreenBridges:   1,
		ScreenLogs:      2,
		ScreenSettings:  3,
	}
	activeIdx := screenToTab[m.screen]

	var rendered []string
	for i, tab := range tabs {
		if i == activeIdx {
			rendered = append(rendered, ActiveTabStyle.Render(tab))
		} else {
			rendered = append(rendered, TabStyle.Render(tab))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func (m Model) renderHelp() string {
	var keys []string

	switch m.screen {
	case ScreenDashboard:
		keys = []string{
			RenderKey("s", "start"),
			RenderKey("x", "stop"),
			RenderKey("r", "restart"),
			RenderKey("e", "element"),
			RenderKey("o", "open"),
		}
	case ScreenBridges:
		keys = []string{
			RenderKey("enter", "toggle"),
			RenderKey("i", "info"),
		}
	case ScreenLogs:
		keys = []string{
			RenderKey("f", "filter"),
			RenderKey("p", "pause"),
		}
	case ScreenSettings:
		// Settings has its own help
		keys = []string{}
	}

	// Add navigation keys
	keys = append(keys,
		RenderKey("d", "dashboard"),
		RenderKey("b", "bridges"),
		RenderKey("l", "logs"),
		RenderKey("c", "settings"),
		RenderKey("q", "quit"),
	)

	return HelpStyle.Render(strings.Join(keys, "  "))
}

// Messages
type statusMsg struct {
	services []docker.ServiceStatus
	err      error
}

type configCreatedMsg struct{}

type autoStartMsg struct {
	services []docker.ServiceStatus
	err      error
}

type tickMsg time.Time

// Commands
func tickCmd() tea.Cmd {
	return tea.Tick(statusRefreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) fetchStatusCmd() tea.Cmd {
	return func() tea.Msg {
		if m.compose == nil {
			return statusMsg{services: nil, err: fmt.Errorf("not initialized")}
		}
		services, err := m.compose.Status()
		return statusMsg{services: services, err: err}
	}
}

func (m Model) autoStartCmd() tea.Cmd {
	return func() tea.Msg {
		if m.compose == nil || m.config == nil {
			return autoStartMsg{services: nil, err: fmt.Errorf("not initialized")}
		}

		// Check if services are already running
		services, _ := m.compose.Status()
		hasRunning := false
		for _, s := range services {
			if s.Running {
				hasRunning = true
				break
			}
		}

		// If no services running, auto-start them
		if !hasRunning {
			profiles := docker.GetProfiles(m.config)
			m.compose.UpQuiet(profiles)
			services, _ = m.compose.Status()
		}

		return autoStartMsg{services: services, err: nil}
	}
}

// Run starts the TUI application
func Run() error {
	p := tea.NewProgram(New(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
