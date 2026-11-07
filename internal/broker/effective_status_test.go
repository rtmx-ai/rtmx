package broker

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/attestation"
	"github.com/rtmx-ai/rtmx/internal/database"
)

func TestStatusIgnoresCompleteWithoutValidAttestation(t *testing.T) {
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if got := EffectiveStatus(database.StatusComplete, nil, pub); got != database.StatusPartial {
		t.Fatalf("missing attestation: got %s, want PARTIAL", got)
	}
	bad := &attestation.Record{
		RequirementID: "REQ-MONO-069f",
		CommitSHA:     "0123456789abcdef0123456789abcdef01234567",
		Command:       "go test",
		OutputHash:    attestation.HashOutput([]byte("x")),
		Signature:     "not-a-signature",
	}
	if got := EffectiveStatus(database.StatusComplete, bad, pub); got != database.StatusPartial {
		t.Fatalf("bad signature: got %s, want PARTIAL", got)
	}
}

func TestStatusTreatsCompleteWhenAttestationVerifies(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	rec, err := attestation.Sign(priv, attestation.Payload{
		RequirementID: "REQ-MONO-069f",
		CommitSHA:     "0123456789abcdef0123456789abcdef01234567",
		Command:       "go test ./internal/broker/ -count=1",
		OutputHash:    attestation.HashOutput([]byte("ok")),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := EffectiveStatus(database.StatusComplete, rec, pub); got != database.StatusComplete {
		t.Fatalf("got %s, want COMPLETE", got)
	}
}
