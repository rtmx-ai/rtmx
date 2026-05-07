package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

func createReleaseTestCmd() *cobra.Command {
	root := &cobra.Command{Use: "rtmx"}
	rel := &cobra.Command{
		Use: "release",
	}

	gate := &cobra.Command{
		Use:  "gate",
		Args: cobra.ExactArgs(1),
		RunE: runReleaseGate,
	}
	gate.Flags().BoolVar(&releaseGateVerify, "verify", false, "")
	gate.Flags().BoolVar(&releaseGateJSON, "json", false, "")
	gate.Flags().BoolVar(&releaseAllowBreak, "allow-breaking", false, "")

	assign := &cobra.Command{
		Use:  "assign",
		Args: cobra.MinimumNArgs(2),
		RunE: runReleaseAssign,
	}
	assign.Flags().BoolVar(&releaseDryRun, "dry-run", false, "")

	unassign := &cobra.Command{
		Use:  "unassign",
		Args: cobra.MinimumNArgs(1),
		RunE: runReleaseUnassign,
	}
	unassign.Flags().BoolVar(&releaseDryRun, "dry-run", false, "")

	rel.AddCommand(gate, assign, unassign)
	root.AddCommand(rel)
	root.PersistentFlags().BoolVar(&noColor, "no-color", true, "")
	return root
}

func setupReleaseTestProject(t *testing.T, dbContent string) string {
	t.Helper()
	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"), []byte("rtmx:\n  database: .rtmx/database.csv\n  schema: core\n"), 0644)
	_ = os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644)
	return tmpDir
}

func TestReleaseGate(t *testing.T) {
	rtmx.Req(t, "REQ-PLAN-005")

	dbContent := testDBHeader +
		"REQ-001,CLI,Foundation,Req one,Pass,mod,TestA,Unit Test,COMPLETE,HIGH,1,,1,,,,v0.3.0,,,,\n" +
		"REQ-002,CLI,Commands,Req two,Pass,mod,TestB,Unit Test,COMPLETE,HIGH,1,,1,,,,v0.3.0,,,,\n" +
		"REQ-003,DATA,Config,Req three,Pass,mod,TestC,Unit Test,MISSING,HIGH,1,,1,,,,v0.4.0,,,,\n"

	t.Run("gate_passes_all_complete", func(t *testing.T) {
		tmpDir := setupReleaseTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createReleaseTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"release", "gate", "v0.3.0", "--no-color"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("gate should pass for v0.3.0 (all complete): %v\nOutput: %s", err, buf.String())
		}
		if !strings.Contains(buf.String(), "PASS") {
			t.Errorf("output should contain PASS, got: %s", buf.String())
		}
	})

	t.Run("gate_fails_incomplete", func(t *testing.T) {
		tmpDir := setupReleaseTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createReleaseTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"release", "gate", "v0.4.0", "--no-color"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("gate should fail for v0.4.0 (has MISSING requirement)")
		}
		if !strings.Contains(buf.String(), "FAIL") {
			t.Errorf("output should contain FAIL, got: %s", buf.String())
		}
	})

	t.Run("gate_fails_no_requirements", func(t *testing.T) {
		tmpDir := setupReleaseTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createReleaseTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"release", "gate", "v9.9.9", "--no-color"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("gate should fail when no requirements assigned")
		}
	})

	t.Run("gate_json_output", func(t *testing.T) {
		tmpDir := setupReleaseTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		releaseGateJSON = true
		defer func() { releaseGateJSON = false }()

		cmd := createReleaseTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"release", "gate", "v0.3.0", "--no-color", "--json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("gate should pass: %v", err)
		}
		if !strings.Contains(buf.String(), `"passed": true`) {
			t.Errorf("JSON should contain passed:true, got: %s", buf.String())
		}
	})
}

func TestReleaseAssign(t *testing.T) {
	rtmx.Req(t, "REQ-PLAN-007")

	dbContent := testDBHeader +
		"REQ-001,CLI,Foundation,Req one,Pass,mod,TestA,Unit Test,MISSING,HIGH,1,,1,,,,,,,,\n" +
		"REQ-002,CLI,Commands,Req two,Pass,mod,TestB,Unit Test,MISSING,HIGH,1,,1,,,,,,,,\n"

	t.Run("assign_sets_version", func(t *testing.T) {
		tmpDir := setupReleaseTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createReleaseTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"release", "assign", "v0.4.0", "REQ-001", "REQ-002", "--no-color"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("assign should succeed: %v\nOutput: %s", err, buf.String())
		}

		// Reload and verify
		dbPath := filepath.Join(tmpDir, ".rtmx", "database.csv")
		db, err := database.Load(dbPath)
		if err != nil {
			t.Fatalf("failed to reload database: %v", err)
		}
		r1 := db.Get("REQ-001")
		if r1.TargetVersion() != "v0.4.0" {
			t.Errorf("REQ-001 version = %q, want v0.4.0", r1.TargetVersion())
		}
		r2 := db.Get("REQ-002")
		if r2.TargetVersion() != "v0.4.0" {
			t.Errorf("REQ-002 version = %q, want v0.4.0", r2.TargetVersion())
		}
	})

	t.Run("assign_unknown_requirement", func(t *testing.T) {
		tmpDir := setupReleaseTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createReleaseTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"release", "assign", "v0.4.0", "REQ-NONEXISTENT", "--no-color"})

		_ = cmd.Execute()
		if !strings.Contains(buf.String(), "Unknown requirement") {
			t.Errorf("should warn about unknown requirement, got: %s", buf.String())
		}
	})

	t.Run("assign_skip_already_assigned", func(t *testing.T) {
		modifiedDB := testDBHeader +
			"REQ-001,CLI,Foundation,Req one,Pass,mod,TestA,Unit Test,MISSING,HIGH,1,,1,,,,v0.4.0,,,,\n"
		tmpDir := setupReleaseTestProject(t, modifiedDB)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createReleaseTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"release", "assign", "v0.4.0", "REQ-001", "--no-color"})

		_ = cmd.Execute()
		if !strings.Contains(buf.String(), "SKIP") {
			t.Errorf("should skip already-assigned, got: %s", buf.String())
		}
	})

	t.Run("unassign_clears_version", func(t *testing.T) {
		modifiedDB := testDBHeader +
			"REQ-001,CLI,Foundation,Req one,Pass,mod,TestA,Unit Test,MISSING,HIGH,1,,1,,,,v0.4.0,,,,\n"
		tmpDir := setupReleaseTestProject(t, modifiedDB)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createReleaseTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"release", "unassign", "REQ-001", "--no-color"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("unassign should succeed: %v\nOutput: %s", err, buf.String())
		}

		// Reload and verify
		dbPath := filepath.Join(tmpDir, ".rtmx", "database.csv")
		db, err := database.Load(dbPath)
		if err != nil {
			t.Fatalf("failed to reload database: %v", err)
		}
		r1 := db.Get("REQ-001")
		if r1.TargetVersion() != "" {
			t.Errorf("REQ-001 version = %q, want empty", r1.TargetVersion())
		}
	})
}
