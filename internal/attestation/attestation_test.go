package attestation

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSignCoversRequirementCommitCommandAndOutputHash(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	p := Payload{
		RequirementID: "REQ-MONO-069a",
		CommitSHA:     "0123456789abcdef0123456789abcdef01234567",
		Command:       "go test ./internal/attestation/ -count=1",
		OutputHash:    HashOutput([]byte("ok")),
	}
	rec, err := Sign(priv, p)
	if err != nil {
		t.Fatal(err)
	}
	if rec.RequirementID != p.RequirementID || rec.CommitSHA != p.CommitSHA ||
		rec.Command != p.Command || rec.OutputHash != p.OutputHash || rec.Signature == "" {
		t.Fatalf("record = %+v", rec)
	}
	body, err := CanonicalJSON(p)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]string
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"requirement_id", "commit_sha", "command", "output_hash"} {
		if decoded[key] == "" {
			t.Fatalf("canonical payload missing %s", key)
		}
	}
	if err := Verify(pub, *rec); err != nil {
		t.Fatal(err)
	}
}

func loadContract(t *testing.T) map[string]any {
	t.Helper()
	// Package dir is rtmx/internal/attestation; monorepo root is three levels up.
	candidates := []string{
		filepath.Join("..", "..", "..", "system", "contracts", "verification-attestation-v1.json"),
		filepath.Join("system", "contracts", "verification-attestation-v1.json"),
	}
	var raw []byte
	var err error
	for _, path := range candidates {
		raw, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("read contract: %v", err)
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	return schema
}

func TestSignForRevisionBindsEventSHA(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	eventSHA := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	rec, err := SignForRevision(
		priv,
		"REQ-MONO-066e",
		"go test ./internal/attestation/ -count=1",
		HashOutput([]byte("ok")),
		RevisionEvent{Forge: "gitea", Ref: "refs/heads/main", CommitSHA: eventSHA},
	)
	if err != nil {
		t.Fatal(err)
	}
	if rec.CommitSHA != eventSHA {
		t.Fatalf("bound sha = %s, want %s", rec.CommitSHA, eventSHA)
	}
	if err := Verify(pub, *rec); err != nil {
		t.Fatal(err)
	}
}

func TestSignForRevisionIgnoresLocalHEAD(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	eventSHA := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	localHEAD := "cccccccccccccccccccccccccccccccccccccccc"
	if eventSHA == localHEAD {
		t.Fatal("fixture shas must differ")
	}
	rec, err := SignForRevision(
		priv,
		"REQ-MONO-066e",
		"go test ./internal/attestation/ -count=1",
		HashOutput([]byte("ok")),
		RevisionEvent{Forge: "github", Ref: "refs/heads/main", CommitSHA: eventSHA},
	)
	if err != nil {
		t.Fatal(err)
	}
	// Local HEAD is intentionally different and is not an argument to SignForRevision.
	_ = localHEAD
	if rec.CommitSHA != eventSHA {
		t.Fatalf("bound sha = %s, want event sha %s (not local HEAD %s)", rec.CommitSHA, eventSHA, localHEAD)
	}
	if rec.CommitSHA == localHEAD {
		t.Fatal("attestation incorrectly bound to local HEAD")
	}
}

func TestVerifyUsesPublicKeyOnly(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	p := Payload{
		RequirementID: "REQ-MONO-069a",
		CommitSHA:     "0123456789abcdef0123456789abcdef01234567",
		Command:       "go test ./internal/attestation/ -count=1",
		OutputHash:    HashOutput([]byte("ok")),
	}
	rec, err := Sign(priv, p)
	if err != nil {
		t.Fatal(err)
	}
	// Discard the private key; only the public key is available to the reader.
	priv = nil
	_ = priv
	if err := Verify(pub, *rec); err != nil {
		t.Fatal(err)
	}
}

func TestMutatedSignedByteFailsVerification(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	p := Payload{
		RequirementID: "REQ-MONO-069a",
		CommitSHA:     "0123456789abcdef0123456789abcdef01234567",
		Command:       "go test ./internal/attestation/ -count=1",
		OutputHash:    HashOutput([]byte("ok")),
	}
	rec, err := Sign(priv, p)
	if err != nil {
		t.Fatal(err)
	}
	mutated := *rec
	mutated.CommitSHA = "ffffffffffffffffffffffffffffffffffffffff"
	if err := Verify(pub, mutated); err == nil {
		t.Fatal("expected verification to fail after mutating commit_sha")
	}
	mutated = *rec
	mutated.Command = mutated.Command + " --mutate"
	if err := Verify(pub, mutated); err == nil {
		t.Fatal("expected verification to fail after mutating command")
	}
	mutated = *rec
	mutated.OutputHash = HashOutput([]byte("different"))
	if err := Verify(pub, mutated); err == nil {
		t.Fatal("expected verification to fail after mutating output_hash")
	}
	mutated = *rec
	mutated.RequirementID = "REQ-MONO-000"
	if err := Verify(pub, mutated); err == nil {
		t.Fatal("expected verification to fail after mutating requirement_id")
	}
}

func TestAttestationContractRequiresSignedFields(t *testing.T) {
	schema := loadContract(t)
	required, ok := schema["required"].([]any)
	if !ok {
		t.Fatal("required missing")
	}
	got := map[string]bool{}
	for _, r := range required {
		got[r.(string)] = true
	}
	for _, want := range []string{"requirement_id", "commit_sha", "command", "output_hash", "signature"} {
		if !got[want] {
			t.Fatalf("missing required %s", want)
		}
	}
}
