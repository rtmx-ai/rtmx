package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx-go/pkg/rtmx"
	"github.com/spf13/cobra"
)

func TestHealthRealCommand(t *testing.T) {
	rtmx.Req(t, "REQ-GO-012")

	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createHealthTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"health"})

	err := rootCmd.Execute()
	// Health may return ExitError for warnings - that's OK
	if err != nil {
		var exitErr *ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("health command failed unexpectedly: %v", err)
		}
	}

	output := buf.String()
	expectedPhrases := []string{
		"Health Check",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Expected output to contain %q, got:\n%s", phrase, output)
		}
	}
}

func TestHealthJSONOutput(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createHealthTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"health", "--json"})

	err := rootCmd.Execute()
	// Health may return ExitError for warnings - that's OK
	if err != nil {
		var exitErr *ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("health --json failed unexpectedly: %v", err)
		}
	}

	output := buf.String()

	// Verify it's valid JSON
	var result HealthResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Errorf("Expected valid JSON output, got parse error: %v\nOutput: %s", err, output)
	}

	// Verify required fields
	if result.Status == "" {
		t.Error("Expected status field in JSON output")
	}
}

// TestHealthCheckByCheckFormat verifies that health shows individual check results
// REQ-GO-051: Go CLI health shall show individual check results like Python
func TestHealthCheckByCheckFormat(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createHealthTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"health"})

	err := rootCmd.Execute()
	// Health may return ExitError for warnings - that's OK
	if err != nil {
		var exitErr *ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("health command failed unexpectedly: %v", err)
		}
	}

	output := buf.String()

	// Verify Python-style check format: [PASS]/[WARN]/[FAIL] check_name: message
	expectedElements := []string{
		"[PASS]",                  // Pass status label
		"rtm_loads:",              // Check name with colon
		"Status:",                 // Status summary line
		"Summary:",                // Summary counts line
		"passed",                  // Summary contains "passed"
		"warnings",                // Summary contains "warnings"
		"failed",                  // Summary contains "failed"
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected health output to contain %q, got:\n%s", element, output)
		}
	}
}

// createHealthTestCmd creates a root command with real health command for testing
func createHealthTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	var jsonOutput bool

	healthCmd := &cobra.Command{
		Use:   "health",
		Short: "Run health check",
		RunE: func(cmd *cobra.Command, args []string) error {
			healthJSON = jsonOutput
			return runHealth(cmd, args)
		},
	}
	healthCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	root.AddCommand(healthCmd)

	return root
}

// setupHealthTestProject creates a temp dir with config and database CSV for health tests.
func setupHealthTestProject(t *testing.T, dbContent string) string {
	t.Helper()
	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"), []byte("database:\n  path: .rtmx/database.csv\n"), 0644)
	_ = os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644)
	return tmpDir
}

const healthDBHeader = "req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id\n"

// TestHealthStatusConsistency verifies the status_consistency check detects
// COMPLETE requirements depending on MISSING/PARTIAL/NOT_STARTED dependencies.
func TestHealthStatusConsistency(t *testing.T) {
	rtmx.Req(t, "REQ-GO-074")

	tests := []struct {
		name           string
		dbContent      string
		wantWarn       bool
		wantIssueCount int
		wantCheckName  string
	}{
		{
			name: "COMPLETE depends on MISSING warns",
			dbContent: healthDBHeader +
				"REQ-001,CAT,Sub,Base requirement,Pass,,,Unit Test,MISSING,HIGH,1,,,,,,,,,\n" +
				"REQ-002,CAT,Sub,Dependent requirement,Pass,,,Unit Test,COMPLETE,HIGH,1,,,REQ-001,,,,,,,,\n",
			wantWarn:       true,
			wantIssueCount: 1,
			wantCheckName:  "status_consistency",
		},
		{
			name: "COMPLETE depends on PARTIAL warns",
			dbContent: healthDBHeader +
				"REQ-001,CAT,Sub,Base requirement,Pass,,,Unit Test,PARTIAL,HIGH,1,,,,,,,,,\n" +
				"REQ-002,CAT,Sub,Dependent requirement,Pass,,,Unit Test,COMPLETE,HIGH,1,,,REQ-001,,,,,,,,\n",
			wantWarn:       true,
			wantIssueCount: 1,
			wantCheckName:  "status_consistency",
		},
		{
			name: "COMPLETE depends on NOT_STARTED warns",
			dbContent: healthDBHeader +
				"REQ-001,CAT,Sub,Base requirement,Pass,,,Unit Test,NOT_STARTED,HIGH,1,,,,,,,,,\n" +
				"REQ-002,CAT,Sub,Dependent requirement,Pass,,,Unit Test,COMPLETE,HIGH,1,,,REQ-001,,,,,,,,\n",
			wantWarn:       true,
			wantIssueCount: 1,
			wantCheckName:  "status_consistency",
		},
		{
			name: "COMPLETE depends on COMPLETE passes",
			dbContent: healthDBHeader +
				"REQ-001,CAT,Sub,Base requirement,Pass,,,Unit Test,COMPLETE,HIGH,1,,,,,,,,,\n" +
				"REQ-002,CAT,Sub,Dependent requirement,Pass,,,Unit Test,COMPLETE,HIGH,1,,,REQ-001,,,,,,,,\n",
			wantWarn:       false,
			wantIssueCount: 0,
			wantCheckName:  "status_consistency",
		},
		{
			name: "MISSING depends on MISSING no warning",
			dbContent: healthDBHeader +
				"REQ-001,CAT,Sub,Base requirement,Pass,,,Unit Test,MISSING,HIGH,1,,,,,,,,,\n" +
				"REQ-002,CAT,Sub,Dependent requirement,Pass,,,Unit Test,MISSING,HIGH,1,,,REQ-001,,,,,,,,\n",
			wantWarn:       false,
			wantIssueCount: 0,
			wantCheckName:  "status_consistency",
		},
		{
			name: "PARTIAL depends on MISSING no warning",
			dbContent: healthDBHeader +
				"REQ-001,CAT,Sub,Base requirement,Pass,,,Unit Test,MISSING,HIGH,1,,,,,,,,,\n" +
				"REQ-002,CAT,Sub,Dependent requirement,Pass,,,Unit Test,PARTIAL,HIGH,1,,,REQ-001,,,,,,,,\n",
			wantWarn:       false,
			wantIssueCount: 0,
			wantCheckName:  "status_consistency",
		},
		{
			name: "multiple COMPLETE depend on MISSING deps",
			dbContent: healthDBHeader +
				"REQ-001,CAT,Sub,Base requirement,Pass,,,Unit Test,MISSING,HIGH,1,,,,,,,,,\n" +
				"REQ-002,CAT,Sub,Another base,Pass,,,Unit Test,PARTIAL,HIGH,1,,,,,,,,,\n" +
				"REQ-003,CAT,Sub,Dep on missing,Pass,,,Unit Test,COMPLETE,HIGH,1,,,REQ-001,,,,,,,,\n" +
				"REQ-004,CAT,Sub,Dep on partial,Pass,,,Unit Test,COMPLETE,HIGH,1,,,REQ-002,,,,,,,,\n",
			wantWarn:       true,
			wantIssueCount: 2,
			wantCheckName:  "status_consistency",
		},
		{
			name:           "empty database no warning",
			dbContent:      healthDBHeader,
			wantWarn:       false,
			wantIssueCount: 0,
			wantCheckName:  "status_consistency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := setupHealthTestProject(t, tt.dbContent)

			origDir, _ := os.Getwd()
			_ = os.Chdir(tmpDir)
			defer func() { _ = os.Chdir(origDir) }()

			cmd := createHealthTestCmd()
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetArgs([]string{"health", "--json"})

			err := cmd.Execute()
			if err != nil {
				var exitErr *ExitError
				if !errors.As(err, &exitErr) {
					t.Fatalf("health --json failed unexpectedly: %v", err)
				}
			}

			out := buf.String()
			var result HealthResult
			if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
				t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, out)
			}

			// Find the status_consistency check
			var found bool
			for _, check := range result.Checks {
				if check.Name == tt.wantCheckName {
					found = true
					if tt.wantWarn && check.Status != CheckWarn {
						t.Errorf("expected status_consistency check status=WARN, got %s", check.Status)
					}
					if !tt.wantWarn && check.Status != CheckPass {
						t.Errorf("expected status_consistency check status=PASS, got %s", check.Status)
					}
					if tt.wantWarn {
						expectedMsg := fmt.Sprintf("%d status consistency issue(s) found", tt.wantIssueCount)
						if check.Message != expectedMsg {
							t.Errorf("expected message %q, got %q", expectedMsg, check.Message)
						}
					}
					break
				}
			}
			if !found {
				t.Errorf("status_consistency check not found in health checks: %v", result.Checks)
			}

			// Verify the stats field
			if result.Stats.StatusConsistency != tt.wantIssueCount {
				t.Errorf("expected stats.status_consistency=%d, got %d",
					tt.wantIssueCount, result.Stats.StatusConsistency)
			}
		})
	}
}

// TestHealthStatusConsistencyTextOutput verifies that text output includes the check.
func TestHealthStatusConsistencyTextOutput(t *testing.T) {
	rtmx.Req(t, "REQ-GO-074")

	// COMPLETE depends on MISSING -- should warn
	dbContent := healthDBHeader +
		"REQ-001,CAT,Sub,Base requirement,Pass,,,Unit Test,MISSING,HIGH,1,,,,,,,,,\n" +
		"REQ-002,CAT,Sub,Dependent,Pass,,,Unit Test,COMPLETE,HIGH,1,,,REQ-001,,,,,,,,\n"

	tmpDir := setupHealthTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := createHealthTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"health"})

	err := cmd.Execute()
	if err != nil {
		var exitErr *ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("health failed unexpectedly: %v", err)
		}
	}

	out := buf.String()

	if !strings.Contains(out, "status_consistency") {
		t.Errorf("expected text output to contain 'status_consistency', got:\n%s", out)
	}
	if !strings.Contains(out, "[WARN]") {
		t.Errorf("expected text output to contain '[WARN]', got:\n%s", out)
	}
}

// TestHealthStatusConsistencyJSONDetails verifies that JSON details are populated.
func TestHealthStatusConsistencyJSONDetails(t *testing.T) {
	rtmx.Req(t, "REQ-GO-074")

	dbContent := healthDBHeader +
		"REQ-001,CAT,Sub,Base requirement,Pass,,,Unit Test,MISSING,HIGH,1,,,,,,,,,\n" +
		"REQ-002,CAT,Sub,Dependent,Pass,,,Unit Test,COMPLETE,HIGH,1,,,REQ-001,,,,,,,,\n"

	tmpDir := setupHealthTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := createHealthTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"health", "--json"})

	err := cmd.Execute()
	if err != nil {
		var exitErr *ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("health --json failed unexpectedly: %v", err)
		}
	}

	// Parse as raw JSON to inspect details
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &raw); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	checks, ok := raw["checks"].([]interface{})
	if !ok {
		t.Fatal("expected checks array in JSON output")
	}

	var consistencyCheck map[string]interface{}
	for _, c := range checks {
		check := c.(map[string]interface{})
		if check["name"] == "status_consistency" {
			consistencyCheck = check
			break
		}
	}

	if consistencyCheck == nil {
		t.Fatal("status_consistency check not found in JSON output")
	}

	details, ok := consistencyCheck["details"].([]interface{})
	if !ok {
		t.Fatal("expected details array in status_consistency check")
	}

	if len(details) != 1 {
		t.Fatalf("expected 1 detail, got %d", len(details))
	}

	detail := details[0].(map[string]interface{})
	if detail["req_id"] != "REQ-002" {
		t.Errorf("expected req_id=REQ-002, got %v", detail["req_id"])
	}
	if detail["status"] != "COMPLETE" {
		t.Errorf("expected status=COMPLETE, got %v", detail["status"])
	}
	if detail["dependency"] != "REQ-001" {
		t.Errorf("expected dependency=REQ-001, got %v", detail["dependency"])
	}
	if detail["dep_status"] != "MISSING" {
		t.Errorf("expected dep_status=MISSING, got %v", detail["dep_status"])
	}
}
