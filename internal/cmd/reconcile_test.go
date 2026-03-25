package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupReconcileProject creates a temp directory with config and database for reconcile tests.
func setupReconcileProject(t *testing.T, csvContent string) string {
	t.Helper()
	dir := t.TempDir()

	// Create .rtmx directory and config
	rtmxDir := filepath.Join(dir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0755); err != nil {
		t.Fatalf("failed to create .rtmx dir: %v", err)
	}

	cfgContent := `rtmx:
  database: .rtmx/database.csv
`
	if err := os.WriteFile(filepath.Join(rtmxDir, "config.yaml"), []byte(cfgContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	if err := os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(csvContent), 0644); err != nil {
		t.Fatalf("failed to write database: %v", err)
	}

	return dir
}

// reconcileCSVHeader is the standard header for reconcile test CSVs.
const reconcileCSVHeader = "req_id,category,requirement_text,status,priority,phase,dependencies,blocks\n"

func TestReconcileDryRunNoIssues(t *testing.T) {
	// All reciprocal relationships are correct: A depends on B, B blocks A
	csv := reconcileCSVHeader +
		"REQ-A,CORE,Feature A,MISSING,MEDIUM,1,REQ-B,\n" +
		"REQ-B,CORE,Feature B,MISSING,MEDIUM,1,,REQ-A\n"

	dir := setupReconcileProject(t, csv)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createReconcileTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"reconcile"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "All dependencies are reciprocal") {
		t.Errorf("expected 'All dependencies are reciprocal' message, got:\n%s", output)
	}
	if !strings.Contains(output, "No fixes needed") {
		t.Errorf("expected 'No fixes needed' message, got:\n%s", output)
	}
}

func TestReconcileDryRunWithIssues(t *testing.T) {
	// REQ-A depends on REQ-B, but REQ-B does not block REQ-A
	csv := reconcileCSVHeader +
		"REQ-A,CORE,Feature A,MISSING,MEDIUM,1,REQ-B,\n" +
		"REQ-B,CORE,Feature B,MISSING,MEDIUM,1,,\n"

	dir := setupReconcileProject(t, csv)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createReconcileTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"reconcile"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "reciprocity issue") {
		t.Errorf("expected reciprocity issue message, got:\n%s", output)
	}
	if !strings.Contains(output, "Dry-run mode") {
		t.Errorf("expected 'Dry-run mode' message, got:\n%s", output)
	}
	if !strings.Contains(output, "REQ-B") {
		t.Errorf("expected REQ-B mentioned in fix, got:\n%s", output)
	}
}

func TestReconcileExecuteMissingBlock(t *testing.T) {
	// REQ-A depends on REQ-B, but REQ-B does not block REQ-A
	// Execute should fix this by adding block to REQ-B
	csv := reconcileCSVHeader +
		"REQ-A,CORE,Feature A,MISSING,MEDIUM,1,REQ-B,\n" +
		"REQ-B,CORE,Feature B,MISSING,MEDIUM,1,,\n"

	dir := setupReconcileProject(t, csv)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createReconcileTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"reconcile", "--execute"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("reconcile --execute failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Applying fixes") {
		t.Errorf("expected 'Applying fixes' message, got:\n%s", output)
	}
	if !strings.Contains(output, "Added block") {
		t.Errorf("expected 'Added block' message, got:\n%s", output)
	}
	if !strings.Contains(output, "Applied") {
		t.Errorf("expected 'Applied N fixes' message, got:\n%s", output)
	}

	// Verify the database was actually modified on disk
	dbContent, err := os.ReadFile(filepath.Join(dir, ".rtmx", "database.csv"))
	if err != nil {
		t.Fatalf("failed to read saved database: %v", err)
	}
	dbStr := string(dbContent)
	// REQ-B should now block REQ-A
	lines := strings.Split(dbStr, "\n")
	var reqBLine string
	for _, line := range lines {
		if strings.HasPrefix(line, "REQ-B,") {
			reqBLine = line
			break
		}
	}
	if reqBLine == "" {
		t.Fatal("could not find REQ-B line in saved database")
	}
	if !strings.Contains(reqBLine, "REQ-A") {
		t.Errorf("expected REQ-B to block REQ-A after fix, got line: %s", reqBLine)
	}
}

func TestReconcileExecuteMissingDependency(t *testing.T) {
	// REQ-A blocks REQ-B, but REQ-B does not depend on REQ-A
	csv := reconcileCSVHeader +
		"REQ-A,CORE,Feature A,MISSING,MEDIUM,1,,REQ-B\n" +
		"REQ-B,CORE,Feature B,MISSING,MEDIUM,1,,\n"

	dir := setupReconcileProject(t, csv)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createReconcileTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"reconcile", "--execute"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("reconcile --execute failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Added dependency") {
		t.Errorf("expected 'Added dependency' message, got:\n%s", output)
	}

	// Verify REQ-B now depends on REQ-A in the saved file
	dbContent, err := os.ReadFile(filepath.Join(dir, ".rtmx", "database.csv"))
	if err != nil {
		t.Fatalf("failed to read saved database: %v", err)
	}
	dbStr := string(dbContent)
	lines := strings.Split(dbStr, "\n")
	var reqBLine string
	for _, line := range lines {
		if strings.HasPrefix(line, "REQ-B,") {
			reqBLine = line
			break
		}
	}
	if reqBLine == "" {
		t.Fatal("could not find REQ-B line in saved database")
	}
	if !strings.Contains(reqBLine, "REQ-A") {
		t.Errorf("expected REQ-B to depend on REQ-A after fix, got line: %s", reqBLine)
	}
}

func TestReconcileMultipleIssues(t *testing.T) {
	// Multiple missing reciprocal relationships
	csv := reconcileCSVHeader +
		"REQ-A,CORE,Feature A,MISSING,MEDIUM,1,REQ-B|REQ-C,\n" +
		"REQ-B,CORE,Feature B,MISSING,MEDIUM,1,,\n" +
		"REQ-C,CORE,Feature C,MISSING,MEDIUM,1,,REQ-D\n" +
		"REQ-D,CORE,Feature D,MISSING,MEDIUM,1,,\n"

	dir := setupReconcileProject(t, csv)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createReconcileTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"reconcile"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}

	output := buf.String()
	// Should find multiple issues:
	// REQ-B should block REQ-A (missing block)
	// REQ-C should block REQ-A (missing block)
	// REQ-D should depend on REQ-C (missing dep)
	if !strings.Contains(output, "missing blocks") || !strings.Contains(output, "missing dependencies") {
		t.Errorf("expected summary with missing blocks and dependencies, got:\n%s", output)
	}
}

func TestReconcileNoDatabase(t *testing.T) {
	dir := t.TempDir()

	// Create config pointing to nonexistent database
	rtmxDir := filepath.Join(dir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0755); err != nil {
		t.Fatalf("failed to create .rtmx dir: %v", err)
	}
	cfgContent := `rtmx:
  database: .rtmx/database.csv
`
	if err := os.WriteFile(filepath.Join(rtmxDir, "config.yaml"), []byte(cfgContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createReconcileTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"reconcile"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when database does not exist")
	}
	if !strings.Contains(err.Error(), "failed to load database") {
		t.Errorf("expected 'failed to load database' error, got: %v", err)
	}
}

func TestReconcileExecuteVerifyDatabaseSaved(t *testing.T) {
	// Comprehensive test: create issues, execute fixes, reload and verify consistency
	csv := reconcileCSVHeader +
		"REQ-A,CORE,Feature A,MISSING,MEDIUM,1,REQ-B,\n" +
		"REQ-B,CORE,Feature B,MISSING,MEDIUM,1,,\n"

	dir := setupReconcileProject(t, csv)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(oldWd) }()

	// Execute reconcile
	rootCmd := createReconcileTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"reconcile", "--execute"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("reconcile --execute failed: %v", err)
	}

	// Now run reconcile again (dry-run) and verify no issues remain
	rootCmd2 := createReconcileTestCmd()
	buf2 := new(bytes.Buffer)
	rootCmd2.SetOut(buf2)
	rootCmd2.SetArgs([]string{"reconcile"})

	err = rootCmd2.Execute()
	if err != nil {
		t.Fatalf("second reconcile failed: %v", err)
	}

	output2 := buf2.String()
	if !strings.Contains(output2, "All dependencies are reciprocal") {
		t.Errorf("expected no issues after executing fixes, got:\n%s", output2)
	}
}
