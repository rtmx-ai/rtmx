package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx-go/pkg/rtmx"
	"github.com/spf13/cobra"
)

func TestDepsRealCommand(t *testing.T) {
	rtmx.Req(t, "REQ-GO-013")

	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createDepsTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"deps"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("deps command failed: %v", err)
	}

	output := buf.String()
	expectedPhrases := []string{
		"Dependencies",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Expected output to contain %q, got:\n%s", phrase, output)
		}
	}
}

// TestDepsTableFormat verifies that deps shows a table with all requirements
// REQ-GO-052: Go CLI deps shall show full requirements table like Python
func TestDepsTableFormat(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createDepsTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"deps"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("deps command failed: %v", err)
	}

	output := buf.String()

	// Verify table format with column headers
	expectedElements := []string{
		"ID",          // Column header
		"Deps",        // Column header
		"Blocks",      // Column header
		"Description", // Column header
		"---",         // Row separator
		"REQ-GO-",     // Should show requirement IDs
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected deps output to contain %q, got:\n%s", element, output)
		}
	}
}

func TestDepsWorkable(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createDepsTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"deps", "--workable"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("deps --workable failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Workable") {
		t.Errorf("Expected workable requirements output, got:\n%s", output)
	}
}

// createDepsTestCmd creates a root command with real deps command for testing
func createDepsTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	var reverse, all, workable bool

	depsTestCmd := &cobra.Command{
		Use:  "deps [req_id]",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			depsReverse = reverse
			depsAll = all
			depsWorkable = workable
			return runDeps(cmd, args)
		},
	}
	depsTestCmd.Flags().BoolVarP(&reverse, "reverse", "r", false, "show dependents")
	depsTestCmd.Flags().BoolVarP(&all, "all", "a", false, "show transitive")
	depsTestCmd.Flags().BoolVarP(&workable, "workable", "w", false, "show workable")
	root.AddCommand(depsTestCmd)

	return root
}

// setupDepsTestProject creates a temp dir with config and a database CSV.
func setupDepsTestProject(t *testing.T, dbContent string) string {
	t.Helper()
	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"), []byte("database:\n  path: .rtmx/database.csv\n"), 0644)
	_ = os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644)
	return tmpDir
}

const depsDBHeader = "req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id\n"

func TestShowReqDeps_DirectDependencies(t *testing.T) {
	// REQ-002 depends on REQ-001. REQ-003 depends on REQ-002.
	dbContent := depsDBHeader +
		"REQ-001,CAT,Sub,Base requirement,Pass,,,Unit Test,COMPLETE,HIGH,1,,,,,,,,,\n" +
		"REQ-002,CAT,Sub,Middle requirement,Pass,,,Unit Test,MISSING,HIGH,1,,,REQ-001,,,,,,,,\n" +
		"REQ-003,CAT,Sub,Top requirement,Pass,,,Unit Test,MISSING,HIGH,1,,,REQ-002,,,,,,,,\n"
	tmpDir := setupDepsTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := createDepsTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"deps", "REQ-002"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("deps REQ-002 failed: %v", err)
	}

	out := buf.String()
	// Should show REQ-002 header
	if !strings.Contains(out, "REQ-002") {
		t.Errorf("expected REQ-002 in output, got:\n%s", out)
	}
	// Should show "Direct Dependencies" section
	if !strings.Contains(out, "Direct Dependencies") {
		t.Errorf("expected 'Direct Dependencies' header, got:\n%s", out)
	}
	// Should show REQ-001 as a dependency
	if !strings.Contains(out, "REQ-001") {
		t.Errorf("expected REQ-001 as dependency, got:\n%s", out)
	}
}

func TestShowReqDeps_NoDependencies(t *testing.T) {
	dbContent := depsDBHeader +
		"REQ-001,CAT,Sub,Standalone requirement,Pass,,,Unit Test,COMPLETE,HIGH,1,,,,,,,,,\n"
	tmpDir := setupDepsTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := createDepsTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"deps", "REQ-001"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("deps REQ-001 failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "(none)") {
		t.Errorf("expected '(none)' for no dependencies, got:\n%s", out)
	}
}

func TestShowReqDeps_NotFound(t *testing.T) {
	dbContent := depsDBHeader +
		"REQ-001,CAT,Sub,Requirement,Pass,,,Unit Test,COMPLETE,HIGH,1,,,,,,,,,\n"
	tmpDir := setupDepsTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := createDepsTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"deps", "REQ-NONEXISTENT"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent requirement")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestShowReqDeps_Reverse(t *testing.T) {
	// REQ-002 depends on REQ-001
	dbContent := depsDBHeader +
		"REQ-001,CAT,Sub,Base requirement,Pass,,,Unit Test,COMPLETE,HIGH,1,,,,,,,,,\n" +
		"REQ-002,CAT,Sub,Dependent requirement,Pass,,,Unit Test,MISSING,HIGH,1,,,REQ-001,,,,,,,,\n"
	tmpDir := setupDepsTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := createDepsTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"deps", "--reverse", "REQ-001"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("deps --reverse REQ-001 failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Direct Dependents") {
		t.Errorf("expected 'Direct Dependents' header, got:\n%s", out)
	}
	if !strings.Contains(out, "REQ-002") {
		t.Errorf("expected REQ-002 as dependent, got:\n%s", out)
	}
}

func TestShowReqDeps_TransitiveAll(t *testing.T) {
	// Chain: REQ-003 -> REQ-002 -> REQ-001
	dbContent := depsDBHeader +
		"REQ-001,CAT,Sub,Root,Pass,,,Unit Test,COMPLETE,HIGH,1,,,,,,,,,\n" +
		"REQ-002,CAT,Sub,Middle,Pass,,,Unit Test,MISSING,HIGH,1,,,REQ-001,,,,,,,,\n" +
		"REQ-003,CAT,Sub,Top,Pass,,,Unit Test,MISSING,HIGH,1,,,REQ-002,,,,,,,,\n"
	tmpDir := setupDepsTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := createDepsTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"deps", "--all", "REQ-003"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("deps --all REQ-003 failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "All Dependencies (transitive)") {
		t.Errorf("expected transitive dependencies header, got:\n%s", out)
	}
	// Should show both REQ-002 and REQ-001
	if !strings.Contains(out, "REQ-002") {
		t.Errorf("expected REQ-002 in transitive deps, got:\n%s", out)
	}
	if !strings.Contains(out, "REQ-001") {
		t.Errorf("expected REQ-001 in transitive deps, got:\n%s", out)
	}
}

func TestShowReqDeps_TransitiveReverse(t *testing.T) {
	// Chain: REQ-003 -> REQ-002 -> REQ-001
	dbContent := depsDBHeader +
		"REQ-001,CAT,Sub,Root,Pass,,,Unit Test,COMPLETE,HIGH,1,,,,,,,,,\n" +
		"REQ-002,CAT,Sub,Middle,Pass,,,Unit Test,MISSING,HIGH,1,,,REQ-001,,,,,,,,\n" +
		"REQ-003,CAT,Sub,Top,Pass,,,Unit Test,MISSING,HIGH,1,,,REQ-002,,,,,,,,\n"
	tmpDir := setupDepsTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := createDepsTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"deps", "--all", "--reverse", "REQ-001"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("deps --all --reverse REQ-001 failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "All Dependents (transitive)") {
		t.Errorf("expected transitive dependents header, got:\n%s", out)
	}
	// Should show both REQ-002 and REQ-003
	if !strings.Contains(out, "REQ-002") {
		t.Errorf("expected REQ-002 in transitive dependents, got:\n%s", out)
	}
	if !strings.Contains(out, "REQ-003") {
		t.Errorf("expected REQ-003 in transitive dependents, got:\n%s", out)
	}
}

func TestShowReqDeps_BlockingInfo(t *testing.T) {
	// REQ-002 depends on REQ-001 which is incomplete
	dbContent := depsDBHeader +
		"REQ-001,CAT,Sub,Incomplete dep,Pass,,,Unit Test,MISSING,HIGH,1,,,,,,,,,\n" +
		"REQ-002,CAT,Sub,Blocked req,Pass,,,Unit Test,MISSING,HIGH,1,,,REQ-001,,,,,,,,\n"
	tmpDir := setupDepsTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := createDepsTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"deps", "REQ-002"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("deps REQ-002 failed: %v", err)
	}

	out := buf.String()
	// Should mention being blocked
	if !strings.Contains(out, "Blocked by") {
		t.Errorf("expected 'Blocked by' message, got:\n%s", out)
	}
	if !strings.Contains(out, "REQ-001") {
		t.Errorf("expected REQ-001 mentioned as blocker, got:\n%s", out)
	}
}

func TestShowReqDeps_AllDepsComplete(t *testing.T) {
	// REQ-002 depends on REQ-001 which is COMPLETE
	dbContent := depsDBHeader +
		"REQ-001,CAT,Sub,Complete dep,Pass,,,Unit Test,COMPLETE,HIGH,1,,,,,,,,,\n" +
		"REQ-002,CAT,Sub,Unblocked req,Pass,,,Unit Test,MISSING,HIGH,1,,,REQ-001,,,,,,,,\n"
	tmpDir := setupDepsTestProject(t, dbContent)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := createDepsTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"deps", "REQ-002"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("deps REQ-002 failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "All dependencies are complete") {
		t.Errorf("expected 'All dependencies are complete', got:\n%s", out)
	}
}
