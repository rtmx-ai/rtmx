package output

import (
	"strings"
	"unicode/utf8"
)

// Table represents an ASCII table for formatted output.
type Table struct {
	headers []string
	rows    [][]string
	widths  []int
}

// NewTable creates a new table with the given headers.
func NewTable(headers ...string) *Table {
	t := &Table{
		headers: headers,
		widths:  make([]int, len(headers)),
	}
	// Initialize widths from headers
	for i, h := range headers {
		t.widths[i] = displayWidth(h)
	}
	return t
}

// AddRow adds a row to the table.
func (t *Table) AddRow(cells ...string) {
	// Ensure correct number of cells
	row := make([]string, len(t.headers))
	for i := range row {
		if i < len(cells) {
			row[i] = cells[i]
		}
	}
	t.rows = append(t.rows, row)
	// Update widths
	for i, cell := range row {
		w := displayWidth(cell)
		if w > t.widths[i] {
			t.widths[i] = w
		}
	}
}

// Render returns the table as a string in tabulate-compatible format.
func (t *Table) Render() string {
	if len(t.headers) == 0 {
		return ""
	}

	var sb strings.Builder

	// Top border
	sb.WriteString(t.renderSeparator("-", "+"))
	sb.WriteString("\n")

	// Header row
	sb.WriteString(t.renderRow(t.headers))
	sb.WriteString("\n")

	// Header separator (with = instead of -)
	sb.WriteString(t.renderSeparator("=", "+"))
	sb.WriteString("\n")

	// Data rows
	for _, row := range t.rows {
		sb.WriteString(t.renderRow(row))
		sb.WriteString("\n")
		sb.WriteString(t.renderSeparator("-", "+"))
		sb.WriteString("\n")
	}

	return sb.String()
}

// RenderCompact returns the table without row separators (only header separator).
func (t *Table) RenderCompact() string {
	if len(t.headers) == 0 {
		return ""
	}

	var sb strings.Builder

	// Top border
	sb.WriteString(t.renderSeparator("-", "+"))
	sb.WriteString("\n")

	// Header row
	sb.WriteString(t.renderRow(t.headers))
	sb.WriteString("\n")

	// Header separator (with = instead of -)
	sb.WriteString(t.renderSeparator("=", "+"))
	sb.WriteString("\n")

	// Data rows (no separators between rows)
	for _, row := range t.rows {
		sb.WriteString(t.renderRow(row))
		sb.WriteString("\n")
	}

	// Bottom border
	sb.WriteString(t.renderSeparator("-", "+"))
	sb.WriteString("\n")

	return sb.String()
}

// renderSeparator creates a line like +-----+-----+
func (t *Table) renderSeparator(fill, corner string) string {
	var parts []string
	for _, w := range t.widths {
		parts = append(parts, strings.Repeat(fill, w+2)) // +2 for padding
	}
	return corner + strings.Join(parts, corner) + corner
}

// renderRow creates a line like | val | val |
func (t *Table) renderRow(cells []string) string {
	var parts []string
	for i, cell := range cells {
		// Handle cells with ANSI codes - we need the display width
		padded := padToWidth(cell, t.widths[i])
		parts = append(parts, " "+padded+" ")
	}
	return "|" + strings.Join(parts, "|") + "|"
}

// displayWidth returns the display width of a string, ignoring ANSI escape codes.
func displayWidth(s string) int {
	// Strip ANSI codes for width calculation
	stripped := stripANSI(s)
	return utf8.RuneCountInString(stripped)
}

// stripANSI removes ANSI escape codes from a string.
func stripANSI(s string) string {
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}

// padToWidth pads a string to the given display width, handling ANSI codes.
func padToWidth(s string, width int) string {
	currentWidth := displayWidth(s)
	if currentWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-currentWidth)
}

// TruncateCell truncates text for table cell, preserving ANSI codes.
func TruncateCell(text string, maxWidth int) string {
	stripped := stripANSI(text)
	if utf8.RuneCountInString(stripped) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return stripped[:maxWidth]
	}
	// Count runes to truncate properly
	runes := []rune(stripped)
	return string(runes[:maxWidth-3]) + "..."
}
