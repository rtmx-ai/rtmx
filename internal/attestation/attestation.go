// Package attestation signs and verifies verification runs (REQ-MONO-069a).
package attestation

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Record is one verification-attestation-v1 object.
type Record struct {
	RequirementID string `json:"requirement_id"`
	CommitSHA     string `json:"commit_sha"`
	Command       string `json:"command"`
	OutputHash    string `json:"output_hash"`
	Signature     string `json:"signature"`
}

// Payload is the signed bytes: requirement id, commit SHA, command, output hash.
type Payload struct {
	RequirementID string `json:"requirement_id"`
	CommitSHA     string `json:"commit_sha"`
	Command       string `json:"command"`
	OutputHash    string `json:"output_hash"`
}

var (
	errInvalidKey = errors.New("attestation: invalid key")
	errVerify     = errors.New("attestation: signature verification failed")
)

// CanonicalJSON returns the deterministic JSON of the signed fields.
func CanonicalJSON(p Payload) ([]byte, error) {
	p.RequirementID = strings.TrimSpace(p.RequirementID)
	p.CommitSHA = strings.ToLower(strings.TrimSpace(p.CommitSHA))
	p.Command = strings.TrimSpace(p.Command)
	p.OutputHash = strings.ToLower(strings.TrimSpace(p.OutputHash))
	if p.RequirementID == "" || p.CommitSHA == "" || p.Command == "" || p.OutputHash == "" {
		return nil, fmt.Errorf("attestation: incomplete payload")
	}
	return json.Marshal(p)
}

// HashOutput returns the lowercase hex SHA-256 of output.
func HashOutput(output []byte) string {
	sum := sha256.Sum256(output)
	return hex.EncodeToString(sum[:])
}

// Sign returns a Record whose signature covers requirement id, commit SHA,
// configured command, and output hash (REQ-MONO-069a AC1).
func Sign(priv ed25519.PrivateKey, p Payload) (*Record, error) {
	if len(priv) != ed25519.PrivateKeySize {
		return nil, errInvalidKey
	}
	body, err := CanonicalJSON(p)
	if err != nil {
		return nil, err
	}
	sig := ed25519.Sign(priv, body)
	return &Record{
		RequirementID: p.RequirementID,
		CommitSHA:     strings.ToLower(strings.TrimSpace(p.CommitSHA)),
		Command:       strings.TrimSpace(p.Command),
		OutputHash:    strings.ToLower(strings.TrimSpace(p.OutputHash)),
		Signature:     base64.StdEncoding.EncodeToString(sig),
	}, nil
}

// RevisionEvent is the forge revision metadata used to pin a verify run
// (REQ-MONO-066e). Only CommitSHA is required for binding.
type RevisionEvent struct {
	Forge     string
	Ref       string
	CommitSHA string
	Paths     []string
}

// SignForRevision signs a verify run bound to the commit SHA named by the
// revision event, even when local HEAD differs (REQ-MONO-066e).
func SignForRevision(priv ed25519.PrivateKey, requirementID, command, outputHash string, event RevisionEvent) (*Record, error) {
	return Sign(priv, Payload{
		RequirementID: requirementID,
		CommitSHA:     event.CommitSHA,
		Command:       command,
		OutputHash:    outputHash,
	})
}

// Verify checks the signature with the public key only.
func Verify(pub ed25519.PublicKey, rec Record) error {
	if len(pub) != ed25519.PublicKeySize {
		return errInvalidKey
	}
	sig, err := base64.StdEncoding.DecodeString(rec.Signature)
	if err != nil {
		return fmt.Errorf("attestation: decode signature: %w", err)
	}
	body, err := CanonicalJSON(Payload{
		RequirementID: rec.RequirementID,
		CommitSHA:     rec.CommitSHA,
		Command:       rec.Command,
		OutputHash:    rec.OutputHash,
	})
	if err != nil {
		return err
	}
	if !ed25519.Verify(pub, body, sig) {
		return errVerify
	}
	return nil
}
