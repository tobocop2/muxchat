package tui

import (
	"bufio"
	"io"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tobias/muxchat/internal/docker"
)

// LogsModel handles the logs screen
type LogsModel struct {
	logs       []logEntry
	paused     bool
	filter     string // Filter by service name (empty = all)
	reader     io.ReadCloser
	streaming  bool
	services   []string // Available services for filtering
	filterIdx  int      // Current filter index (0 = all)
}

type logEntry struct {
	service string
	message string
	time    time.Time
}

// NewLogsModel creates a new logs model
func NewLogsModel() *LogsModel {
	return &LogsModel{
		logs:   []logEntry{},
		paused: false,
	}
}

// Init initializes the logs screen
func (m *LogsModel) Init() tea.Cmd {
	return nil
}

// StartStreaming begins streaming logs from Docker
func (m *LogsModel) StartStreaming(compose *docker.Compose) tea.Cmd {
	if compose == nil || m.streaming {
		return nil
	}

	return func() tea.Msg {
		reader, err := compose.LogsReader("", true, "100")
		if err != nil {
			return logsErrorMsg{err: err}
		}
		return logsStartedMsg{reader: reader}
	}
}

// Update handles logs screen events
func (m *LogsModel) Update(msg tea.Msg, compose *docker.Compose) (*LogsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "p":
			m.paused = !m.paused
		case "f":
			// Cycle through filters
			m.filterIdx++
			if m.filterIdx > len(m.services) {
				m.filterIdx = 0
			}
			if m.filterIdx == 0 {
				m.filter = ""
			} else {
				m.filter = m.services[m.filterIdx-1]
			}
		case "c":
			m.logs = []logEntry{}
		}

	case logsStartedMsg:
		m.reader = msg.reader
		m.streaming = true
		return m, m.readLogsCmd()

	case logsBatchMsg:
		if !m.paused {
			m.logs = append(m.logs, msg.entries...)
			for _, e := range msg.entries {
				m.addService(e.service)
			}
			// Keep last 500 lines
			if len(m.logs) > 500 {
				m.logs = m.logs[len(m.logs)-500:]
			}
		}
		// Continue reading
		if m.streaming {
			return m, m.readLogsCmd()
		}

	case logsErrorMsg:
		m.streaming = false
		// Could show error, but for now just stop streaming

	case logsStopMsg:
		m.streaming = false
		if m.reader != nil {
			m.reader.Close()
			m.reader = nil
		}
	}
	return m, nil
}

func (m *LogsModel) addService(service string) {
	for _, s := range m.services {
		if s == service {
			return
		}
	}
	m.services = append(m.services, service)
}

// readLogsCmd reads the next batch of log lines
func (m *LogsModel) readLogsCmd() tea.Cmd {
	if m.reader == nil {
		return nil
	}

	return func() tea.Msg {
		scanner := bufio.NewScanner(m.reader)
		var entries []logEntry

		// Read up to 10 lines at a time
		for i := 0; i < 10 && scanner.Scan(); i++ {
			line := scanner.Text()
			entry := parseLine(line)
			entries = append(entries, entry)
		}

		if err := scanner.Err(); err != nil {
			return logsErrorMsg{err: err}
		}

		if len(entries) == 0 {
			// No more data right now, wait a bit then try again
			time.Sleep(100 * time.Millisecond)
			return logsBatchMsg{entries: nil}
		}

		return logsBatchMsg{entries: entries}
	}
}

// parseLine extracts service name and message from a Docker Compose log line
func parseLine(line string) logEntry {
	// Docker Compose format: "container-name  | message"
	// or sometimes: "container-name | message"
	parts := strings.SplitN(line, "|", 2)
	if len(parts) == 2 {
		service := strings.TrimSpace(parts[0])
		// Parse service name (e.g., "muxchat-synapse-1" -> "synapse")
		service = docker.ParseServiceName(service)
		message := strings.TrimSpace(parts[1])
		return logEntry{
			service: service,
			message: message,
			time:    time.Now(),
		}
	}
	return logEntry{
		service: "unknown",
		message: line,
		time:    time.Now(),
	}
}

// View renders the logs screen
func (m *LogsModel) View() string {
	var s string

	s += TitleStyle.Render("Service Logs") + "\n\n"

	// Status line
	var statusParts []string
	if m.paused {
		statusParts = append(statusParts, StatusWarning.Render("[PAUSED]"))
	}
	if m.filter != "" {
		statusParts = append(statusParts, SubtitleStyle.Render("Filter: "+m.filter))
	} else {
		statusParts = append(statusParts, SubtitleStyle.Render("Filter: all"))
	}
	if m.streaming {
		statusParts = append(statusParts, SuccessStyle.Render("● streaming"))
	}
	s += strings.Join(statusParts, "  ") + "\n\n"

	if len(m.logs) == 0 {
		s += SubtitleStyle.Render("No logs yet. Start services to see logs.") + "\n"
	} else {
		// Show last 20 lines (filtered)
		filtered := m.filteredLogs()
		start := 0
		if len(filtered) > 20 {
			start = len(filtered) - 20
		}
		for _, entry := range filtered[start:] {
			// Color code by service
			svcStyle := getServiceStyle(entry.service)
			s += svcStyle.Render(padRight(entry.service, 20)) + " " + entry.message + "\n"
		}
	}

	s += "\n"
	s += HelpStyle.Render(RenderKey("p", "pause") + "  " + RenderKey("f", "filter") + "  " + RenderKey("c", "clear"))

	return s
}

func (m *LogsModel) filteredLogs() []logEntry {
	if m.filter == "" {
		return m.logs
	}
	var filtered []logEntry
	for _, entry := range m.logs {
		if entry.service == m.filter {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func padRight(s string, length int) string {
	if len(s) >= length {
		return s[:length]
	}
	return s + strings.Repeat(" ", length-len(s))
}

// getServiceStyle returns a color style based on service name
func getServiceStyle(service string) lipgloss.Style {
	// Assign colors based on service type
	switch {
	case strings.Contains(service, "synapse"):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("39")) // Blue
	case strings.Contains(service, "postgres"):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("33")) // Cyan
	case strings.Contains(service, "element"):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")) // Green
	case strings.Contains(service, "whatsapp"):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("46")) // Bright green
	case strings.Contains(service, "telegram"):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("75")) // Light blue
	case strings.Contains(service, "discord"):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("99")) // Purple
	case strings.Contains(service, "slack"):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // Orange
	case strings.Contains(service, "signal"):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("69")) // Blue-purple
	case strings.Contains(service, "gmessages"):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("82")) // Bright green
	case strings.Contains(service, "googlechat"):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow
	case strings.Contains(service, "gvoice"):
		return lipgloss.NewStyle().Foreground(lipgloss.Color("214")) // Gold
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("245")) // Gray
	}
}

// Message types
type logsBatchMsg struct {
	entries []logEntry
}

type logsStartedMsg struct {
	reader io.ReadCloser
}

type logsErrorMsg struct {
	err error
}

type logsStopMsg struct{}
