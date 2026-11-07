package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rtmx-ai/rtmx/internal/database"
)

// evidenceRecord is one line of .rtmx/evidence.jsonl (verification-evidence-v1).
type evidenceRecord struct {
	RequirementID string `json:"requirement_id"`
	TestID        string `json:"test_id"`
	CommitSHA     string `json:"commit_sha"`
	Result        string `json:"result"`
	Timestamp     string `json:"timestamp"`
}

// recordVerifyEvidence appends evidence for verify results when HEAD resolves
// to a 40-hex commit in root. A directory that is not a git checkout, or a
// git directory with no commit, writes nothing (REQ-MONO-064b).
func recordVerifyEvidence(root string, db *database.Database, results []VerificationResult) (int, error) {
	sha, ok := headCommitSHA(root)
	if !ok {
		return 0, nil
	}
	now := time.Now().UTC().Format(time.RFC3339)
	var recs []evidenceRecord
	for _, r := range results {
		req := db.Get(r.ReqID)
		if req == nil {
			continue
		}
		testID := evidenceTestID(req)
		if testID == "" || StructuralRejection(r) != "" {
			continue
		}
		result := evidenceResult(r)
		if testID == "" || result == "" {
			continue
		}
		recs = append(recs, evidenceRecord{
			RequirementID: req.ReqID,
			TestID:        testID,
			CommitSHA:     sha,
			Result:        result,
			Timestamp:     now,
		})
	}
	if len(recs) == 0 {
		return 0, nil
	}
	path := filepath.Join(root, ".rtmx", "evidence.jsonl")
	if err := appendEvidenceJSONL(path, recs); err != nil {
		return 0, err
	}
	return len(recs), nil
}

func evidenceTestID(req *database.Requirement) string {
	mod := strings.TrimSpace(req.TestModule)
	fn := strings.TrimSpace(req.TestFunction)
	switch {
	case mod != "" && fn != "":
		return mod + "::" + fn
	case mod != "":
		return mod
	default:
		return fn
	}
}

// StructuralRejection reports why a result is not verifying evidence.
// An empty string means the result may be recorded.
func StructuralRejection(r VerificationResult) string {
	if r.EmptyBody {
		return "empty-body"
	}
	if r.ZeroAssertions {
		return "zero-assertions"
	}
	if r.TestsSkipped > 0 && r.TestsPassed == 0 && r.TestsFailed == 0 {
		return "skipped"
	}
	return ""
}

func evidenceResult(r VerificationResult) string {
	if r.TestsFailed > 0 {
		return "fail"
	}
	if r.TestsPassed > 0 {
		return "pass"
	}
	return ""
}

func headCommitSHA(dir string) (string, bool) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--verify", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}
	sha := strings.ToLower(strings.TrimSpace(string(out)))
	if len(sha) != 40 {
		return "", false
	}
	for _, c := range sha {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return "", false
		}
	}
	return sha, true
}

// appendEvidenceJSONL adds records. A later SHA does not rewrite evidence
// already recorded for an earlier SHA (REQ-MONO-064c).
func appendEvidenceJSONL(path string, recs []evidenceRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create evidence dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open evidence log: %w", err)
	}
	defer func() { _ = f.Close() }()
	enc := json.NewEncoder(f)
	for _, rec := range recs {
		if err := enc.Encode(rec); err != nil {
			return fmt.Errorf("write evidence: %w", err)
		}
	}
	return nil
}
