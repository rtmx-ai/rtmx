package broker

import (
	"crypto/ed25519"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/attestation"
)

func TestPrivateSigningKeyReadableOnlyByVerifierAccount(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "verify.key")
	if err := os.WriteFile(path, []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsurePrivateKeyMode600(path); err != nil {
		t.Fatal(err)
	}
	readable, err := PrivateKeyReadableByOthers(path)
	if err != nil {
		t.Fatal(err)
	}
	if readable {
		t.Fatal("private key must not be readable by group/other")
	}
	acl := DefaultWindowsKeyACL(`DESKTOP\developer`)
	if !acl.GrantsVerifierDeniesDeveloper(`DESKTOP\developer`) {
		t.Fatal("Windows ACL must grant verifier and deny developer")
	}
}

func TestDeveloperVerifiesWithPublicKeyCannotReadPrivateKey(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "verify.key")
	if err := os.WriteFile(keyPath, priv, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := EnsurePrivateKeyMode600(keyPath); err != nil {
		t.Fatal(err)
	}
	readable, err := PrivateKeyReadableByOthers(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	if readable {
		t.Fatal("developer must not read the private key")
	}
	rec, err := attestation.Sign(priv, attestation.Payload{
		RequirementID: "REQ-MONO-069b",
		CommitSHA:     "0123456789abcdef0123456789abcdef01234567",
		Command:       "go test ./internal/broker/ -count=1",
		OutputHash:    attestation.HashOutput([]byte("ok")),
	})
	if err != nil {
		t.Fatal(err)
	}
	// Developer verifies with the public key only.
	if err := attestation.Verify(pub, *rec); err != nil {
		t.Fatal(err)
	}
}
