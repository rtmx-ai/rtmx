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

func TestTuiCommand(t *testing.T) {
	rtmx.Req(t, "REQ-GO-038")

	t.Run("renders_dashboard", func(t *testing.T) {
		dbContent := testDBHeader +
			"REQ-A,CLI,Commands,Feature A,Pass,mod,TestA,Unit Test,COMPLETE,HIGH,1,,1.0,,,,,,,\n" +
			"REQ-B,CLI,Commands,Feature B,Pass,mod,TestB,Unit Test,MISSING,HIGH,1,,1.0,,,,,,,\n" +
			"REQ-C,DATA,Config,Feature C,Pass,mod,TestC,Unit Test,PARTIAL,MEDIUM,1,,0.5,,,,,,,\n"

		tmpDir := t.TempDir()
		rtmxDir := filepath.Join(tmpDir, ".rtmx")
		_ = os.MkdirAll(rtmxDir, 0755)
		_ = os.WriteFile(filepath.Join(tmpDir, "rtmx.yaml"),
			[]byte("rtmx:\n  database: .rtmx/database.csv\n  schema: core\n"), 0644)
		_ = os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644)

		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		root := &cobra.Command{Use: "rtmx", SilenceUsage: true, SilenceErrors: true}
		tui := &cobra.Command{Use: "tui", RunE: runTui}
		root.AddCommand(tui)

		buf := new(bytes.Buffer)
		root.SetOut(buf)
		root.SetArgs([]string{"tui"})

		err := root.Execute()
		if err != nil {
			t.Fatalf("tui failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "Dashboard") {
			t.Error("expected 'Dashboard' header in output")
		}
		if !strings.Contains(out, "complete") {
			t.Error("expected 'complete' in status output")
		}
		if !strings.Contains(out, "%") {
			t.Error("expected percentage in output")
		}
	})

	t.Run("flags_registered", func(t *testing.T) {
		if tuiCmd.Flags().Lookup("once") == nil {
			t.Error("tui should have --once flag")
		}
	})
}
