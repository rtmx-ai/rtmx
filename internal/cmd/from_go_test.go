package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFromGoCommand tests the from-go command.
// REQ-LANG-003: Go testing integration with helper functions
func TestFromGoCommand(t *testing.T) {
	// Save original flag values
	origUpdate := fromGoUpdate
	origDryRun := fromGoDryRun
	origVerbose := fromGoVerbose

	defer func() {
		// Restore original flag values
		fromGoUpdate = origUpdate
		fromGoDryRun = origDryRun
		fromGoVerbose = origVerbose
	}()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a sample rtmx-results.json file
	resultsContent := `[
  {
    "marker": {
      "req_id": "REQ-TEST-001",
      "scope": "unit",
      "technique": "nominal",
      "env": "",
      "test_name": "TestFeature",
      "test_file": "feature_test.go",
      "line": 10
    },
    "passed": true,
    "duration_ms": 5.2,
    "error": "",
    "timestamp": "2026-02-17T10:00:00Z"
  },
  {
    "marker": {
      "req_id": "REQ-TEST-002",
      "scope": "integration",
      "technique": "",
      "env": "",
      "test_name": "TestOther",
      "test_file": "other_test.go",
      "line": 25
    },
    "passed": false,
    "duration_ms": 10.5,
    "error": "assertion failed",
    "timestamp": "2026-02-17T10:00:01Z"
  }
]`
	resultsPath := filepath.Join(tmpDir, "rtmx-results.json")
	if err := os.WriteFile(resultsPath, []byte(resultsContent), 0644); err != nil {
		t.Fatalf("Failed to write results file: %v", err)
	}

	tests := []struct {
		name        string
		verbose     bool
		dryRun      bool
		wantContain string
	}{
		{
			name:        "basic import",
			verbose:     false,
			dryRun:      false,
			wantContain: "Total results: 2",
		},
		{
			name:        "verbose output",
			verbose:     true,
			dryRun:      false,
			wantContain: "REQ-TEST-001",
		},
		{
			name:        "dry run",
			verbose:     false,
			dryRun:      true,
			wantContain: "DRY RUN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set flags
			fromGoUpdate = false
			fromGoDryRun = tt.dryRun
			fromGoVerbose = tt.verbose

			// Capture output
			var buf bytes.Buffer
			fromGoCmd.SetOut(&buf)
			fromGoCmd.SetErr(&buf)

			// Run command
			err := runFromGo(fromGoCmd, []string{resultsPath})
			if err != nil {
				t.Errorf("runFromGo failed: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, tt.wantContain) {
				t.Errorf("Output should contain %q, got: %s", tt.wantContain, output)
			}
		})
	}
}

// TestFromGoWithDatabase tests importing results into a database.
func TestFromGoWithDatabase(t *testing.T) {
	// Save original flag values
	origUpdate := fromGoUpdate
	origDryRun := fromGoDryRun
	origVerbose := fromGoVerbose

	defer func() {
		fromGoUpdate = origUpdate
		fromGoDryRun = origDryRun
		fromGoVerbose = origVerbose
	}()

	// Create temp directory with rtmx structure
	tmpDir, err := os.MkdirTemp("", "rtmx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .rtmx directory
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0755); err != nil {
		t.Fatalf("Failed to create .rtmx dir: %v", err)
	}

	// Create database with a matching requirement
	dbContent := `req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id
REQ-TEST-001,TEST,,Test requirement,,,TestFeature,Unit Test,MISSING,HIGH,1,,0.5,,,,,,,,.rtmx/requirements/TEST/REQ-TEST-001.md,
`
	dbPath := filepath.Join(rtmxDir, "database.csv")
	if err := os.WriteFile(dbPath, []byte(dbContent), 0644); err != nil {
		t.Fatalf("Failed to write database: %v", err)
	}

	// Create results file
	resultsContent := `[
  {
    "marker": {
      "req_id": "REQ-TEST-001",
      "scope": "unit",
      "technique": "",
      "env": "",
      "test_name": "TestFeature",
      "test_file": "feature_test.go",
      "line": 10
    },
    "passed": true,
    "duration_ms": 5.2,
    "error": "",
    "timestamp": "2026-02-17T10:00:00Z"
  }
]`
	resultsPath := filepath.Join(tmpDir, "rtmx-results.json")
	if err := os.WriteFile(resultsPath, []byte(resultsContent), 0644); err != nil {
		t.Fatalf("Failed to write results file: %v", err)
	}

	// Change to temp directory
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Set flags
	fromGoUpdate = true
	fromGoDryRun = false
	fromGoVerbose = true

	// Capture output
	var buf bytes.Buffer
	fromGoCmd.SetOut(&buf)
	fromGoCmd.SetErr(&buf)

	// Run command
	err = runFromGo(fromGoCmd, []string{resultsPath})
	if err != nil {
		t.Fatalf("runFromGo failed: %v", err)
	}

	output := buf.String()

	// Should show status change
	if !strings.Contains(output, "COMPLETE") {
		t.Errorf("Output should show COMPLETE status, got: %s", output)
	}

	// Read updated database
	updatedDB, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("Failed to read updated database: %v", err)
	}

	if !strings.Contains(string(updatedDB), "COMPLETE") {
		t.Errorf("Database should be updated to COMPLETE, got: %s", string(updatedDB))
	}
}

// TestFromGoEmptyResults tests handling of empty results file.
func TestFromGoEmptyResults(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create empty results file
	resultsPath := filepath.Join(tmpDir, "empty-results.json")
	if err := os.WriteFile(resultsPath, []byte("[]"), 0644); err != nil {
		t.Fatalf("Failed to write results file: %v", err)
	}

	// Capture output
	var buf bytes.Buffer
	fromGoCmd.SetOut(&buf)
	fromGoCmd.SetErr(&buf)

	err = runFromGo(fromGoCmd, []string{resultsPath})
	if err != nil {
		t.Errorf("runFromGo failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No results to import") {
		t.Errorf("Should say no results to import, got: %s", output)
	}
}

// TestFromGoInvalidJSON tests handling of invalid JSON.
func TestFromGoInvalidJSON(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create invalid JSON file
	resultsPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(resultsPath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("Failed to write results file: %v", err)
	}

	err = runFromGo(fromGoCmd, []string{resultsPath})
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

// TestFromGoFileNotFound tests handling of missing results file.
func TestFromGoFileNotFound(t *testing.T) {
	err := runFromGo(fromGoCmd, []string{"/nonexistent/path/results.json"})
	if err == nil {
		t.Error("Expected error for missing file")
	}
}

// TestCountPassed tests the countPassed helper function.
func TestCountPassed(t *testing.T) {
	tests := []struct {
		name    string
		results []GoTestResult
		want    int
	}{
		{
			name:    "empty",
			results: []GoTestResult{},
			want:    0,
		},
		{
			name: "all passed",
			results: []GoTestResult{
				{Passed: true},
				{Passed: true},
			},
			want: 2,
		},
		{
			name: "mixed",
			results: []GoTestResult{
				{Passed: true},
				{Passed: false},
				{Passed: true},
			},
			want: 2,
		},
		{
			name: "all failed",
			results: []GoTestResult{
				{Passed: false},
				{Passed: false},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countPassed(tt.results)
			if got != tt.want {
				t.Errorf("countPassed() = %d, want %d", got, tt.want)
			}
		})
	}
}
