package tui

import "github.com/charmbracelet/lipgloss"

// Status colors
var (
	ColorComplete   = lipgloss.Color("#16a34a") // green
	ColorPartial    = lipgloss.Color("#d97706") // amber
	ColorMissing    = lipgloss.Color("#dc2626") // red
	ColorNotStarted = lipgloss.Color("#71717a") // zinc dim
	ColorAccent     = lipgloss.Color("#3b82f6") // blue
)

// Component styles
var (
	StyleTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorAccent).
			Underline(true).
			Padding(0, 2)

	StyleTabInactive = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#a1a1aa")).
				Padding(0, 2)

	StyleStatusBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#fafafa")).
			Background(lipgloss.Color("#27272a")).
			Padding(0, 1)

	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#fafafa"))

	StyleHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#a1a1aa"))

	StyleComplete   = lipgloss.NewStyle().Foreground(ColorComplete)
	StylePartial    = lipgloss.NewStyle().Foreground(ColorPartial)
	StyleMissing    = lipgloss.NewStyle().Foreground(ColorMissing)
	StyleNotStarted = lipgloss.NewStyle().Foreground(ColorNotStarted)
)

// StatusStyle returns the appropriate style for a requirement status string.
func StatusStyle(status string) lipgloss.Style {
	switch status {
	case "COMPLETE":
		return StyleComplete
	case "PARTIAL":
		return StylePartial
	case "MISSING":
		return StyleMissing
	default:
		return StyleNotStarted
	}
}
