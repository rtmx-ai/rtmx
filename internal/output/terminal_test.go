package output

import (
	"strings"
	"testing"
)

func TestTerminalWidthDefault(t *testing.T) {
	// When override is 0, TerminalWidth should return a positive value.
	// In test environments (piped stdout), it should fall back to DefaultTerminalWidth.
	SetTerminalWidthOverride(0)
	defer SetTerminalWidthOverride(0)

	w := TerminalWidth()
	if w <= 0 {
		t.Errorf("TerminalWidth() = %d, want > 0", w)
	}
}

func TestTerminalWidthOverride(t *testing.T) {
	tests := []int{40, 80, 120, 200}
	for _, width := range tests {
		SetTerminalWidthOverride(width)
		got := TerminalWidth()
		if got != width {
			t.Errorf("TerminalWidth() with override %d = %d", width, got)
		}
	}
	SetTerminalWidthOverride(0)
}

func TestTerminalWidthFallback(t *testing.T) {
	// With override=0 in a non-terminal (test) environment,
	// TerminalWidth should return DefaultTerminalWidth.
	SetTerminalWidthOverride(0)
	defer SetTerminalWidthOverride(0)

	w := TerminalWidth()
	// In CI / test environment, stdout is not a terminal, so we expect the default.
	if w != DefaultTerminalWidth {
		// It's also valid if we are in a real terminal and detect a width.
		if w < MinTerminalWidth {
			t.Errorf("TerminalWidth() = %d, want >= %d", w, MinTerminalWidth)
		}
	}
}

func TestPhaseProgressLineWidth40(t *testing.T) {
	DisableColor()
	defer EnableColor()

	line := PhaseProgressLine(1, "Foundation", 100.0, 3, 0, 0, 40)
	dw := displayWidth(line)

	// At width 40, the line should not exceed 40 display chars
	// (it may be shorter due to minimum bar constraints)
	if dw > 50 { // allow some slack for minimum bar
		t.Errorf("PhaseProgressLine at width 40: display width = %d, want <= 50", dw)
	}

	// Should contain key elements
	if !strings.Contains(line, "Phase 1") {
		t.Error("expected 'Phase 1' in output")
	}
	if !strings.Contains(line, "100.0%") {
		t.Error("expected '100.0%' in output")
	}
	if !strings.Contains(line, "Complete") {
		t.Error("expected 'Complete' in output")
	}
}

func TestPhaseProgressLineWidth80(t *testing.T) {
	DisableColor()
	defer EnableColor()

	line := PhaseProgressLine(13, "Zero-Trust", 66.7, 2, 0, 1, 80)
	dw := displayWidth(line)

	if dw > 80 {
		t.Errorf("PhaseProgressLine at width 80: display width = %d, want <= 80", dw)
	}
	if !strings.Contains(line, "Phase 13") {
		t.Error("expected 'Phase 13' in output")
	}
	if !strings.Contains(line, "Zero-Trust") {
		t.Error("expected 'Zero-Trust' in output")
	}
	if !strings.Contains(line, "66.7%") {
		t.Error("expected '66.7%' in output")
	}
	if !strings.Contains(line, "In Progress") {
		t.Error("expected 'In Progress' in output")
	}
}

func TestPhaseProgressLineWidth120(t *testing.T) {
	DisableColor()
	defer EnableColor()

	line := PhaseProgressLine(1, "Foundation", 100.0, 5, 1, 2, 120)
	dw := displayWidth(line)

	if dw > 120 {
		t.Errorf("PhaseProgressLine at width 120: display width = %d, want <= 120", dw)
	}
	// Wider terminal should produce longer progress bar
	line80 := PhaseProgressLine(1, "Foundation", 100.0, 5, 1, 2, 80)
	if displayWidth(line) <= displayWidth(line80) {
		t.Error("expected wider terminal to produce wider output")
	}
}

func TestPhaseProgressLineNotStarted(t *testing.T) {
	DisableColor()
	defer EnableColor()

	line := PhaseProgressLine(18, "Language Extensions", 0.0, 0, 0, 4, 80)

	if !strings.Contains(line, "Not Started") {
		t.Error("expected 'Not Started' for 0% completion")
	}
	if !strings.Contains(line, "0.0%") {
		t.Error("expected '0.0%' in output")
	}
}

func TestPhaseNameTruncation(t *testing.T) {
	DisableColor()
	defer EnableColor()

	// Very long phase name at narrow width should be truncated or dropped
	line60 := PhaseProgressLine(18, "Language Extensions and Cross-Platform Support", 50.0, 1, 1, 1, 60)

	// Phase number must always be present
	if !strings.Contains(line60, "Phase 18") {
		t.Error("expected 'Phase 18' even when name is truncated")
	}
	// The full name should NOT appear
	if strings.Contains(line60, "Cross-Platform Support") {
		t.Error("expected long name to be truncated at narrow width")
	}

	// At a medium width (90), the name should be truncated with "..."
	line90 := PhaseProgressLine(18, "Language Extensions and Cross-Platform Support", 50.0, 1, 1, 1, 90)
	if !strings.Contains(line90, "Phase 18") {
		t.Error("expected 'Phase 18' at width 90")
	}
	if strings.Contains(line90, "Cross-Platform Support") {
		t.Error("expected truncation at width 90")
	}
	if !strings.Contains(line90, "...") {
		t.Error("expected '...' truncation marker at width 90")
	}
}

func TestPhaseNameNoTruncationAtWideWidth(t *testing.T) {
	DisableColor()
	defer EnableColor()

	line := PhaseProgressLine(18, "Language Extensions", 50.0, 1, 1, 1, 120)

	// At wide width, full name should be preserved
	if !strings.Contains(line, "Language Extensions") {
		t.Error("expected full phase name at wide width")
	}
	if strings.Contains(line, "...") {
		t.Error("did not expect truncation at width 120")
	}
}

func TestPhaseProgressLineNoName(t *testing.T) {
	DisableColor()
	defer EnableColor()

	line := PhaseProgressLine(5, "", 75.0, 3, 0, 1, 80)

	if !strings.Contains(line, "Phase 5:") {
		t.Error("expected 'Phase 5:' when no name provided")
	}
	if strings.Contains(line, "()") {
		t.Error("should not contain empty parens when no name")
	}
}

func TestFormatPctFixed(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{100.0, "100.0%"},
		{66.7, " 66.7%"},
		{0.0, "  0.0%"},
		{5.5, "  5.5%"},
		{99.9, " 99.9%"},
	}

	for _, tt := range tests {
		got := formatPctFixed(tt.input)
		if got != tt.expected {
			t.Errorf("formatPctFixed(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFormatCounts(t *testing.T) {
	DisableColor()
	defer EnableColor()

	got := formatCounts(3, 1, 2)
	if got != "(3v 1~ 2x)" {
		t.Errorf("formatCounts(3,1,2) = %q, want %q", got, "(3v 1~ 2x)")
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{100, "100"},
	}

	for _, tt := range tests {
		got := itoa(tt.input)
		if got != tt.expected {
			t.Errorf("itoa(%d) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFormatFloat1(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0.0, "0.0"},
		{100.0, "100.0"},
		{66.7, "66.7"},
		{33.3, "33.3"},
		{99.9, "99.9"},
	}

	for _, tt := range tests {
		got := formatFloat1(tt.input)
		if got != tt.expected {
			t.Errorf("formatFloat1(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestProgressBarScaling(t *testing.T) {
	DisableColor()
	defer EnableColor()

	// At 50%, half the bar should be filled
	bar := ProgressBar(50.0, 20)
	// Strip brackets
	inner := bar[1 : len(bar)-1]
	filled := strings.Count(inner, "█")
	empty := strings.Count(inner, "░")

	if filled != 10 {
		t.Errorf("ProgressBar(50%%, 20): filled = %d, want 10", filled)
	}
	if empty != 10 {
		t.Errorf("ProgressBar(50%%, 20): empty = %d, want 10", empty)
	}

	// At 100%, full bar
	bar100 := ProgressBar(100.0, 20)
	inner100 := bar100[1 : len(bar100)-1]
	if strings.Count(inner100, "█") != 20 {
		t.Errorf("ProgressBar(100%%, 20): expected 20 filled chars")
	}

	// At 0%, empty bar
	bar0 := ProgressBar(0.0, 20)
	inner0 := bar0[1 : len(bar0)-1]
	if strings.Count(inner0, "░") != 20 {
		t.Errorf("ProgressBar(0%%, 20): expected 20 empty chars")
	}
}
