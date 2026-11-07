package broker

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/attestation"
)

// Config is broker-owned. The client cannot supply the test command.
type Config struct {
	WorkDir string
	Command []string // argv the broker executes after checkout
	Key     ed25519.PrivateKey
}

// GateResult is the structural quality outcome for a run (REQ-MONO-064e).
type GateResult struct {
	TestsPassed    int
	TestsFailed    int
	TestsSkipped   int
	EmptyBody      bool
	ZeroAssertions bool
}

// Request is what a client may submit. Command from the client is ignored.
type Request struct {
	RequirementID string
	CommitSHA     string
	// ClientCommand is rejected: the broker never executes it (REQ-MONO-069c).
	ClientCommand []string
	// Gate is the structural quality outcome. When Rejected is non-empty, the
	// broker does not sign (REQ-MONO-069c / REQ-MONO-064e).
	Gate GateResult
}

// Result is a signed attestation, or an error explaining refusal.
type Result struct {
	Attestation *attestation.Record
	Output      []byte
}

func gateRejection(g GateResult) string {
	if g.EmptyBody {
		return "empty-body"
	}
	if g.ZeroAssertions {
		return "zero-assertions"
	}
	if g.TestsSkipped > 0 && g.TestsPassed == 0 && g.TestsFailed == 0 {
		return "skipped"
	}
	return ""
}

// Run checks out the submitted commit, runs the configured test command, and
// signs when the structural gate accepts the run.
func (c Config) Run(req Request) (*Result, error) {
	if len(c.Command) == 0 {
		return nil, fmt.Errorf("broker: configured test command is required")
	}
	if strings.TrimSpace(req.CommitSHA) == "" {
		return nil, fmt.Errorf("broker: commit SHA is required")
	}
	if len(req.ClientCommand) > 0 {
		return nil, fmt.Errorf("broker: client-supplied test command is not allowed")
	}
	if reason := gateRejection(req.Gate); reason != "" {
		return nil, fmt.Errorf("broker: structural gate rejected run: %s", reason)
	}

	work := c.WorkDir
	if work == "" {
		return nil, fmt.Errorf("broker: work dir is required")
	}
	if err := checkoutCommit(work, req.CommitSHA); err != nil {
		return nil, err
	}

	argv := append([]string{}, c.Command...)
	cmdExec := exec.Command(argv[0], argv[1:]...)
	cmdExec.Dir = work
	var buf bytes.Buffer
	cmdExec.Stdout = &buf
	cmdExec.Stderr = &buf
	runErr := cmdExec.Run()
	out := buf.Bytes()
	if runErr != nil {
		return nil, fmt.Errorf("broker: configured command failed: %w\n%s", runErr, out)
	}

	rec, err := attestation.Sign(c.Key, attestation.Payload{
		RequirementID: req.RequirementID,
		CommitSHA:     strings.ToLower(strings.TrimSpace(req.CommitSHA)),
		Command:       strings.Join(c.Command, " "),
		OutputHash:    attestation.HashOutput(out),
	})
	if err != nil {
		return nil, err
	}
	return &Result{Attestation: rec, Output: out}, nil
}

func checkoutCommit(dir, sha string) error {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		return fmt.Errorf("broker: work dir is not a git repo: %w", err)
	}
	cmd := exec.Command("git", "-C", dir, "checkout", "--detach", sha)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("broker: checkout %s: %w\n%s", sha, err, out)
	}
	return nil
}
