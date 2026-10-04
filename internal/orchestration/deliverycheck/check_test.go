package deliverycheck

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckGoodAndBad(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		c := exec.Command("git", args...)
		c.Dir = dir
		c.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@e.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@e.com",
		)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "f"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "f")
	git("commit", "-m", "init")
	git("checkout", "-b", "feat")
	if err := os.WriteFile(filepath.Join(dir, "g"), []byte("y\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "g")
	git("commit", "-m", "REQ-T-1 AC-1: good")

	ok, err := Check(Input{Dir: dir, Base: "main", ReqID: "REQ-T-1"})
	if err != nil || !ok.OK {
		t.Fatalf("good: %+v %v", ok, err)
	}

	bad, err := Check(Input{
		Dir: dir, Base: "main",
		Title: "REQ-T-1 and REQ-T-2", Body: "multi",
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(bad.Errors, " ")
	if bad.OK || !strings.Contains(joined, "exactly one REQ") {
		t.Fatalf("expected multi-REQ error: %+v", bad)
	}
}
