package broker

import (
	"crypto/ed25519"
	"fmt"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/attestation"
)

// CISigner produces the same attestation shape as the local broker
// (REQ-MONO-069h). Public CI uses this when no local service is installed.
type CISigner struct {
	Key     ed25519.PrivateKey
	Command []string // configured test command (same field as Config.Command)
}

// SignRun signs a CI verify run. The payload fields match Config.Run:
// requirement id, commit SHA, configured command, and output hash.
func (s CISigner) SignRun(requirementID, commitSHA string, output []byte) (*attestation.Record, error) {
	if len(s.Command) == 0 {
		return nil, fmt.Errorf("ci: configured test command is required")
	}
	if strings.TrimSpace(commitSHA) == "" {
		return nil, fmt.Errorf("ci: commit SHA is required")
	}
	return attestation.Sign(s.Key, attestation.Payload{
		RequirementID: requirementID,
		CommitSHA:     strings.ToLower(strings.TrimSpace(commitSHA)),
		Command:       strings.Join(s.Command, " "),
		OutputHash:    attestation.HashOutput(output),
	})
}
