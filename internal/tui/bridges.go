package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tobias/muxbee/internal/bridges"
	"github.com/tobias/muxbee/internal/config"
	"github.com/tobias/muxbee/internal/docker"
	"github.com/tobias/muxbee/internal/generator"
)

// BridgesModel handles the bridges screen
type BridgesModel struct {
	cursor      int
	bridges     []bridges.BridgeInfo
	showInfo    bool
	infoBridge  *bridges.BridgeInfo
	isLoading   bool
	loadingStep string
	spinnerIdx  int
	lastError   error
	resultChan  chan bridgeToggledMsg

	// Credential input state
	credentialInput  bool
	credentialBridge string
	credentialStep   int // 0 = API ID, 1 = API Hash
	apiIDInput       textinput.Model
	apiHashInput     textinput.Model
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NewBridgesModel creates a new bridges model
func NewBridgesModel() *BridgesModel {
	apiID := textinput.New()
	apiID.Placeholder = "Enter API ID (numbers)"
	apiID.CharLimit = 20

	apiHash := textinput.New()
	apiHash.Placeholder = "Enter API Hash"
	apiHash.CharLimit = 64

	return &BridgesModel{
		cursor:       0,
		bridges:      bridges.List(),
		resultChan:   make(chan bridgeToggledMsg, 1),
		apiIDInput:   apiID,
		apiHashInput: apiHash,
	}
}

// Init initializes the bridges screen
func (m *BridgesModel) Init() tea.Cmd {
	return nil
}

// Update handles bridges screen events
func (m *BridgesModel) Update(msg tea.Msg, cfg *config.Config) (*BridgesModel, tea.Cmd) {
	// Handle credential input mode
	if m.credentialInput {
		return m.updateCredentialInput(msg, cfg)
	}

	switch msg := msg.(type) {
	case bridgeProgressMsg:
		m.loadingStep = msg.step
		return m, nil

	case bridgeSpinnerMsg:
		if m.isLoading {
			m.spinnerIdx = (m.spinnerIdx + 1) % len(spinnerFrames)
			return m, m.spinnerTick()
		}
		return m, nil

	case bridgeToggledMsg:
		m.isLoading = false
		m.loadingStep = ""
		m.lastError = msg.err
		return m, nil

	case bridgeCheckResultMsg:
		// Check if there's a result waiting
		select {
		case result := <-m.resultChan:
			m.isLoading = false
			m.loadingStep = ""
			m.lastError = result.err
			return m, nil
		default:
			// No result yet, keep checking
			if m.isLoading {
				return m, tea.Batch(m.spinnerTick(), m.checkResult())
			}
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.bridges)-1 {
				m.cursor++
			}
		case "enter", " ":
			// Block starting new operations while one is in progress
			if m.isLoading {
				return m, nil
			}
			if cfg != nil && m.cursor < len(m.bridges) {
				bridge := m.bridges[m.cursor]

				// Check if bridge requires API credentials and they're not configured
				if !cfg.IsBridgeEnabled(bridge.Name) && bridge.RequiresAPICredentials {
					if !m.hasRequiredCredentials(cfg, bridge.Name) {
						// Enter credential input mode
						m.startCredentialInput(bridge.Name)
						return m, textinput.Blink
					}
				}

				m.startBridgeToggle(cfg, bridge.Name)
				return m, tea.Batch(m.spinnerTick(), m.checkResult())
			}
		case "i":
			if m.cursor < len(m.bridges) {
				m.showInfo = !m.showInfo
				if m.showInfo {
					b := m.bridges[m.cursor]
					m.infoBridge = &b
				} else {
					m.infoBridge = nil
				}
			}
		case "esc":
			m.showInfo = false
			m.infoBridge = nil
		}
	}
	return m, nil
}

// updateCredentialInput handles input when in credential entry mode
func (m *BridgesModel) updateCredentialInput(msg tea.Msg, cfg *config.Config) (*BridgesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Cancel credential input
			m.credentialInput = false
			m.credentialStep = 0
			m.apiIDInput.SetValue("")
			m.apiHashInput.SetValue("")
			return m, nil

		case "enter":
			if m.credentialStep == 0 {
				// Move to API Hash input
				apiID := strings.TrimSpace(m.apiIDInput.Value())
				if apiID == "" {
					m.lastError = errEmptyAPIID
					return m, nil
				}
				m.credentialStep = 1
				m.apiIDInput.Blur()
				m.apiHashInput.Focus()
				m.lastError = nil
				return m, textinput.Blink
			} else {
				// Submit credentials
				apiHash := strings.TrimSpace(m.apiHashInput.Value())
				if apiHash == "" {
					m.lastError = errEmptyAPIHash
					return m, nil
				}

				// Save credentials to config immediately
				apiID := strings.TrimSpace(m.apiIDInput.Value())
				m.saveCredentials(cfg, m.credentialBridge, apiID, apiHash)

				// Save config to disk before starting the bridge
				if err := cfg.Save(); err != nil {
					m.lastError = err
					return m, nil
				}

				// Exit credential mode and start bridge toggle
				m.credentialInput = false
				m.credentialStep = 0
				m.apiIDInput.SetValue("")
				m.apiHashInput.SetValue("")
				m.lastError = nil

				m.startBridgeToggle(cfg, m.credentialBridge)
				return m, tea.Batch(m.spinnerTick(), m.checkResult())
			}

		case "tab", "shift+tab":
			// Toggle between inputs
			if m.credentialStep == 0 {
				m.credentialStep = 1
				m.apiIDInput.Blur()
				m.apiHashInput.Focus()
			} else {
				m.credentialStep = 0
				m.apiHashInput.Blur()
				m.apiIDInput.Focus()
			}
			return m, textinput.Blink
		}
	}

	// Update the active text input
	var cmd tea.Cmd
	if m.credentialStep == 0 {
		m.apiIDInput, cmd = m.apiIDInput.Update(msg)
	} else {
		m.apiHashInput, cmd = m.apiHashInput.Update(msg)
	}
	return m, cmd
}

// startCredentialInput enters credential input mode
func (m *BridgesModel) startCredentialInput(bridgeName string) {
	m.credentialInput = true
	m.credentialBridge = bridgeName
	m.credentialStep = 0
	m.lastError = nil
	m.apiIDInput.SetValue("")
	m.apiHashInput.SetValue("")
	m.apiIDInput.Focus()
	m.apiHashInput.Blur()
}

// saveCredentials saves the entered credentials to config
func (m *BridgesModel) saveCredentials(cfg *config.Config, bridgeName, apiID, apiHash string) {
	switch bridgeName {
	case "telegram":
		cfg.Telegram = &config.TelegramConfig{
			APIID:   apiID,
			APIHash: apiHash,
		}
	}
}

// startBridgeToggle starts the async bridge toggle operation
func (m *BridgesModel) startBridgeToggle(cfg *config.Config, bridgeName string) {
	m.isLoading = true
	m.lastError = nil
	m.spinnerIdx = 0
	if cfg.IsBridgeEnabled(bridgeName) {
		m.loadingStep = "Stopping " + bridgeName
	} else {
		m.loadingStep = "Pulling & starting " + bridgeName
	}
	go m.toggleBridgeBackground(cfg, bridgeName)
}

// spinnerTick returns a command that ticks the spinner
func (m *BridgesModel) spinnerTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return bridgeSpinnerMsg{}
	})
}

// checkResult returns a command that checks for async results
func (m *BridgesModel) checkResult() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return bridgeCheckResultMsg{}
	})
}

// View renders the bridges screen
func (m *BridgesModel) View(cfg *config.Config) string {
	// Credential input view
	if m.credentialInput {
		return m.viewCredentialInput()
	}

	var s string

	s += TitleStyle.Render("Messaging Bridges") + "\n\n"

	if m.isLoading {
		spinner := spinnerFrames[m.spinnerIdx]
		s += SubtitleStyle.Render(spinner+" "+m.loadingStep+"...") + "\n\n"
	}

	if m.lastError != nil {
		s += ErrorStyle.Render("Error: "+m.lastError.Error()) + "\n\n"
	}

	if m.showInfo && m.infoBridge != nil {
		return m.viewBridgeInfo(cfg)
	}

	// List bridges
	for i, bridge := range m.bridges {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		enabled := " "
		if cfg != nil && cfg.IsBridgeEnabled(bridge.Name) {
			enabled = SuccessStyle.Render("[x]")
		} else {
			enabled = "[ ]"
		}

		if i == m.cursor {
			s += ListItemSelectedStyle.Render(cursor+" "+enabled+" "+bridge.Name) + "\n"
		} else {
			s += ListItemStyle.Render(cursor+" "+enabled+" "+bridge.Name) + "\n"
		}
	}

	s += "\n"
	s += HelpStyle.Render(RenderKey("i", "info") + "  " + RenderKey("enter", "toggle"))

	return s
}

// viewCredentialInput renders the credential input form
func (m *BridgesModel) viewCredentialInput() string {
	var s string

	s += TitleStyle.Render("Configure "+m.credentialBridge) + "\n\n"

	switch m.credentialBridge {
	case "telegram":
		s += "Telegram requires API credentials from https://my.telegram.org\n\n"
		s += "1. Go to https://my.telegram.org\n"
		s += "2. Log in with your phone number\n"
		s += "3. Go to 'API development tools'\n"
		s += "4. Create a new application\n"
		s += "5. Copy the api_id and api_hash\n\n"
	}

	if m.lastError != nil {
		s += ErrorStyle.Render("Error: "+m.lastError.Error()) + "\n\n"
	}

	// API ID input
	if m.credentialStep == 0 {
		s += SubtitleStyle.Render("API ID:") + "\n"
	} else {
		s += "API ID:\n"
	}
	s += m.apiIDInput.View() + "\n\n"

	// API Hash input
	if m.credentialStep == 1 {
		s += SubtitleStyle.Render("API Hash:") + "\n"
	} else {
		s += "API Hash:\n"
	}
	s += m.apiHashInput.View() + "\n\n"

	s += HelpStyle.Render(RenderKey("enter", "next/submit") + "  " + RenderKey("tab", "switch field") + "  " + RenderKey("esc", "cancel"))

	return s
}

func (m *BridgesModel) viewBridgeInfo(cfg *config.Config) string {
	var s string
	b := m.infoBridge

	s += TitleStyle.Render(b.Name) + "\n"
	s += SubtitleStyle.Render(b.Description) + "\n\n"

	if b.HasNote() {
		s += SubtitleStyle.Render("Note:") + "\n"
		s += b.Note + "\n\n"
	}

	s += TitleStyle.Render("Login Instructions:") + "\n"
	s += b.LoginInstructions + "\n"

	s += "\n" + HelpStyle.Render(RenderKey("esc", "back"))

	return s
}

type bridgeToggledMsg struct {
	err error
}

type bridgeProgressMsg struct {
	step string
}

type bridgeSpinnerMsg struct{}

type bridgeCheckResultMsg struct{}

// Custom errors
type credentialError string

func (e credentialError) Error() string { return string(e) }

var (
	errEmptyAPIID   = credentialError("API ID is required")
	errEmptyAPIHash = credentialError("API Hash is required")
)

// toggleBridgeBackground runs the toggle operation in a goroutine
func (m *BridgesModel) toggleBridgeBackground(cfg *config.Config, bridgeName string) {
	var err error

	enabling := !cfg.IsBridgeEnabled(bridgeName)

	// Update config
	if enabling {
		cfg.EnableBridge(bridgeName)
	} else {
		cfg.DisableBridge(bridgeName)
	}

	// Save config
	if err = cfg.Save(); err != nil {
		m.resultChan <- bridgeToggledMsg{err: err}
		return
	}

	// Regenerate all configs
	gen := generator.New()
	if err = gen.GenerateAll(cfg); err != nil {
		m.resultChan <- bridgeToggledMsg{err: err}
		return
	}

	d := docker.New(cfg)

	if enabling {
		// IMPORTANT: Restart Synapse FIRST so it loads the new registration file
		// before the bridge tries to connect
		if err = d.RestartQuiet("synapse"); err != nil {
			m.resultChan <- bridgeToggledMsg{err: err}
			return
		}

		// Give Synapse a moment to become ready
		time.Sleep(3 * time.Second)

		// Pull the new bridge image quietly
		profiles := []string{bridgeName}
		d.PullQuiet(profiles) // Ignore error, Up will pull if needed

		// Start the bridge service (Synapse already knows about it)
		allProfiles := docker.GetProfiles(cfg)
		if err = d.UpQuiet(allProfiles); err != nil {
			m.resultChan <- bridgeToggledMsg{err: err}
			return
		}

		// Verify bridge started successfully
		time.Sleep(5 * time.Second)
		serviceName := "mautrix-" + bridgeName
		if !d.IsServiceRunning(serviceName) {
			m.resultChan <- bridgeToggledMsg{err: fmt.Errorf("%s failed to start - check logs with 'muxbee logs %s'", bridgeName, serviceName)}
			return
		}
	} else {
		// Stop and remove the bridge container
		serviceName := "mautrix-" + bridgeName
		d.StopService(serviceName) // Ignore error, might not be running

		// Restart Synapse to update bridge registrations
		if err = d.RestartQuiet("synapse"); err != nil {
			m.resultChan <- bridgeToggledMsg{err: err}
			return
		}
	}

	m.resultChan <- bridgeToggledMsg{}
}

// hasRequiredCredentials checks if required API credentials are configured
func (m *BridgesModel) hasRequiredCredentials(cfg *config.Config, bridgeName string) bool {
	switch bridgeName {
	case "telegram":
		return cfg.Telegram != nil && cfg.Telegram.APIID != "" && cfg.Telegram.APIHash != ""
	}
	return true
}
