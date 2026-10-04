package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

func createPlanParallelTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	var jsonOut bool
	pp := &cobra.Command{
		Use:  "plan-parallel",
		RunE: func(cmd *cobra.Command, args []string) error {
			planParallelJSON = jsonOut
			return runPlanParallel(cmd, args)
		},
	}
	pp.Flags().BoolVar(&jsonOut, "json", false, "")
	root.AddCommand(pp)
	return root
}

func TestPlanParallelCommand(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-014")

	dbContent := testDBHeader +
		"REQ-A,CLI,Commands,Feature A,Pass,mod,TestA,Unit Test,MISSING,P0,1,,1.0,,,,,,,\n" +
		"REQ-B,CLI,Commands,Feature B,Pass,mod,TestB,Unit Test,MISSING,HIGH,1,,2.0,REQ-A,,,,,,\n" +
		"REQ-C,DATA,Config,Feature C,Pass,mod,TestC,Unit Test,MISSING,MEDIUM,1,,0.5,,,,,,,\n"

	t.Run("table_output", func(t *testing.T) {
		tmpDir := setupNextTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createPlanParallelTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"plan-parallel"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("plan-parallel failed: %v\nOutput: %s", err, buf.String())
		}

		out := buf.String()
		if !strings.Contains(out, "Parallel Execution Plan") {
			t.Errorf("expected 'Parallel Execution Plan' header, got:\n%s", out)
		}
		if !strings.Contains(out, "Phase") {
			t.Errorf("expected phase details, got:\n%s", out)
		}
		if !strings.Contains(out, "Parallelism") {
			t.Errorf("expected parallelism factor, got:\n%s", out)
		}
	})

	t.Run("json_output", func(t *testing.T) {
		tmpDir := setupNextTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createPlanParallelTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"plan-parallel", "--json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("plan-parallel --json failed: %v\nOutput: %s", err, buf.String())
		}

		out := buf.String()
		if !strings.Contains(out, "\"phases\"") {
			t.Errorf("expected JSON with 'phases' key, got:\n%s", out)
		}
		if !strings.Contains(out, "\"critical_path_effort\"") {
			t.Errorf("expected JSON with 'critical_path_effort' key, got:\n%s", out)
		}
		if !strings.Contains(out, "\"parallelism_factor\"") {
			t.Errorf("expected JSON with 'parallelism_factor' key, got:\n%s", out)
		}
		if !strings.Contains(out, "\"merge_order\"") {
			t.Errorf("expected JSON with 'merge_order' key, got:\n%s", out)
		}
	})

	t.Run("shows_web_assignments", func(t *testing.T) {
		tmpDir := setupNextTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createPlanParallelTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"plan-parallel"})

		_ = cmd.Execute()
		out := buf.String()
		if !strings.Contains(out, "Web") {
			t.Errorf("expected web assignments, got:\n%s", out)
		}
		if !strings.Contains(out, "Merge order") {
			t.Errorf("expected merge order, got:\n%s", out)
		}
	})
}
