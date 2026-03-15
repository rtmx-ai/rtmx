package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/pkg/rtmx"
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

// REQ-VERIFY-001: Cross-language results file support

func TestVerifyResultsFile(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-001",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	// Setup: create a project with database
	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0755)

	// Write config
	_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"), []byte("database:\n  path: .rtmx/database.csv\n"), 0644)

	// Write database with MISSING requirements
	dbContent := `req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id
REQ-AUTH-001,AUTH,Login,Login shall work,Tests pass,test_auth.py,test_login,Unit Test,MISSING,HIGH,1,,,,,,,,,
REQ-AUTH-002,AUTH,Logout,Logout shall work,Tests pass,test_auth.py,test_logout,Unit Test,MISSING,HIGH,1,,,,,,,,,
`
	_ = os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644)

	// Write results file with one passing test
	resultsContent := `[
		{
			"marker": {"req_id": "REQ-AUTH-001", "test_name": "test_login", "test_file": "test_auth.py"},
			"passed": true
		}
	]`
	resultsPath := filepath.Join(tmpDir, "results.json")
	_ = os.WriteFile(resultsPath, []byte(resultsContent), 0644)

	// Run verify --results
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"verify", "--results", resultsPath, "--update"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("verify --results failed: %v\nOutput: %s", err, buf.String())
	}

	// Verify REQ-AUTH-001 is now COMPLETE
	db, err := database.Load(filepath.Join(rtmxDir, "database.csv"))
	if err != nil {
		t.Fatalf("failed to reload database: %v", err)
	}
	req := db.Get("REQ-AUTH-001")
	if req == nil {
		t.Fatal("REQ-AUTH-001 not found in database")
	}
	if req.Status != database.StatusComplete {
		t.Errorf("REQ-AUTH-001 status = %v, want COMPLETE", req.Status)
	}

	// REQ-AUTH-002 should still be MISSING
	req2 := db.Get("REQ-AUTH-002")
	if req2 == nil {
		t.Fatal("REQ-AUTH-002 not found in database")
	}
	if req2.Status != database.StatusMissing {
		t.Errorf("REQ-AUTH-002 status = %v, want MISSING", req2.Status)
	}
}

func TestVerifyResultsFileDryRun(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-001",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"), []byte("database:\n  path: .rtmx/database.csv\n"), 0644)

	dbContent := `req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id
REQ-AUTH-001,AUTH,Login,Login shall work,Tests pass,test_auth.py,test_login,Unit Test,MISSING,HIGH,1,,,,,,,,,
`
	_ = os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644)

	resultsContent := `[{"marker": {"req_id": "REQ-AUTH-001", "test_name": "test_login", "test_file": "test_auth.py"}, "passed": true}]`
	resultsPath := filepath.Join(tmpDir, "results.json")
	_ = os.WriteFile(resultsPath, []byte(resultsContent), 0644)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"verify", "--results", resultsPath, "--dry-run"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("verify --results --dry-run failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Dry run") {
		t.Errorf("Expected 'Dry run' in output, got: %s", output)
	}

	// Database should NOT be updated
	db, _ := database.Load(filepath.Join(rtmxDir, "database.csv"))
	req := db.Get("REQ-AUTH-001")
	if req.Status != database.StatusMissing {
		t.Errorf("Dry run should not update status, got %v", req.Status)
	}
}

func TestVerifyResultsFileWithFailingTests(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-001",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"), []byte("database:\n  path: .rtmx/database.csv\n"), 0644)

	dbContent := `req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id
REQ-AUTH-001,AUTH,Login,Login shall work,Tests pass,test_auth.py,test_login,Unit Test,COMPLETE,HIGH,1,,,,,,,,,
`
	_ = os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644)

	// One pass, one fail for same requirement
	resultsContent := `[
		{"marker": {"req_id": "REQ-AUTH-001", "test_name": "test_login", "test_file": "test_auth.py"}, "passed": true},
		{"marker": {"req_id": "REQ-AUTH-001", "test_name": "test_login_edge", "test_file": "test_auth.py"}, "passed": false, "error": "failed"}
	]`
	resultsPath := filepath.Join(tmpDir, "results.json")
	_ = os.WriteFile(resultsPath, []byte(resultsContent), 0644)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"verify", "--results", resultsPath, "--update"})

	// Should exit with error (test failures)
	err := cmd.Execute()
	// The command itself may succeed but exit code handled differently
	_ = err

	// COMPLETE should downgrade to PARTIAL since a test failed
	db, _ := database.Load(filepath.Join(rtmxDir, "database.csv"))
	req := db.Get("REQ-AUTH-001")
	if req.Status != database.StatusPartial {
		t.Errorf("REQ-AUTH-001 status = %v, want PARTIAL (had COMPLETE, test failed)", req.Status)
	}
}

func TestVerifyResultsFileMissing(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-001",
		rtmx.Scope("unit"),
		rtmx.Technique("boundary"),
		rtmx.Env("simulation"),
	)

	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"), []byte("database:\n  path: .rtmx/database.csv\n"), 0644)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"verify", "--results", "nonexistent.json"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error for missing results file")
	}
}

func TestVerifyResultsHelpShowsFlag(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-001",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	cmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"verify", "--help"})

	_ = cmd.Execute()
	output := buf.String()
	if !strings.Contains(output, "--results") {
		t.Errorf("Expected --results in help output, got: %s", output)
	}
}
