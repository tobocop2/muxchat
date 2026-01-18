package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tobias/muxchat/internal/config"
	"github.com/tobias/muxchat/internal/docker"
	"github.com/tobias/muxchat/internal/generator"
)

// SettingsModel handles the settings screen
type SettingsModel struct {
	cursor      int
	mode        int // 0=local, 1=private, 2=public
	serverInput textinput.Model
	domainInput textinput.Model
	emailInput  textinput.Model
	editing     bool      // true when editing a text field
	editField   int       // which field is being edited
	err         error
	saved       bool
}

// NewSettingsModel creates a new settings model
func NewSettingsModel(cfg *config.Config) *SettingsModel {
	serverInput := textinput.New()
	serverInput.Placeholder = "192.168.1.50 or hostname"
	serverInput.CharLimit = 256
	serverInput.Width = 40

	domainInput := textinput.New()
	domainInput.Placeholder = "chat.example.com"
	domainInput.CharLimit = 256
	domainInput.Width = 40

	emailInput := textinput.New()
	emailInput.Placeholder = "you@example.com"
	emailInput.CharLimit = 256
	emailInput.Width = 40

	m := &SettingsModel{
		serverInput: serverInput,
		domainInput: domainInput,
		emailInput:  emailInput,
	}

	// Initialize from config
	if cfg != nil {
		switch cfg.ConnectivityMode {
		case "local":
			m.mode = 0
		case "private":
			m.mode = 1
			m.serverInput.SetValue(cfg.ServerName)
		case "public":
			m.mode = 2
			m.domainInput.SetValue(cfg.ServerName)
			m.emailInput.SetValue(cfg.HTTPS.Email)
		}
	}

	return m
}

// Init initializes the settings
func (m *SettingsModel) Init() tea.Cmd {
	return nil
}

// Update handles settings events
func (m *SettingsModel) Update(msg tea.Msg, cfg *config.Config, compose *docker.Compose) (*SettingsModel, tea.Cmd) {
	// Handle text input when editing
	if m.editing {
		return m.updateEditing(msg, cfg, compose)
	}

	switch msg := msg.(type) {
	case settingsSavedMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.saved = true
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			maxCursor := 0
			switch m.mode {
			case 0: // local
				maxCursor = 0
			case 1: // private
				maxCursor = 1
			case 2: // public
				maxCursor = 2
			}
			if m.cursor < maxCursor {
				m.cursor++
			}
		case "left", "h":
			if m.cursor == 0 && m.mode > 0 {
				m.mode--
				m.saved = false
			}
		case "right", "l":
			if m.cursor == 0 && m.mode < 2 {
				m.mode++
				m.saved = false
			}
		case "enter":
			// Edit the selected field
			if m.cursor > 0 {
				m.editing = true
				m.editField = m.cursor
				m.saved = false
				if m.mode == 1 && m.cursor == 1 {
					m.serverInput.Focus()
					return m, textinput.Blink
				} else if m.mode == 2 {
					if m.cursor == 1 {
						m.domainInput.Focus()
					} else if m.cursor == 2 {
						m.emailInput.Focus()
					}
					return m, textinput.Blink
				}
			}
		case "s":
			// Save settings
			return m, m.saveCmd(cfg, compose)
		case "esc", "q":
			// Return to dashboard handled by app.go
		}
	}
	return m, nil
}

func (m *SettingsModel) updateEditing(msg tea.Msg, cfg *config.Config, compose *docker.Compose) (*SettingsModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "esc":
			m.editing = false
			m.serverInput.Blur()
			m.domainInput.Blur()
			m.emailInput.Blur()
			return m, nil
		}
	}

	// Update the active input
	if m.mode == 1 {
		m.serverInput, cmd = m.serverInput.Update(msg)
	} else if m.mode == 2 {
		if m.editField == 1 {
			m.domainInput, cmd = m.domainInput.Update(msg)
		} else {
			m.emailInput, cmd = m.emailInput.Update(msg)
		}
	}

	return m, cmd
}

// View renders the settings
func (m *SettingsModel) View() string {
	var s string

	s += TitleStyle.Render("Settings") + "\n\n"

	// Mode selector
	modes := []string{"Local", "Private", "Public HTTPS"}
	s += "Mode: "
	for i, mode := range modes {
		if i == m.mode {
			s += ListItemSelectedStyle.Render("[" + mode + "]")
		} else {
			s += SubtitleStyle.Render(" " + mode + " ")
		}
		if i < len(modes)-1 {
			s += " "
		}
	}
	if m.cursor == 0 {
		s += "  " + SubtitleStyle.Render("← →")
	}
	s += "\n\n"

	// Mode-specific fields
	switch m.mode {
	case 0: // local
		s += SubtitleStyle.Render("Server: localhost") + "\n"
	case 1: // private
		cursor := "  "
		if m.cursor == 1 {
			cursor = "> "
		}
		s += cursor + "Server: "
		if m.editing && m.editField == 1 {
			s += m.serverInput.View()
		} else {
			val := m.serverInput.Value()
			if val == "" {
				val = SubtitleStyle.Render("(not set)")
			}
			s += val
		}
		s += "\n"
	case 2: // public
		// Domain
		cursor := "  "
		if m.cursor == 1 {
			cursor = "> "
		}
		s += cursor + "Domain: "
		if m.editing && m.editField == 1 {
			s += m.domainInput.View()
		} else {
			val := m.domainInput.Value()
			if val == "" {
				val = SubtitleStyle.Render("(not set)")
			}
			s += val
		}
		s += "\n"

		// Email
		cursor = "  "
		if m.cursor == 2 {
			cursor = "> "
		}
		s += cursor + "Email:  "
		if m.editing && m.editField == 2 {
			s += m.emailInput.View()
		} else {
			val := m.emailInput.Value()
			if val == "" {
				val = SubtitleStyle.Render("(not set)")
			}
			s += val
		}
		s += "\n"
	}

	s += "\n"

	if m.err != nil {
		s += ErrorStyle.Render(m.err.Error()) + "\n\n"
	}

	if m.saved {
		s += SuccessStyle.Render("Saved! Restart services to apply.") + "\n\n"
	}

	// Help
	if m.editing {
		s += HelpStyle.Render(RenderKey("enter", "done") + "  " + RenderKey("esc", "cancel"))
	} else {
		s += HelpStyle.Render(RenderKey("↑/↓", "navigate") + "  " + RenderKey("enter", "edit") + "  " + RenderKey("s", "save") + "  " + RenderKey("esc", "back"))
	}

	return s
}

type settingsSavedMsg struct {
	err error
}

func (m *SettingsModel) saveCmd(cfg *config.Config, compose *docker.Compose) tea.Cmd {
	return func() tea.Msg {
		if cfg == nil {
			return settingsSavedMsg{err: nil}
		}

		// Update config based on mode
		switch m.mode {
		case 0: // local
			cfg.ConnectivityMode = "local"
			cfg.ServerName = "localhost"
			cfg.HTTPS.Enabled = false
			cfg.HTTPS.Domain = ""
			cfg.HTTPS.Email = ""
		case 1: // private
			server := strings.TrimSpace(m.serverInput.Value())
			if server == "" {
				server = "localhost"
			}
			cfg.ConnectivityMode = "private"
			cfg.ServerName = server
			cfg.HTTPS.Enabled = false
			cfg.HTTPS.Domain = ""
			cfg.HTTPS.Email = ""
		case 2: // public
			domain := strings.TrimSpace(m.domainInput.Value())
			email := strings.TrimSpace(m.emailInput.Value())
			if domain == "" || email == "" {
				return settingsSavedMsg{err: nil} // Silently skip if not complete
			}
			cfg.ConnectivityMode = "public"
			cfg.ServerName = domain
			cfg.HTTPS.Enabled = true
			cfg.HTTPS.Domain = domain
			cfg.HTTPS.Email = email
		}

		// Save config
		if err := cfg.Save(); err != nil {
			return settingsSavedMsg{err: err}
		}

		// Regenerate all config files
		gen := generator.New()
		if err := gen.GenerateAll(cfg); err != nil {
			return settingsSavedMsg{err: err}
		}

		// Write docker-compose
		if compose != nil {
			if err := compose.WriteComposeFile(); err != nil {
				return settingsSavedMsg{err: err}
			}
		}

		return settingsSavedMsg{}
	}
}
