package broker

import (
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"crypto/ed25519"
	"crypto/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

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
	if err := os.WriteFile(filepath.Join(dir, "ok.sh"), []byte("#!/bin/sh\necho ok\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	run("add", "ok.sh")
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
		Command: []string{"./ok.sh"},
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
	if res.Attestation.Command != "./ok.sh" {
		t.Fatalf("command = %q", res.Attestation.Command)
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
		Command: []string{"./ok.sh"},
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
		Command: []string{"./ok.sh"},
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
