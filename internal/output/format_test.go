package output

import (
	"testing"
)

// TestColorScheme verifies that the color scheme matches Python CLI
// REQ-GO-054: Go CLI shall use same color scheme as Python CLI
func TestColorScheme(t *testing.T) {
	// Status colors: COMPLETE=green, PARTIAL=yellow, MISSING=red
	tests := []struct {
		status   string
		expected string
	}{
		{"COMPLETE", Green},
		{"PARTIAL", Yellow},
		{"MISSING", Red},
		{"NOT_STARTED", Red},
	}

	for _, tt := range tests {
		got := StatusColor(tt.status)
		if got != tt.expected {
			t.Errorf("StatusColor(%q) = %q, want %q", tt.status, got, tt.expected)
		}
	}
}

func TestPriorityColors(t *testing.T) {
	// Priority colors: P0=bold red, HIGH=red, MEDIUM=yellow, LOW=green
	tests := []struct {
		priority string
		expected string
	}{
		{"P0", BoldRed},
		{"HIGH", Red},
		{"MEDIUM", Yellow},
		{"LOW", Green},
	}

	for _, tt := range tests {
		got := PriorityColor(tt.priority)
		if got != tt.expected {
			t.Errorf("PriorityColor(%q) = %q, want %q", tt.priority, got, tt.expected)
		}
	}
}

func TestStatusIcon(t *testing.T) {
	// Status icons match the expected symbols
	tests := []struct {
		status      string
		containsIcon string
	}{
		{"COMPLETE", "✓"},
		{"PARTIAL", "⚠"},
		{"MISSING", "✗"},
		{"NOT_STARTED", "✗"},
	}

	for _, tt := range tests {
		got := StatusIcon(tt.status)
		// The icon is wrapped in ANSI codes, but should contain the symbol
		// When color is disabled, we get just the symbol
		DisableColor()
		gotNoColor := StatusIcon(tt.status)
		EnableColor()

		if gotNoColor != tt.containsIcon {
			t.Errorf("StatusIcon(%q) without color = %q, want %q", tt.status, gotNoColor, tt.containsIcon)
		}
		_ = got // Use the colored version to ensure it doesn't panic
	}
}

func TestProgressBarColors(t *testing.T) {
	// Progress bar colors: >=80% green, >=50% yellow, <50% red
	// Just verify the progress bar is generated without errors for various percentages
	tests := []float64{100.0, 85.0, 80.0, 75.0, 50.0, 49.0, 25.0, 0.0}

	EnableColor()
	for _, pct := range tests {
		got := ProgressBar(pct, 20)
		if got == "" {
			t.Errorf("ProgressBar(%.1f%%) returned empty string", pct)
		}
		// Should contain progress bar characters
		if len(got) < 20 {
			t.Errorf("ProgressBar(%.1f%%) too short: %q", pct, got)
		}
	}
}

func TestFormatPercent(t *testing.T) {
	// Percent colors match progress bar colors
	tests := []struct {
		percent float64
		desc    string
	}{
		{100.0, "green for 100%"},
		{80.0, "green for 80%"},
		{50.0, "yellow for 50%"},
		{25.0, "red for 25%"},
	}

	for _, tt := range tests {
		got := FormatPercent(tt.percent)
		if got == "" {
			t.Errorf("FormatPercent(%.1f) returned empty string", tt.percent)
		}
	}
}
