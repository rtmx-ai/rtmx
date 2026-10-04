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

type scriptedMerged struct {
	batches [][]PullRequest
	n       int
}

func (s *scriptedMerged) ListMerged() ([]PullRequest, error) {
	if len(s.batches) == 0 {
		return nil, nil
	}
	if s.n >= len(s.batches) {
		return s.batches[len(s.batches)-1], nil
	}
	out := s.batches[s.n]
	s.n++
	return out, nil
}

func withLoopWatch(t *testing.T, src PRSource, polls int) {
	t.Helper()
	prevSrc := loopMergedSource
	prevWait := loopWait
	prevLimit := loopPollLimit
	prevEdges := loopEdgesOverride
	prevPR := openPRSource
	loopMergedSource = src
	loopPollLimit = polls
	loopEdgesOverride = nil
	if openPRSource == nil {
		openPRSource = staticOpenPRs{}
	}
	waits := 0
	loopWait = func() { waits++ }
	t.Cleanup(func() {
		loopMergedSource = prevSrc
		loopWait = prevWait
		loopPollLimit = prevLimit
		loopEdgesOverride = prevEdges
		openPRSource = prevPR
	})
	_ = waits
}

func runLoopCmd(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	root := &cobra.Command{Use: "rtmx", SilenceUsage: true, SilenceErrors: true}
	var once, jsonOut bool
	loop := &cobra.Command{
		Use: "loop",
		RunE: func(cmd *cobra.Command, args []string) error {
			loopOnce = once
			loopJSON = jsonOut
			return runLoop(cmd, args)
		},
	}
	loop.Flags().BoolVar(&once, "once", false, "")
	loop.Flags().BoolVar(&jsonOut, "json", false, "")
	root.AddCommand(loop)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestLoopOneTickPerEdge(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-019a")
	src := &scriptedMerged{batches: [][]PullRequest{{
		{Number: 1, Title: "feat(REQ-EX-020): a", Merged: true, SHA: "aaa"},
		{Number: 2, Title: "feat(REQ-EX-021): b", Merged: true, SHA: "bbb"},
	}}}
	withLoopWatch(t, src, 1)

	db := testDBHeader +
		"REQ-EX-020,CLI,Commands,One,Pass,mod,TestA,Unit Test,MISSING,HIGH,1,,1.0,,,,,,,\n" +
		"REQ-EX-021,CLI,Commands,Two,Pass,mod,TestB,Unit Test,MISSING,HIGH,1,,1.0,,,,,,,\n"
	dir := setupNextTestProject(t, db)
	writeLoopReqMD(t, dir, "REQ-EX-020", 2)
	writeLoopReqMD(t, dir, "REQ-EX-021", 2)
	if err := SaveLoopCursor(LoopCursorPath(dir), LoopCursor{}); err != nil {
		t.Fatal(err)
	}

	out, err := runLoopCmd(t, dir, "loop")
	if err != nil {
		t.Fatalf("loop: %v\n%s", err, out)
	}
	if strings.Count(out, `"req_id"`) != 2 {
		t.Fatalf("expected one tick per edge, got:\n%s", out)
	}
	claims, _ := filepath.Glob(filepath.Join(dir, ".rtmx", "claims", "*.json"))
	if len(claims) != 2 {
		t.Fatalf("claims: %v", claims)
	}
}

func TestLoopBootstrapThenWait(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-019a")
	waits := 0
	prevWait := loopWait
	loopWait = func() { waits++ }
	t.Cleanup(func() { loopWait = prevWait })
	withLoopWatch(t, &scriptedMerged{}, 1)
	loopWait = func() { waits++ }

	db := testDBHeader +
		"REQ-EX-030,CLI,Commands,Boot,Pass,mod,TestA,Unit Test,MISSING,HIGH,1,,1.0,,,,,,,\n"
	dir := setupNextTestProject(t, db)
	writeLoopReqMD(t, dir, "REQ-EX-030", 2)

	out, err := runLoopCmd(t, dir, "loop")
	if err != nil {
		t.Fatalf("loop: %v\n%s", err, out)
	}
	if strings.Count(out, `"req_id"`) != 1 {
		t.Fatalf("bootstrap should tick once, got:\n%s", out)
	}
	if waits < 1 {
		t.Fatal("runner should wait for edges after bootstrap")
	}
	if _, err := os.Stat(LoopCursorPath(dir)); err != nil {
		t.Fatal(err)
	}

	waits = 0
	out, err = runLoopCmd(t, dir, "loop")
	if err != nil {
		t.Fatalf("second loop: %v\n%s", err, out)
	}
	if strings.Contains(out, `"req_id"`) {
		t.Fatalf("bootstrap must not repeat: %s", out)
	}
	claims, _ := filepath.Glob(filepath.Join(dir, ".rtmx", "claims", "*.json"))
	if len(claims) != 1 {
		t.Fatalf("claims after second wait: %v", claims)
	}
}

func TestLoopOnceIdleAndAtMostOne(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-019a")
	dir := t.TempDir()
	out, err := runLoopCmd(t, dir, "loop", "--once")
	if err != nil {
		t.Fatalf("empty --once should exit 0: %v\n%s", err, out)
	}
	if strings.TrimSpace(out) != `{"idle":true}` {
		t.Fatalf("idle line: %q", out)
	}

	edges := []LoopEdge{
		{ReqID: "REQ-EX-040", Source: "pr", PRNumber: 1},
		{ReqID: "REQ-EX-041", Source: "pr", PRNumber: 2},
	}
	prev := loopEdgesOverride
	loopEdgesOverride = &edges
	t.Cleanup(func() { loopEdgesOverride = prev })
	if openPRSource == nil {
		openPRSource = staticOpenPRs{}
	}

	db := testDBHeader +
		"REQ-EX-040,CLI,Commands,Only,Pass,mod,TestA,Unit Test,MISSING,HIGH,1,,1.0,,,,,,,\n"
	proj := setupNextTestProject(t, db)
	writeLoopReqMD(t, proj, "REQ-EX-040", 2)
	out, err = runLoopCmd(t, proj, "loop", "--once")
	if err != nil {
		t.Fatalf("once: %v\n%s", err, out)
	}
	if strings.Count(out, `"req_id"`) != 1 {
		t.Fatalf("--once must tick at most once:\n%s", out)
	}
}

func newInstallCmd() *cobra.Command {
	root := &cobra.Command{Use: "rtmx", SilenceUsage: true, SilenceErrors: true}
	var dry, force bool
	loop := &cobra.Command{Use: "loop"}
	inst := &cobra.Command{
		Use: "install",
		RunE: func(cmd *cobra.Command, args []string) error {
			loopInstallDry = dry
			loopInstallForce = force
			return runLoopInstall(cmd, args)
		},
	}
	inst.Flags().BoolVar(&dry, "dry-run", false, "")
	inst.Flags().BoolVar(&force, "force", false, "")
	loop.AddCommand(inst)
	root.AddCommand(loop)
	return root
}

func TestLoopInstall(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-019a")
	prevOS := loopInstallGOOS
	prevEn := loopEnable
	t.Cleanup(func() {
		loopInstallGOOS = prevOS
		loopEnable = prevEn
		loopInstallDry = false
		loopInstallForce = false
	})

	t.Run("dry_run_writes_nothing", func(t *testing.T) {
		loopInstallGOOS = "darwin"
		dir := t.TempDir()
		orig, _ := os.Getwd()
		if err := os.Chdir(dir); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chdir(orig) })
		cmd := newInstallCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"loop", "install", "--dry-run"})
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
		out := buf.String()
		if !strings.Contains(out, "rtmx") || !strings.Contains(out, "loop") || !strings.Contains(out, "Dry run") {
			t.Fatalf("dry-run output: %s", out)
		}
		if !strings.Contains(out, "launchctl") {
			t.Fatalf("enable command missing: %s", out)
		}
		if _, err := os.Stat(filepath.Join(dir, ".rtmx", "loop")); !os.IsNotExist(err) {
			t.Fatal("dry-run wrote a unit")
		}
	})

	t.Run("darwin_unit_not_enabled", func(t *testing.T) {
		loopInstallGOOS = "darwin"
		called := false
		loopEnable = func(name string, args ...string) error {
			called = true
			return nil
		}
		dir := t.TempDir()
		orig, _ := os.Getwd()
		if err := os.Chdir(dir); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chdir(orig) })
		cmd := newInstallCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"loop", "install"})
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
		body, err := os.ReadFile(filepath.Join(dir, ".rtmx", "loop", "ai.rtmx.loop.plist"))
		if err != nil {
			t.Fatal(err)
		}
		text := string(body)
		if !strings.Contains(text, "rtmx") || !strings.Contains(text, "loop") || !strings.Contains(text, dir) {
			t.Fatalf("plist: %s", text)
		}
		if called {
			t.Fatal("install enabled the unit")
		}
		if strings.Contains(buf.String(), "launchctl") && strings.Contains(buf.String(), "does not load") {
			return
		}
		t.Fatalf("enable docs: %s", buf.String())
	})

	t.Run("linux_unit", func(t *testing.T) {
		loopInstallGOOS = "linux"
		dir := t.TempDir()
		orig, _ := os.Getwd()
		if err := os.Chdir(dir); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chdir(orig) })
		cmd := newInstallCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"loop", "install"})
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
		body, err := os.ReadFile(filepath.Join(dir, ".rtmx", "loop", "rtmx-loop.service"))
		if err != nil {
			t.Fatal(err)
		}
		text := string(body)
		if !strings.Contains(text, "ExecStart=rtmx loop") || !strings.Contains(text, dir) {
			t.Fatalf("service: %s", text)
		}
		if !strings.Contains(buf.String(), "systemctl --user") {
			t.Fatalf("enable command: %s", buf.String())
		}
	})

	t.Run("refuses_overwrite", func(t *testing.T) {
		loopInstallGOOS = "linux"
		dir := t.TempDir()
		orig, _ := os.Getwd()
		if err := os.Chdir(dir); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chdir(orig) })
		cmd := newInstallCmd()
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"loop", "install"})
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
		cmd = newInstallCmd()
		buf.Reset()
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"loop", "install"})
		if err := cmd.Execute(); err == nil {
			t.Fatal("second install should refuse")
		}
		cmd = newInstallCmd()
		buf.Reset()
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{"loop", "install", "--force"})
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
		body, err := os.ReadFile(filepath.Join(dir, ".rtmx", "loop", "rtmx-loop.service"))
		if err != nil || !strings.Contains(string(body), "ExecStart=rtmx loop") {
			t.Fatalf("force rewrite: %v %s", err, body)
		}
	})
}
