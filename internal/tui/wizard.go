package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tobocop2/muxbee/internal/config"
	"github.com/tobocop2/muxbee/internal/docker"
	"github.com/tobocop2/muxbee/internal/generator"
)

// WizardStep represents a step in the setup wizard
type WizardStep int

const (
	WizardStepWelcome WizardStep = iota
	WizardStepConnectivity
	WizardStepServerConfig // New: input for domain/email/server name
	WizardStepBridges
	WizardStepConfirm
	WizardStepComplete
)

// WizardModel handles the setup wizard
type WizardModel struct {
	step             WizardStep
	cursor           int
	connectivityMode string
	serverName       string
	domain           string
	email            string
	selectedBridges  []string
	err              error

	// Text inputs for server config
	serverInput textinput.Model
	domainInput textinput.Model
	emailInput  textinput.Model
	inputFocus  int // 0 = domain, 1 = email (for public mode)
}

// NewWizardModel creates a new wizard model
func NewWizardModel() *WizardModel {
	// Server name input (for private network mode)
	serverInput := textinput.New()
	serverInput.Placeholder = "192.168.1.50 or hostname.local"
	serverInput.CharLimit = 256
	serverInput.Width = 40

	// Domain input (for public HTTPS mode)
	domainInput := textinput.New()
	domainInput.Placeholder = "chat.example.com"
	domainInput.CharLimit = 256
	domainInput.Width = 40

	// Email input (for Let's Encrypt)
	emailInput := textinput.New()
	emailInput.Placeholder = "you@example.com"
	emailInput.CharLimit = 256
	emailInput.Width = 40

	return &WizardModel{
		step:             WizardStepWelcome,
		cursor:           0,
		connectivityMode: "local",
		serverName:       "localhost",
		selectedBridges:  []string{},
		serverInput:      serverInput,
		domainInput:      domainInput,
		emailInput:       emailInput,
	}
}

// Init initializes the wizard
func (m *WizardModel) Init() tea.Cmd {
	return nil
}

// Update handles wizard events
func (m *WizardModel) Update(msg tea.Msg) (*WizardModel, tea.Cmd) {
	// Handle text input in server config step
	if m.step == WizardStepServerConfig {
		return m.updateServerConfig(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m.nextStep()
		case "esc":
			return m.prevStep()
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			m.cursor++
		case " ":
			m.toggleSelection()
		case "q":
			if m.step == WizardStepWelcome {
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m *WizardModel) updateServerConfig(msg tea.Msg) (*WizardModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Validate and move to next step
			if m.connectivityMode == "private" {
				if strings.TrimSpace(m.serverInput.Value()) == "" {
					m.err = nil // Clear error, let them try again
					return m, nil
				}
				m.serverName = strings.TrimSpace(m.serverInput.Value())
			} else if m.connectivityMode == "public" {
				if m.inputFocus == 0 {
					// Move to email input
					m.inputFocus = 1
					m.domainInput.Blur()
					m.emailInput.Focus()
					return m, textinput.Blink
				}
				// Validate both fields
				domain := strings.TrimSpace(m.domainInput.Value())
				email := strings.TrimSpace(m.emailInput.Value())
				if domain == "" || email == "" {
					m.err = nil
					return m, nil
				}
				m.domain = domain
				m.email = email
				m.serverName = domain
			}
			m.step = WizardStepBridges
			m.cursor = 0
			m.err = nil
			return m, nil
		case "esc":
			return m.prevStep()
		case "tab", "shift+tab":
			if m.connectivityMode == "public" {
				if m.inputFocus == 0 {
					m.inputFocus = 1
					m.domainInput.Blur()
					m.emailInput.Focus()
				} else {
					m.inputFocus = 0
					m.emailInput.Blur()
					m.domainInput.Focus()
				}
				return m, textinput.Blink
			}
		}
	}

	// Update the focused input
	if m.connectivityMode == "private" {
		m.serverInput, cmd = m.serverInput.Update(msg)
	} else if m.connectivityMode == "public" {
		if m.inputFocus == 0 {
			m.domainInput, cmd = m.domainInput.Update(msg)
		} else {
			m.emailInput, cmd = m.emailInput.Update(msg)
		}
	}

	return m, cmd
}

func (m *WizardModel) nextStep() (*WizardModel, tea.Cmd) {
	switch m.step {
	case WizardStepWelcome:
		m.step = WizardStepConnectivity
		m.cursor = 0
	case WizardStepConnectivity:
		switch m.cursor {
		case 0:
			m.connectivityMode = "local"
			m.serverName = "localhost"
			// Skip server config for local mode
			m.step = WizardStepBridges
		case 1:
			m.connectivityMode = "private"
			m.step = WizardStepServerConfig
			m.serverInput.Focus()
			return m, textinput.Blink
		case 2:
			m.connectivityMode = "public"
			m.step = WizardStepServerConfig
			m.inputFocus = 0
			m.domainInput.Focus()
			return m, textinput.Blink
		}
		m.cursor = 0
	case WizardStepBridges:
		m.step = WizardStepConfirm
		m.cursor = 0
	case WizardStepConfirm:
		return m, m.createConfigCmd()
	case WizardStepComplete:
		return m, func() tea.Msg { return configCreatedMsg{} }
	}
	return m, nil
}

func (m *WizardModel) prevStep() (*WizardModel, tea.Cmd) {
	switch m.step {
	case WizardStepConnectivity:
		m.step = WizardStepWelcome
	case WizardStepServerConfig:
		m.step = WizardStepConnectivity
		m.serverInput.Blur()
		m.domainInput.Blur()
		m.emailInput.Blur()
	case WizardStepBridges:
		if m.connectivityMode == "local" {
			m.step = WizardStepConnectivity
		} else {
			m.step = WizardStepServerConfig
			if m.connectivityMode == "private" {
				m.serverInput.Focus()
			} else {
				m.inputFocus = 0
				m.domainInput.Focus()
			}
			return m, textinput.Blink
		}
	case WizardStepConfirm:
		m.step = WizardStepBridges
	}
	m.cursor = 0
	return m, nil
}

func (m *WizardModel) toggleSelection() {
	if m.step == WizardStepBridges {
		bridges := []string{"whatsapp", "signal", "discord", "telegram", "gmessages"}
		if m.cursor < len(bridges) {
			bridge := bridges[m.cursor]
			found := false
			for i, b := range m.selectedBridges {
				if b == bridge {
					m.selectedBridges = append(m.selectedBridges[:i], m.selectedBridges[i+1:]...)
					found = true
					break
				}
			}
			if !found {
				m.selectedBridges = append(m.selectedBridges, bridge)
			}
		}
	}
}

// View renders the wizard
func (m *WizardModel) View() string {
	var s string

	switch m.step {
	case WizardStepWelcome:
		s = m.viewWelcome()
	case WizardStepConnectivity:
		s = m.viewConnectivity()
	case WizardStepServerConfig:
		s = m.viewServerConfig()
	case WizardStepBridges:
		s = m.viewBridges()
	case WizardStepConfirm:
		s = m.viewConfirm()
	case WizardStepComplete:
		s = m.viewComplete()
	}

	return BoxStyle.Render(s)
}

func (m *WizardModel) viewWelcome() string {
	s := TitleStyle.Render("muxbee") + "\n\n"
	s += "Self-hosted Matrix server with messaging bridges.\n"
	s += "This wizard will configure your server.\n\n"
	s += HelpStyle.Render(RenderKey("enter", "start") + "  " + RenderKey("q", "quit"))
	return s
}

func (m *WizardModel) viewConnectivity() string {
	s := TitleStyle.Render("How will you access this?") + "\n\n"

	options := []struct {
		name string
		desc string
	}{
		{"Local", "localhost only, good for testing"},
		{"Private network", "VPN, Tailscale, or LAN IP"},
		{"Public HTTPS", "domain with automatic SSL"},
	}

	for i, opt := range options {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		if i == m.cursor {
			s += ListItemSelectedStyle.Render(cursor+opt.name) + "\n"
			s += "     " + SubtitleStyle.Render(opt.desc) + "\n"
		} else {
			s += ListItemStyle.Render(cursor+opt.name) + "\n"
		}
	}

	s += "\n" + HelpStyle.Render(RenderKey("↑/↓", "move") + "  " + RenderKey("enter", "select") + "  " + RenderKey("esc", "back"))
	return s
}

func (m *WizardModel) viewServerConfig() string {
	var s string

	if m.connectivityMode == "private" {
		s = TitleStyle.Render("Server Address") + "\n\n"
		s += "Enter your server's IP or hostname:\n\n"
		s += m.serverInput.View() + "\n\n"
		s += SubtitleStyle.Render("Examples: 192.168.1.50, myserver.local, machine.tailnet.ts.net") + "\n"
	} else {
		s = TitleStyle.Render("HTTPS Setup") + "\n\n"
		s += "Domain:\n"
		s += m.domainInput.View() + "\n\n"
		s += "Email (for Let's Encrypt):\n"
		s += m.emailInput.View() + "\n\n"
		s += SubtitleStyle.Render("Requires ports 80 + 443 open, DNS pointing to this server") + "\n"
	}

	if m.err != nil {
		s += "\n" + ErrorStyle.Render(m.err.Error()) + "\n"
	}

	s += "\n" + HelpStyle.Render(RenderKey("tab", "next field") + "  " + RenderKey("enter", "continue") + "  " + RenderKey("esc", "back"))
	return s
}

func (m *WizardModel) viewBridges() string {
	s := TitleStyle.Render("Select Bridges") + "\n\n"

	bridges := []struct {
		name string
		desc string
	}{
		{"whatsapp", "WhatsApp (QR code)"},
		{"signal", "Signal (QR code)"},
		{"discord", "Discord (QR or token)"},
		{"telegram", "Telegram (requires API key)"},
		{"gmessages", "Google Messages (QR code)"},
	}

	for i, b := range bridges {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		checked := "[ ]"
		for _, sel := range m.selectedBridges {
			if sel == b.name {
				checked = "[x]"
				break
			}
		}

		line := cursor + checked + " " + b.name
		if i == m.cursor {
			s += ListItemSelectedStyle.Render(line) + "  " + SubtitleStyle.Render(b.desc) + "\n"
		} else {
			s += ListItemStyle.Render(line) + "\n"
		}
	}

	s += "\n" + HelpStyle.Render(RenderKey("space", "toggle") + "  " + RenderKey("enter", "continue") + "  " + RenderKey("esc", "back"))
	return s
}

func (m *WizardModel) viewConfirm() string {
	s := TitleStyle.Render("Confirm") + "\n\n"

	s += "Mode:    " + m.connectivityMode + "\n"
	s += "Server:  " + m.serverName + "\n"
	if m.connectivityMode == "public" {
		s += "Email:   " + m.email + "\n"
	}

	s += "Bridges: "
	if len(m.selectedBridges) == 0 {
		s += "none"
	} else {
		s += strings.Join(m.selectedBridges, ", ")
	}
	s += "\n"

	if m.err != nil {
		s += "\n" + ErrorStyle.Render(m.err.Error()) + "\n"
	}

	s += "\n" + HelpStyle.Render(RenderKey("enter", "create") + "  " + RenderKey("esc", "back"))
	return s
}

func (m *WizardModel) viewComplete() string {
	s := SuccessStyle.Render("Done!") + "\n\n"
	s += "Configuration created.\n\n"
	s += "Press " + HelpKeyStyle.Render("enter") + " to continue.\n"
	return s
}

type wizardCompleteMsg struct {
	err error
}

func (m *WizardModel) createConfigCmd() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.NewDefaultConfig()
		if err != nil {
			return wizardCompleteMsg{err: err}
		}

		cfg.ServerName = m.serverName
		cfg.ConnectivityMode = m.connectivityMode
		cfg.EnabledBridges = m.selectedBridges

		if m.connectivityMode == "public" {
			cfg.HTTPS.Enabled = true
			cfg.HTTPS.Domain = m.domain
			cfg.HTTPS.Email = m.email
		}

		cfg.EnsureAvailablePorts()

		if err := cfg.Save(); err != nil {
			return wizardCompleteMsg{err: err}
		}

		gen := generator.New()
		if err := gen.GenerateAll(cfg); err != nil {
			return wizardCompleteMsg{err: err}
		}

		compose := docker.New(cfg)
		if err := compose.WriteComposeFile(); err != nil {
			return wizardCompleteMsg{err: err}
		}

		m.step = WizardStepComplete
		m.err = nil
		return wizardCompleteMsg{}
	}
}
