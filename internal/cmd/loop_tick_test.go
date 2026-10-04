package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

func newLoopTestCmd() *cobra.Command {
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
	return root
}

func injectLoopEdge(t *testing.T, id string) {
	t.Helper()
	prevE := loopEdgesOverride
	prevP := openPRSource
	edges := []LoopEdge{{ReqID: id, Source: "pr", PRNumber: 1}}
	loopEdgesOverride = &edges
	if openPRSource == nil {
		openPRSource = staticOpenPRs{}
	}
	t.Cleanup(func() {
		loopEdgesOverride = prevE
		openPRSource = prevP
	})
}

func writeLoopReqMD(t *testing.T, dir, id string, acs int) {
	t.Helper()
	path := filepath.Join(dir, ".rtmx", "requirements", "CLI", id+".md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	b.WriteString("# " + id + "\n\n## Acceptance Criteria\n\n")
	for i := 1; i <= acs; i++ {
		b.WriteString("- [ ] criterion " + string(rune('0'+i)) + "\n")
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runLoopOnceIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if len(args) == 0 {
		args = []string{"loop", "--once", "--json"}
	}
	cmd := newLoopTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("loop: %v\n%s", err, buf.String())
	}
	return buf.String()
}

func TestLoopTickDecomposes(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-019c")
	injectLoopEdge(t, "REQ-MERGED")

	db := testDBHeader +
		"REQ-EX-010,CLI,Commands,Coarse requirement,Pass,mod,TestA,Unit Test,MISSING,HIGH,1,,1.0,,,,,,,\n"
	dir := setupNextTestProject(t, db)
	writeLoopReqMD(t, dir, "REQ-EX-010", 6)

	gitEnv := []string{
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@example.com",
	}
	git := func(args ...string) string {
		t.Helper()
		c := exec.Command("git", args...)
		c.Dir = dir
		c.Env = append(os.Environ(), gitEnv...)
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-b", "main")
	git("add", "-A")
	git("commit", "-m", "init")
	head := git("rev-parse", "HEAD")
	branches := git("branch")

	out := runLoopOnceIn(t, dir)
	var plan loopPlan
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &plan); err != nil {
		t.Fatalf("plan json: %v\n%s", err, out)
	}
	if plan.ReqID != "REQ-EX-010" && !strings.HasPrefix(plan.ReqID, "REQ-EX-010") {
		t.Fatalf("plan req_id: %+v", plan)
	}
	if !plan.Decomposed || len(plan.Children) < 6 {
		t.Fatalf("expected children, got %+v", plan)
	}
	if plan.PRPolicy != "one_pr_per_requirement" || plan.CommitPolicy != "one_commit_per_ac" {
		t.Fatalf("policies: %+v", plan)
	}
	if !strings.Contains(plan.Note, "parent is not the implementation target") {
		t.Fatalf("parent should not be the implementation target: %+v", plan)
	}
	if plan.ImplementationTarget == "REQ-EX-010" {
		t.Fatalf("parent named as implementation target: %+v", plan)
	}
	claims, err := filepath.Glob(filepath.Join(dir, ".rtmx", "claims", "*.json"))
	if err != nil || len(claims) != 1 {
		t.Fatalf("expected one claim, got %v %v", claims, err)
	}
	if git("rev-parse", "HEAD") != head {
		t.Fatal("tick created a commit")
	}
	if git("branch") != branches {
		t.Fatal("tick created a branch")
	}
}

func TestLoopTickAtomic(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-019c")
	injectLoopEdge(t, "REQ-MERGED")
	openPRSource = staticOpenPRs{}

	db := testDBHeader +
		"REQ-EX-011,CLI,Commands,Atomic requirement,Pass,mod,TestA,Unit Test,MISSING,HIGH,1,,1.0,,,,,,,\n"
	dir := setupNextTestProject(t, db)
	writeLoopReqMD(t, dir, "REQ-EX-011", 2)

	out := runLoopOnceIn(t, dir)
	var plan loopPlan
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &plan); err != nil {
		t.Fatalf("plan: %v\n%s", err, out)
	}
	if plan.ReqID != "REQ-EX-011" || plan.Decomposed || !plan.Atomic {
		t.Fatalf("atomic plan: %+v", plan)
	}
	if plan.PRTarget != "REQ-EX-011" || plan.ImplementationTarget != "REQ-EX-011" {
		t.Fatalf("pr target: %+v", plan)
	}
	if !strings.Contains(plan.Note, "already atomic") {
		t.Fatalf("note: %+v", plan)
	}
	if len(plan.Children) != 0 {
		t.Fatalf("atomic children: %+v", plan.Children)
	}
}

func TestLoopTickIdleLeavesClaims(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-019c")
	injectLoopEdge(t, "REQ-MERGED")
	openPRSource = staticOpenPRs{}

	db := testDBHeader +
		"REQ-EX-012,CLI,Commands,Done,Pass,mod,TestA,Unit Test,COMPLETE,HIGH,1,,1.0,,,,,,,\n"
	dir := setupNextTestProject(t, db)
	claims := filepath.Join(dir, ".rtmx", "claims")
	if err := os.MkdirAll(claims, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(claims, "keep.txt")
	if err := os.WriteFile(marker, []byte("same"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := runLoopOnceIn(t, dir)
	if strings.TrimSpace(out) != `{"idle":true}` {
		t.Fatalf("idle plan: %s", out)
	}
	body, err := os.ReadFile(marker)
	if err != nil || string(body) != "same" {
		t.Fatalf("claim store changed: %v %s", err, body)
	}
	got, _ := filepath.Glob(filepath.Join(claims, "*.json"))
	if len(got) != 0 {
		t.Fatalf("idle created claims: %v", got)
	}
}

func TestLoopTickSkipsOpenPR(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-019c")
	injectLoopEdge(t, "REQ-MERGED")
	openPRSource = staticOpenPRs{{Title: "feat(REQ-EX-013): open", Body: ""}}

	db := testDBHeader +
		"REQ-EX-013,CLI,Commands,Open PR,Pass,mod,TestA,Unit Test,MISSING,P0,1,,1.0,,,,,,,\n" +
		"REQ-EX-014,CLI,Commands,Free,Pass,mod,TestB,Unit Test,MISSING,MEDIUM,1,,1.0,,,,,,,\n"
	dir := setupNextTestProject(t, db)
	writeLoopReqMD(t, dir, "REQ-EX-013", 2)
	writeLoopReqMD(t, dir, "REQ-EX-014", 2)

	out := runLoopOnceIn(t, dir)
	var plan loopPlan
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &plan); err != nil {
		t.Fatalf("plan: %v\n%s", err, out)
	}
	if plan.ReqID != "REQ-EX-014" {
		t.Fatalf("expected skip of open PR, got %+v", plan)
	}
}
