package output

import (
	"os"
	"syscall"
	"unsafe"
)

// DefaultTerminalWidth is the fallback width when detection fails.
const DefaultTerminalWidth = 80

// MinTerminalWidth is the minimum supported width for rendering.
const MinTerminalWidth = 40

// termWidthOverride allows tests to override terminal width detection.
// When non-zero, TerminalWidth() returns this value instead of detecting.
var termWidthOverride int

// SetTerminalWidthOverride sets a fixed terminal width for testing.
// Pass 0 to restore automatic detection.
func SetTerminalWidthOverride(width int) {
	termWidthOverride = width
}

// TerminalWidth returns the current terminal width.
// It attempts to detect the width via ioctl on stdout. If detection fails
// (e.g., stdout is not a terminal), it returns DefaultTerminalWidth.
func TerminalWidth() int {
	if termWidthOverride > 0 {
		return termWidthOverride
	}

	width := detectTerminalWidth()
	if width <= 0 {
		return DefaultTerminalWidth
	}
	return width
}

// detectTerminalWidth uses the TIOCGWINSZ ioctl to get the terminal size.
func detectTerminalWidth() int {
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}

	var ws winsize
	fd := os.Stdout.Fd()
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		syscall.TIOCGWINSZ,
		uintptr(unsafe.Pointer(&ws)),
	)
	if errno != 0 {
		return 0
	}
	return int(ws.Col)
}

// PhaseProgressLine renders a single phase progress line that fits within the given width.
//
// Format: "Phase N (Name):  [████████░░░░]  66.7%  status  (2v 0w 1x)"
//
// Parameters:
//   - phase: phase number
//   - name: phase description (may be truncated)
//   - pct: completion percentage (0-100)
//   - complete, partial, missing: counts for each status
//   - totalWidth: total available character width
func PhaseProgressLine(phase int, name string, pct float64, complete, partial, missing, totalWidth int) string {
	if totalWidth < MinTerminalWidth {
		totalWidth = MinTerminalWidth
	}

	// Build the status text and suffix (these have fixed width once computed)
	var statusLabel string
	switch {
	case pct >= 100:
		statusLabel = Color("v Complete", Green)
	case pct > 0:
		statusLabel = Color("~ In Progress", Yellow)
	default:
		statusLabel = Color("x Not Started", Red)
	}

	// Suffix: "(2v 0~ 1x)"
	suffix := formatCounts(complete, partial, missing)

	// Percent: "100.0%" (6 chars max)
	pctStr := formatPctFixed(pct)

	// Calculate widths of fixed-size parts (separators + pct + status + suffix)
	// Layout: prefix + " " + bar + " " + pct + " " + status + " " + suffix
	statusDW := displayWidth(statusLabel)
	suffixDW := displayWidth(suffix)
	// Spacers: 1+1+1+1 = 4 single spaces between components
	// bar brackets: 2 ([ and ])
	// pctStr: 6
	fixedWidth := 4 + 2 + 6 + statusDW + suffixDW

	// Start with the natural prefix width
	minBarWidth := 10
	prefixBase := formatPhasePrefix(phase, name, totalWidth)
	prefixDW := displayWidth(prefixBase)

	availableForBar := totalWidth - fixedWidth - prefixDW
	if availableForBar < minBarWidth {
		// Shrink prefix to make room
		maxPrefixWidth := totalWidth - fixedWidth - minBarWidth
		if maxPrefixWidth < 8 {
			maxPrefixWidth = 8
		}
		prefixBase = formatPhasePrefixTruncated(phase, name, maxPrefixWidth)
		prefixDW = displayWidth(prefixBase)
		availableForBar = totalWidth - fixedWidth - prefixDW
	}

	if availableForBar < minBarWidth {
		availableForBar = minBarWidth
	}

	bar := ProgressBar(pct, availableForBar)

	return prefixBase + " " + bar + " " + pctStr + " " + statusLabel + " " + suffix
}

// formatPhasePrefix creates the "Phase N (Name):" label, padded for alignment.
// The padWidth parameter controls maximum padding width for alignment.
func formatPhasePrefix(phase int, name string, totalWidth int) string {
	var prefix string
	if name != "" {
		prefix = "Phase " + itoa(phase) + " (" + name + "):"
	} else {
		prefix = "Phase " + itoa(phase) + ":"
	}
	// Pad to align across phases, but cap padding based on terminal width
	padTarget := 30
	if totalWidth < 80 {
		padTarget = 20
	}
	if len(prefix) < padTarget {
		return padToWidth(prefix, padTarget)
	}
	return prefix
}

// formatPhasePrefixTruncated creates the prefix, truncating name if needed.
func formatPhasePrefixTruncated(phase int, name string, maxWidth int) string {
	if name == "" {
		prefix := "Phase " + itoa(phase) + ":"
		if len(prefix) > maxWidth {
			return prefix[:maxWidth]
		}
		return padToWidth(prefix, maxWidth)
	}

	// "Phase N (" + "..." + "):" = overhead
	overhead := len("Phase ") + len(itoa(phase)) + len(" (") + len("):")
	availForName := maxWidth - overhead
	if availForName < 4 {
		// Too small for name, just show phase number
		prefix := "Phase " + itoa(phase) + ":"
		if len(prefix) > maxWidth && maxWidth > 3 {
			return prefix[:maxWidth-3] + "..."
		}
		return prefix
	}

	truncName := Truncate(name, availForName)
	return "Phase " + itoa(phase) + " (" + truncName + "):"
}

// formatPctFixed formats a percentage into a fixed-width 7-char string: " 100.0%"
func formatPctFixed(pct float64) string {
	s := formatFloat1(pct) + "%"
	// Pad to 6 chars (e.g. "100.0%" is 6, " 66.7%" is 6, "  0.0%" is 6)
	for len(s) < 6 {
		s = " " + s
	}
	return s
}

// formatCounts builds the "(Nv Nw Nx)" suffix with colored symbols.
func formatCounts(complete, partial, missing int) string {
	return "(" +
		itoa(complete) + Color("v", Green) + " " +
		itoa(partial) + Color("~", Yellow) + " " +
		itoa(missing) + Color("x", Red) + ")"
}

// itoa converts an int to a string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// formatFloat1 formats a float to 1 decimal place without importing fmt.
func formatFloat1(f float64) string {
	// Use simple approach: multiply by 10, round, format
	if f < 0 {
		return "-" + formatFloat1(-f)
	}
	whole := int(f)
	frac := int((f - float64(whole)) * 10 + 0.5)
	if frac >= 10 {
		whole++
		frac = 0
	}
	return itoa(whole) + "." + itoa(frac)
}
