package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx-go/internal/config"
)

// TestBootstrapCommand tests the bootstrap command.
// REQ-GO-027: Go CLI shall implement bootstrap command with tests at internal/cmd/bootstrap_test.go
func TestBootstrapCommand(t *testing.T) {
	// Save original flag values
	origFromTests := bootstrapFromTests
	origFromGitHub := bootstrapFromGitHub
	origFromJira := bootstrapFromJira
	origMerge := bootstrapMerge
	origDryRun := bootstrapDryRun
	origPrefix := bootstrapPrefix

	defer func() {
		// Restore original flag values
		bootstrapFromTests = origFromTests
		bootstrapFromGitHub = origFromGitHub
		bootstrapFromJira = origFromJira
		bootstrapMerge = origMerge
		bootstrapDryRun = origDryRun
		bootstrapPrefix = origPrefix
	}()

	tests := []struct {
		name        string
		fromTests   bool
		fromGitHub  bool
		fromJira    bool
		dryRun      bool
		wantErr     bool
		wantContain string
	}{
		{
			name:        "no source specified",
			fromTests:   false,
			fromGitHub:  false,
			fromJira:    false,
			wantErr:     true,
			wantContain: "No source specified",
		},
		{
			name:        "dry run from tests",
			fromTests:   true,
			fromGitHub:  false,
			fromJira:    false,
			dryRun:      true,
			wantErr:     false,
			wantContain: "DRY RUN",
		},
		{
			name:        "from github without config",
			fromTests:   false,
			fromGitHub:  true,
			fromJira:    false,
			wantErr:     false,
			wantContain: "GitHub adapter not enabled",
		},
		{
			name:        "from jira without config",
			fromTests:   false,
			fromGitHub:  false,
			fromJira:    true,
			wantErr:     false,
			wantContain: "Jira adapter not enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set flags
			bootstrapFromTests = tt.fromTests
			bootstrapFromGitHub = tt.fromGitHub
			bootstrapFromJira = tt.fromJira
			bootstrapMerge = false
			bootstrapDryRun = tt.dryRun
			bootstrapPrefix = "REQ"

			// Create a temp directory for testing
			tmpDir, err := os.MkdirTemp("", "rtmx-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Change to temp directory
			origDir, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(origDir)

			// Capture output
			var buf bytes.Buffer
			bootstrapCmd.SetOut(&buf)
			bootstrapCmd.SetErr(&buf)

			err = runBootstrap(bootstrapCmd, nil)

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

// TestBootstrapFromTestFiles tests bootstrapping from test files.
func TestBootstrapFromTestFiles(t *testing.T) {
	// Create a temp directory with test files
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
    """Test that has a marker."""
    pass

def test_without_marker():
    """Test that does not have a marker."""
    pass

def test_another_unmarked():
    pass
`
	if err := os.WriteFile(filepath.Join(testsDir, "test_sample.py"), []byte(testFile), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Run bootstrap
	reqs := bootstrapFromTestFiles(tmpDir, "REQ")

	// Should find the unmarked tests
	if len(reqs) != 2 {
		t.Errorf("Expected 2 requirements (for unmarked tests), got %d", len(reqs))
	}

	// Check requirement properties
	for _, req := range reqs {
		if req.Source != "test" {
			t.Errorf("Expected source 'test', got %q", req.Source)
		}
		if !strings.HasPrefix(req.ID, "REQ-") {
			t.Errorf("Expected ID to start with 'REQ-', got %q", req.ID)
		}
	}
}

// TestInferCategoryFromPath tests category inference from file paths.
func TestInferCategoryFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"tests/test_models.py", "MODELS"},
		{"tests/cli/test_status.py", "CLI"},
		{"test/integration/test_api.py", "INTEGRATION"},
		{"tests/test_auth.py", "AUTH"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := inferCategoryFromPath(tt.path)
			if result != tt.expected {
				t.Errorf("inferCategoryFromPath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

// TestInferRequirementText tests requirement text inference from function names.
func TestInferRequirementText(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		funcLine int
		funcName string
		expected string
	}{
		{
			name: "from docstring",
			lines: []string{
				"def test_user_login():",
				`    """User can log in with valid credentials."""`,
				"    pass",
			},
			funcLine: 0,
			funcName: "test_user_login",
			expected: "User can log in with valid credentials.",
		},
		{
			name: "from function name",
			lines: []string{
				"def test_user_can_login():",
				"    pass",
			},
			funcLine: 0,
			funcName: "test_user_can_login",
			expected: "User can login",
		},
		{
			name: "single quote docstring",
			lines: []string{
				"def test_feature():",
				`    '''Feature works correctly.'''`,
				"    pass",
			},
			funcLine: 0,
			funcName: "test_feature",
			expected: "Feature works correctly.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferRequirementText(tt.lines, tt.funcLine, tt.funcName)
			if result != tt.expected {
				t.Errorf("inferRequirementText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestTruncateString tests string truncation.
func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly ten", 11, "exactly ten"},
		{"this is a longer string", 10, "this is..."},
		{"", 5, ""},
		{"abc", 3, "abc"},
		{"abcd", 3, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestWriteBootstrapRequirements tests writing requirements to CSV.
func TestWriteBootstrapRequirements(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config
	cfg := config.DefaultConfig()

	reqs := []BootstrapRequirement{
		{
			ID:         "REQ-TEST-001",
			Category:   "TEST",
			Text:       "First requirement",
			TestModule: "tests/test_sample.py",
			TestFunc:   "test_first",
			Source:     "test",
		},
		{
			ID:         "REQ-TEST-002",
			Category:   "TEST",
			Text:       "Second requirement with, comma",
			TestModule: "tests/test_sample.py",
			TestFunc:   "test_second",
			Source:     "test",
		},
	}

	err = writeBootstrapRequirements(tmpDir, cfg, reqs, false)
	if err != nil {
		t.Fatalf("writeBootstrapRequirements failed: %v", err)
	}

	// Read and verify
	dbPath := filepath.Join(tmpDir, ".rtmx", "database.csv")
	content, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("Failed to read database: %v", err)
	}

	contentStr := string(content)

	// Check header exists
	if !strings.Contains(contentStr, "req_id,category") {
		t.Error("Database should contain header")
	}

	// Check requirements exist
	if !strings.Contains(contentStr, "REQ-TEST-001") {
		t.Error("Database should contain REQ-TEST-001")
	}
	if !strings.Contains(contentStr, "REQ-TEST-002") {
		t.Error("Database should contain REQ-TEST-002")
	}

	// Check comma is properly escaped
	if !strings.Contains(contentStr, `"Second requirement with, comma"`) {
		t.Error("Comma in text should be quoted")
	}
}

// TestWriteBootstrapRequirementsMerge tests merging with existing database.
func TestWriteBootstrapRequirementsMerge(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create existing database
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0755); err != nil {
		t.Fatalf("Failed to create .rtmx dir: %v", err)
	}

	existingContent := "req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id\nREQ-EXIST-001,EXIST,,Existing requirement,,,,,COMPLETE,HIGH,1,,,,,,,,,.rtmx/requirements/EXIST/REQ-EXIST-001.md,\n"
	dbPath := filepath.Join(rtmxDir, "database.csv")
	if err := os.WriteFile(dbPath, []byte(existingContent), 0644); err != nil {
		t.Fatalf("Failed to write existing database: %v", err)
	}

	// Create config
	cfg := config.DefaultConfig()

	reqs := []BootstrapRequirement{
		{
			ID:         "REQ-NEW-001",
			Category:   "NEW",
			Text:       "New requirement",
			TestModule: "tests/test_new.py",
			TestFunc:   "test_new",
			Source:     "test",
		},
	}

	err = writeBootstrapRequirements(tmpDir, cfg, reqs, true)
	if err != nil {
		t.Fatalf("writeBootstrapRequirements failed: %v", err)
	}

	// Read and verify
	content, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("Failed to read database: %v", err)
	}

	contentStr := string(content)

	// Check both old and new requirements exist
	if !strings.Contains(contentStr, "REQ-EXIST-001") {
		t.Error("Database should still contain existing requirement")
	}
	if !strings.Contains(contentStr, "REQ-NEW-001") {
		t.Error("Database should contain new requirement")
	}
}

// TestBootstrapCommandFlags tests that all flags are properly registered.
func TestBootstrapCommandFlags(t *testing.T) {
	flags := bootstrapCmd.Flags()

	expectedFlags := []string{
		"from-tests",
		"from-github",
		"from-jira",
		"merge",
		"dry-run",
		"prefix",
	}

	for _, name := range expectedFlags {
		if flags.Lookup(name) == nil {
			t.Errorf("Missing flag: %s", name)
		}
	}
}

// TestBootstrapEmptyTestsDir tests bootstrap with no tests directory.
func TestBootstrapEmptyTestsDir(t *testing.T) {
	// Create a temp directory without test files
	tmpDir, err := os.MkdirTemp("", "rtmx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Run bootstrap
	reqs := bootstrapFromTestFiles(tmpDir, "REQ")

	// Should find no requirements
	if len(reqs) != 0 {
		t.Errorf("Expected 0 requirements, got %d", len(reqs))
	}
}

// TestBootstrapSubdirectory tests bootstrap with nested test directories.
func TestBootstrapSubdirectory(t *testing.T) {
	// Create a temp directory with nested test structure
	tmpDir, err := os.MkdirTemp("", "rtmx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create nested tests directory
	cliDir := filepath.Join(tmpDir, "tests", "cli")
	if err := os.MkdirAll(cliDir, 0755); err != nil {
		t.Fatalf("Failed to create tests/cli dir: %v", err)
	}

	// Create test file in subdirectory
	testFile := `def test_cli_command():
    """Test CLI command."""
    pass
`
	if err := os.WriteFile(filepath.Join(cliDir, "test_commands.py"), []byte(testFile), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Run bootstrap
	reqs := bootstrapFromTestFiles(tmpDir, "REQ")

	// Should find one requirement
	if len(reqs) != 1 {
		t.Errorf("Expected 1 requirement, got %d", len(reqs))
	}

	// Category should be CLI (from subdirectory)
	if len(reqs) > 0 && reqs[0].Category != "CLI" {
		t.Errorf("Expected category 'CLI', got %q", reqs[0].Category)
	}
}
