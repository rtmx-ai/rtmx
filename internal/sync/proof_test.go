package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// testKey returns a deterministic test signing key.
func testKey(id string) []byte {
	h := sha256.Sum256([]byte("test-secret-" + id))
	return h[:]
}

// testHash returns a deterministic SHA-256 hash string for test output.
func testHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}

func TestProofOfVerification(t *testing.T) {
	rtmx.Req(t, "REQ-INT-002")

	const (
		reqID     = "REQ-AUTH-001"
		oldStatus = "MISSING"
		newStatus = "COMPLETE"
		verifier  = "dev@rtmx.ai"
		timestamp = "2026-03-28T12:00:00Z"
	)

	signingKey := testKey(verifier)
	outputHash := testHash("PASSED test_auth.py::test_login")

	trustedKeys := map[string][]byte{
		verifier: signingKey,
	}

	t.Run("generate_and_verify_proof_succeeds", func(t *testing.T) {
		proof := generateProofWithTimestamp(reqID, oldStatus, newStatus, outputHash, verifier, signingKey, timestamp)

		if proof.TestOutputHash != outputHash {
			t.Errorf("expected test_output_hash %q, got %q", outputHash, proof.TestOutputHash)
		}
		if proof.VerifierID != verifier {
			t.Errorf("expected verifier_id %q, got %q", verifier, proof.VerifierID)
		}
		if proof.TestTimestamp != timestamp {
			t.Errorf("expected test_timestamp %q, got %q", timestamp, proof.TestTimestamp)
		}
		if proof.Signature == "" {
			t.Fatal("expected non-empty signature")
		}

		op := &StatusChangeOp{
			ReqID:     reqID,
			OldStatus: oldStatus,
			NewStatus: newStatus,
			Timestamp: timestamp,
			Proof:     proof,
		}

		if err := VerifyProof(op, trustedKeys); err != nil {
			t.Errorf("expected valid proof, got error: %v", err)
		}
	})

	t.Run("tampered_proof_rejected", func(t *testing.T) {
		proof := generateProofWithTimestamp(reqID, oldStatus, newStatus, outputHash, verifier, signingKey, timestamp)

		// Tamper with the signature by flipping the last character.
		tampered := proof.Signature[:len(proof.Signature)-1] + "0"
		if tampered == proof.Signature {
			tampered = proof.Signature[:len(proof.Signature)-1] + "1"
		}
		proof.Signature = tampered

		op := &StatusChangeOp{
			ReqID:     reqID,
			OldStatus: oldStatus,
			NewStatus: newStatus,
			Timestamp: timestamp,
			Proof:     proof,
		}

		err := VerifyProof(op, trustedKeys)
		if err != ErrInvalidProof {
			t.Errorf("expected ErrInvalidProof, got: %v", err)
		}
	})

	t.Run("untrusted_key_rejected", func(t *testing.T) {
		unknownVerifier := "attacker@evil.com"
		attackerKey := testKey(unknownVerifier)

		proof := generateProofWithTimestamp(reqID, oldStatus, newStatus, outputHash, unknownVerifier, attackerKey, timestamp)

		op := &StatusChangeOp{
			ReqID:     reqID,
			OldStatus: oldStatus,
			NewStatus: newStatus,
			Timestamp: timestamp,
			Proof:     proof,
		}

		// trustedKeys does NOT contain the attacker's verifier ID.
		err := VerifyProof(op, trustedKeys)
		if err != ErrUntrustedKey {
			t.Errorf("expected ErrUntrustedKey, got: %v", err)
		}
	})

	t.Run("missing_proof_rejected", func(t *testing.T) {
		op := &StatusChangeOp{
			ReqID:     reqID,
			OldStatus: oldStatus,
			NewStatus: newStatus,
			Timestamp: timestamp,
			Proof:     nil,
		}

		err := VerifyProof(op, trustedKeys)
		if err != ErrMissingProof {
			t.Errorf("expected ErrMissingProof, got: %v", err)
		}
	})

	t.Run("proof_with_wrong_test_hash_rejected", func(t *testing.T) {
		proof := generateProofWithTimestamp(reqID, oldStatus, newStatus, outputHash, verifier, signingKey, timestamp)

		op := &StatusChangeOp{
			ReqID:     reqID,
			OldStatus: oldStatus,
			NewStatus: newStatus,
			Timestamp: timestamp,
			Proof:     proof,
		}

		wrongHash := testHash("FAILED test_auth.py::test_login")
		err := VerifyProofWithHash(op, trustedKeys, wrongHash)
		if err != ErrHashMismatch {
			t.Errorf("expected ErrHashMismatch, got: %v", err)
		}
	})

	t.Run("proof_with_correct_test_hash_accepted", func(t *testing.T) {
		proof := generateProofWithTimestamp(reqID, oldStatus, newStatus, outputHash, verifier, signingKey, timestamp)

		op := &StatusChangeOp{
			ReqID:     reqID,
			OldStatus: oldStatus,
			NewStatus: newStatus,
			Timestamp: timestamp,
			Proof:     proof,
		}

		err := VerifyProofWithHash(op, trustedKeys, outputHash)
		if err != nil {
			t.Errorf("expected valid proof with matching hash, got error: %v", err)
		}
	})

	t.Run("tampered_req_id_invalidates_proof", func(t *testing.T) {
		proof := generateProofWithTimestamp(reqID, oldStatus, newStatus, outputHash, verifier, signingKey, timestamp)

		op := &StatusChangeOp{
			ReqID:     "REQ-AUTH-999", // Different req ID
			OldStatus: oldStatus,
			NewStatus: newStatus,
			Timestamp: timestamp,
			Proof:     proof,
		}

		err := VerifyProof(op, trustedKeys)
		if err != ErrInvalidProof {
			t.Errorf("expected ErrInvalidProof when req_id tampered, got: %v", err)
		}
	})

	t.Run("tampered_status_invalidates_proof", func(t *testing.T) {
		proof := generateProofWithTimestamp(reqID, oldStatus, newStatus, outputHash, verifier, signingKey, timestamp)

		op := &StatusChangeOp{
			ReqID:     reqID,
			OldStatus: oldStatus,
			NewStatus: "IN_PROGRESS", // Different new status
			Timestamp: timestamp,
			Proof:     proof,
		}

		err := VerifyProof(op, trustedKeys)
		if err != ErrInvalidProof {
			t.Errorf("expected ErrInvalidProof when status tampered, got: %v", err)
		}
	})

	t.Run("invalid_hex_signature_rejected", func(t *testing.T) {
		proof := generateProofWithTimestamp(reqID, oldStatus, newStatus, outputHash, verifier, signingKey, timestamp)
		proof.Signature = "not-valid-hex!!!"

		op := &StatusChangeOp{
			ReqID:     reqID,
			OldStatus: oldStatus,
			NewStatus: newStatus,
			Timestamp: timestamp,
			Proof:     proof,
		}

		err := VerifyProof(op, trustedKeys)
		if err != ErrInvalidProof {
			t.Errorf("expected ErrInvalidProof for invalid hex, got: %v", err)
		}
	})

	t.Run("GenerateProof_uses_current_time", func(t *testing.T) {
		proof := GenerateProof(reqID, oldStatus, newStatus, outputHash, verifier, signingKey)

		if proof.TestTimestamp == "" {
			t.Fatal("expected non-empty test_timestamp from GenerateProof")
		}
		if proof.Signature == "" {
			t.Fatal("expected non-empty signature from GenerateProof")
		}

		// Verify the generated proof is valid.
		op := &StatusChangeOp{
			ReqID:     reqID,
			OldStatus: oldStatus,
			NewStatus: newStatus,
			Timestamp: proof.TestTimestamp,
			Proof:     proof,
		}

		if err := VerifyProof(op, trustedKeys); err != nil {
			t.Errorf("expected GenerateProof output to verify, got: %v", err)
		}
	})

	t.Run("proof_is_portable_across_nodes", func(t *testing.T) {
		// Simulate: Node A generates a proof, Node B verifies it with the same trusted keys.
		// This validates acceptance criterion 4: proofs are portable.
		nodeAKey := signingKey
		proof := generateProofWithTimestamp(reqID, oldStatus, newStatus, outputHash, verifier, nodeAKey, timestamp)

		op := &StatusChangeOp{
			ReqID:     reqID,
			OldStatus: oldStatus,
			NewStatus: newStatus,
			Timestamp: timestamp,
			Proof:     proof,
		}

		// Node B has the same trusted keys (shared secret distributed out-of-band).
		nodeBTrustedKeys := map[string][]byte{
			verifier: signingKey,
		}

		if err := VerifyProof(op, nodeBTrustedKeys); err != nil {
			t.Errorf("proof should be portable across nodes, got: %v", err)
		}
	})
}
