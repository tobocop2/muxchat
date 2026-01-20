package tui

import "github.com/charmbracelet/lipgloss"

// Colors - more subtle palette
var (
	ColorPrimary   = lipgloss.Color("#8B5CF6") // Softer purple
	ColorSecondary = lipgloss.Color("#10B981") // Green
	ColorWarning   = lipgloss.Color("#F59E0B") // Amber
	ColorDanger    = lipgloss.Color("#EF4444") // Red
	ColorMuted     = lipgloss.Color("#9CA3AF") // Lighter gray
	ColorText      = lipgloss.Color("#E5E7EB") // Light gray
	ColorDim       = lipgloss.Color("#6B7280") // Dim gray
)

// Styles
var (
	// Title - clean, not too heavy
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorText)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// Box - subtle border
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDim).
			Padding(1, 2)

	// Status
	StatusRunning = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	StatusStopped = lipgloss.NewStyle().
			Foreground(ColorDanger)

	StatusWarning = lipgloss.NewStyle().
			Foreground(ColorWarning)

	// Tabs
	TabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(ColorMuted)

	ActiveTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(ColorPrimary).
			Bold(true)

	// Help - subtle
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// List items
	ListItemStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	ListItemSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				Bold(true)

	// Feedback
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorDanger)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	// Version - subtle dim color
	VersionStyle = lipgloss.NewStyle().
			Foreground(ColorDim)
)

// Helper functions
func RenderKey(key, description string) string {
	return HelpKeyStyle.Render(key) + " " + HelpStyle.Render(description)
}

func RenderStatus(running bool, health string) string {
	if running {
		if health == "healthy" || health == "" {
			return StatusRunning.Render("running")
		}
		return StatusWarning.Render("running (" + health + ")")
	}
	return StatusStopped.Render("stopped")
}
