package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx-go/pkg/rtmx"
	"github.com/spf13/cobra"
)

func TestCyclesRealCommand(t *testing.T) {
	rtmx.Req(t, "REQ-GO-013")

	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createCyclesTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"cycles"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("cycles command failed: %v", err)
	}

	output := buf.String()
	expectedPhrases := []string{
		"Circular Dependency Analysis",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Expected output to contain %q, got:\n%s", phrase, output)
		}
	}
}

// TestCyclesDetailFormat verifies that cycles shows statistics and recommendations
// REQ-GO-053: Go CLI cycles shall show stats paths and recommendations
func TestCyclesDetailFormat(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createCyclesTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"cycles"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("cycles command failed: %v", err)
	}

	output := buf.String()

	// Verify statistics section exists
	expectedElements := []string{
		"RTM Statistics:",
		"Total requirements:",
		"Total dependencies:",
		"Average dependencies per requirement:",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected cycles output to contain %q, got:\n%s", element, output)
		}
	}
}

// createCyclesTestCmd creates a root command with real cycles command for testing
func createCyclesTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	var jsonOutput bool

	cyclesTestCmd := &cobra.Command{
		Use: "cycles",
		RunE: func(cmd *cobra.Command, args []string) error {
			cyclesJSON = jsonOutput
			return runCycles(cmd, args)
		},
	}
	cyclesTestCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	root.AddCommand(cyclesTestCmd)

	return root
}

// setupCyclesTestProject creates a temp dir with config and database CSV.
func setupCyclesTestProject(t *testing.T, dbContent string) string {
	t.Helper()
	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"), []byte("database:\n  path: .rtmx/database.csv\n"), 0644)
	_ = os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644)
	return tmpDir
}

const cyclesDBHeader = "req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id\n"

func TestOutputCyclesJSON_NoCycles(t *testing.T) {
	// No circular dependencies: REQ-002 depends on REQ-001 (acyclic)
	dbContent := cyclesDBHeader +
		"REQ-001,CAT,Sub,Base requirement,Pass,,,Unit Test,COMPLETE,HIGH,1,,,,,,,,,\n" +
		"REQ-002,CAT,Sub,Dependent requirement,Pass,,,Unit Test,MISSING,HIGH,1,,,REQ-001,,,,,,,,\n"
	tmpDir := setupCyclesTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := createCyclesTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"cycles", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cycles --json failed: %v", err)
	}

	out := buf.String()

	// Parse the JSON output
	var result cycleResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, out)
	}

	if result.Found {
		t.Error("expected Found=false for acyclic graph")
	}
	if result.Count != 0 {
		t.Errorf("expected Count=0, got %d", result.Count)
	}
	if result.Cycles != nil && len(result.Cycles) > 0 {
		t.Errorf("expected empty cycles, got %v", result.Cycles)
	}
}

func TestOutputCyclesJSON_EmptyDatabase(t *testing.T) {
	dbContent := cyclesDBHeader
	tmpDir := setupCyclesTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := createCyclesTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"cycles", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cycles --json failed: %v", err)
	}

	out := buf.String()
	var result cycleResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, out)
	}

	if result.Found {
		t.Error("expected Found=false for empty database")
	}
	if result.Count != 0 {
		t.Errorf("expected Count=0, got %d", result.Count)
	}
}

func TestCyclesText_NoCycles(t *testing.T) {
	dbContent := cyclesDBHeader +
		"REQ-001,CAT,Sub,Base requirement,Pass,,,Unit Test,COMPLETE,HIGH,1,,,,,,,,,\n" +
		"REQ-002,CAT,Sub,Dependent requirement,Pass,,,Unit Test,MISSING,HIGH,1,,,REQ-001,,,,,,,,\n"
	tmpDir := setupCyclesTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := createCyclesTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"cycles"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cycles failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "No circular dependencies found") {
		t.Errorf("expected 'No circular dependencies found', got:\n%s", out)
	}
	if !strings.Contains(out, "acyclic (DAG)") {
		t.Errorf("expected 'acyclic (DAG)' message, got:\n%s", out)
	}
}

func TestCyclesText_Statistics(t *testing.T) {
	dbContent := cyclesDBHeader +
		"REQ-001,CAT,Sub,Req one,Pass,,,Unit Test,COMPLETE,HIGH,1,,,,,,,,,\n" +
		"REQ-002,CAT,Sub,Req two,Pass,,,Unit Test,MISSING,HIGH,1,,,REQ-001,,,,,,,,\n" +
		"REQ-003,CAT,Sub,Req three,Pass,,,Unit Test,MISSING,HIGH,1,,,REQ-001,,,,,,,,\n"
	tmpDir := setupCyclesTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := createCyclesTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"cycles"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cycles failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Total requirements: 3") {
		t.Errorf("expected 'Total requirements: 3', got:\n%s", out)
	}
	if !strings.Contains(out, "Total dependencies: 2") {
		t.Errorf("expected 'Total dependencies: 2', got:\n%s", out)
	}
	if !strings.Contains(out, "Average dependencies per requirement:") {
		t.Errorf("expected average deps in output, got:\n%s", out)
	}
}

func TestCyclesJSON_SingleNodeNoCycle(t *testing.T) {
	// A single requirement with no deps has no cycles
	dbContent := cyclesDBHeader +
		"REQ-001,CAT,Sub,Only requirement,Pass,,,Unit Test,COMPLETE,HIGH,1,,,,,,,,,\n"
	tmpDir := setupCyclesTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := createCyclesTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"cycles", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cycles --json failed: %v", err)
	}

	var result cycleResult
	if err := json.Unmarshal([]byte(buf.String()), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if result.Found {
		t.Error("single node should have no cycles")
	}
}

func TestCyclesJSON_LongChainNoCycles(t *testing.T) {
	// Linear chain: REQ-005 -> REQ-004 -> REQ-003 -> REQ-002 -> REQ-001
	dbContent := cyclesDBHeader +
		"REQ-001,CAT,Sub,Req 1,Pass,,,Unit Test,COMPLETE,HIGH,1,,,,,,,,,\n" +
		"REQ-002,CAT,Sub,Req 2,Pass,,,Unit Test,MISSING,HIGH,1,,,REQ-001,,,,,,,,\n" +
		"REQ-003,CAT,Sub,Req 3,Pass,,,Unit Test,MISSING,HIGH,1,,,REQ-002,,,,,,,,\n" +
		"REQ-004,CAT,Sub,Req 4,Pass,,,Unit Test,MISSING,HIGH,1,,,REQ-003,,,,,,,,\n" +
		"REQ-005,CAT,Sub,Req 5,Pass,,,Unit Test,MISSING,HIGH,1,,,REQ-004,,,,,,,,\n"
	tmpDir := setupCyclesTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := createCyclesTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"cycles", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("cycles --json failed: %v", err)
	}

	var result cycleResult
	if err := json.Unmarshal([]byte(buf.String()), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if result.Found {
		t.Error("linear chain should have no cycles")
	}
	if result.Count != 0 {
		t.Errorf("expected 0 cycles, got %d", result.Count)
	}
}

func TestCycles_MissingConfig(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := createCyclesTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"cycles"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when no config exists")
	}
}

func TestCycles_MissingDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"), []byte("database:\n  path: .rtmx/database.csv\n"), 0644)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := createCyclesTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"cycles"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when database file is missing")
	}
	if !strings.Contains(err.Error(), "failed to load database") {
		t.Errorf("expected 'failed to load database' error, got: %v", err)
	}
}
