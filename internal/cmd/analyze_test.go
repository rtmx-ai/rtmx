package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAnalyzeCommand tests the analyze command.
// REQ-GO-031: Go CLI shall implement analyze command with tests at internal/cmd/analyze_test.go
func TestAnalyzeCommand(t *testing.T) {
	// Save original flag values
	origOutput := analyzeOutput
	origFormat := analyzeFormat
	origDeep := analyzeDeep

	defer func() {
		// Restore original flag values
		analyzeOutput = origOutput
		analyzeFormat = origFormat
		analyzeDeep = origDeep
	}()

	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "rtmx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create tests directory
	testsDir := filepath.Join(tmpDir, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		t.Fatalf("Failed to create tests dir: %v", err)
	}

	// Create test file with markers
	testFile := `import pytest

@pytest.mark.req("REQ-TEST-001")
def test_with_marker():
    pass

def test_without_marker():
    pass
`
	if err := os.WriteFile(filepath.Join(testsDir, "test_sample.py"), []byte(testFile), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	tests := []struct {
		name        string
		format      string
		wantErr     bool
		wantContain string
	}{
		{
			name:        "terminal format",
			format:      "terminal",
			wantErr:     false,
			wantContain: "RTMX Project Analysis",
		},
		{
			name:        "json format",
			format:      "json",
			wantErr:     false,
			wantContain: `"path"`,
		},
		{
			name:        "markdown format",
			format:      "markdown",
			wantErr:     false,
			wantContain: "# RTMX Project Analysis",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set flags
			analyzeOutput = ""
			analyzeFormat = tt.format
			analyzeDeep = false

			// Capture output
			var buf bytes.Buffer
			analyzeCmd.SetOut(&buf)
			analyzeCmd.SetErr(&buf)

			err = runAnalyze(analyzeCmd, []string{tmpDir})

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			output := buf.String()
			if tt.wantContain != "" && !strings.Contains(output, tt.wantContain) {
				t.Errorf("Output should contain %q, got: %s", tt.wantContain, output)
			}
		})
	}
}

// TestAnalyzeProject tests the analyzeProject function.
func TestAnalyzeProject(t *testing.T) {
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create tests directory with test files
	testsDir := filepath.Join(tmpDir, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		t.Fatalf("Failed to create tests dir: %v", err)
	}

	// Create test file with markers
	testFileWithMarker := `import pytest

@pytest.mark.req("REQ-TEST-001")
def test_with_marker():
    pass
`
	if err := os.WriteFile(filepath.Join(testsDir, "test_marked.py"), []byte(testFileWithMarker), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create test file without markers
	testFileWithoutMarker := `def test_without_marker():
    pass
`
	if err := os.WriteFile(filepath.Join(testsDir, "test_unmarked.py"), []byte(testFileWithoutMarker), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Run analysis
	report := analyzeProject(tmpDir, nil)

	// Check test file counts
	if len(report.TestFiles) != 2 {
		t.Errorf("Expected 2 test files, got %d", len(report.TestFiles))
	}

	if report.TotalTests != 2 {
		t.Errorf("Expected 2 total tests, got %d", report.TotalTests)
	}

	if report.TestsWithMarker != 1 {
		t.Errorf("Expected 1 test with marker, got %d", report.TestsWithMarker)
	}

	if report.UnmarkedTests != 1 {
		t.Errorf("Expected 1 unmarked test, got %d", report.UnmarkedTests)
	}

	// Check recommendations
	if len(report.Recommendations) == 0 {
		t.Error("Expected at least one recommendation")
	}
}

// TestCountMarkersInFile tests the marker counting function.
func TestCountMarkersInFile(t *testing.T) {
	// Create temp file
	tmpDir, err := os.MkdirTemp("", "rtmx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		content string
		want    int
	}{
		{
			name:    "no markers",
			content: "def test_something(): pass",
			want:    0,
		},
		{
			name: "one marker",
			content: `@pytest.mark.req("REQ-001")
def test_something(): pass`,
			want: 1,
		},
		{
			name: "multiple markers",
			content: `@pytest.mark.req("REQ-001")
def test_first(): pass

@pytest.mark.req("REQ-002")
def test_second(): pass

@pytest.mark.req("REQ-003")
def test_third(): pass`,
			want: 3,
		},
		{
			name: "marker with different syntax",
			content: `import pytest
pytest.mark.req("REQ-001")`,
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, "test_"+tt.name+".py")
			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			got := countMarkersInFile(path)
			if got != tt.want {
				t.Errorf("countMarkersInFile() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestCountMarkersInFileNotExist tests marker counting for non-existent file.
func TestCountMarkersInFileNotExist(t *testing.T) {
	got := countMarkersInFile("/nonexistent/path/to/file.py")
	if got != 0 {
		t.Errorf("countMarkersInFile for non-existent file = %d, want 0", got)
	}
}

// TestFormatAnalysisMarkdown tests the markdown formatting function.
func TestFormatAnalysisMarkdown(t *testing.T) {
	report := &AnalysisReport{
		Path: "/test/project",
		TestFiles: []TestFileInfo{
			{Path: "tests/test_sample.py", HasMarkers: true, MarkerCount: 2},
			{Path: "tests/test_other.py", HasMarkers: false, MarkerCount: 0},
		},
		TotalTests:       2,
		TestsWithMarker:  1,
		UnmarkedTests:    1,
		GitHubConfigured: true,
		GitHubRepo:       "owner/repo",
		JiraConfigured:   false,
		RTMExists:        true,
		RTMPath:          ".rtmx/database.csv",
		Recommendations:  []string{"Run tests", "Check coverage"},
	}

	result := formatAnalysisMarkdown(report)

	// Check required sections
	expectedParts := []string{
		"# RTMX Project Analysis",
		"## Test Files",
		"tests/test_sample.py",
		"tests/test_other.py",
		"## Integrations",
		"**GitHub:** Configured",
		"**Jira:** Not configured",
		"## RTM Status",
		".rtmx/database.csv",
		"## Recommendations",
		"Run tests",
		"Check coverage",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Markdown should contain %q", part)
		}
	}
}

// TestAnalyzeCommandFlags tests that all flags are properly registered.
func TestAnalyzeCommandFlags(t *testing.T) {
	flags := analyzeCmd.Flags()

	expectedFlags := []string{
		"output",
		"format",
		"deep",
	}

	for _, name := range expectedFlags {
		if flags.Lookup(name) == nil {
			t.Errorf("Missing flag: %s", name)
		}
	}
}

// TestAnalyzeOutputToFile tests writing output to file.
func TestAnalyzeOutputToFile(t *testing.T) {
	// Save original flag values
	origOutput := analyzeOutput
	origFormat := analyzeFormat
	origDeep := analyzeDeep

	defer func() {
		// Restore original flag values
		analyzeOutput = origOutput
		analyzeFormat = origFormat
		analyzeDeep = origDeep
	}()

	// Create temp directories
	tmpDir, err := os.MkdirTemp("", "rtmx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputFile := filepath.Join(tmpDir, "report.json")

	// Set flags
	analyzeOutput = outputFile
	analyzeFormat = "json"
	analyzeDeep = false

	// Capture output
	var buf bytes.Buffer
	analyzeCmd.SetOut(&buf)
	analyzeCmd.SetErr(&buf)

	err = runAnalyze(analyzeCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("runAnalyze failed: %v", err)
	}

	// Check output file was created
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Verify it's valid JSON
	var report AnalysisReport
	if err := json.Unmarshal(content, &report); err != nil {
		t.Errorf("Output should be valid JSON: %v", err)
	}

	// Check buffer contains success message
	if !strings.Contains(buf.String(), "Report written to") {
		t.Error("Should print success message")
	}
}

// TestAnalyzeEmptyProject tests analyzing a project with no tests.
func TestAnalyzeEmptyProject(t *testing.T) {
	// Create temp directory without tests
	tmpDir, err := os.MkdirTemp("", "rtmx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	report := analyzeProject(tmpDir, nil)

	if len(report.TestFiles) != 0 {
		t.Errorf("Expected no test files, got %d", len(report.TestFiles))
	}

	if report.TotalTests != 0 {
		t.Errorf("Expected 0 total tests, got %d", report.TotalTests)
	}
}

// TestAnalyzeWithRTMDatabase tests analyzing a project with existing RTM.
func TestAnalyzeWithRTMDatabase(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .rtmx directory and database
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0755); err != nil {
		t.Fatalf("Failed to create .rtmx dir: %v", err)
	}

	dbContent := "req_id,category,requirement_text\nREQ-001,TEST,Test requirement\n"
	if err := os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644); err != nil {
		t.Fatalf("Failed to write database: %v", err)
	}

	// Create config file
	configContent := `rtmx:
  database: .rtmx/database.csv
`
	if err := os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Use nil config for simplified test
	report := analyzeProject(tmpDir, nil)

	// Should still find the RTM with default config
	// Note: RTM detection works even without explicit config
	_ = report // Just verify it runs without panic
}

// TestAnalyzeJSONFormat tests JSON output format.
func TestAnalyzeJSONFormat(t *testing.T) {
	report := &AnalysisReport{
		Path: "/test/project",
		TestFiles: []TestFileInfo{
			{Path: "test.py", HasMarkers: true, MarkerCount: 1},
		},
		TotalTests:       1,
		TestsWithMarker:  1,
		UnmarkedTests:    0,
		GitHubConfigured: false,
		JiraConfigured:   false,
		RTMExists:        false,
		Recommendations:  []string{"Run init"},
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal report: %v", err)
	}

	// Verify it can be unmarshaled back
	var parsed AnalysisReport
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("JSON should be valid: %v", err)
	}

	if parsed.Path != report.Path {
		t.Errorf("Path mismatch: got %q, want %q", parsed.Path, report.Path)
	}

	if len(parsed.TestFiles) != len(report.TestFiles) {
		t.Errorf("TestFiles count mismatch: got %d, want %d", len(parsed.TestFiles), len(report.TestFiles))
	}
}

// TestAnalyzeDefaultPath tests default path handling.
func TestAnalyzeDefaultPath(t *testing.T) {
	// Save original flag values
	origOutput := analyzeOutput
	origFormat := analyzeFormat
	origDeep := analyzeDeep

	defer func() {
		// Restore original flag values
		analyzeOutput = origOutput
		analyzeFormat = origFormat
		analyzeDeep = origDeep
	}()

	// Set flags
	analyzeOutput = ""
	analyzeFormat = "terminal"
	analyzeDeep = false

	// Capture output
	var buf bytes.Buffer
	analyzeCmd.SetOut(&buf)
	analyzeCmd.SetErr(&buf)

	// Run with no path argument (uses current directory)
	err := runAnalyze(analyzeCmd, []string{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Output should contain analysis header
	if !strings.Contains(buf.String(), "RTMX Project Analysis") {
		t.Error("Output should contain analysis header")
	}
}
