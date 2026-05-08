// Package orchestration provides multi-agent coordination for RTMX,
// including atomic requirement claiming and release.
package orchestration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Claim represents an active claim on a requirement by an agent.
type Claim struct {
	ReqID     string    `json:"req_id"`
	AgentID   string    `json:"agent_id"`
	ClaimedAt time.Time `json:"claimed_at"`
}

// ClaimStore manages file-based claims in a directory.
// Each claim is a separate file: .rtmx/claims/{req_id}.json
// Atomicity is provided by O_CREATE|O_EXCL on file creation.
type ClaimStore struct {
	dir string
}

// NewClaimStore creates a ClaimStore backed by the given directory.
// The directory is created if it does not exist.
func NewClaimStore(dir string) (*ClaimStore, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create claims directory: %w", err)
	}
	return &ClaimStore{dir: dir}, nil
}

// Claim atomically claims a requirement for an agent.
// Returns ErrAlreadyClaimed if the requirement is already claimed.
func (s *ClaimStore) Claim(reqID, agentID string) (*Claim, error) {
	c := &Claim{
		ReqID:     reqID,
		AgentID:   agentID,
		ClaimedAt: time.Now().UTC(),
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal claim: %w", err)
	}

	path := s.claimPath(reqID)

	// O_CREATE|O_EXCL ensures atomic creation -- fails if file exists.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			// Read who holds it for a useful error message
			existing, readErr := s.Get(reqID)
			if readErr == nil && existing != nil {
				return nil, &AlreadyClaimedError{
					ReqID:   reqID,
					HeldBy:  existing.AgentID,
					HeldSince: existing.ClaimedAt,
				}
			}
			return nil, &AlreadyClaimedError{ReqID: reqID}
		}
		return nil, fmt.Errorf("failed to create claim file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.Write(data); err != nil {
		// Clean up on write failure
		_ = os.Remove(path)
		return nil, fmt.Errorf("failed to write claim: %w", err)
	}

	return c, nil
}

// Release removes a claim. Only the agent that holds the claim can release it.
// Returns ErrNotClaimed if no claim exists, ErrNotOwner if another agent holds it.
func (s *ClaimStore) Release(reqID, agentID string) error {
	existing, err := s.Get(reqID)
	if err != nil {
		return err
	}
	if existing == nil {
		return &NotClaimedError{ReqID: reqID}
	}
	if existing.AgentID != agentID {
		return &NotOwnerError{
			ReqID:  reqID,
			Owner:  existing.AgentID,
			Caller: agentID,
		}
	}

	if err := os.Remove(s.claimPath(reqID)); err != nil {
		return fmt.Errorf("failed to remove claim file: %w", err)
	}
	return nil
}

// ForceRelease removes a claim regardless of owner. Use for admin cleanup.
func (s *ClaimStore) ForceRelease(reqID string) error {
	path := s.claimPath(reqID)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return &NotClaimedError{ReqID: reqID}
		}
		return fmt.Errorf("failed to remove claim file: %w", err)
	}
	return nil
}

// Get returns the claim for a requirement, or nil if unclaimed.
func (s *ClaimStore) Get(reqID string) (*Claim, error) {
	data, err := os.ReadFile(s.claimPath(reqID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read claim: %w", err)
	}

	var c Claim
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("failed to parse claim: %w", err)
	}
	return &c, nil
}

// List returns all active claims.
func (s *ClaimStore) List() ([]*Claim, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read claims directory: %w", err)
	}

	var claims []*Claim
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, entry.Name()))
		if err != nil {
			continue
		}
		var c Claim
		if err := json.Unmarshal(data, &c); err != nil {
			continue
		}
		claims = append(claims, &c)
	}
	return claims, nil
}

func (s *ClaimStore) claimPath(reqID string) string {
	return filepath.Join(s.dir, reqID+".json")
}

// AlreadyClaimedError is returned when attempting to claim an already-claimed requirement.
type AlreadyClaimedError struct {
	ReqID     string
	HeldBy    string
	HeldSince time.Time
}

func (e *AlreadyClaimedError) Error() string {
	if e.HeldBy != "" {
		return fmt.Sprintf("%s already claimed by %s since %s",
			e.ReqID, e.HeldBy, e.HeldSince.Format(time.RFC3339))
	}
	return fmt.Sprintf("%s already claimed", e.ReqID)
}

// NotClaimedError is returned when releasing an unclaimed requirement.
type NotClaimedError struct {
	ReqID string
}

func (e *NotClaimedError) Error() string {
	return fmt.Sprintf("%s is not claimed", e.ReqID)
}

// NotOwnerError is returned when an agent tries to release another agent's claim.
type NotOwnerError struct {
	ReqID  string
	Owner  string
	Caller string
}

func (e *NotOwnerError) Error() string {
	return fmt.Sprintf("%s is claimed by %s, not %s", e.ReqID, e.Owner, e.Caller)
}
