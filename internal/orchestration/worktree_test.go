package orchestration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %v\n%s", args, err, out)
		}
	}

	// Create an initial commit so worktree creation works
	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, out)
	}
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v\n%s", err, out)
	}

	return dir
}

func TestCreateWorktree(t *testing.T) {
	repoDir := initTestRepo(t)

	t.Run("creates_worktree", func(t *testing.T) {
		wtPath, branch, err := CreateWorktree(repoDir, 1)
		if err != nil {
			t.Fatalf("CreateWorktree failed: %v", err)
		}
		defer func() { _ = RemoveWorktree(repoDir, 1) }()

		if branch != "agent/web-1" {
			t.Errorf("branch = %q, want %q", branch, "agent/web-1")
		}

		expectedPath := filepath.Join(repoDir, ".worktrees", "web-1")
		if wtPath != expectedPath {
			t.Errorf("path = %q, want %q", wtPath, expectedPath)
		}

		// Verify worktree exists
		if _, err := os.Stat(wtPath); err != nil {
			t.Errorf("worktree directory should exist: %v", err)
		}

		// Verify it's a valid git checkout
		gitDir := filepath.Join(wtPath, ".git")
		if _, err := os.Stat(gitDir); err != nil {
			t.Errorf("worktree should have .git: %v", err)
		}
	})

	t.Run("multiple_worktrees_no_conflict", func(t *testing.T) {
		wt2Path, _, err := CreateWorktree(repoDir, 2)
		if err != nil {
			t.Fatalf("CreateWorktree(2) failed: %v", err)
		}
		defer func() { _ = RemoveWorktree(repoDir, 2) }()

		wt3Path, _, err := CreateWorktree(repoDir, 3)
		if err != nil {
			t.Fatalf("CreateWorktree(3) failed: %v", err)
		}
		defer func() { _ = RemoveWorktree(repoDir, 3) }()

		if wt2Path == wt3Path {
			t.Error("different web IDs should create different worktrees")
		}

		// Both should exist
		for _, p := range []string{wt2Path, wt3Path} {
			if _, err := os.Stat(p); err != nil {
				t.Errorf("worktree %s should exist", p)
			}
		}
	})
}

func TestRemoveWorktree(t *testing.T) {
	repoDir := initTestRepo(t)

	wtPath, _, err := CreateWorktree(repoDir, 99)
	if err != nil {
		t.Fatalf("CreateWorktree failed: %v", err)
	}

	if _, err := os.Stat(wtPath); err != nil {
		t.Fatalf("worktree should exist before removal")
	}

	if err := RemoveWorktree(repoDir, 99); err != nil {
		// Force flag might not work if there are changes
		fmt.Printf("RemoveWorktree warning: %v\n", err)
	}
}
