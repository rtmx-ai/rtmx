package output

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
// It attempts to detect the width via platform-specific methods.
// If detection fails (e.g., stdout is not a terminal), it returns DefaultTerminalWidth.
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

// PhaseProgressLine renders a single phase progress line that fits within the given width.
func PhaseProgressLine(phase int, name string, pct float64, complete, partial, missing, totalWidth int) string {
	if totalWidth < MinTerminalWidth {
		totalWidth = MinTerminalWidth
	}

	var statusLabel string
	switch {
	case pct >= 100:
		statusLabel = Color("v Complete", Green)
	case pct > 0:
		statusLabel = Color("~ In Progress", Yellow)
	default:
		statusLabel = Color("x Not Started", Red)
	}

	suffix := formatCounts(complete, partial, missing)
	pctStr := formatPctFixed(pct)

	statusDW := displayWidth(statusLabel)
	suffixDW := displayWidth(suffix)
	fixedWidth := 4 + 2 + 6 + statusDW + suffixDW

	minBarWidth := 10
	prefixBase := formatPhasePrefix(phase, name, totalWidth)
	prefixDW := displayWidth(prefixBase)

	availableForBar := totalWidth - fixedWidth - prefixDW
	if availableForBar < minBarWidth {
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

func formatPhasePrefix(phase int, name string, totalWidth int) string {
	var prefix string
	if name != "" {
		prefix = "Phase " + itoa(phase) + " (" + name + "):"
	} else {
		prefix = "Phase " + itoa(phase) + ":"
	}
	padTarget := 30
	if totalWidth < 80 {
		padTarget = 20
	}
	if len(prefix) < padTarget {
		return padToWidth(prefix, padTarget)
	}
	return prefix
}

func formatPhasePrefixTruncated(phase int, name string, maxWidth int) string {
	if name == "" {
		prefix := "Phase " + itoa(phase) + ":"
		if len(prefix) > maxWidth {
			return prefix[:maxWidth]
		}
		return padToWidth(prefix, maxWidth)
	}

	overhead := len("Phase ") + len(itoa(phase)) + len(" (") + len("):")
	availForName := maxWidth - overhead
	if availForName < 4 {
		prefix := "Phase " + itoa(phase) + ":"
		if len(prefix) > maxWidth && maxWidth > 3 {
			return prefix[:maxWidth-3] + "..."
		}
		return prefix
	}

	truncName := Truncate(name, availForName)
	return "Phase " + itoa(phase) + " (" + truncName + "):"
}

func formatPctFixed(pct float64) string {
	s := formatFloat1(pct) + "%"
	for len(s) < 6 {
		s = " " + s
	}
	return s
}

func formatCounts(complete, partial, missing int) string {
	return "(" +
		itoa(complete) + Color("v", Green) + " " +
		itoa(partial) + Color("~", Yellow) + " " +
		itoa(missing) + Color("x", Red) + ")"
}

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

func formatFloat1(f float64) string {
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
