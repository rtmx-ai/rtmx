package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

func createMergeGateTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	var jsonOut bool
	var webIdx int
	mg := &cobra.Command{
		Use:  "merge-gate",
		RunE: func(cmd *cobra.Command, args []string) error {
			mergeGateJSON = jsonOut
			mergeGateWebIndex = webIdx
			return runMergeGate(cmd, args)
		},
	}
	mg.Flags().IntVar(&webIdx, "web", -1, "")
	mg.Flags().BoolVar(&jsonOut, "json", false, "")
	root.AddCommand(mg)
	return root
}

func TestMergeGateCommand(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-015")

	t.Run("incomplete_web_fails", func(t *testing.T) {
		dbContent := testDBHeader +
			"REQ-A,CLI,Commands,Feature A,Pass,mod,TestA,Unit Test,MISSING,P0,1,,1.0,,,,,,,\n"

		tmpDir := setupNextTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createMergeGateTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"merge-gate", "--web", "0"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for incomplete web")
		}
		out := buf.String()
		if !strings.Contains(out, "NOT safe") {
			t.Errorf("incomplete web should report NOT safe, got:\n%s", out)
		}
	})

	t.Run("missing_web_flag_fails", func(t *testing.T) {
		dbContent := testDBHeader +
			"REQ-A,CLI,Commands,Feature A,Pass,mod,TestA,Unit Test,MISSING,P0,1,,1.0,,,,,,,\n"

		tmpDir := setupNextTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createMergeGateTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"merge-gate"})

		err := cmd.Execute()
		if err == nil {
			t.Error("expected error when --web not provided")
		}
	})

	t.Run("json_output", func(t *testing.T) {
		dbContent := testDBHeader +
			"REQ-A,CLI,Commands,Feature A,Pass,mod,TestA,Unit Test,MISSING,P0,1,,1.0,,,,,,,\n"

		tmpDir := setupNextTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createMergeGateTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"merge-gate", "--web", "0", "--json"})

		_ = cmd.Execute()
		out := buf.String()
		if !strings.Contains(out, "\"safe\"") {
			t.Errorf("expected JSON with 'safe' key, got:\n%s", out)
		}
		if !strings.Contains(out, "\"incomplete_ids\"") {
			t.Errorf("expected JSON with 'incomplete_ids' key, got:\n%s", out)
		}
	})

	t.Run("out_of_range_web_fails", func(t *testing.T) {
		dbContent := testDBHeader +
			"REQ-A,CLI,Commands,Feature A,Pass,mod,TestA,Unit Test,MISSING,P0,1,,1.0,,,,,,,\n"

		tmpDir := setupNextTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createMergeGateTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"merge-gate", "--web", "99"})

		err := cmd.Execute()
		if err == nil {
			t.Error("expected error for out-of-range web index")
		}
	})
}
