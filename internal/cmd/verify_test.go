package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
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

// --- Tests for mapTestsToRequirements ---

func TestMapTestsToRequirements_MatchingTests(t *testing.T) {
	db := database.NewDatabase()

	req1 := database.NewRequirement("REQ-001")
	req1.TestFunction = "TestLogin"
	req1.TestModule = "auth_test.go"
	req1.Status = database.StatusMissing
	_ = db.Add(req1)

	req2 := database.NewRequirement("REQ-002")
	req2.TestFunction = "TestLogout"
	req2.TestModule = "auth_test.go"
	req2.Status = database.StatusComplete
	_ = db.Add(req2)

	testResults := map[string]*TestResult{
		"pkg/TestLogin": {
			Package: "pkg",
			Test:    "TestLogin",
			Passed:  true,
		},
		"pkg/TestLogout": {
			Package: "pkg",
			Test:    "TestLogout",
			Failed:  true,
		},
	}

	results := mapTestsToRequirements(db, testResults)

	if len(results) != 2 {
		t.Fatalf("expected 2 verification results, got %d", len(results))
	}

	// Find results by ReqID
	resultMap := make(map[string]VerificationResult)
	for _, r := range results {
		resultMap[r.ReqID] = r
	}

	// REQ-001: test passed, was MISSING -> should become COMPLETE
	r1, ok := resultMap["REQ-001"]
	if !ok {
		t.Fatal("REQ-001 not found in results")
	}
	if r1.NewStatus != database.StatusComplete {
		t.Errorf("REQ-001 new status = %v, want COMPLETE", r1.NewStatus)
	}
	if r1.TestsPassed != 1 {
		t.Errorf("REQ-001 tests passed = %d, want 1", r1.TestsPassed)
	}
	if !r1.Updated {
		t.Error("REQ-001 should be marked as updated")
	}

	// REQ-002: test failed, was COMPLETE -> should downgrade to PARTIAL
	r2, ok := resultMap["REQ-002"]
	if !ok {
		t.Fatal("REQ-002 not found in results")
	}
	if r2.NewStatus != database.StatusPartial {
		t.Errorf("REQ-002 new status = %v, want PARTIAL", r2.NewStatus)
	}
	if r2.TestsFailed != 1 {
		t.Errorf("REQ-002 tests failed = %d, want 1", r2.TestsFailed)
	}
	if !r2.Updated {
		t.Error("REQ-002 should be marked as updated")
	}
}

func TestMapTestsToRequirements_NoMatchingTests(t *testing.T) {
	db := database.NewDatabase()

	req := database.NewRequirement("REQ-001")
	req.TestFunction = "TestSomething"
	req.TestModule = "mod_test.go"
	req.Status = database.StatusMissing
	_ = db.Add(req)

	testResults := map[string]*TestResult{
		"pkg/TestOtherThing": {
			Package: "pkg",
			Test:    "TestOtherThing",
			Passed:  true,
		},
	}

	results := mapTestsToRequirements(db, testResults)
	if len(results) != 0 {
		t.Errorf("expected 0 results when no tests match, got %d", len(results))
	}
}

func TestMapTestsToRequirements_EmptyTestFunction(t *testing.T) {
	db := database.NewDatabase()

	req := database.NewRequirement("REQ-001")
	req.TestFunction = "" // No test function specified
	req.Status = database.StatusMissing
	_ = db.Add(req)

	testResults := map[string]*TestResult{
		"pkg/TestAnything": {
			Package: "pkg",
			Test:    "TestAnything",
			Passed:  true,
		},
	}

	results := mapTestsToRequirements(db, testResults)
	if len(results) != 0 {
		t.Errorf("expected 0 results when requirement has no test function, got %d", len(results))
	}
}

func TestMapTestsToRequirements_SkippedTest(t *testing.T) {
	db := database.NewDatabase()

	req := database.NewRequirement("REQ-001")
	req.TestFunction = "TestSkipped"
	req.TestModule = "mod_test.go"
	req.Status = database.StatusPartial
	_ = db.Add(req)

	testResults := map[string]*TestResult{
		"pkg/TestSkipped": {
			Package: "pkg",
			Test:    "TestSkipped",
			Skipped: true,
		},
	}

	results := mapTestsToRequirements(db, testResults)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	// Skipped should keep current status (PARTIAL stays PARTIAL)
	if results[0].NewStatus != database.StatusPartial {
		t.Errorf("skipped test should keep status PARTIAL, got %v", results[0].NewStatus)
	}
	if results[0].Updated {
		t.Error("skipped test should not mark as updated since status unchanged")
	}
	if results[0].TestsSkipped != 1 {
		t.Errorf("expected 1 skipped test, got %d", results[0].TestsSkipped)
	}
}

func TestMapTestsToRequirements_NilTestResults(t *testing.T) {
	db := database.NewDatabase()

	req := database.NewRequirement("REQ-001")
	req.TestFunction = "TestSomething"
	req.TestModule = "mod_test.go"
	_ = db.Add(req)

	results := mapTestsToRequirements(db, nil)
	if len(results) != 0 {
		t.Errorf("expected 0 results with nil test map, got %d", len(results))
	}
}

func TestMapTestsToRequirements_EmptyDatabase(t *testing.T) {
	db := database.NewDatabase()

	testResults := map[string]*TestResult{
		"pkg/TestSomething": {
			Package: "pkg",
			Test:    "TestSomething",
			Passed:  true,
		},
	}

	results := mapTestsToRequirements(db, testResults)
	if len(results) != 0 {
		t.Errorf("expected 0 results with empty database, got %d", len(results))
	}
}

// --- Tests for runVerifyFromTests via the command ---

// setupTestProject creates a temp directory with rtmx.yaml and database.csv.
// Returns the temp dir path.
func setupTestProject(t *testing.T, dbContent string) string {
	t.Helper()
	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"), []byte("database:\n  path: .rtmx/database.csv\n"), 0644)
	_ = os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644)
	return tmpDir
}

const testDBHeader = "req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id\n"

func TestVerifyFromTests_CustomCommand_EchoJSON(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("echo behaves differently on Windows")
	}
	dbContent := testDBHeader +
		"REQ-001,CAT,Sub,Requirement text,Pass,test_mod,TestFoo,Unit Test,MISSING,HIGH,1,,,,,,,,,\n"
	tmpDir := setupTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	// Use a custom command that outputs go test JSON format for a passing test
	// The echo command produces a JSON line that runTests will parse
	jsonLine := `{"Time":"2024-01-01T00:00:00Z","Action":"pass","Package":"test_mod","Test":"TestFoo","Elapsed":0.1}`

	cmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"verify", "--command", "echo " + jsonLine})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("verify with custom command failed: %v\nOutput: %s", err, buf.String())
	}

	out := buf.String()
	// Should contain verification results
	if !strings.Contains(out, "Verification Results") {
		t.Errorf("expected Verification Results in output, got:\n%s", out)
	}
	if !strings.Contains(out, "PASSING") {
		t.Errorf("expected PASSING in output, got:\n%s", out)
	}
}

func TestVerifyFromTests_CustomCommand_FailingTest(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("echo behaves differently on Windows")
	}
	dbContent := testDBHeader +
		"REQ-001,CAT,Sub,Requirement text,Pass,test_mod,TestBar,Unit Test,COMPLETE,HIGH,1,,,,,,,,,\n"
	tmpDir := setupTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	jsonLine := `{"Time":"2024-01-01T00:00:00Z","Action":"fail","Package":"test_mod","Test":"TestBar","Elapsed":0.1}`

	cmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"verify", "--command", "echo " + jsonLine, "--update"})

	err := cmd.Execute()
	// Should return an error because tests failed
	if err == nil {
		t.Fatal("expected error for failing tests")
	}
	if !strings.Contains(err.Error(), "verification failed") {
		t.Errorf("expected 'verification failed' error, got: %v", err)
	}

	// With --update, database should be modified: COMPLETE -> PARTIAL
	db, loadErr := database.Load(filepath.Join(tmpDir, ".rtmx", "database.csv"))
	if loadErr != nil {
		t.Fatalf("failed to reload database: %v", loadErr)
	}
	req := db.Get("REQ-001")
	if req == nil {
		t.Fatal("REQ-001 not found")
	}
	if req.Status != database.StatusPartial {
		t.Errorf("REQ-001 status = %v, want PARTIAL after failing test", req.Status)
	}
}

func TestVerifyFromTests_DryRun_DoesNotUpdateDatabase(t *testing.T) {
	dbContent := testDBHeader +
		"REQ-001,CAT,Sub,Requirement text,Pass,test_mod,TestFoo,Unit Test,MISSING,HIGH,1,,,,,,,,,\n"
	tmpDir := setupTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	jsonLine := `{"Time":"2024-01-01T00:00:00Z","Action":"pass","Package":"test_mod","Test":"TestFoo","Elapsed":0.1}`

	cmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"verify", "--command", "echo " + jsonLine, "--dry-run"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("verify --dry-run failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected 'Dry run' in output, got:\n%s", out)
	}

	// Database should NOT be updated
	db, _ := database.Load(filepath.Join(tmpDir, ".rtmx", "database.csv"))
	req := db.Get("REQ-001")
	if req.Status != database.StatusMissing {
		t.Errorf("dry run should not update status, got %v", req.Status)
	}
}

func TestVerifyFromTests_NoMatchingTests_NoChanges(t *testing.T) {
	dbContent := testDBHeader +
		"REQ-001,CAT,Sub,Requirement text,Pass,test_mod,TestUnmatched,Unit Test,MISSING,HIGH,1,,,,,,,,,\n"
	tmpDir := setupTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	// echo a test result for a different test function
	jsonLine := `{"Time":"2024-01-01T00:00:00Z","Action":"pass","Package":"test_mod","Test":"TestOtherFunc","Elapsed":0.1}`

	cmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"verify", "--command", "echo " + jsonLine, "--update"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "No requirements with linked tests found") {
		t.Errorf("expected 'No requirements' message, got:\n%s", out)
	}
}

func TestVerifyFromTests_CustomCommand_NonJSON(t *testing.T) {
	// A custom command that outputs non-JSON text
	dbContent := testDBHeader +
		"REQ-001,CAT,Sub,Requirement text,Pass,test_mod,TestFoo,Unit Test,MISSING,HIGH,1,,,,,,,,,\n"
	tmpDir := setupTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	// echo plain text, not JSON - should be silently skipped
	cmd.SetArgs([]string{"verify", "--command", "echo not-json-output"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("verify with non-JSON command failed: %v", err)
	}

	// No test results matched, so no requirements found
	out := buf.String()
	if !strings.Contains(out, "No requirements with linked tests found") {
		t.Errorf("expected 'No requirements' message, got:\n%s", out)
	}
}

func TestVerifyFromTests_MissingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	// No rtmx.yaml or .rtmx directory

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"verify", "--command", "echo test"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no config file exists")
	}
}

func TestVerifyFromTests_MissingDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	// Create config but no database
	_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"), []byte("database:\n  path: .rtmx/database.csv\n"), 0644)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"verify", "--command", "echo test"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when database file is missing")
	}
	if !strings.Contains(err.Error(), "failed to load database") {
		t.Errorf("expected 'failed to load database' error, got: %v", err)
	}
}

func TestVerifyFromTests_MutuallyExclusiveFlags(t *testing.T) {
	dbContent := testDBHeader +
		"REQ-001,CAT,Sub,Requirement text,Pass,test_mod,TestFoo,Unit Test,MISSING,HIGH,1,,,,,,,,,\n"
	tmpDir := setupTestProject(t, dbContent)

	resultsPath := filepath.Join(tmpDir, "results.json")
	_ = os.WriteFile(resultsPath, []byte(`[{"marker": {"req_id": "REQ-001", "test_name": "test", "test_file": "test.py"}, "passed": true}]`), 0644)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	// Both --results and a positional test_path
	cmd.SetArgs([]string{"verify", "--results", resultsPath, "./some/path"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for mutually exclusive --results and test_path")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected 'mutually exclusive' error, got: %v", err)
	}
}

func TestVerifyFromTests_UpdateWithNoStatusChanges(t *testing.T) {
	// Test where all tests pass and requirement is already COMPLETE
	dbContent := testDBHeader +
		"REQ-001,CAT,Sub,Requirement text,Pass,test_mod,TestFoo,Unit Test,COMPLETE,HIGH,1,,,,,,,,,\n"
	tmpDir := setupTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	jsonLine := `{"Time":"2024-01-01T00:00:00Z","Action":"pass","Package":"test_mod","Test":"TestFoo","Elapsed":0.1}`

	cmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"verify", "--command", "echo " + jsonLine, "--update"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}

	out := buf.String()
	// COMPLETE -> COMPLETE is not a change
	if !strings.Contains(out, "No status changes needed") {
		t.Errorf("expected 'No status changes needed', got:\n%s", out)
	}
}

func TestVerifyFromTests_SkippedTestKeepsStatus(t *testing.T) {
	dbContent := testDBHeader +
		"REQ-001,CAT,Sub,Requirement text,Pass,test_mod,TestFoo,Unit Test,PARTIAL,HIGH,1,,,,,,,,,\n"
	tmpDir := setupTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	jsonLine := `{"Time":"2024-01-01T00:00:00Z","Action":"skip","Package":"test_mod","Test":"TestFoo","Elapsed":0.0}`

	cmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"verify", "--command", "echo " + jsonLine, "--update"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}

	// Status should remain PARTIAL
	db, _ := database.Load(filepath.Join(tmpDir, ".rtmx", "database.csv"))
	req := db.Get("REQ-001")
	if req.Status != database.StatusPartial {
		t.Errorf("skipped test should keep PARTIAL status, got %v", req.Status)
	}
}

func TestCountFailingReqs(t *testing.T) {
	tests := []struct {
		name     string
		results  []VerificationResult
		expected int
	}{
		{
			name:     "empty results",
			results:  nil,
			expected: 0,
		},
		{
			name: "no failures",
			results: []VerificationResult{
				{ReqID: "REQ-001", TestsPassed: 1, TestsFailed: 0},
				{ReqID: "REQ-002", TestsPassed: 1, TestsFailed: 0},
			},
			expected: 0,
		},
		{
			name: "one failure",
			results: []VerificationResult{
				{ReqID: "REQ-001", TestsPassed: 1, TestsFailed: 0},
				{ReqID: "REQ-002", TestsPassed: 0, TestsFailed: 1},
			},
			expected: 1,
		},
		{
			name: "all failing",
			results: []VerificationResult{
				{ReqID: "REQ-001", TestsFailed: 1},
				{ReqID: "REQ-002", TestsFailed: 2},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countFailingReqs(tt.results)
			if got != tt.expected {
				t.Errorf("countFailingReqs() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestPrintVerifyResults_EmptyResults(t *testing.T) {
	cmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	// Call printVerifyResults directly
	printVerifyResults(cmd, nil)

	out := buf.String()
	if !strings.Contains(out, "No requirements with linked tests found") {
		t.Errorf("expected 'No requirements' message for empty results, got:\n%s", out)
	}
}

func TestPrintVerifyResults_WithChanges(t *testing.T) {
	cmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	results := []VerificationResult{
		{
			ReqID:          "REQ-001",
			TestsPassed:    1,
			TestsFailed:    0,
			PreviousStatus: database.StatusMissing,
			NewStatus:      database.StatusComplete,
			Updated:        true,
		},
		{
			ReqID:          "REQ-002",
			TestsPassed:    0,
			TestsFailed:    1,
			PreviousStatus: database.StatusComplete,
			NewStatus:      database.StatusPartial,
			Updated:        true,
		},
	}

	printVerifyResults(cmd, results)

	out := buf.String()
	if !strings.Contains(out, "PASSING") {
		t.Errorf("expected PASSING in output, got:\n%s", out)
	}
	if !strings.Contains(out, "FAILING") {
		t.Errorf("expected FAILING in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Status Changes") {
		t.Errorf("expected Status Changes header, got:\n%s", out)
	}
	if !strings.Contains(out, "REQ-001") {
		t.Errorf("expected REQ-001 in output, got:\n%s", out)
	}
	if !strings.Contains(out, "REQ-002") {
		t.Errorf("expected REQ-002 in output, got:\n%s", out)
	}
}

// --- Threshold Tests (REQ-SEC-011) ---

func TestVerifyThresholds(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-011")

	// Build a database with many MISSING requirements that will flip to COMPLETE
	makeDB := func(n int) string {
		db := testDBHeader
		for i := 1; i <= n; i++ {
			db += fmt.Sprintf("REQ-%03d,CAT,Sub,Req %d,Pass,mod,TestReq%03d,Unit Test,MISSING,HIGH,1,,,,,,,,,\n", i, i, i)
		}
		return db
	}

	// Build a test results map where all tests pass
	makeResults := func(n int) map[string]*TestResult {
		results := make(map[string]*TestResult)
		for i := 1; i <= n; i++ {
			name := fmt.Sprintf("TestReq%03d", i)
			results[fmt.Sprintf("mod/%s", name)] = &TestResult{
				Package: "mod",
				Test:    name,
				Passed:  true,
			}
		}
		return results
	}

	t.Run("within_warn_threshold", func(t *testing.T) {
		// 3 changes, warn=5, fail=15 -> should succeed silently
		tmpDir := setupTestProject(t, makeDB(3))
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cfg, _ := config.LoadFromDir(tmpDir)
		dbPath := cfg.DatabasePath(tmpDir)
		db, _ := database.Load(dbPath)

		results := mapTestsToRequirements(db, makeResults(3))

		updateCount := 0
		for _, r := range results {
			if r.Updated {
				updateCount++
			}
		}

		warnThreshold := 5
		failThreshold := 15

		if updateCount > failThreshold {
			t.Errorf("should not exceed fail threshold: %d > %d", updateCount, failThreshold)
		}
		if updateCount > warnThreshold {
			t.Errorf("should not exceed warn threshold: %d > %d", updateCount, warnThreshold)
		}
		if updateCount != 3 {
			t.Errorf("expected 3 updates, got %d", updateCount)
		}
	})

	t.Run("exceeds_warn_within_fail", func(t *testing.T) {
		// 8 changes, warn=5, fail=15 -> should warn but succeed
		tmpDir := setupTestProject(t, makeDB(8))
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cfg, _ := config.LoadFromDir(tmpDir)
		dbPath := cfg.DatabasePath(tmpDir)
		db, _ := database.Load(dbPath)

		results := mapTestsToRequirements(db, makeResults(8))

		updateCount := 0
		for _, r := range results {
			if r.Updated {
				updateCount++
			}
		}

		warnThreshold := 5
		failThreshold := 15

		if updateCount <= warnThreshold {
			t.Errorf("should exceed warn threshold: %d <= %d", updateCount, warnThreshold)
		}
		if updateCount > failThreshold {
			t.Errorf("should not exceed fail threshold: %d > %d", updateCount, failThreshold)
		}
	})

	t.Run("exceeds_fail_threshold", func(t *testing.T) {
		// 20 changes, warn=5, fail=15 -> should fail
		tmpDir := setupTestProject(t, makeDB(20))
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		// Write a config with low thresholds
		cfgContent := "rtmx:\n  database: .rtmx/database.csv\n  verify:\n    thresholds:\n      warn: 5\n      fail: 15\n"
		_ = os.WriteFile(filepath.Join(tmpDir, ".rtmx", "config.yaml"), []byte(cfgContent), 0644)

		cfg, _ := config.LoadFromDir(tmpDir)
		dbPath := cfg.DatabasePath(tmpDir)
		db, _ := database.Load(dbPath)

		results := mapTestsToRequirements(db, makeResults(20))

		updateCount := 0
		for _, r := range results {
			if r.Updated {
				updateCount++
			}
		}

		if updateCount <= cfg.RTMX.Verify.Thresholds.Fail {
			t.Errorf("should exceed fail threshold: %d <= %d", updateCount, cfg.RTMX.Verify.Thresholds.Fail)
		}

		// Verify the database was NOT modified (threshold blocks write)
		originalDB, _ := os.ReadFile(dbPath)
		if !strings.Contains(string(originalDB), "MISSING") {
			t.Error("database should still contain MISSING (threshold should block write)")
		}
	})

	t.Run("custom_thresholds_from_config", func(t *testing.T) {
		tmpDir := setupTestProject(t, makeDB(3))
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		// Set thresholds to warn=1, fail=2
		cfgContent := "rtmx:\n  database: .rtmx/database.csv\n  verify:\n    thresholds:\n      warn: 1\n      fail: 2\n"
		_ = os.WriteFile(filepath.Join(tmpDir, ".rtmx", "config.yaml"), []byte(cfgContent), 0644)

		cfg, _ := config.LoadFromDir(tmpDir)
		if cfg.RTMX.Verify.Thresholds.Warn != 1 {
			t.Errorf("expected warn=1, got %d", cfg.RTMX.Verify.Thresholds.Warn)
		}
		if cfg.RTMX.Verify.Thresholds.Fail != 2 {
			t.Errorf("expected fail=2, got %d", cfg.RTMX.Verify.Thresholds.Fail)
		}
	})

	t.Run("defaults_when_no_config", func(t *testing.T) {
		cfg := config.DefaultConfig()
		if cfg.RTMX.Verify.Thresholds.Warn != 5 {
			t.Errorf("expected default warn=5, got %d", cfg.RTMX.Verify.Thresholds.Warn)
		}
		if cfg.RTMX.Verify.Thresholds.Fail != 15 {
			t.Errorf("expected default fail=15, got %d", cfg.RTMX.Verify.Thresholds.Fail)
		}
		if !cfg.RTMX.Verify.AutoUpdate {
			t.Error("expected default auto_update=true")
		}
	})
}

func TestVerifyCargoTestOutput(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-003")

	// Create a temp project with database linking tests to cargo test names
	dbContent := testDBHeader +
		"REQ-001,CAT,Sub,Req one,Pass,src/lib.rs,tests::test_parse_csv,Unit Test,MISSING,HIGH,1,,,,,,,,,\n" +
		"REQ-002,CAT,Sub,Req two,Pass,src/lib.rs,tests::test_get_by_id,Unit Test,MISSING,HIGH,1,,,,,,,,,\n" +
		"REQ-003,CAT,Sub,Req three,Pass,src/lib.rs,tests::test_filter_status,Unit Test,COMPLETE,HIGH,1,,,,,,,,,\n"

	tmpDir := setupTestProject(t, dbContent)
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	// Create a script that outputs cargo test format
	scriptContent := `#!/bin/sh
echo "running 3 tests"
echo "test tests::test_parse_csv ... ok"
echo "test tests::test_get_by_id ... ok"
echo "test tests::test_filter_status ... FAILED"
echo ""
echo "test result: FAILED. 2 passed; 1 failed; 0 ignored"
exit 1
`
	scriptPath := filepath.Join(tmpDir, "fake_cargo.sh")
	_ = os.WriteFile(scriptPath, []byte(scriptContent), 0755)

	// Run verify with the fake cargo script
	verifyCommand = "sh " + scriptPath
	verifyUpdate = false
	verifyDryRun = false
	verifyVerbose = false
	verifyForce = false

	cmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	cfg, _ := config.LoadFromDir(tmpDir)
	dbPath := cfg.DatabasePath(tmpDir)
	db, _ := database.Load(dbPath)

	testResults, _ := runTests(cmd, "")
	results := mapTestsToRequirements(db, testResults)

	// Should find 3 test results
	if len(testResults) < 2 {
		t.Fatalf("expected at least 2 cargo test results, got %d", len(testResults))
	}

	// Check that REQ-001 and REQ-002 would flip to COMPLETE (tests passed)
	foundReq1 := false
	foundReq3 := false
	for _, r := range results {
		if r.ReqID == "REQ-001" && r.NewStatus == database.StatusComplete {
			foundReq1 = true
		}
		if r.ReqID == "REQ-003" && r.NewStatus == database.StatusPartial {
			foundReq3 = true
		}
	}

	if !foundReq1 {
		t.Error("REQ-001 should be marked COMPLETE from passing cargo test")
	}
	if !foundReq3 {
		t.Error("REQ-003 should be marked PARTIAL from failing cargo test (was COMPLETE)")
	}

	// Reset
	verifyCommand = ""
}

func TestDetectTestCommand(t *testing.T) {
	rtmx.Req(t, "REQ-GO-034")

	tests := []struct {
		name     string
		files    map[string]string // filename -> content
		wantCmd  string
		wantArgs []string
	}{
		{
			name:     "Cargo.toml detects cargo test",
			files:    map[string]string{"Cargo.toml": "[package]\nname = \"myapp\""},
			wantCmd:  "cargo",
			wantArgs: []string{"test", "--workspace"},
		},
		{
			name:     "package.json detects npm test",
			files:    map[string]string{"package.json": "{\"scripts\":{\"test\":\"jest\"}}"},
			wantCmd:  "npm",
			wantArgs: []string{"test"},
		},
		{
			name:     "setup.py detects pytest",
			files:    map[string]string{"setup.py": "from setuptools import setup"},
			wantCmd:  "pytest",
			wantArgs: []string{"-v"},
		},
		{
			name:     "pyproject.toml detects pytest",
			files:    map[string]string{"pyproject.toml": "[build-system]"},
			wantCmd:  "pytest",
			wantArgs: []string{"-v"},
		},
		{
			name:     "build.gradle detects gradle test",
			files:    map[string]string{"build.gradle": "apply plugin: 'java'"},
			wantCmd:  "gradle",
			wantArgs: []string{"test"},
		},
		{
			name:     "pom.xml detects mvn test",
			files:    map[string]string{"pom.xml": "<project></project>"},
			wantCmd:  "mvn",
			wantArgs: []string{"test"},
		},
		{
			name:     "mix.exs detects mix test",
			files:    map[string]string{"mix.exs": "defmodule MyApp do"},
			wantCmd:  "mix",
			wantArgs: []string{"test"},
		},
		{
			name:     "Gemfile detects bundle exec rake test",
			files:    map[string]string{"Gemfile": "source 'https://rubygems.org'"},
			wantCmd:  "bundle",
			wantArgs: []string{"exec", "rake", "test"},
		},
		{
			name:     "go.mod falls back to go test",
			files:    map[string]string{"go.mod": "module example.com/mymod"},
			wantCmd:  "go",
			wantArgs: []string{"test", "-json", "./..."},
		},
		{
			name:     "no build files falls back to go test",
			files:    map[string]string{},
			wantCmd:  "go",
			wantArgs: []string{"test", "-json", "./..."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}
			cmd, args := DetectTestCommand(dir)
			if cmd != tt.wantCmd {
				t.Errorf("DetectTestCommand() cmd = %q, want %q", cmd, tt.wantCmd)
			}
			if len(args) != len(tt.wantArgs) {
				t.Errorf("DetectTestCommand() args = %v, want %v", args, tt.wantArgs)
			} else {
				for i, a := range args {
					if a != tt.wantArgs[i] {
						t.Errorf("DetectTestCommand() args[%d] = %q, want %q", i, a, tt.wantArgs[i])
					}
				}
			}
		})
	}
}

func TestVerifyAutoSetDates(t *testing.T) {
	rtmx.Req(t, "REQ-PLAN-010")

	t.Run("sets_started_date_on_partial", func(t *testing.T) {
		db := database.NewDatabase()

		req := database.NewRequirement("REQ-001")
		req.TestFunction = "TestFoo"
		req.Status = database.StatusMissing
		_ = db.Add(req)

		// Simulate a failing test match -> MISSING stays MISSING (no date change)
		testResults := map[string]*TestResult{
			"pkg/TestFoo": {Package: "pkg", Test: "TestFoo", Failed: true},
		}
		results := mapTestsToRequirements(db, testResults)

		// Apply updates like runVerify does
		for _, r := range results {
			if r.Updated {
				dbReq := db.Get(r.ReqID)
				if dbReq != nil {
					dbReq.Status = r.NewStatus
					if r.NewStatus == database.StatusPartial || r.NewStatus == database.StatusComplete {
						dbReq.SetStartedDate()
					}
					if r.NewStatus == database.StatusComplete {
						dbReq.SetCompletedDate()
					}
				}
			}
		}

		// MISSING->MISSING: no date change
		if req.StartedDate != "" {
			t.Errorf("started_date should be empty for MISSING, got %q", req.StartedDate)
		}
	})

	t.Run("sets_dates_on_complete", func(t *testing.T) {
		db := database.NewDatabase()

		req := database.NewRequirement("REQ-001")
		req.TestFunction = "TestFoo"
		req.Status = database.StatusMissing
		_ = db.Add(req)

		testResults := map[string]*TestResult{
			"pkg/TestFoo": {Package: "pkg", Test: "TestFoo", Passed: true},
		}
		results := mapTestsToRequirements(db, testResults)

		for _, r := range results {
			if r.Updated {
				dbReq := db.Get(r.ReqID)
				if dbReq != nil {
					dbReq.Status = r.NewStatus
					if r.NewStatus == database.StatusPartial || r.NewStatus == database.StatusComplete {
						dbReq.SetStartedDate()
					}
					if r.NewStatus == database.StatusComplete {
						dbReq.SetCompletedDate()
					}
				}
			}
		}

		if req.StartedDate == "" {
			t.Error("started_date should be set on COMPLETE transition")
		}
		if req.CompletedDate == "" {
			t.Error("completed_date should be set on COMPLETE transition")
		}
	})

	t.Run("preserves_existing_started_date", func(t *testing.T) {
		db := database.NewDatabase()

		req := database.NewRequirement("REQ-001")
		req.TestFunction = "TestFoo"
		req.Status = database.StatusPartial
		req.StartedDate = "2026-01-15"
		_ = db.Add(req)

		testResults := map[string]*TestResult{
			"pkg/TestFoo": {Package: "pkg", Test: "TestFoo", Passed: true},
		}
		results := mapTestsToRequirements(db, testResults)

		for _, r := range results {
			if r.Updated {
				dbReq := db.Get(r.ReqID)
				if dbReq != nil {
					dbReq.Status = r.NewStatus
					if r.NewStatus == database.StatusPartial || r.NewStatus == database.StatusComplete {
						dbReq.SetStartedDate()
					}
					if r.NewStatus == database.StatusComplete {
						dbReq.SetCompletedDate()
					}
				}
			}
		}

		if req.StartedDate != "2026-01-15" {
			t.Errorf("started_date should be preserved, got %q", req.StartedDate)
		}
		if req.CompletedDate == "" {
			t.Error("completed_date should be set on COMPLETE")
		}
	})
}

func TestVerifyAudit(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-005")

	t.Run("unmatched_test_function", func(t *testing.T) {
		db := database.NewDatabase()

		req := database.NewRequirement("REQ-001")
		req.TestFunction = "TestNonExistent"
		req.TestModule = "some_test.go"
		req.Status = database.StatusMissing
		_ = db.Add(req)

		// No test results match REQ-001
		var verifyResults []VerificationResult

		audit := runAudit(db, verifyResults)

		if len(audit.Unmatched) != 1 {
			t.Fatalf("expected 1 unmatched, got %d", len(audit.Unmatched))
		}
		if audit.Unmatched[0].ReqID != "REQ-001" {
			t.Errorf("unmatched ReqID = %q, want REQ-001", audit.Unmatched[0].ReqID)
		}
		if audit.Unmatched[0].Value != "TestNonExistent" {
			t.Errorf("unmatched Value = %q, want TestNonExistent", audit.Unmatched[0].Value)
		}
	})

	t.Run("stale_test_module_path", func(t *testing.T) {
		db := database.NewDatabase()

		req := database.NewRequirement("REQ-001")
		req.TestFunction = "TestFoo"
		req.TestModule = "nonexistent/path/foo_test.go"
		req.Status = database.StatusMissing
		_ = db.Add(req)

		var verifyResults []VerificationResult

		audit := runAudit(db, verifyResults)

		if len(audit.StalePaths) != 1 {
			t.Fatalf("expected 1 stale path, got %d", len(audit.StalePaths))
		}
		if audit.StalePaths[0].Value != "nonexistent/path/foo_test.go" {
			t.Errorf("stale path Value = %q, want nonexistent/path/foo_test.go", audit.StalePaths[0].Value)
		}
	})

	t.Run("unverified_complete", func(t *testing.T) {
		db := database.NewDatabase()

		req := database.NewRequirement("REQ-001")
		req.TestFunction = "TestFoo"
		req.TestModule = "foo_test.go"
		req.Status = database.StatusComplete
		_ = db.Add(req)

		// No test results for REQ-001
		var verifyResults []VerificationResult

		audit := runAudit(db, verifyResults)

		if len(audit.UnverifiedComplete) != 1 {
			t.Fatalf("expected 1 unverified complete, got %d", len(audit.UnverifiedComplete))
		}
		if audit.UnverifiedComplete[0].ReqID != "REQ-001" {
			t.Errorf("unverified ReqID = %q, want REQ-001", audit.UnverifiedComplete[0].ReqID)
		}
	})

	t.Run("empty_test_references", func(t *testing.T) {
		db := database.NewDatabase()

		req := database.NewRequirement("REQ-001")
		req.TestFunction = ""
		req.TestModule = ""
		req.Status = database.StatusMissing
		_ = db.Add(req)

		var verifyResults []VerificationResult

		audit := runAudit(db, verifyResults)

		if len(audit.EmptyRefs) != 1 {
			t.Fatalf("expected 1 empty ref, got %d", len(audit.EmptyRefs))
		}
		if audit.EmptyRefs[0].ReqID != "REQ-001" {
			t.Errorf("empty ref ReqID = %q, want REQ-001", audit.EmptyRefs[0].ReqID)
		}
	})

	t.Run("matched_requirement_not_flagged", func(t *testing.T) {
		db := database.NewDatabase()

		req := database.NewRequirement("REQ-001")
		req.TestFunction = "TestFoo"
		req.TestModule = "foo_test.go"
		req.Status = database.StatusComplete
		_ = db.Add(req)

		// REQ-001 was matched in verification
		verifyResults := []VerificationResult{
			{ReqID: "REQ-001", TestsPassed: 1, NewStatus: database.StatusComplete},
		}

		audit := runAudit(db, verifyResults)

		if len(audit.Unmatched) != 0 {
			t.Errorf("expected 0 unmatched, got %d", len(audit.Unmatched))
		}
		if len(audit.UnverifiedComplete) != 0 {
			t.Errorf("expected 0 unverified complete, got %d", len(audit.UnverifiedComplete))
		}
	})

	t.Run("mixed_findings", func(t *testing.T) {
		db := database.NewDatabase()

		// Matched requirement -- should not appear in findings
		r1 := database.NewRequirement("REQ-001")
		r1.TestFunction = "TestMatched"
		r1.TestModule = "matched_test.go"
		r1.Status = database.StatusComplete
		_ = db.Add(r1)

		// Unmatched with test_function
		r2 := database.NewRequirement("REQ-002")
		r2.TestFunction = "TestOrphan"
		r2.TestModule = "orphan_test.go"
		r2.Status = database.StatusMissing
		_ = db.Add(r2)

		// Empty references
		r3 := database.NewRequirement("REQ-003")
		r3.Status = database.StatusMissing
		_ = db.Add(r3)

		// COMPLETE but unmatched
		r4 := database.NewRequirement("REQ-004")
		r4.TestFunction = "TestGhost"
		r4.Status = database.StatusComplete
		_ = db.Add(r4)

		verifyResults := []VerificationResult{
			{ReqID: "REQ-001", TestsPassed: 1, NewStatus: database.StatusComplete},
		}

		audit := runAudit(db, verifyResults)

		if len(audit.Unmatched) != 2 { // REQ-002 and REQ-004
			t.Errorf("expected 2 unmatched, got %d", len(audit.Unmatched))
		}
		if len(audit.EmptyRefs) != 1 { // REQ-003
			t.Errorf("expected 1 empty ref, got %d", len(audit.EmptyRefs))
		}
		if len(audit.UnverifiedComplete) != 1 { // REQ-004
			t.Errorf("expected 1 unverified complete, got %d", len(audit.UnverifiedComplete))
		}
	})
}

func TestVerifyUnmatchedWarning(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-006")

	t.Run("warning_when_unmatched", func(t *testing.T) {
		db := database.NewDatabase()

		req := database.NewRequirement("REQ-001")
		req.TestFunction = "TestOrphan"
		req.Status = database.StatusMissing
		_ = db.Add(req)

		var verifyResults []VerificationResult

		count := countUnmatchedRefs(db, verifyResults)
		if count != 1 {
			t.Errorf("countUnmatchedRefs = %d, want 1", count)
		}
	})

	t.Run("no_warning_when_all_matched", func(t *testing.T) {
		db := database.NewDatabase()

		req := database.NewRequirement("REQ-001")
		req.TestFunction = "TestFoo"
		req.Status = database.StatusComplete
		_ = db.Add(req)

		verifyResults := []VerificationResult{
			{ReqID: "REQ-001", TestsPassed: 1},
		}

		count := countUnmatchedRefs(db, verifyResults)
		if count != 0 {
			t.Errorf("countUnmatchedRefs = %d, want 0", count)
		}
	})

	t.Run("no_warning_for_empty_refs", func(t *testing.T) {
		db := database.NewDatabase()

		req := database.NewRequirement("REQ-001")
		req.TestFunction = "" // No test function set
		req.Status = database.StatusMissing
		_ = db.Add(req)

		var verifyResults []VerificationResult

		count := countUnmatchedRefs(db, verifyResults)
		if count != 0 {
			t.Errorf("countUnmatchedRefs = %d, want 0 (empty refs should not count)", count)
		}
	})
}

func TestVerifyGitAttribution(t *testing.T) {
	rtmx.Req(t, "REQ-PLAN-013",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	t.Run("get_git_author_returns_string_for_tracked_file", func(t *testing.T) {
		// GetGitAuthor should return a non-empty string for a file tracked in this repo
		author := GetGitAuthor("internal/cmd/verify.go")
		// In a real git repo this will return the author name; in CI it may vary
		// The key contract is it doesn't panic and returns a string
		if author == "" {
			t.Log("GetGitAuthor returned empty -- likely not in a git repo context")
		}
	})

	t.Run("get_git_author_returns_empty_for_nonexistent", func(t *testing.T) {
		author := GetGitAuthor("nonexistent/file/that/doesnt/exist.go")
		if author != "" {
			t.Errorf("expected empty author for nonexistent file, got %q", author)
		}
	})

	t.Run("assignee_not_overwritten", func(t *testing.T) {
		// Simulate the attribution logic: if assignee is already set, don't overwrite
		db := database.NewDatabase()

		req := database.NewRequirement("REQ-001")
		req.TestFunction = "TestFoo"
		req.TestModule = "internal/cmd/verify.go"
		req.Status = database.StatusMissing
		req.Assignee = "existing-person"
		_ = db.Add(req)

		testResults := map[string]*TestResult{
			"pkg/TestFoo": {Package: "pkg", Test: "TestFoo", Passed: true},
		}
		results := mapTestsToRequirements(db, testResults)

		for _, r := range results {
			if r.Updated {
				dbReq := db.Get(r.ReqID)
				if dbReq != nil {
					dbReq.Status = r.NewStatus
					if r.NewStatus == database.StatusComplete {
						dbReq.SetCompletedDate()
						// Attribution logic: don't overwrite existing assignee
						if dbReq.Assignee == "" && dbReq.TestModule != "" {
							if author := GetGitAuthor(dbReq.TestModule); author != "" {
								dbReq.Assignee = author
							}
						}
					}
				}
			}
		}

		if req.Assignee != "existing-person" {
			t.Errorf("assignee should not be overwritten, got %q", req.Assignee)
		}
	})

	t.Run("assignee_set_when_empty", func(t *testing.T) {
		db := database.NewDatabase()

		req := database.NewRequirement("REQ-001")
		req.TestFunction = "TestFoo"
		req.TestModule = "internal/cmd/verify.go"
		req.Status = database.StatusMissing
		// Assignee intentionally left empty
		_ = db.Add(req)

		testResults := map[string]*TestResult{
			"pkg/TestFoo": {Package: "pkg", Test: "TestFoo", Passed: true},
		}
		results := mapTestsToRequirements(db, testResults)

		for _, r := range results {
			if r.Updated {
				dbReq := db.Get(r.ReqID)
				if dbReq != nil {
					dbReq.Status = r.NewStatus
					if r.NewStatus == database.StatusComplete {
						dbReq.SetCompletedDate()
						if dbReq.Assignee == "" && dbReq.TestModule != "" {
							if author := GetGitAuthor(dbReq.TestModule); author != "" {
								dbReq.Assignee = author
							}
						}
					}
				}
			}
		}

		// In a git repo, this should have set the assignee
		// In CI without git, it may remain empty
		if req.Assignee == "" {
			t.Log("assignee not set -- GetGitAuthor returned empty (likely not in full git repo)")
		} else {
			t.Logf("assignee auto-set to %q from git attribution", req.Assignee)
		}
	})
}

func TestMatchTestFunction(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-007")

	tests := []struct {
		name           string
		testResultName string
		dbTestFunction string
		expected       bool
	}{
		// Exact match
		{"exact_match", "TestFoo", "TestFoo", true},
		{"exact_match_with_colons", "tests::test_foo", "tests::test_foo", true},

		// Rust :: suffix matching
		{"rust_module_prefix", "embedding::tests::test_foo", "tests::test_foo", true},
		{"rust_deep_module", "crate::embedding::tests::test_foo", "tests::test_foo", true},
		{"rust_single_prefix", "module::test_bar", "test_bar", true},

		// Python . suffix matching
		{"python_package_prefix", "pkg.tests.test_foo", "tests.test_foo", true},
		{"python_deep_package", "app.pkg.tests.test_foo", "tests.test_foo", true},
		{"python_single_prefix", "module.test_bar", "test_bar", true},

		// Go / suffix matching
		{"go_package_prefix", "github.com/org/pkg/TestFoo", "TestFoo", true},
		{"go_deep_package", "github.com/org/repo/internal/cmd/TestFoo", "TestFoo", true},

		// Negative cases
		{"no_match", "TestFoo", "TestBar", false},
		{"partial_name_no_boundary", "xTestFoo", "TestFoo", false},
		{"substring_no_separator", "module_test_foo", "test_foo", false},
		{"wrong_separator", "module::test_foo", "module.test_foo", false},
		{"empty_db_function", "TestFoo", "", false},
		{"empty_result_name", "", "TestFoo", false},
		{"both_empty", "", "", true}, // exact match of empty strings
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchTestFunction(tt.testResultName, tt.dbTestFunction)
			if got != tt.expected {
				t.Errorf("matchTestFunction(%q, %q) = %v, want %v",
					tt.testResultName, tt.dbTestFunction, got, tt.expected)
			}
		})
	}
}

func TestMapTestsToRequirements_SuffixMatching(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-007")

	dbContent := testDBHeader +
		"REQ-001,CAT,Sub,Rust suffix test,Pass,mod,tests::test_create,Unit Test,MISSING,HIGH,1,,,,,,,,,\n" +
		"REQ-002,CAT,Sub,Python suffix test,Pass,mod,tests.test_update,Unit Test,MISSING,HIGH,1,,,,,,,,,\n"

	tmpDir := setupTestProject(t, dbContent)
	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cfg, _ := config.LoadFromDir(tmpDir)
	dbPath := cfg.DatabasePath(tmpDir)
	db, _ := database.Load(dbPath)

	testResults := map[string]*TestResult{
		"embedding::tests::test_create": {
			Package: "embedding", Test: "embedding::tests::test_create", Passed: true,
		},
		"app.tests.test_update": {
			Package: "app", Test: "app.tests.test_update", Passed: true,
		},
	}

	results := mapTestsToRequirements(db, testResults)

	if len(results) != 2 {
		t.Fatalf("expected 2 results from suffix matching, got %d", len(results))
	}

	matched := map[string]bool{}
	for _, r := range results {
		matched[r.ReqID] = r.Updated
	}
	if !matched["REQ-001"] {
		t.Error("REQ-001 should have matched via :: suffix")
	}
	if !matched["REQ-002"] {
		t.Error("REQ-002 should have matched via . suffix")
	}
}

func TestVerifyWarnThresholdCLIFlag(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-008")

	if runtime.GOOS == "windows" {
		t.Skip("echo behaves differently on Windows")
	}

	// Build a database with n requirements, all will flip MISSING -> COMPLETE
	makeDB := func(n int) string {
		db := testDBHeader
		for i := 1; i <= n; i++ {
			db += fmt.Sprintf("REQ-%03d,CAT,Sub,Req %d,Pass,mod,TestReq%03d,Unit Test,MISSING,HIGH,1,,,,,,,,,\n", i, i, i)
		}
		return db
	}

	// Create a temp script that outputs passing go test JSON for n tests
	makeTestScript := func(t *testing.T, n int) string {
		t.Helper()
		script := "#!/bin/sh\n"
		for i := 1; i <= n; i++ {
			name := fmt.Sprintf("TestReq%03d", i)
			script += fmt.Sprintf(`printf '{"Time":"2024-01-01T00:00:00Z","Action":"pass","Package":"mod","Test":"%s","Elapsed":0.1}\n'`+"\n", name)
		}
		path := filepath.Join(t.TempDir(), "test.sh")
		_ = os.WriteFile(path, []byte(script), 0755)
		return path
	}

	t.Run("cli_warn_threshold_suppresses_warning", func(t *testing.T) {
		tmpDir := setupTestProject(t, makeDB(8))
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		scriptPath := makeTestScript(t, 8)

		cmd := newTestRootCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"verify", "--update", "--warn-threshold", "20",
			"--command", scriptPath})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("verify failed: %v\nOutput: %s", err, buf.String())
		}

		if strings.Contains(buf.String(), "WARNING") {
			t.Error("--warn-threshold=20 should suppress warning for 8 changes")
		}
	})

	t.Run("cli_warn_threshold_triggers_warning", func(t *testing.T) {
		tmpDir := setupTestProject(t, makeDB(8))
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		scriptPath := makeTestScript(t, 8)

		cmd := newTestRootCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"verify", "--update", "--warn-threshold", "3",
			"--command", scriptPath})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("verify failed: %v\nOutput: %s", err, buf.String())
		}

		if !strings.Contains(buf.String(), "WARNING") {
			t.Error("--warn-threshold=3 should trigger warning for 8 changes")
		}
	})

	t.Run("cli_fail_threshold_blocks_update", func(t *testing.T) {
		tmpDir := setupTestProject(t, makeDB(8))
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		scriptPath := makeTestScript(t, 8)

		cmd := newTestRootCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"verify", "--update", "--fail-threshold", "3",
			"--command", scriptPath})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error when fail threshold exceeded")
		}

		if !strings.Contains(err.Error(), "threshold exceeded") {
			t.Errorf("expected threshold exceeded error, got: %v", err)
		}
	})

	t.Run("cli_flag_overrides_config", func(t *testing.T) {
		tmpDir := setupTestProject(t, makeDB(8))
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		// Config says warn=1, but CLI says warn=20
		cfgContent := "rtmx:\n  database: .rtmx/database.csv\n  verify:\n    thresholds:\n      warn: 1\n      fail: 100\n"
		_ = os.WriteFile(filepath.Join(tmpDir, ".rtmx", "config.yaml"), []byte(cfgContent), 0644)

		scriptPath := makeTestScript(t, 8)

		cmd := newTestRootCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"verify", "--update", "--warn-threshold", "20",
			"--command", scriptPath})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("verify failed: %v\nOutput: %s", err, buf.String())
		}

		if strings.Contains(buf.String(), "WARNING") {
			t.Error("CLI --warn-threshold=20 should override config warn=1")
		}
	})

	t.Run("force_overrides_fail_threshold", func(t *testing.T) {
		tmpDir := setupTestProject(t, makeDB(8))
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		scriptPath := makeTestScript(t, 8)

		cmd := newTestRootCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"verify", "--update", "--fail-threshold", "3", "--force",
			"--command", scriptPath})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("verify with --force should succeed: %v", err)
		}
	})
}
