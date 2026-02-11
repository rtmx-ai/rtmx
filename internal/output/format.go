// Package output provides formatting and display utilities for RTMX.
package output

import (
	"fmt"
	"os"
	"strings"
)

// ANSI color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Magenta   = "\033[35m"
	Cyan      = "\033[36m"
	White     = "\033[37m"
	BoldRed   = "\033[1;31m"
	BoldGreen = "\033[1;32m"
)

var useColor = true

// DisableColor disables colored output.
func DisableColor() {
	useColor = false
}

// EnableColor enables colored output.
func EnableColor() {
	useColor = true
}

// IsColorEnabled returns whether color output is enabled.
func IsColorEnabled() bool {
	return useColor && isTerminal()
}

// isTerminal checks if stdout is a terminal.
func isTerminal() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// Color applies a color to text if color is enabled.
func Color(text, color string) string {
	if !IsColorEnabled() {
		return text
	}
	return color + text + Reset
}

// StatusColor returns the appropriate color for a status.
func StatusColor(status string) string {
	switch strings.ToUpper(status) {
	case "COMPLETE":
		return Green
	case "PARTIAL":
		return Yellow
	case "MISSING", "NOT_STARTED":
		return Red
	default:
		return White
	}
}

// PriorityColor returns the appropriate color for a priority.
func PriorityColor(priority string) string {
	switch strings.ToUpper(priority) {
	case "P0":
		return BoldRed
	case "HIGH":
		return Red
	case "MEDIUM":
		return Yellow
	case "LOW":
		return Green
	default:
		return White
	}
}

// ProgressBar creates a visual progress bar.
func ProgressBar(percent float64, width int) string {
	filled := int(percent / 100.0 * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	// Color the bar based on completion
	var color string
	switch {
	case percent >= 80:
		color = Green
	case percent >= 50:
		color = Yellow
	default:
		color = Red
	}

	return Color("["+bar+"]", color)
}

// Header creates a formatted header line.
func Header(text string, width int) string {
	padding := (width - len(text) - 2) / 2
	if padding < 0 {
		padding = 0
	}
	line := strings.Repeat("=", padding) + " " + text + " " + strings.Repeat("=", padding)
	// Ensure exact width
	for len(line) < width {
		line += "="
	}
	return Color(line, Bold)
}

// SubHeader creates a formatted subheader line.
func SubHeader(text string, width int) string {
	padding := (width - len(text) - 2) / 2
	if padding < 0 {
		padding = 0
	}
	line := strings.Repeat("-", padding) + " " + text + " " + strings.Repeat("-", padding)
	for len(line) < width {
		line += "-"
	}
	return Color(line, Dim)
}

// Checkmark returns a colored checkmark or X.
func Checkmark(ok bool) string {
	if ok {
		return Color("✓", Green)
	}
	return Color("✗", Red)
}

// StatusIcon returns a colored icon for a status.
func StatusIcon(status string) string {
	switch strings.ToUpper(status) {
	case "COMPLETE":
		return Color("✓", Green)
	case "PARTIAL":
		return Color("⚠", Yellow)
	case "MISSING", "NOT_STARTED":
		return Color("✗", Red)
	default:
		return "?"
	}
}

// FormatPercent formats a percentage with color.
func FormatPercent(percent float64) string {
	text := fmt.Sprintf("%.1f%%", percent)
	var color string
	switch {
	case percent >= 80:
		color = Green
	case percent >= 50:
		color = Yellow
	default:
		color = Red
	}
	return Color(text, color)
}

// Truncate truncates text to a maximum width with ellipsis.
func Truncate(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return text[:maxWidth]
	}
	return text[:maxWidth-3] + "..."
}

// PadRight pads text to a minimum width.
func PadRight(text string, width int) string {
	if len(text) >= width {
		return text
	}
	return text + strings.Repeat(" ", width-len(text))
}

// PadLeft pads text to a minimum width.
func PadLeft(text string, width int) string {
	if len(text) >= width {
		return text
	}
	return strings.Repeat(" ", width-len(text)) + text
}

// Center centers text in a given width.
func Center(text string, width int) string {
	if len(text) >= width {
		return text
	}
	leftPad := (width - len(text)) / 2
	rightPad := width - len(text) - leftPad
	return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
}
