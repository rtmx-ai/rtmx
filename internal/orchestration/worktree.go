package orchestration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// CreateWorktree creates a git worktree for a work web.
// Returns the worktree path and branch name.
func CreateWorktree(repoRoot string, webID int) (string, string, error) {
	branch := fmt.Sprintf("agent/web-%d", webID)
	wtPath := filepath.Join(repoRoot, ".worktrees", fmt.Sprintf("web-%d", webID))

	cmd := exec.Command("git", "worktree", "add", wtPath, "-b", branch)
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("git worktree add failed: %w\n%s", err, string(out))
	}

	return wtPath, branch, nil
}

// WorktreeAssignment tracks a worktree assigned to an agent for a web.
type WorktreeAssignment struct {
	WebID     int       `json:"web_id"`
	AgentID   string    `json:"agent_id"`
	Branch    string    `json:"branch"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

// WorktreeRegistry persists worktree assignments to disk.
type WorktreeRegistry struct {
	mu   sync.Mutex
	path string
}

// NewWorktreeRegistry creates a registry backed by the given file path.
func NewWorktreeRegistry(path string) *WorktreeRegistry {
	return &WorktreeRegistry{path: path}
}

// AssignWorktree records a worktree assignment atomically.
func (r *WorktreeRegistry) AssignWorktree(assignment WorktreeAssignment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	assignments, err := r.readLocked()
	if err != nil {
		assignments = []WorktreeAssignment{}
	}

	// Replace existing assignment for this web
	found := false
	for i, a := range assignments {
		if a.WebID == assignment.WebID {
			assignments[i] = assignment
			found = true
			break
		}
	}
	if !found {
		assignments = append(assignments, assignment)
	}

	return r.writeLocked(assignments)
}

// ReleaseWorktree removes a worktree assignment.
func (r *WorktreeRegistry) ReleaseWorktree(webID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	assignments, err := r.readLocked()
	if err != nil {
		return nil // nothing to release
	}

	var filtered []WorktreeAssignment
	for _, a := range assignments {
		if a.WebID != webID {
			filtered = append(filtered, a)
		}
	}

	return r.writeLocked(filtered)
}

// ListActiveWorktrees returns all current worktree assignments.
func (r *WorktreeRegistry) ListActiveWorktrees() ([]WorktreeAssignment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.readLocked()
}

func (r *WorktreeRegistry) readLocked() ([]WorktreeAssignment, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var assignments []WorktreeAssignment
	if err := json.Unmarshal(data, &assignments); err != nil {
		return nil, err
	}
	return assignments, nil
}

func (r *WorktreeRegistry) writeLocked(assignments []WorktreeAssignment) error {
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(assignments, "", "  ")
	if err != nil {
		return err
	}
	// Atomic write via temp file
	tmpPath := r.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, r.path)
}

// RemoveWorktree removes a git worktree.
func RemoveWorktree(repoRoot string, webID int) error {
	wtPath := filepath.Join(repoRoot, ".worktrees", fmt.Sprintf("web-%d", webID))

	cmd := exec.Command("git", "worktree", "remove", wtPath, "--force")
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree remove failed: %w\n%s", err, string(out))
	}

	return nil
}
