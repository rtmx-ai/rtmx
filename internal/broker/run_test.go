package broker

import (
	"crypto/ed25519"
	"crypto/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// okCommand is a cross-platform no-op command checked into the temp git repo.
func okCommand() []string {
	if runtime.GOOS == "windows" {
		return []string{"cmd", "/c", "echo ok"}
	}
	return []string{"./ok.sh"}
}

func gitRepo(t *testing.T) (dir, sha string) {
	t.Helper()
	dir = t.TempDir()
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("init")
	run("config", "user.email", "broker@example.com")
	run("config", "user.name", "broker")
	if runtime.GOOS != "windows" {
		if err := os.WriteFile(filepath.Join(dir, "ok.sh"), []byte("#!/bin/sh\necho ok\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		run("add", "ok.sh")
	} else if err := os.WriteFile(filepath.Join(dir, "README"), []byte("ok\n"), 0o644); err != nil {
		t.Fatal(err)
	} else {
		run("add", "README")
	}
	run("-c", "commit.gpgsign=false", "commit", "-m", "init")
	return dir, run("rev-parse", "HEAD")
}

func TestBrokerChecksOutCommitAndRunsConfiguredCommand(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-069c")

	dir, sha := gitRepo(t)
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	cfg := Config{
		WorkDir: dir,
		Command: okCommand(),
		Key:     priv,
	}
	res, err := cfg.Run(Request{
		RequirementID: "REQ-MONO-069c",
		CommitSHA:     sha,
		Gate:          GateResult{TestsPassed: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Attestation == nil || res.Attestation.CommitSHA != sha {
		t.Fatalf("attestation = %+v", res.Attestation)
	}
	if !strings.Contains(string(res.Output), "ok") {
		t.Fatalf("output = %q", res.Output)
	}
	wantCmd := strings.Join(okCommand(), " ")
	if res.Attestation.Command != wantCmd {
		t.Fatalf("command = %q, want %q", res.Attestation.Command, wantCmd)
	}
}

func TestBrokerRejectsClientSuppliedCommand(t *testing.T) {
	dir, sha := gitRepo(t)
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	cfg := Config{
		WorkDir: dir,
		Command: okCommand(),
		Key:     priv,
	}
	_, err = cfg.Run(Request{
		RequirementID: "REQ-MONO-069c",
		CommitSHA:     sha,
		ClientCommand: []string{"/bin/evil"},
		Gate:          GateResult{TestsPassed: 1},
	})
	if err == nil || !strings.Contains(err.Error(), "client-supplied") {
		t.Fatalf("err = %v, want client-supplied rejection", err)
	}
}

func TestBrokerDoesNotSignWhenStructuralGateRejects(t *testing.T) {
	dir, sha := gitRepo(t)
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	cfg := Config{
		WorkDir: dir,
		Command: okCommand(),
		Key:     priv,
	}
	for _, gate := range []GateResult{
		{TestsSkipped: 1},
		{TestsPassed: 1, EmptyBody: true},
		{TestsPassed: 1, ZeroAssertions: true},
	} {
		res, err := cfg.Run(Request{
			RequirementID: "REQ-MONO-069c",
			CommitSHA:     sha,
			Gate:          gate,
		})
		if err == nil || !strings.Contains(err.Error(), "structural gate") {
			t.Fatalf("gate=%+v err=%v, want structural gate rejection", gate, err)
		}
		if res != nil && res.Attestation != nil {
			t.Fatalf("signed despite gate %+v", gate)
		}
	}
}
