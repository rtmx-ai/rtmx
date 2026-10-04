package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/orchestration/deliverycheck"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	c.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func setupDeliveryRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitIn(t, dir, "init", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "add", "README")
	gitIn(t, dir, "commit", "-m", "init")
	return dir
}

func newDeliveryCheckTestCmd() *cobra.Command {
	root := &cobra.Command{Use: "rtmx", SilenceUsage: true, SilenceErrors: true}
	var req, base, title, body string
	var strict, jsonOut bool
	c := &cobra.Command{
		Use: "delivery-check",
		RunE: func(cmd *cobra.Command, args []string) error {
			deliveryCheckReq = req
			deliveryCheckBase = base
			deliveryCheckStrict = strict
			deliveryCheckJSON = jsonOut
			deliveryCheckTitle = title
			deliveryCheckBody = body
			return runDeliveryCheck(cmd, args)
		},
	}
	c.Flags().StringVar(&req, "req", "", "")
	c.Flags().StringVar(&base, "base", "main", "")
	c.Flags().BoolVar(&strict, "strict", false, "")
	c.Flags().BoolVar(&jsonOut, "json", false, "")
	c.Flags().StringVar(&title, "title", "", "")
	c.Flags().StringVar(&body, "body", "", "")
	root.AddCommand(c)
	return root
}

func runDeliveryCheckIn(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()
	buf := new(bytes.Buffer)
	cmd := newDeliveryCheckTestCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestDeliveryCheck(t *testing.T) {
	rtmx.Req(t, "REQ-ORCH-023a")

	t.Run("good_commits_ok", func(t *testing.T) {
		dir := setupDeliveryRepo(t)
		gitIn(t, dir, "checkout", "-b", "feat")
		if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		gitIn(t, dir, "add", "a.txt")
		gitIn(t, dir, "commit", "-m", "REQ-FOO-001 AC-1: add a")
		if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		gitIn(t, dir, "add", "b.txt")
		gitIn(t, dir, "commit", "-m", "REQ-FOO-001 AC-2: add b")

		out, err := runDeliveryCheckIn(t, dir, "delivery-check", "--req", "REQ-FOO-001", "--json")
		if err != nil {
			t.Fatalf("unexpected err: %v\n%s", err, out)
		}
		var res deliverycheck.Result
		if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &res); err != nil {
			t.Fatalf("json: %v\n%s", err, out)
		}
		if !res.OK || len(res.Commits) != 2 {
			t.Fatalf("want ok with 2 commits: %+v", res)
		}
	})

	t.Run("missing_ac_warns_exit_0", func(t *testing.T) {
		dir := setupDeliveryRepo(t)
		gitIn(t, dir, "checkout", "-b", "feat")
		if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		gitIn(t, dir, "add", "a.txt")
		gitIn(t, dir, "commit", "-m", "REQ-FOO-001: stuff without ac")

		out, err := runDeliveryCheckIn(t, dir, "delivery-check", "--req", "REQ-FOO-001")
		if err != nil {
			t.Fatalf("default should warn exit 0: %v\n%s", err, out)
		}
		if !strings.Contains(out, "warnings") && !strings.Contains(out, "missing AC") {
			t.Fatalf("expected warnings:\n%s", out)
		}
	})

	t.Run("strict_fails_multi_req", func(t *testing.T) {
		dir := setupDeliveryRepo(t)
		gitIn(t, dir, "checkout", "-b", "feat")
		if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		gitIn(t, dir, "add", "a.txt")
		gitIn(t, dir, "commit", "-m", "REQ-FOO-001 AC-1: a")

		out, err := runDeliveryCheckIn(t, dir, "delivery-check", "--strict",
			"--title", "REQ-FOO-001 and REQ-BAR-002", "--body", "both")
		if err == nil {
			t.Fatalf("expected strict failure, got:\n%s", out)
		}
		if ee, ok := err.(*ExitError); !ok || ee.Code != 1 {
			t.Fatalf("want ExitError 1, got %v", err)
		}
	})

	t.Run("package_check_unit", func(t *testing.T) {
		dir := setupDeliveryRepo(t)
		gitIn(t, dir, "checkout", "-b", "feat")
		if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		gitIn(t, dir, "add", "a.txt")
		gitIn(t, dir, "commit", "-m", "REQ-Z-1 AC1 do it")
		res, err := deliverycheck.Check(deliverycheck.Input{Dir: dir, Base: "main", ReqID: "REQ-Z-1"})
		if err != nil || !res.OK {
			t.Fatalf("check: %+v %v", res, err)
		}
	})
}
