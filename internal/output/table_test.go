package output

import (
	"strings"
	"testing"
)

func TestNewTable(t *testing.T) {
	table := NewTable("A", "B", "C")
	if table == nil {
		t.Fatal("NewTable returned nil")
	}
	if len(table.headers) != 3 {
		t.Errorf("Expected 3 headers, got %d", len(table.headers))
	}
}

func TestTableAddRow(t *testing.T) {
	table := NewTable("Name", "Value")
	table.AddRow("foo", "bar")
	table.AddRow("longer name", "x")

	if len(table.rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(table.rows))
	}
}

func TestTableRender(t *testing.T) {
	table := NewTable("#", "Status", "ID")
	table.AddRow("1", "✓", "REQ-001")
	table.AddRow("2", "✗", "REQ-002")

	output := table.Render()

	// Check for table structure
	expectedElements := []string{
		"+---+",    // Column separator
		"|",        // Row borders
		"| # |",    // Header cell
		"+===+",    // Header separator
		"| 1 |",    // Data cell
		"| 2 |",    // Data cell
		"REQ-001",  // Cell content
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected %q in output:\n%s", element, output)
		}
	}
}

func TestTableRenderCompact(t *testing.T) {
	table := NewTable("A", "B")
	table.AddRow("1", "2")
	table.AddRow("3", "4")

	output := table.RenderCompact()

	// Compact should have top border, header separator, and bottom border
	// but no separators between data rows
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have: top border, header, header sep, row1, row2, bottom border = 6 lines
	if len(lines) != 6 {
		t.Errorf("Expected 6 lines for compact table, got %d:\n%s", len(lines), output)
	}
}

func TestDisplayWidth(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"hello", 5},
		{"✓", 1},
		{"\033[32mgreen\033[0m", 5}, // ANSI colored "green"
		{"", 0},
	}

	for _, tt := range tests {
		got := displayWidth(tt.input)
		if got != tt.expected {
			t.Errorf("displayWidth(%q) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"\033[32mgreen\033[0m", "green"},
		{"\033[1;31mred bold\033[0m", "red bold"},
		{"no codes", "no codes"},
	}

	for _, tt := range tests {
		got := stripANSI(tt.input)
		if got != tt.expected {
			t.Errorf("stripANSI(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestTruncateCell(t *testing.T) {
	tests := []struct {
		input    string
		maxWidth int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},
	}

	for _, tt := range tests {
		got := TruncateCell(tt.input, tt.maxWidth)
		if got != tt.expected {
			t.Errorf("TruncateCell(%q, %d) = %q, want %q", tt.input, tt.maxWidth, got, tt.expected)
		}
	}
}
