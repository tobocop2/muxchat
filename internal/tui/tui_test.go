package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tobias/muxbee/internal/config"
	"github.com/tobias/muxbee/internal/docker"
)

// Test Dashboard Model

func TestNewDashboardModel(t *testing.T) {
	m := NewDashboardModel()
	if m == nil {
		t.Fatal("expected non-nil model")
	}
	if m.isLoading {
		t.Error("expected isLoading to be false initially")
	}
	if len(m.services) != 0 {
		t.Errorf("expected empty services, got %d", len(m.services))
	}
}

func TestDashboardModel_Update_KeyS_StartsLoading(t *testing.T) {
	m := NewDashboardModel()
	cfg := &config.Config{ServerName: "test.local"}

	// Create a mock compose (we won't actually call docker)
	// The model just needs a non-nil compose to proceed

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	newM, _ := m.Update(msg, cfg, nil)

	// With nil compose, it shouldn't start loading
	if newM.isLoading {
		t.Error("expected isLoading to be false with nil compose")
	}
}

func TestDashboardModel_Update_ServicesUpdated(t *testing.T) {
	m := NewDashboardModel()
	m.isLoading = true
	m.loadingOp = "Starting"

	services := []docker.ServiceStatus{
		{Name: "muxbee-synapse-1", Running: true, Health: "healthy"},
		{Name: "muxbee-postgres-1", Running: true},
	}

	msg := servicesUpdatedMsg{services: services}
	newM, _ := m.Update(msg, nil, nil)

	if newM.isLoading {
		t.Error("expected isLoading to be false after servicesUpdatedMsg")
	}
	if len(newM.services) != 2 {
		t.Errorf("expected 2 services, got %d", len(newM.services))
	}
	if newM.lastUpdated.IsZero() {
		t.Error("expected lastUpdated to be set")
	}
}

func TestDashboardModel_View_NoConfig(t *testing.T) {
	m := NewDashboardModel()
	view := m.View(nil)

	if !strings.Contains(view, "Config not loaded") {
		t.Error("expected view to show config not loaded message")
	}
}

func TestDashboardModel_View_WithConfig(t *testing.T) {
	m := NewDashboardModel()
	cfg := &config.Config{
		ServerName:       "test.local",
		ConnectivityMode: "local",
		EnabledBridges:   []string{"whatsapp", "signal"},
	}

	view := m.View(cfg)

	if !strings.Contains(view, "muxbee") {
		t.Error("expected view to contain 'muxbee'")
	}
	if !strings.Contains(view, "test.local") {
		t.Error("expected view to contain server name")
	}
	if !strings.Contains(view, "local") {
		t.Error("expected view to contain connectivity mode")
	}
	if !strings.Contains(view, "whatsapp") {
		t.Error("expected view to contain enabled bridges")
	}
}

func TestDashboardModel_View_Loading(t *testing.T) {
	m := NewDashboardModel()
	m.isLoading = true
	m.loadingOp = "Starting"

	cfg := &config.Config{
		ServerName:       "test.local",
		ConnectivityMode: "local",
	}

	view := m.View(cfg)

	if !strings.Contains(view, "Starting...") {
		t.Error("expected view to show loading operation")
	}
}

func TestDashboardModel_View_Services(t *testing.T) {
	m := NewDashboardModel()
	m.services = []docker.ServiceStatus{
		{Name: "muxbee-synapse-1", Running: true, Health: "healthy"},
		{Name: "muxbee-postgres-1", Running: false},
	}

	cfg := &config.Config{
		ServerName:       "test.local",
		ConnectivityMode: "local",
	}

	view := m.View(cfg)

	if !strings.Contains(view, "synapse") {
		t.Error("expected view to contain synapse service")
	}
	if !strings.Contains(view, "postgres") {
		t.Error("expected view to contain postgres service")
	}
}

// Test Bridges Model

func TestNewBridgesModel(t *testing.T) {
	m := NewBridgesModel()
	if m == nil {
		t.Fatal("expected non-nil model")
	}
	if len(m.bridges) == 0 {
		t.Error("expected bridges to be loaded")
	}
	if m.cursor != 0 {
		t.Error("expected cursor to start at 0")
	}
}

func TestBridgesModel_Update_Navigation(t *testing.T) {
	m := NewBridgesModel()
	numBridges := len(m.bridges)

	// Test down navigation
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newM, _ := m.Update(msg, nil)
	if newM.cursor != 1 {
		t.Errorf("expected cursor to be 1 after j, got %d", newM.cursor)
	}

	// Test down with 'down' key
	msg = tea.KeyMsg{Type: tea.KeyDown}
	newM, _ = newM.Update(msg, nil)
	if newM.cursor != 2 {
		t.Errorf("expected cursor to be 2 after down, got %d", newM.cursor)
	}

	// Test up navigation
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newM, _ = newM.Update(msg, nil)
	if newM.cursor != 1 {
		t.Errorf("expected cursor to be 1 after k, got %d", newM.cursor)
	}

	// Test boundary - can't go below 0
	newM.cursor = 0
	msg = tea.KeyMsg{Type: tea.KeyUp}
	newM, _ = newM.Update(msg, nil)
	if newM.cursor != 0 {
		t.Errorf("expected cursor to stay at 0, got %d", newM.cursor)
	}

	// Test boundary - can't go above max
	newM.cursor = numBridges - 1
	msg = tea.KeyMsg{Type: tea.KeyDown}
	newM, _ = newM.Update(msg, nil)
	if newM.cursor != numBridges-1 {
		t.Errorf("expected cursor to stay at %d, got %d", numBridges-1, newM.cursor)
	}
}

func TestBridgesModel_Update_ToggleInfo(t *testing.T) {
	m := NewBridgesModel()

	// Press 'i' to show info
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}
	newM, _ := m.Update(msg, nil)
	if !newM.showInfo {
		t.Error("expected showInfo to be true after i")
	}
	if newM.infoBridge == nil {
		t.Error("expected infoBridge to be set")
	}

	// Press 'i' again to hide
	newM, _ = newM.Update(msg, nil)
	if newM.showInfo {
		t.Error("expected showInfo to be false after second i")
	}

	// Press 'i' then 'esc'
	newM, _ = newM.Update(msg, nil)
	if !newM.showInfo {
		t.Error("expected showInfo to be true")
	}
	msg = tea.KeyMsg{Type: tea.KeyEscape}
	newM, _ = newM.Update(msg, nil)
	if newM.showInfo {
		t.Error("expected showInfo to be false after esc")
	}
}

func TestBridgesModel_View(t *testing.T) {
	m := NewBridgesModel()
	cfg := &config.Config{
		EnabledBridges: []string{"whatsapp"},
	}

	view := m.View(cfg)

	if !strings.Contains(view, "Messaging Bridges") {
		t.Error("expected view to contain title")
	}
	if !strings.Contains(view, "whatsapp") {
		t.Error("expected view to contain whatsapp")
	}
	if !strings.Contains(view, "[x]") {
		t.Error("expected view to show whatsapp as enabled")
	}
}

func TestBridgesModel_View_Loading(t *testing.T) {
	m := NewBridgesModel()
	m.isLoading = true
	m.loadingStep = "Pulling & starting whatsapp"

	view := m.View(nil)

	if !strings.Contains(view, "Pulling & starting whatsapp") {
		t.Error("expected view to show loading step")
	}
}

func TestBridgesModel_View_Error(t *testing.T) {
	m := NewBridgesModel()
	m.lastError = errEmptyAPIID

	view := m.View(nil)

	if !strings.Contains(view, "API ID is required") {
		t.Error("expected view to show error message")
	}
}

func TestBridgesModel_HasRequiredCredentials(t *testing.T) {
	m := NewBridgesModel()

	// Telegram without config
	cfg := &config.Config{}
	if m.hasRequiredCredentials(cfg, "telegram") {
		t.Error("expected false for telegram without credentials")
	}

	// Telegram with config
	cfg.Telegram = &config.TelegramConfig{
		APIID:   "12345",
		APIHash: "abcdef",
	}
	if !m.hasRequiredCredentials(cfg, "telegram") {
		t.Error("expected true for telegram with credentials")
	}

	// Non-telegram bridge doesn't require credentials
	if !m.hasRequiredCredentials(cfg, "whatsapp") {
		t.Error("expected true for whatsapp (no credentials required)")
	}
}

func TestBridgesModel_CredentialInput(t *testing.T) {
	m := NewBridgesModel()

	// Start credential input
	m.startCredentialInput("telegram")

	if !m.credentialInput {
		t.Error("expected credentialInput to be true")
	}
	if m.credentialBridge != "telegram" {
		t.Errorf("expected credentialBridge to be 'telegram', got '%s'", m.credentialBridge)
	}
	if m.credentialStep != 0 {
		t.Error("expected credentialStep to be 0")
	}
}

func TestBridgesModel_ViewCredentialInput(t *testing.T) {
	m := NewBridgesModel()
	m.startCredentialInput("telegram")

	view := m.viewCredentialInput()

	if !strings.Contains(view, "Configure telegram") {
		t.Error("expected view to contain bridge name")
	}
	if !strings.Contains(view, "my.telegram.org") {
		t.Error("expected view to contain telegram instructions")
	}
	if !strings.Contains(view, "API ID") {
		t.Error("expected view to contain API ID field")
	}
	if !strings.Contains(view, "API Hash") {
		t.Error("expected view to contain API Hash field")
	}
}

// Test Logs Model

func TestNewLogsModel(t *testing.T) {
	m := NewLogsModel()
	if m == nil {
		t.Fatal("expected non-nil model")
	}
	if len(m.logs) != 0 {
		t.Error("expected empty logs initially")
	}
	if m.paused {
		t.Error("expected paused to be false initially")
	}
}

func TestLogsModel_Update_TogglePause(t *testing.T) {
	m := NewLogsModel()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	newM, _ := m.Update(msg, nil)
	if !newM.paused {
		t.Error("expected paused to be true after p")
	}

	newM, _ = newM.Update(msg, nil)
	if newM.paused {
		t.Error("expected paused to be false after second p")
	}
}

func TestLogsModel_Update_ClearLogs(t *testing.T) {
	m := NewLogsModel()
	m.logs = []logEntry{
		{service: "synapse", message: "test log"},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	newM, _ := m.Update(msg, nil)

	if len(newM.logs) != 0 {
		t.Error("expected logs to be cleared after c")
	}
}

func TestLogsModel_Update_CycleFilter(t *testing.T) {
	m := NewLogsModel()
	m.services = []string{"synapse", "postgres", "whatsapp"}

	// First filter press: select first service
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
	newM, _ := m.Update(msg, nil)
	if newM.filter != "synapse" {
		t.Errorf("expected filter 'synapse', got '%s'", newM.filter)
	}

	// Second: select second service
	newM, _ = newM.Update(msg, nil)
	if newM.filter != "postgres" {
		t.Errorf("expected filter 'postgres', got '%s'", newM.filter)
	}

	// Keep pressing until wrap around
	newM, _ = newM.Update(msg, nil) // whatsapp
	newM, _ = newM.Update(msg, nil) // back to all
	if newM.filter != "" {
		t.Errorf("expected filter to be empty (all), got '%s'", newM.filter)
	}
}

func TestLogsModel_View(t *testing.T) {
	m := NewLogsModel()

	view := m.View()

	if !strings.Contains(view, "Service Logs") {
		t.Error("expected view to contain title")
	}
	if !strings.Contains(view, "Filter: all") {
		t.Error("expected view to show filter status")
	}
}

func TestLogsModel_View_WithLogs(t *testing.T) {
	m := NewLogsModel()
	m.logs = []logEntry{
		{service: "synapse", message: "Server started"},
		{service: "postgres", message: "Ready to accept connections"},
	}

	view := m.View()

	if !strings.Contains(view, "synapse") {
		t.Error("expected view to contain synapse log")
	}
	if !strings.Contains(view, "postgres") {
		t.Error("expected view to contain postgres log")
	}
}

func TestLogsModel_View_Paused(t *testing.T) {
	m := NewLogsModel()
	m.paused = true

	view := m.View()

	if !strings.Contains(view, "PAUSED") {
		t.Error("expected view to show PAUSED status")
	}
}

func TestLogsModel_View_Filtered(t *testing.T) {
	m := NewLogsModel()
	m.filter = "synapse"

	view := m.View()

	if !strings.Contains(view, "Filter: synapse") {
		t.Error("expected view to show filter name")
	}
}

func TestLogsModel_FilteredLogs(t *testing.T) {
	m := NewLogsModel()
	m.logs = []logEntry{
		{service: "synapse", message: "log 1"},
		{service: "postgres", message: "log 2"},
		{service: "synapse", message: "log 3"},
	}

	// No filter - all logs
	filtered := m.filteredLogs()
	if len(filtered) != 3 {
		t.Errorf("expected 3 logs without filter, got %d", len(filtered))
	}

	// With filter
	m.filter = "synapse"
	filtered = m.filteredLogs()
	if len(filtered) != 2 {
		t.Errorf("expected 2 synapse logs, got %d", len(filtered))
	}
}

func TestParseLine(t *testing.T) {
	tests := []struct {
		input       string
		wantService string
		wantMessage string
	}{
		{
			input:       "muxbee-synapse-1  | Server starting",
			wantService: "synapse",
			wantMessage: "Server starting",
		},
		{
			input:       "muxbee-postgres-1 | Ready",
			wantService: "postgres",
			wantMessage: "Ready",
		},
		{
			input:       "no pipe here",
			wantService: "unknown",
			wantMessage: "no pipe here",
		},
	}

	for _, tt := range tests {
		entry := parseLine(tt.input)
		if entry.service != tt.wantService {
			t.Errorf("parseLine(%q).service = %q, want %q", tt.input, entry.service, tt.wantService)
		}
		if entry.message != tt.wantMessage {
			t.Errorf("parseLine(%q).message = %q, want %q", tt.input, entry.message, tt.wantMessage)
		}
	}
}

// Test Styles

func TestRenderKey(t *testing.T) {
	result := RenderKey("q", "quit")
	if !strings.Contains(result, "q") {
		t.Error("expected result to contain key")
	}
	if !strings.Contains(result, "quit") {
		t.Error("expected result to contain description")
	}
}

func TestRenderStatus(t *testing.T) {
	// Running and healthy
	result := RenderStatus(true, "healthy")
	if !strings.Contains(result, "running") {
		t.Error("expected 'running' for healthy service")
	}

	// Running but unhealthy
	result = RenderStatus(true, "unhealthy")
	if !strings.Contains(result, "unhealthy") {
		t.Error("expected 'unhealthy' in status")
	}

	// Stopped
	result = RenderStatus(false, "")
	if !strings.Contains(result, "stopped") {
		t.Error("expected 'stopped' for non-running service")
	}
}

// Test Wizard Model

func TestNewWizardModel(t *testing.T) {
	m := NewWizardModel()
	if m == nil {
		t.Fatal("expected non-nil model")
	}
	if m.step != WizardStepWelcome {
		t.Errorf("expected step to be WizardStepWelcome, got %d", m.step)
	}
	if m.connectivityMode != "local" {
		t.Errorf("expected connectivityMode to be 'local', got '%s'", m.connectivityMode)
	}
	if m.serverName != "localhost" {
		t.Errorf("expected serverName to be 'localhost', got '%s'", m.serverName)
	}
	if len(m.selectedBridges) != 0 {
		t.Errorf("expected empty selectedBridges, got %v", m.selectedBridges)
	}
}

func TestWizardModel_NextStep_WelcomeToConnectivity(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepWelcome

	newM, _ := m.nextStep()
	if newM.step != WizardStepConnectivity {
		t.Errorf("expected step to be WizardStepConnectivity, got %d", newM.step)
	}
	if newM.cursor != 0 {
		t.Errorf("expected cursor to reset to 0, got %d", newM.cursor)
	}
}

func TestWizardModel_NextStep_ConnectivityLocal(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepConnectivity
	m.cursor = 0 // Local mode

	newM, _ := m.nextStep()
	if newM.step != WizardStepBridges {
		t.Errorf("expected step to skip to WizardStepBridges for local, got %d", newM.step)
	}
	if newM.connectivityMode != "local" {
		t.Errorf("expected connectivityMode to be 'local', got '%s'", newM.connectivityMode)
	}
	if newM.serverName != "localhost" {
		t.Errorf("expected serverName to be 'localhost', got '%s'", newM.serverName)
	}
}

func TestWizardModel_NextStep_ConnectivityPrivate(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepConnectivity
	m.cursor = 1 // Private mode

	newM, _ := m.nextStep()
	if newM.step != WizardStepServerConfig {
		t.Errorf("expected step to be WizardStepServerConfig, got %d", newM.step)
	}
	if newM.connectivityMode != "private" {
		t.Errorf("expected connectivityMode to be 'private', got '%s'", newM.connectivityMode)
	}
}

func TestWizardModel_NextStep_ConnectivityPublic(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepConnectivity
	m.cursor = 2 // Public mode

	newM, _ := m.nextStep()
	if newM.step != WizardStepServerConfig {
		t.Errorf("expected step to be WizardStepServerConfig, got %d", newM.step)
	}
	if newM.connectivityMode != "public" {
		t.Errorf("expected connectivityMode to be 'public', got '%s'", newM.connectivityMode)
	}
}

func TestWizardModel_NextStep_BridgesToConfirm(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepBridges

	newM, _ := m.nextStep()
	if newM.step != WizardStepConfirm {
		t.Errorf("expected step to be WizardStepConfirm, got %d", newM.step)
	}
}

func TestWizardModel_PrevStep_ConnectivityToWelcome(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepConnectivity

	newM, _ := m.prevStep()
	if newM.step != WizardStepWelcome {
		t.Errorf("expected step to be WizardStepWelcome, got %d", newM.step)
	}
}

func TestWizardModel_PrevStep_BridgesToConnectivity_LocalMode(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepBridges
	m.connectivityMode = "local"

	newM, _ := m.prevStep()
	if newM.step != WizardStepConnectivity {
		t.Errorf("expected step to be WizardStepConnectivity, got %d", newM.step)
	}
}

func TestWizardModel_PrevStep_BridgesToServerConfig_PrivateMode(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepBridges
	m.connectivityMode = "private"

	newM, _ := m.prevStep()
	if newM.step != WizardStepServerConfig {
		t.Errorf("expected step to be WizardStepServerConfig, got %d", newM.step)
	}
}

func TestWizardModel_PrevStep_ConfirmToBridges(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepConfirm

	newM, _ := m.prevStep()
	if newM.step != WizardStepBridges {
		t.Errorf("expected step to be WizardStepBridges, got %d", newM.step)
	}
}

func TestWizardModel_ToggleSelection(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepBridges
	m.cursor = 0 // whatsapp

	// Toggle on
	m.toggleSelection()
	if len(m.selectedBridges) != 1 || m.selectedBridges[0] != "whatsapp" {
		t.Errorf("expected whatsapp to be selected, got %v", m.selectedBridges)
	}

	// Toggle off
	m.toggleSelection()
	if len(m.selectedBridges) != 0 {
		t.Errorf("expected empty selectedBridges after toggle off, got %v", m.selectedBridges)
	}
}

func TestWizardModel_ToggleSelection_MultipleBridges(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepBridges

	// Select whatsapp (cursor 0)
	m.cursor = 0
	m.toggleSelection()

	// Select signal (cursor 1)
	m.cursor = 1
	m.toggleSelection()

	// Select discord (cursor 2)
	m.cursor = 2
	m.toggleSelection()

	if len(m.selectedBridges) != 3 {
		t.Errorf("expected 3 bridges selected, got %d", len(m.selectedBridges))
	}

	// Deselect signal
	m.cursor = 1
	m.toggleSelection()
	if len(m.selectedBridges) != 2 {
		t.Errorf("expected 2 bridges after deselect, got %d", len(m.selectedBridges))
	}

	// Verify signal is gone
	for _, b := range m.selectedBridges {
		if b == "signal" {
			t.Error("expected signal to be deselected")
		}
	}
}

func TestWizardModel_Update_Navigation(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepConnectivity

	// Test down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newM, _ := m.Update(msg)
	if newM.cursor != 1 {
		t.Errorf("expected cursor to be 1 after j, got %d", newM.cursor)
	}

	// Test up
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newM, _ = newM.Update(msg)
	if newM.cursor != 0 {
		t.Errorf("expected cursor to be 0 after k, got %d", newM.cursor)
	}
}

func TestWizardModel_Update_SpaceToggle(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepBridges
	m.cursor = 0

	msg := tea.KeyMsg{Type: tea.KeySpace}
	newM, _ := m.Update(msg)

	if len(newM.selectedBridges) != 1 {
		t.Errorf("expected 1 bridge selected after space, got %d", len(newM.selectedBridges))
	}
}

func TestWizardModel_Update_QuitOnWelcome(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepWelcome

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)

	// Should return quit command
	if cmd == nil {
		t.Error("expected quit command on q at welcome step")
	}
}

func TestWizardModel_Update_QuitIgnoredOtherSteps(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepBridges

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)

	// Should not quit on other steps
	if cmd != nil {
		t.Error("expected no quit command on q at bridges step")
	}
}

func TestWizardModel_View_ReturnsContentForEachStep(t *testing.T) {
	m := NewWizardModel()

	// Test each step has view content
	steps := []WizardStep{
		WizardStepWelcome,
		WizardStepConnectivity,
		WizardStepBridges,
		WizardStepConfirm,
		WizardStepComplete,
	}

	for _, step := range steps {
		m.step = step
		view := m.View()
		if view == "" {
			t.Errorf("expected non-empty view for step %d", step)
		}
	}
}

func TestWizardModel_View_WelcomeContent(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepWelcome
	view := m.View()

	if !strings.Contains(view, "muxbee") {
		t.Error("expected welcome view to contain 'muxbee'")
	}
}

func TestWizardModel_View_ConnectivityContent(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepConnectivity
	view := m.View()

	if !strings.Contains(view, "Local") {
		t.Error("expected connectivity view to contain 'Local'")
	}
	if !strings.Contains(view, "Private") {
		t.Error("expected connectivity view to contain 'Private'")
	}
	if !strings.Contains(view, "Public") {
		t.Error("expected connectivity view to contain 'Public'")
	}
}

func TestWizardModel_View_BridgesContent(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepBridges
	view := m.View()

	if !strings.Contains(view, "whatsapp") {
		t.Error("expected bridges view to contain 'whatsapp'")
	}
	if !strings.Contains(view, "signal") {
		t.Error("expected bridges view to contain 'signal'")
	}
}

func TestWizardModel_View_ConfirmContent(t *testing.T) {
	m := NewWizardModel()
	m.step = WizardStepConfirm
	m.connectivityMode = "local"
	m.serverName = "localhost"
	m.selectedBridges = []string{"whatsapp", "signal"}
	view := m.View()

	if !strings.Contains(view, "local") {
		t.Error("expected confirm view to show mode")
	}
	if !strings.Contains(view, "localhost") {
		t.Error("expected confirm view to show server")
	}
	if !strings.Contains(view, "whatsapp") {
		t.Error("expected confirm view to show selected bridges")
	}
}

// Test Settings Model

func TestNewSettingsModel(t *testing.T) {
	m := NewSettingsModel(nil)
	if m == nil {
		t.Fatal("expected non-nil model")
	}
	if m.mode != 0 {
		t.Error("expected mode to default to 0 (local)")
	}
}

func TestNewSettingsModel_WithConfig(t *testing.T) {
	cfg := &config.Config{
		ConnectivityMode: "private",
		ServerName:       "192.168.1.50",
	}
	m := NewSettingsModel(cfg)

	if m.mode != 1 {
		t.Errorf("expected mode to be 1 (private), got %d", m.mode)
	}
	if m.serverInput.Value() != "192.168.1.50" {
		t.Errorf("expected serverInput to be '192.168.1.50', got '%s'", m.serverInput.Value())
	}
}

func TestNewSettingsModel_WithPublicConfig(t *testing.T) {
	cfg := &config.Config{
		ConnectivityMode: "public",
		ServerName:       "chat.example.com",
		HTTPS: config.HTTPSConfig{
			Enabled: true,
			Domain:  "chat.example.com",
			Email:   "admin@example.com",
		},
	}
	m := NewSettingsModel(cfg)

	if m.mode != 2 {
		t.Errorf("expected mode to be 2 (public), got %d", m.mode)
	}
	if m.domainInput.Value() != "chat.example.com" {
		t.Errorf("expected domainInput to be 'chat.example.com', got '%s'", m.domainInput.Value())
	}
	if m.emailInput.Value() != "admin@example.com" {
		t.Errorf("expected emailInput to be 'admin@example.com', got '%s'", m.emailInput.Value())
	}
}

func TestSettingsModel_Update_ModeSwitch(t *testing.T) {
	m := NewSettingsModel(nil)
	m.cursor = 0 // On mode selector
	m.mode = 0   // Local

	// Press right to switch to private
	msg := tea.KeyMsg{Type: tea.KeyRight}
	newM, _ := m.Update(msg, nil, nil)
	if newM.mode != 1 {
		t.Errorf("expected mode to be 1 after right, got %d", newM.mode)
	}

	// Press right again to switch to public
	newM, _ = newM.Update(msg, nil, nil)
	if newM.mode != 2 {
		t.Errorf("expected mode to be 2 after second right, got %d", newM.mode)
	}

	// Press right again - should stay at 2
	newM, _ = newM.Update(msg, nil, nil)
	if newM.mode != 2 {
		t.Errorf("expected mode to stay at 2, got %d", newM.mode)
	}

	// Press left to go back to private
	msg = tea.KeyMsg{Type: tea.KeyLeft}
	newM, _ = newM.Update(msg, nil, nil)
	if newM.mode != 1 {
		t.Errorf("expected mode to be 1 after left, got %d", newM.mode)
	}
}

func TestSettingsModel_Update_Navigation(t *testing.T) {
	m := NewSettingsModel(nil)
	m.mode = 2 // Public mode has 3 fields (mode, domain, email)

	// Navigate down
	msg := tea.KeyMsg{Type: tea.KeyDown}
	newM, _ := m.Update(msg, nil, nil)
	if newM.cursor != 1 {
		t.Errorf("expected cursor to be 1, got %d", newM.cursor)
	}

	newM, _ = newM.Update(msg, nil, nil)
	if newM.cursor != 2 {
		t.Errorf("expected cursor to be 2, got %d", newM.cursor)
	}

	// Can't go beyond max
	newM, _ = newM.Update(msg, nil, nil)
	if newM.cursor != 2 {
		t.Errorf("expected cursor to stay at 2, got %d", newM.cursor)
	}

	// Navigate up
	msg = tea.KeyMsg{Type: tea.KeyUp}
	newM, _ = newM.Update(msg, nil, nil)
	if newM.cursor != 1 {
		t.Errorf("expected cursor to be 1 after up, got %d", newM.cursor)
	}
}

func TestSettingsModel_Update_EnterEditing(t *testing.T) {
	m := NewSettingsModel(nil)
	m.mode = 1   // Private mode
	m.cursor = 1 // Server field

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newM, _ := m.Update(msg, nil, nil)

	if !newM.editing {
		t.Error("expected editing to be true after enter on field")
	}
	if newM.editField != 1 {
		t.Errorf("expected editField to be 1, got %d", newM.editField)
	}
}

func TestSettingsModel_UpdateEditing_ExitOnEsc(t *testing.T) {
	m := NewSettingsModel(nil)
	m.editing = true
	m.mode = 1
	m.editField = 1

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newM, _ := m.updateEditing(msg, nil, nil)

	if newM.editing {
		t.Error("expected editing to be false after esc")
	}
}

func TestSettingsModel_Update_SavedMessage(t *testing.T) {
	m := NewSettingsModel(nil)

	msg := settingsSavedMsg{err: nil}
	newM, _ := m.Update(msg, nil, nil)

	if !newM.saved {
		t.Error("expected saved to be true after settingsSavedMsg")
	}
}

func TestSettingsModel_View_LocalMode(t *testing.T) {
	m := NewSettingsModel(nil)
	m.mode = 0
	view := m.View()

	if !strings.Contains(view, "Settings") {
		t.Error("expected view to contain 'Settings'")
	}
	if !strings.Contains(view, "Local") {
		t.Error("expected view to show Local mode")
	}
	if !strings.Contains(view, "localhost") {
		t.Error("expected view to show localhost for local mode")
	}
}

func TestSettingsModel_View_PublicMode(t *testing.T) {
	m := NewSettingsModel(nil)
	m.mode = 2
	view := m.View()

	if !strings.Contains(view, "Domain") {
		t.Error("expected view to show Domain field")
	}
	if !strings.Contains(view, "Email") {
		t.Error("expected view to show Email field")
	}
}

func TestSettingsModel_View_SavedMessage(t *testing.T) {
	m := NewSettingsModel(nil)
	m.saved = true
	view := m.View()

	if !strings.Contains(view, "Saved") {
		t.Error("expected view to show saved message")
	}
}

// Test pad function
func TestPadRight(t *testing.T) {
	tests := []struct {
		input  string
		length int
		want   string
	}{
		{"abc", 5, "abc  "},
		{"abcdef", 3, "abc"},
		{"abc", 3, "abc"},
	}

	for _, tt := range tests {
		got := padRight(tt.input, tt.length)
		if got != tt.want {
			t.Errorf("padRight(%q, %d) = %q, want %q", tt.input, tt.length, got, tt.want)
		}
	}
}
