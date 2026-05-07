package cmd

import (
	"bytes"
	"os"
	"os/exec"
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

func TestReleaseGateVersionPolicy(t *testing.T) {
	rtmx.Req(t, "REQ-PLAN-014")

	t.Run("policy_warn_insufficient_bump", func(t *testing.T) {
		tmpDir := t.TempDir()
		rtmxDir := filepath.Join(tmpDir, ".rtmx")
		_ = os.MkdirAll(rtmxDir, 0755)

		cfgContent := `rtmx:
  database: .rtmx/database.csv
  schema: core
  version_policy:
    enforcement: warn
    default: patch
    categories:
      CLI: minor
      DATA: major
      BENCH: none
`
		_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"), []byte(cfgContent), 0644)

		dbContent := testDBHeader +
			"REQ-001,CLI,Commands,New command,Pass,mod,TestA,Unit Test,COMPLETE,HIGH,1,,1,,,,v0.4.0,,,,\n"
		_ = os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644)

		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createReleaseTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"release", "gate", "v0.4.0", "--no-color"})

		err := cmd.Execute()
		// Gate should pass (warn mode, all complete)
		if err != nil {
			t.Fatalf("gate should pass in warn mode: %v\nOutput: %s", err, buf.String())
		}
		out := buf.String()
		if !strings.Contains(out, "Version policy") || !strings.Contains(out, "policy check") {
			t.Errorf("expected version policy output, got:\n%s", out)
		}
		if !strings.Contains(out, "minor") {
			t.Errorf("expected 'minor' bump level for CLI category, got:\n%s", out)
		}
	})

	t.Run("policy_enforce_blocks_release", func(t *testing.T) {
		tmpDir := t.TempDir()
		rtmxDir := filepath.Join(tmpDir, ".rtmx")
		_ = os.MkdirAll(rtmxDir, 0755)

		cfgContent := `rtmx:
  database: .rtmx/database.csv
  schema: core
  version_policy:
    enforcement: enforce
    default: patch
    categories:
      DATA: major
`
		_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"), []byte(cfgContent), 0644)

		// DATA category = major, but version bump is v0.3.0 -> v0.3.1 (patch)
		dbContent := testDBHeader +
			"REQ-001,DATA,Config,Schema change,Pass,mod,TestA,Unit Test,COMPLETE,HIGH,1,,1,,,,v0.3.1,,,,\n"
		_ = os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644)

		// Create a git repo with a tag so the policy can compute the bump
		initGitWithTag(t, tmpDir, "v0.3.0")

		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createReleaseTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"release", "gate", "v0.3.1", "--no-color"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("gate should fail in enforce mode with insufficient bump")
		}
		out := buf.String()
		if !strings.Contains(out, "FAIL") {
			t.Errorf("expected FAIL in output, got:\n%s", out)
		}
	})

	t.Run("policy_off_skips_check", func(t *testing.T) {
		tmpDir := t.TempDir()
		rtmxDir := filepath.Join(tmpDir, ".rtmx")
		_ = os.MkdirAll(rtmxDir, 0755)

		cfgContent := `rtmx:
  database: .rtmx/database.csv
  schema: core
  version_policy:
    enforcement: off
`
		_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"), []byte(cfgContent), 0644)

		dbContent := testDBHeader +
			"REQ-001,CLI,Commands,Feature,Pass,mod,TestA,Unit Test,COMPLETE,HIGH,1,,1,,,,v0.5.0,,,,\n"
		_ = os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644)

		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createReleaseTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"release", "gate", "v0.5.0", "--no-color"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("gate should pass with policy off: %v", err)
		}
		if strings.Contains(buf.String(), "Version policy") {
			t.Errorf("should not show version policy output when off, got:\n%s", buf.String())
		}
	})
}

// initGitWithTag creates a git repo with an initial commit, tag, then a second
// commit so that HEAD~1 resolves to the tagged commit.
func initGitWithTag(t *testing.T, dir, tag string) {
	t.Helper()
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "add", "."},
		{"git", "commit", "-m", "initial", "--no-gpg-sign"},
		{"git", "tag", tag},
	}
	for _, args := range cmds {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git command %v failed: %v\n%s", args, err, out)
		}
	}
	// Create a second commit so HEAD~1 resolves to the tagged commit
	markerPath := filepath.Join(dir, ".release-marker")
	_ = os.WriteFile(markerPath, []byte("release"), 0644)
	for _, args := range [][]string{
		{"git", "add", ".release-marker"},
		{"git", "commit", "-m", "post-tag", "--no-gpg-sign"},
	} {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git command %v failed: %v\n%s", args, err, out)
		}
	}
}
