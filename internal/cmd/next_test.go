package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

func createNextTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	var one, jsonOut bool
	next := &cobra.Command{
		Use:  "next",
		RunE: func(cmd *cobra.Command, args []string) error {
			nextOne = one
			nextJSON = jsonOut
			return runNext(cmd, args)
		},
	}
	next.Flags().BoolVar(&one, "one", false, "")
	next.Flags().BoolVar(&jsonOut, "json", false, "")
	root.AddCommand(next)
	return root
}

func setupNextTestProject(t *testing.T, dbContent string) string {
	t.Helper()
	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"),
		[]byte("rtmx:\n  database: .rtmx/database.csv\n  schema: core\n"), 0644)
	_ = os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644)
	return tmpDir
}

func TestNextShow(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-002")

	dbContent := testDBHeader +
		"REQ-A,CLI,Commands,Feature A,Pass,mod,TestA,Unit Test,MISSING,P0,1,,1.0,,,,,,,\n" +
		"REQ-B,CLI,Commands,Feature B,Pass,mod,TestB,Unit Test,MISSING,HIGH,1,,2.0,REQ-A,,,,,,\n" +
		"REQ-C,DATA,Config,Feature C,Pass,mod,TestC,Unit Test,MISSING,MEDIUM,1,,0.5,,,,,,,\n"

	t.Run("shows_webs_with_stats", func(t *testing.T) {
		tmpDir := setupNextTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createNextTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"next"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("next failed: %v\nOutput: %s", err, buf.String())
		}

		out := buf.String()
		if !strings.Contains(out, "Work Webs") {
			t.Errorf("expected 'Work Webs' header, got:\n%s", out)
		}
		if !strings.Contains(out, "2 web(s)") {
			t.Errorf("expected 2 webs (A-B connected, C isolated), got:\n%s", out)
		}
		if !strings.Contains(out, "REQ-A") {
			t.Errorf("expected REQ-A in output, got:\n%s", out)
		}
		if !strings.Contains(out, "REQ-C") {
			t.Errorf("expected REQ-C in output, got:\n%s", out)
		}
	})

	t.Run("blocked_shown", func(t *testing.T) {
		tmpDir := setupNextTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createNextTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"next"})

		_ = cmd.Execute()
		out := buf.String()
		// REQ-B depends on REQ-A (incomplete) so should be blocked
		if !strings.Contains(out, "blocked") {
			t.Errorf("expected 'blocked' marker for REQ-B, got:\n%s", out)
		}
	})

	t.Run("no_incomplete_requirements", func(t *testing.T) {
		allComplete := testDBHeader +
			"REQ-A,CLI,Commands,Feature A,Pass,mod,TestA,Unit Test,COMPLETE,HIGH,1,,1.0,,,,,,,\n"
		tmpDir := setupNextTestProject(t, allComplete)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createNextTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"next"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("next should succeed with no incomplete: %v", err)
		}
		if !strings.Contains(buf.String(), "No incomplete") {
			t.Errorf("expected 'No incomplete' message, got:\n%s", buf.String())
		}
	})
}

func TestNextOne(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-003")

	dbContent := testDBHeader +
		"REQ-A,CLI,Commands,Feature A,Pass,mod,TestA,Unit Test,MISSING,MEDIUM,1,,2.0,,,,,,,\n" +
		"REQ-B,DATA,Config,Feature B,Pass,mod,TestB,Unit Test,MISSING,P0,1,,1.0,,,,,,,\n" +
		"REQ-C,PLAN,Release,Feature C,Pass,mod,TestC,Unit Test,MISSING,HIGH,1,,0.5,REQ-B,,,,,,\n"

	t.Run("picks_highest_priority_unblocked", func(t *testing.T) {
		tmpDir := setupNextTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createNextTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"next", "--one"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("next --one failed: %v\nOutput: %s", err, buf.String())
		}

		out := buf.String()
		// REQ-B is P0 and unblocked, should be picked
		if !strings.Contains(out, "REQ-B") {
			t.Errorf("expected REQ-B (P0, unblocked), got:\n%s", out)
		}
		// REQ-C is HIGH but blocked by REQ-B, should not be picked
		if strings.Contains(out, "REQ-C") && strings.Contains(out, "Requirement:") {
			t.Errorf("REQ-C is blocked, should not be the primary pick")
		}
	})

	t.Run("json_output", func(t *testing.T) {
		tmpDir := setupNextTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createNextTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"next", "--one", "--json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("next --one --json failed: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, `"req_id":"REQ-B"`) {
			t.Errorf("JSON should contain REQ-B, got:\n%s", out)
		}
	})

	t.Run("tiebreak_by_effort", func(t *testing.T) {
		tieDB := testDBHeader +
			"REQ-X,CLI,Commands,Big feature,Pass,mod,TestX,Unit Test,MISSING,P0,1,,5.0,,,,,,,\n" +
			"REQ-Y,CLI,Commands,Small feature,Pass,mod,TestY,Unit Test,MISSING,P0,1,,0.5,,,,,,,\n"
		tmpDir := setupNextTestProject(t, tieDB)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createNextTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"next", "--one"})

		_ = cmd.Execute()
		out := buf.String()
		// Same priority, should pick smaller effort (REQ-Y)
		if !strings.Contains(out, "REQ-Y") {
			t.Errorf("expected REQ-Y (smaller effort), got:\n%s", out)
		}
	})
}
