package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

func newLoopTickCLICmd() *cobra.Command {
	root := &cobra.Command{Use: "rtmx", SilenceUsage: true, SilenceErrors: true}
	var agentID string
	var strict bool
	loop := &cobra.Command{Use: "loop"}
	tick := &cobra.Command{
		Use: "tick",
		RunE: func(cmd *cobra.Command, args []string) error {
			loopTickAgentID = agentID
			loopTickStrict = strict
			return runLoopTickCLI(cmd, args)
		},
	}
	tick.Flags().StringVar(&agentID, "agent-id", "", "")
	tick.Flags().BoolVar(&strict, "strict", false, "")
	loop.AddCommand(tick)
	root.AddCommand(loop)
	return root
}

func TestLoopTickCLI(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-023b")

	t.Run("claims_and_returns_plan", func(t *testing.T) {
		db := testDBHeader +
			"REQ-EX-020,CLI,Commands,Atomic work,Pass,mod,TestA,Unit Test,MISSING,HIGH,1,,0.5,,,,,,,\n"
		dir := setupNextTestProject(t, db)
		writeLoopReqMD(t, dir, "REQ-EX-020", 2)
		openPRSource = staticOpenPRs{}
		t.Cleanup(func() { openPRSource = nil })

		orig, _ := os.Getwd()
		if err := os.Chdir(dir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(orig) }()

		buf := new(bytes.Buffer)
		cmd := newLoopTickCLICmd()
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"loop", "tick", "--agent-id", "agent-1"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("tick: %v\n%s", err, buf.String())
		}
		var plan loopPlan
		if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &plan); err != nil {
			t.Fatalf("json: %v\n%s", err, buf.String())
		}
		if plan.ReqID != "REQ-EX-020" || !plan.Claimed {
			t.Fatalf("plan: %+v", plan)
		}
		if plan.PRPolicy == "" || plan.CommitPolicy == "" {
			t.Fatalf("missing policies: %+v", plan)
		}
		if _, err := os.Stat(filepath.Join(dir, ".rtmx", "claims", "REQ-EX-020.json")); err != nil {
			t.Fatalf("claim file: %v", err)
		}
	})

	t.Run("redecompose_from_delivery_note", func(t *testing.T) {
		db := testDBHeader +
			"REQ-EX-030,CLI,Commands,Next sibling,Pass,mod,TestA,Unit Test,MISSING,HIGH,1,,0.5,,,,,,,\n" +
			"REQ-EX-031,CLI,Commands,Coarse follow-on,Pass,mod,TestB,Unit Test,MISSING,MEDIUM,1,,1.0,,,,,,,\n"
		dir := setupNextTestProject(t, db)
		writeLoopReqMD(t, dir, "REQ-EX-030", 2)
		writeLoopReqMD(t, dir, "REQ-EX-031", 6)
		delDir := filepath.Join(dir, ".rtmx", "delivery")
		if err := os.MkdirAll(delDir, 0o755); err != nil {
			t.Fatal(err)
		}
		note := "# Prior\n\n## Follow-on decomposition\n\n- REQ-EX-031 needs split\n"
		if err := os.WriteFile(filepath.Join(delDir, "REQ-PRIOR.md"), []byte(note), 0o644); err != nil {
			t.Fatal(err)
		}
		openPRSource = staticOpenPRs{}
		t.Cleanup(func() { openPRSource = nil })

		orig, _ := os.Getwd()
		if err := os.Chdir(dir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(orig) }()

		buf := new(bytes.Buffer)
		cmd := newLoopTickCLICmd()
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"loop", "tick", "--agent-id", "agent-2"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("tick: %v\n%s", err, buf.String())
		}
		var plan loopPlan
		if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &plan); err != nil {
			t.Fatalf("json: %v\n%s", err, buf.String())
		}
		found := false
		for _, id := range plan.Redecomposed {
			if id == "REQ-EX-031" {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected REQ-EX-031 in redecomposed: %+v", plan)
		}
	})

	t.Run("idle_when_empty", func(t *testing.T) {
		db := testDBHeader +
			"REQ-EX-040,CLI,Commands,Done,Pass,mod,TestA,Unit Test,COMPLETE,HIGH,1,,0.5,,,,,,,\n"
		dir := setupNextTestProject(t, db)
		openPRSource = staticOpenPRs{}
		t.Cleanup(func() { openPRSource = nil })
		orig, _ := os.Getwd()
		if err := os.Chdir(dir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(orig) }()
		buf := new(bytes.Buffer)
		cmd := newLoopTickCLICmd()
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"loop", "tick", "--agent-id", "agent-3"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("tick: %v\n%s", err, buf.String())
		}
		if !strings.Contains(buf.String(), `"idle":true`) {
			t.Fatalf("want idle: %s", buf.String())
		}
	})
}
