package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx-go/internal/database"
)

func TestVerifyCommandHelp(t *testing.T) {
	rootCmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"verify", "--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("verify --help failed: %v", err)
	}

	output := buf.String()
	expectedPhrases := []string{
		"verify",
		"--update",
		"--dry-run",
		"--verbose",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Expected help to contain %q, got: %s", phrase, output)
		}
	}
}

func TestVerifyDetermineNewStatus(t *testing.T) {
	tests := []struct {
		name     string
		result   *TestResult
		current  database.Status
		expected database.Status
	}{
		{
			name:     "passing test completes requirement",
			result:   &TestResult{Passed: true},
			current:  database.StatusMissing,
			expected: database.StatusComplete,
		},
		{
			name:     "failing test downgrades complete to partial",
			result:   &TestResult{Failed: true},
			current:  database.StatusComplete,
			expected: database.StatusPartial,
		},
		{
			name:     "failing test keeps missing as missing",
			result:   &TestResult{Failed: true},
			current:  database.StatusMissing,
			expected: database.StatusMissing,
		},
		{
			name:     "skipped test keeps current status",
			result:   &TestResult{Skipped: true},
			current:  database.StatusMissing,
			expected: database.StatusMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineNewStatus(tt.result, tt.current)
			if got != tt.expected {
				t.Errorf("determineNewStatus() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBoolToInt(t *testing.T) {
	if boolToInt(true) != 1 {
		t.Error("boolToInt(true) should be 1")
	}
	if boolToInt(false) != 0 {
		t.Error("boolToInt(false) should be 0")
	}
}
