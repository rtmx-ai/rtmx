package output

import (
	"strings"
	"testing"
)

// TestHeader tests the Header function.
func TestHeader(t *testing.T) {
	DisableColor()
	defer EnableColor()

	header := Header("Test Header", 40)
	if !strings.Contains(header, "Test Header") {
		t.Errorf("Header should contain text, got %q", header)
	}
	if len(header) < 40 {
		t.Errorf("Header should be at least width 40, got %d", len(header))
	}
	if !strings.Contains(header, "=") {
		t.Errorf("Header should contain '=' padding, got %q", header)
	}
}

// TestHeaderSmall tests Header with a small width.
func TestHeaderSmall(t *testing.T) {
	DisableColor()
	defer EnableColor()

	header := Header("Very Long Header Text That Exceeds Width", 10)
	if !strings.Contains(header, "Very Long Header") {
		t.Errorf("Header should contain text even when small, got %q", header)
	}
}

// TestSubHeader tests the SubHeader function.
func TestSubHeader(t *testing.T) {
	DisableColor()
	defer EnableColor()

	sub := SubHeader("Sub Header", 40)
	if !strings.Contains(sub, "Sub Header") {
		t.Errorf("SubHeader should contain text, got %q", sub)
	}
	if !strings.Contains(sub, "-") {
		t.Errorf("SubHeader should contain '-' padding, got %q", sub)
	}
}

// TestSubHeaderSmall tests SubHeader with width smaller than text.
func TestSubHeaderSmall(t *testing.T) {
	DisableColor()
	defer EnableColor()

	sub := SubHeader("Very Long Sub Header Text", 5)
	if !strings.Contains(sub, "Very Long") {
		t.Errorf("SubHeader should contain text, got %q", sub)
	}
}

// TestCheckmark tests the Checkmark function.
func TestCheckmark(t *testing.T) {
	DisableColor()
	defer EnableColor()

	got := Checkmark(true)
	if got != "\xe2\x9c\x93" { // UTF-8 for checkmark
		t.Errorf("Checkmark(true) = %q, want checkmark symbol", got)
	}

	got = Checkmark(false)
	if got != "\xe2\x9c\x97" { // UTF-8 for X
		t.Errorf("Checkmark(false) = %q, want X symbol", got)
	}
}

// TestTruncate tests the Truncate function.
func TestTruncate(t *testing.T) {
	tests := []struct {
		text     string
		maxWidth int
		expected string
	}{
		{"short", 10, "short"},
		{"a long text here", 10, "a long ..."},
		{"ab", 2, "ab"},
		{"abcdef", 3, "abc"},
		{"abcdef", 4, "a..."},
		{"hello world", 11, "hello world"},
	}

	for _, tt := range tests {
		got := Truncate(tt.text, tt.maxWidth)
		if got != tt.expected {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tt.text, tt.maxWidth, got, tt.expected)
		}
	}
}

// TestPadRight tests the PadRight function.
func TestPadRight(t *testing.T) {
	tests := []struct {
		text     string
		width    int
		expected string
	}{
		{"hi", 5, "hi   "},
		{"hello", 5, "hello"},
		{"hello!", 5, "hello!"},
	}

	for _, tt := range tests {
		got := PadRight(tt.text, tt.width)
		if got != tt.expected {
			t.Errorf("PadRight(%q, %d) = %q, want %q", tt.text, tt.width, got, tt.expected)
		}
	}
}

// TestPadLeft tests the PadLeft function.
func TestPadLeft(t *testing.T) {
	tests := []struct {
		text     string
		width    int
		expected string
	}{
		{"hi", 5, "   hi"},
		{"hello", 5, "hello"},
		{"hello!", 5, "hello!"},
	}

	for _, tt := range tests {
		got := PadLeft(tt.text, tt.width)
		if got != tt.expected {
			t.Errorf("PadLeft(%q, %d) = %q, want %q", tt.text, tt.width, got, tt.expected)
		}
	}
}

// TestCenter tests the Center function.
func TestCenter(t *testing.T) {
	tests := []struct {
		text     string
		width    int
		expected string
	}{
		{"hi", 6, "  hi  "},
		{"hi", 5, " hi  "},
		{"hello", 5, "hello"},
		{"hello!", 5, "hello!"},
	}

	for _, tt := range tests {
		got := Center(tt.text, tt.width)
		if got != tt.expected {
			t.Errorf("Center(%q, %d) = %q, want %q", tt.text, tt.width, got, tt.expected)
		}
	}
}

// TestColorDisabled tests Color when color is disabled.
func TestColorDisabled(t *testing.T) {
	DisableColor()
	defer EnableColor()

	got := Color("text", Green)
	if got != "text" {
		t.Errorf("Color with disabled color = %q, want 'text'", got)
	}
}

// TestStatusColorUnknown tests StatusColor with unknown status.
func TestStatusColorUnknown(t *testing.T) {
	got := StatusColor("UNKNOWN")
	if got != White {
		t.Errorf("StatusColor(UNKNOWN) = %q, want White", got)
	}
}

// TestPriorityColorUnknown tests PriorityColor with unknown priority.
func TestPriorityColorUnknown(t *testing.T) {
	got := PriorityColor("UNKNOWN")
	if got != White {
		t.Errorf("PriorityColor(UNKNOWN) = %q, want White", got)
	}
}

// TestStatusIconUnknown tests StatusIcon with unknown status.
func TestStatusIconUnknown(t *testing.T) {
	got := StatusIcon("UNKNOWN")
	if got != "?" {
		t.Errorf("StatusIcon(UNKNOWN) = %q, want '?'", got)
	}
}

// TestProgressBarEdgeCases tests ProgressBar with edge cases.
func TestProgressBarEdgeCases(t *testing.T) {
	DisableColor()
	defer EnableColor()

	// Negative percent
	bar := ProgressBar(-10, 10)
	if bar == "" {
		t.Error("ProgressBar should handle negative percent")
	}

	// Over 100%
	bar = ProgressBar(150, 10)
	if bar == "" {
		t.Error("ProgressBar should handle >100%")
	}

	// Zero width
	bar = ProgressBar(50, 0)
	if bar == "" {
		t.Error("ProgressBar should handle zero width")
	}
}

// TestIsColorEnabled tests the IsColorEnabled function.
func TestIsColorEnabled(t *testing.T) {
	DisableColor()
	if IsColorEnabled() {
		t.Error("IsColorEnabled should return false when color is disabled")
	}
	EnableColor()
	// Note: IsColorEnabled also checks isTerminal(), so in test context
	// it may still return false (not a real terminal), which is expected.
}

// TestStatusColorCaseInsensitive tests that StatusColor is case-insensitive.
func TestStatusColorCaseInsensitive(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"complete", Green},
		{"COMPLETE", Green},
		{"Complete", Green},
		{"partial", Yellow},
		{"missing", Red},
		{"not_started", Red},
	}

	for _, tt := range tests {
		got := StatusColor(tt.status)
		if got != tt.expected {
			t.Errorf("StatusColor(%q) = %q, want %q", tt.status, got, tt.expected)
		}
	}
}
