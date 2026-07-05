package orchestration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
