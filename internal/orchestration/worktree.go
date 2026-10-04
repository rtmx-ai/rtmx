package orchestration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// gitCommandEnv returns the current environment with the per-invocation git
// variables removed. Git exports GIT_DIR, GIT_INDEX_FILE, GIT_WORK_TREE, etc.
// into hook and alias subprocesses; if they leak into a `git worktree add`
// child, git resolves the new worktree's git dir/index against the PARENT
// repository and fails ("fatal: .git/index: index file open failed"). Stripping
// them lets git compute the worktree's own paths from cmd.Dir, so worktree
// operations work even when rtmx is invoked from a git hook.
func gitCommandEnv() []string {
	drop := map[string]bool{
		"GIT_DIR": true, "GIT_INDEX_FILE": true, "GIT_WORK_TREE": true,
		"GIT_COMMON_DIR": true, "GIT_PREFIX": true,
	}
	env := os.Environ()
	out := env[:0]
	for _, kv := range env {
		if k, _, ok := strings.Cut(kv, "="); ok && drop[k] {
			continue
		}
		out = append(out, kv)
	}
	return out
}

// CreateWorktree creates a git worktree for a work web.
// Returns the worktree path and branch name.
func CreateWorktree(repoRoot string, webID int) (string, string, error) {
	branch := fmt.Sprintf("agent/web-%d", webID)
	wtPath := filepath.Join(repoRoot, ".worktrees", fmt.Sprintf("web-%d", webID))

	cmd := exec.Command("git", "worktree", "add", wtPath, "-b", branch)
	cmd.Dir = repoRoot
	cmd.Env = gitCommandEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("git worktree add failed: %w\n%s", err, string(out))
	}

	return wtPath, branch, nil
}

// RemoveWorktree removes a git worktree.
func RemoveWorktree(repoRoot string, webID int) error {
	wtPath := filepath.Join(repoRoot, ".worktrees", fmt.Sprintf("web-%d", webID))

	cmd := exec.Command("git", "worktree", "remove", wtPath, "--force")
	cmd.Dir = repoRoot
	cmd.Env = gitCommandEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree remove failed: %w\n%s", err, string(out))
	}

	return nil
}

// WorktreeAssignment records which agent holds which web worktree.
type WorktreeAssignment struct {
	WebID          int       `json:"web_id"`
	AgentID        string    `json:"agent_id"`
	Branch         string    `json:"branch,omitempty"`
	Path           string    `json:"path,omitempty"`
	WorktreePath   string    `json:"worktree_path,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	LastHeartbeat  time.Time `json:"last_heartbeat,omitempty"`
}

// WorktreeRegistry persists assignments in a JSON file (REQ-ORCH-016).
type WorktreeRegistry struct {
	path string
	mu   sync.Mutex
}

// NewWorktreeRegistry returns a registry backed by path.
func NewWorktreeRegistry(path string) *WorktreeRegistry {
	return &WorktreeRegistry{path: path}
}

// AssignWorktree creates or replaces the assignment for a web.
func (r *WorktreeRegistry) AssignWorktree(a WorktreeAssignment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	items, err := r.load()
	if err != nil {
		return err
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	a.LastHeartbeat = time.Now().UTC()
	replaced := false
	for i, existing := range items {
		if existing.WebID == a.WebID {
			items[i] = a
			replaced = true
			break
		}
	}
	if !replaced {
		items = append(items, a)
	}
	return r.save(items)
}

// GetWorktreeState returns the assignment for webID, or nil if none.
func (r *WorktreeRegistry) GetWorktreeState(webID int) (*WorktreeAssignment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items, err := r.load()
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].WebID == webID {
			cp := items[i]
			return &cp, nil
		}
	}
	return nil, nil
}

// ListActiveWorktrees returns all stored assignments.
func (r *WorktreeRegistry) ListActiveWorktrees() ([]WorktreeAssignment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.load()
}

// ReleaseWorktree removes the assignment for webID.
func (r *WorktreeRegistry) ReleaseWorktree(webID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	items, err := r.load()
	if err != nil {
		return err
	}
	out := items[:0]
	for _, a := range items {
		if a.WebID != webID {
			out = append(out, a)
		}
	}
	return r.save(out)
}

func (r *WorktreeRegistry) load() ([]WorktreeAssignment, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []WorktreeAssignment{}, nil
		}
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return []WorktreeAssignment{}, nil
	}
	var items []WorktreeAssignment
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("worktree registry: %w", err)
	}
	if items == nil {
		items = []WorktreeAssignment{}
	}
	return items, nil
}

func (r *WorktreeRegistry) save(items []WorktreeAssignment) error {
	if items == nil {
		items = []WorktreeAssignment{}
	}
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "worktrees-*.json")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, r.path)
}
