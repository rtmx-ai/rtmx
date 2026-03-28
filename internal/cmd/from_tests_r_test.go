package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestExtractRMarkers(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-022")

	tests := []struct {
		name     string
		content  string
		expected []TestRequirement
	}{
		{
			name: "comment marker # rtmx:req REQ-ID with test_that",
			content: `# rtmx:req REQ-STAT-001
test_that("mean is calculated correctly", {
  expect_equal(mean(1:10), 5.5)
})
`,
			expected: []TestRequirement{
				{ReqID: "REQ-STAT-001", TestFunction: "mean is calculated correctly"},
			},
		},
		{
			name: "comment marker with function assignment",
			content: `# rtmx:req REQ-STAT-002
test_variance <- function(x) {
  var(x)
}
`,
			expected: []TestRequirement{
				{ReqID: "REQ-STAT-002", TestFunction: "test_variance"},
			},
		},
		{
			name: "multiple markers",
			content: `# rtmx:req REQ-STAT-001
test_that("addition works", {
  expect_equal(1 + 1, 2)
})

# rtmx:req REQ-STAT-002
test_that("subtraction works", {
  expect_equal(2 - 1, 1)
})
`,
			expected: []TestRequirement{
				{ReqID: "REQ-STAT-001", TestFunction: "addition works"},
				{ReqID: "REQ-STAT-002", TestFunction: "subtraction works"},
			},
		},
		{
			name: "no markers",
			content: `test_that("no marker", {
  expect_true(TRUE)
})
`,
			expected: nil,
		},
		{
			name: "empty file",
			content: ``,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "rtmx-r-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			testFile := filepath.Join(tmpDir, "test_stats.R")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			markers, err := extractRMarkersFromFile(testFile)
			if err != nil {
				t.Fatalf("extractRMarkersFromFile failed: %v", err)
			}

			if len(markers) != len(tt.expected) {
				t.Fatalf("Expected %d markers, got %d", len(tt.expected), len(markers))
			}

			for i, exp := range tt.expected {
				got := markers[i]
				if got.ReqID != exp.ReqID {
					t.Errorf("Marker %d: expected ReqID %q, got %q", i, exp.ReqID, got.ReqID)
				}
				if got.TestFunction != exp.TestFunction {
					t.Errorf("Marker %d: expected TestFunction %q, got %q", i, exp.TestFunction, got.TestFunction)
				}
				if got.TestFile != testFile {
					t.Errorf("Marker %d: expected TestFile %q, got %q", i, testFile, got.TestFile)
				}
				if got.LineNumber == 0 {
					t.Errorf("Marker %d: expected non-zero line number", i)
				}
			}
		})
	}
}

func TestExtractRMarkersFileNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-022")

	_, err := extractRMarkersFromFile("/nonexistent/test_stats.R")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestIsRTestFile(t *testing.T) {
	rtmx.Req(t, "REQ-LANG-022")

	tests := []struct {
		path     string
		expected bool
	}{
		{"test_stats.R", true},
		{"stats_test.R", true},
		{"test-stats.R", true},
		{"stats.R", false},
		{"test_stats.py", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isRTestFile(tt.path); got != tt.expected {
				t.Errorf("isRTestFile(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
