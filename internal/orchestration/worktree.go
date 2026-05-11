package orchestration

import (
	"fmt"
	"os/exec"
	"path/filepath"
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
