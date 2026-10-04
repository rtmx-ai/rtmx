package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

func createWebsTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	var jsonOut bool
	webs := &cobra.Command{
		Use:  "webs",
		RunE: func(cmd *cobra.Command, args []string) error {
			websJSON = jsonOut
			return runWebs(cmd, args)
		},
	}
	webs.Flags().BoolVar(&jsonOut, "json", false, "")
	root.AddCommand(webs)
	return root
}

func TestWebsCommand(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-013")

	dbContent := testDBHeader +
		"REQ-A,CLI,Commands,Feature A,Pass,mod,TestA,Unit Test,MISSING,P0,1,,1.0,,,,,,,\n" +
		"REQ-B,CLI,Commands,Feature B,Pass,mod,TestB,Unit Test,MISSING,HIGH,1,,2.0,REQ-A,,,,,,\n" +
		"REQ-C,DATA,Config,Feature C,Pass,mod,TestC,Unit Test,MISSING,MEDIUM,1,,0.5,,,,,,,\n"

	t.Run("table_output", func(t *testing.T) {
		tmpDir := setupNextTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createWebsTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"webs"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("webs failed: %v\nOutput: %s", err, buf.String())
		}

		out := buf.String()
		if !strings.Contains(out, "Work Webs") {
			t.Errorf("expected 'Work Webs' header, got:\n%s", out)
		}
		if !strings.Contains(out, "parallel groups") {
			t.Errorf("expected 'parallel groups' in output, got:\n%s", out)
		}
	})

	t.Run("json_output", func(t *testing.T) {
		tmpDir := setupNextTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createWebsTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"webs", "--json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("webs --json failed: %v\nOutput: %s", err, buf.String())
		}

		out := buf.String()
		if !strings.Contains(out, "\"webs\"") {
			t.Errorf("expected JSON with 'webs' key, got:\n%s", out)
		}
		if !strings.Contains(out, "\"parallel_groups\"") {
			t.Errorf("expected JSON with 'parallel_groups' key, got:\n%s", out)
		}
		if !strings.Contains(out, "\"dependencies\"") {
			t.Errorf("expected JSON with 'dependencies' key, got:\n%s", out)
		}
		if !strings.Contains(out, "\"overlaps\"") {
			t.Errorf("expected JSON with 'overlaps' key, got:\n%s", out)
		}
	})

	t.Run("shows_web_details", func(t *testing.T) {
		tmpDir := setupNextTestProject(t, dbContent)
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		cmd := createWebsTestCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"webs"})

		_ = cmd.Execute()
		out := buf.String()
		if !strings.Contains(out, "Web 0") {
			t.Errorf("expected web details, got:\n%s", out)
		}
		if !strings.Contains(out, "IDs:") {
			t.Errorf("expected IDs listing, got:\n%s", out)
		}
	})
}
