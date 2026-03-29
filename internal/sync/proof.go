package sync

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// ProofErrors returned by proof validation.
var (
	ErrMissingProof   = errors.New("status change operation missing proof")
	ErrInvalidProof   = errors.New("proof signature is invalid")
	ErrUntrustedKey   = errors.New("proof signed by untrusted verifier")
	ErrHashMismatch   = errors.New("proof test output hash does not match")
)

// TrustPolicy defines which verifier keys are accepted.
type TrustPolicy string

const (
	TrustPolicySelf       TrustPolicy = "self"
	TrustPolicyTeam       TrustPolicy = "team"
	TrustPolicyDelegated  TrustPolicy = "delegated"
	TrustPolicyWebOfTrust TrustPolicy = "web-of-trust"
)

// StatusChangeProof is the cryptographic proof attached to a status change operation.
type StatusChangeProof struct {
	TestOutputHash string `json:"test_output_hash"`
	TestTimestamp  string `json:"test_timestamp"`
	VerifierID     string `json:"verifier_id"`
	Signature      string `json:"signature"`
}

// StatusChangeOp represents a status change operation with its proof.
type StatusChangeOp struct {
	ReqID     string             `json:"req_id"`
	OldStatus string             `json:"old_status"`
	NewStatus string             `json:"new_status"`
	Timestamp string             `json:"timestamp"`
	Proof     *StatusChangeProof `json:"proof,omitempty"`
}

// proofPayload constructs the canonical byte string that gets signed.
// This ensures the signature covers the requirement ID, status transition,
// and test output hash -- binding the proof to a specific verified change.
func proofPayload(reqID, oldStatus, newStatus, testOutputHash, testTimestamp string) []byte {
	canonical := fmt.Sprintf(
		"rtmx-proof:v1\nreq_id:%s\nold_status:%s\nnew_status:%s\ntest_output_hash:%s\ntest_timestamp:%s",
		reqID, oldStatus, newStatus, testOutputHash, testTimestamp,
	)
	return []byte(canonical)
}

// GenerateProof creates a StatusChangeProof using HMAC-SHA256.
//
// The signingKey is the shared secret for the verifier. In a future release
// (REQ-SEC-002), this will be upgraded to Ed25519 asymmetric signing.
func GenerateProof(reqID, oldStatus, newStatus, testOutputHash string, verifierID string, signingKey []byte) *StatusChangeProof {
	now := time.Now().UTC().Format(time.RFC3339)
	payload := proofPayload(reqID, oldStatus, newStatus, testOutputHash, now)

	mac := hmac.New(sha256.New, signingKey)
	mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))

	return &StatusChangeProof{
		TestOutputHash: testOutputHash,
		TestTimestamp:  now,
		VerifierID:     verifierID,
		Signature:      sig,
	}
}

// generateProofWithTimestamp is an internal helper for testing that allows
// specifying the timestamp rather than using time.Now().
func generateProofWithTimestamp(reqID, oldStatus, newStatus, testOutputHash string, verifierID string, signingKey []byte, timestamp string) *StatusChangeProof {
	payload := proofPayload(reqID, oldStatus, newStatus, testOutputHash, timestamp)

	mac := hmac.New(sha256.New, signingKey)
	mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))

	return &StatusChangeProof{
		TestOutputHash: testOutputHash,
		TestTimestamp:  timestamp,
		VerifierID:     verifierID,
		Signature:      sig,
	}
}

// VerifyProof validates a StatusChangeProof against a set of trusted keys.
//
// trustedKeys maps verifier IDs to their HMAC signing keys.
// The proof is verified by recomputing the HMAC and comparing.
//
// Returns nil if the proof is valid, or one of:
//   - ErrMissingProof if proof is nil
//   - ErrUntrustedKey if verifier_id is not in trustedKeys
//   - ErrInvalidProof if the signature does not match
func VerifyProof(op *StatusChangeOp, trustedKeys map[string][]byte) error {
	if op.Proof == nil {
		return ErrMissingProof
	}

	key, ok := trustedKeys[op.Proof.VerifierID]
	if !ok {
		return ErrUntrustedKey
	}

	payload := proofPayload(op.ReqID, op.OldStatus, op.NewStatus, op.Proof.TestOutputHash, op.Proof.TestTimestamp)

	mac := hmac.New(sha256.New, key)
	mac.Write(payload)
	expected := mac.Sum(nil)

	sig, err := hex.DecodeString(op.Proof.Signature)
	if err != nil {
		return ErrInvalidProof
	}

	if !hmac.Equal(sig, expected) {
		return ErrInvalidProof
	}

	return nil
}

// VerifyProofWithHash validates a proof and additionally checks that the
// test output hash in the proof matches an expected hash value. This is
// used when the verifier independently computed the hash of the test output.
func VerifyProofWithHash(op *StatusChangeOp, trustedKeys map[string][]byte, expectedHash string) error {
	if err := VerifyProof(op, trustedKeys); err != nil {
		return err
	}

	if op.Proof.TestOutputHash != expectedHash {
		return ErrHashMismatch
	}

	return nil
}
