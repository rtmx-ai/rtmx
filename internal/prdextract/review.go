package prdextract

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/database"
)

// Decision is a human accept or reject of one candidate (REQ-MONO-067b).
type Decision string

const (
	DecisionAccept Decision = "accept"
	DecisionReject Decision = "reject"
)

// ApplyDecision writes an RTM row only when the human accepts the candidate.
// Rejecting writes nothing. Accepted rows are never COMPLETE (REQ-MONO-067d).
func ApplyDecision(dbPath string, c Candidate, d Decision) (string, error) {
	switch d {
	case DecisionReject:
		return "", nil
	case DecisionAccept:
		return writeAccepted(dbPath, c)
	default:
		return "", fmt.Errorf("unknown decision: %q", d)
	}
}

func writeAccepted(dbPath string, c Candidate) (string, error) {
	id := StableID(c)
	db, err := loadOrNew(dbPath)
	if err != nil {
		return "", err
	}
	if existing := db.Get(id); existing != nil {
		return id, nil
	}
	req := database.NewRequirement(id)
	req.RequirementText = c.Statement
	req.Status = database.StatusNotStarted
	// At least one acceptance criterion (REQ-MONO-067c).
	req.Notes = "Acceptance Criteria:\n1. [ ] " + defaultAC(c.Statement)
	if err := db.Add(req); err != nil {
		return "", err
	}
	if err := db.Save(dbPath); err != nil {
		return "", err
	}
	return id, nil
}

// StableID returns a deterministic requirement ID for a candidate (REQ-MONO-067c).
func StableID(c Candidate) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(c.Statement)))
	return "REQ-PRD-" + strings.ToUpper(hex.EncodeToString(sum[:4]))
}

func defaultAC(statement string) string {
	s := strings.TrimSpace(statement)
	if s == "" {
		return "The accepted requirement is satisfied."
	}
	return s
}

func loadOrNew(path string) (*database.Database, error) {
	if _, err := os.Stat(path); err == nil {
		return database.Load(path)
	}
	db := database.NewDatabase()
	db.SetPath(path)
	return db, nil
}
