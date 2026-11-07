package broker

import (
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/attestation"
)

func TestCIAndLocalBrokerVerifyUnderSameRules(t *testing.T) {
	rtmx.Req(t, "REQ-MONO-069h")

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	dir, sha := gitRepo(t)
	cmd := []string{"./ok.sh"}
	output := []byte("ok\n")

	local, err := (Config{WorkDir: dir, Command: cmd, Key: priv}).Run(Request{
		RequirementID: "REQ-MONO-069h",
		CommitSHA:     sha,
		Gate:          GateResult{TestsPassed: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	ci, err := (CISigner{Key: priv, Command: cmd}).SignRun("REQ-MONO-069h", sha, output)
	if err != nil {
		t.Fatal(err)
	}
	if err := attestation.Verify(pub, *local.Attestation); err != nil {
		t.Fatalf("local: %v", err)
	}
	if err := attestation.Verify(pub, *ci); err != nil {
		t.Fatalf("ci: %v", err)
	}
}

func TestCIAndLocalAttestationsCoverSHACommandAndOutputHash(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	dir, sha := gitRepo(t)
	cmd := []string{"./ok.sh"}
	local, err := (Config{WorkDir: dir, Command: cmd, Key: priv}).Run(Request{
		RequirementID: "REQ-MONO-069h",
		CommitSHA:     sha,
		Gate:          GateResult{TestsPassed: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	ci, err := (CISigner{Key: priv, Command: cmd}).SignRun("REQ-MONO-069h", sha, local.Output)
	if err != nil {
		t.Fatal(err)
	}
	wantCmd := "./ok.sh"
	wantHash := attestation.HashOutput(local.Output)
	for _, rec := range []*attestation.Record{local.Attestation, ci} {
		if rec.CommitSHA != sha {
			t.Fatalf("sha = %s, want %s", rec.CommitSHA, sha)
		}
		if rec.Command != wantCmd {
			t.Fatalf("command = %q, want %q", rec.Command, wantCmd)
		}
		if rec.OutputHash != wantHash {
			t.Fatalf("output_hash = %s, want %s", rec.OutputHash, wantHash)
		}
	}
}
