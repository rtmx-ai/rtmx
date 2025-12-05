"""RTMX git operations for integration workflows.

Provides git worktree and branch operations for safe, reversible
integration of rtmx into existing projects.
"""

from __future__ import annotations

import subprocess
import sys
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path

from rtmx.formatting import Colors


class GitError(Exception):
    """Git operation failed."""

    pass


@dataclass
class GitStatus:
    """Git repository status."""

    is_clean: bool
    branch: str
    uncommitted_files: list[str]
    commit_sha: str


def check_git_installed() -> bool:
    """Check if git is available."""
    try:
        subprocess.run(
            ["git", "--version"],
            capture_output=True,
            check=True,
        )
        return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        return False


def is_git_repo(path: Path) -> bool:
    """Check if path is inside a git repository."""
    try:
        subprocess.run(
            ["git", "rev-parse", "--git-dir"],
            cwd=path,
            capture_output=True,
            check=True,
        )
        return True
    except subprocess.CalledProcessError:
        return False


def get_git_status(path: Path) -> GitStatus:
    """Get current git status.

    Args:
        path: Path to git repository

    Returns:
        GitStatus with current state

    Raises:
        GitError: If not a git repository or git command fails
    """
    if not is_git_repo(path):
        raise GitError(f"Not a git repository: {path}")

    # Get uncommitted files
    result = subprocess.run(
        ["git", "status", "--porcelain"],
        cwd=path,
        capture_output=True,
        text=True,
        check=True,
    )
    uncommitted = [line.strip() for line in result.stdout.strip().split("\n") if line.strip()]

    # Get current branch
    result = subprocess.run(
        ["git", "branch", "--show-current"],
        cwd=path,
        capture_output=True,
        text=True,
        check=True,
    )
    branch = result.stdout.strip()

    # If detached HEAD, get commit SHA
    if not branch:
        branch = "(detached HEAD)"

    # Get current commit SHA
    result = subprocess.run(
        ["git", "rev-parse", "HEAD"],
        cwd=path,
        capture_output=True,
        text=True,
        check=True,
    )
    commit_sha = result.stdout.strip()

    return GitStatus(
        is_clean=len(uncommitted) == 0,
        branch=branch,
        uncommitted_files=uncommitted,
        commit_sha=commit_sha,
    )


def create_rollback_point(path: Path) -> str:
    """Create a rollback point (returns current commit SHA).

    Args:
        path: Path to git repository

    Returns:
        Current commit SHA for rollback

    Raises:
        GitError: If git command fails
    """
    status = get_git_status(path)
    return status.commit_sha


def generate_branch_name(prefix: str = "integration/rtmx") -> str:
    """Generate timestamped branch name.

    Args:
        prefix: Branch name prefix

    Returns:
        Branch name like 'integration/rtmx-20251204-143522'
    """
    timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
    return f"{prefix}-{timestamp}"


def create_branch(path: Path, branch_name: str) -> None:
    """Create and checkout a new branch.

    Args:
        path: Path to git repository
        branch_name: Name of branch to create

    Raises:
        GitError: If branch creation fails
    """
    try:
        subprocess.run(
            ["git", "checkout", "-b", branch_name],
            cwd=path,
            capture_output=True,
            text=True,
            check=True,
        )
    except subprocess.CalledProcessError as e:
        raise GitError(f"Failed to create branch '{branch_name}': {e.stderr}") from e


def checkout_branch(path: Path, branch_name: str) -> None:
    """Checkout an existing branch.

    Args:
        path: Path to git repository
        branch_name: Name of branch to checkout

    Raises:
        GitError: If checkout fails
    """
    try:
        subprocess.run(
            ["git", "checkout", branch_name],
            cwd=path,
            capture_output=True,
            text=True,
            check=True,
        )
    except subprocess.CalledProcessError as e:
        raise GitError(f"Failed to checkout branch '{branch_name}': {e.stderr}") from e


def delete_branch(path: Path, branch_name: str, force: bool = False) -> None:
    """Delete a branch.

    Args:
        path: Path to git repository
        branch_name: Name of branch to delete
        force: Force delete even if not merged

    Raises:
        GitError: If branch deletion fails
    """
    flag = "-D" if force else "-d"
    try:
        subprocess.run(
            ["git", "branch", flag, branch_name],
            cwd=path,
            capture_output=True,
            text=True,
            check=True,
        )
    except subprocess.CalledProcessError as e:
        raise GitError(f"Failed to delete branch '{branch_name}': {e.stderr}") from e


def create_worktree(
    path: Path,
    worktree_path: Path,
    branch_name: str | None = None,
) -> Path:
    """Create a git worktree for isolated integration.

    Args:
        path: Path to main git repository
        worktree_path: Path for the new worktree
        branch_name: Optional branch name (auto-generated if None)

    Returns:
        Path to created worktree

    Raises:
        GitError: If worktree creation fails
    """
    if branch_name is None:
        branch_name = generate_branch_name()

    try:
        subprocess.run(
            ["git", "worktree", "add", str(worktree_path), "-b", branch_name],
            cwd=path,
            capture_output=True,
            text=True,
            check=True,
        )
        return worktree_path
    except subprocess.CalledProcessError as e:
        raise GitError(f"Failed to create worktree: {e.stderr}") from e


def remove_worktree(path: Path, worktree_path: Path, force: bool = False) -> None:
    """Remove a git worktree.

    Args:
        path: Path to main git repository
        worktree_path: Path to worktree to remove
        force: Force removal even with uncommitted changes

    Raises:
        GitError: If worktree removal fails
    """
    cmd = ["git", "worktree", "remove", str(worktree_path)]
    if force:
        cmd.append("--force")

    try:
        subprocess.run(
            cmd,
            cwd=path,
            capture_output=True,
            text=True,
            check=True,
        )
    except subprocess.CalledProcessError as e:
        raise GitError(f"Failed to remove worktree: {e.stderr}") from e


def list_worktrees(path: Path) -> list[dict[str, str]]:
    """List all worktrees for a repository.

    Args:
        path: Path to git repository

    Returns:
        List of worktree info dicts with 'path', 'branch', 'commit'

    Raises:
        GitError: If listing fails
    """
    try:
        result = subprocess.run(
            ["git", "worktree", "list", "--porcelain"],
            cwd=path,
            capture_output=True,
            text=True,
            check=True,
        )
    except subprocess.CalledProcessError as e:
        raise GitError(f"Failed to list worktrees: {e.stderr}") from e

    worktrees: list[dict[str, str]] = []
    current: dict[str, str] = {}

    for line in result.stdout.strip().split("\n"):
        if not line:
            if current:
                worktrees.append(current)
                current = {}
        elif line.startswith("worktree "):
            current["path"] = line[9:]
        elif line.startswith("HEAD "):
            current["commit"] = line[5:]
        elif line.startswith("branch "):
            current["branch"] = line[7:]
        elif line == "detached":
            current["branch"] = "(detached)"

    if current:
        worktrees.append(current)

    return worktrees


def rollback_to_commit(path: Path, commit_sha: str, hard: bool = True) -> None:
    """Rollback to a specific commit.

    Args:
        path: Path to git repository
        commit_sha: Commit SHA to rollback to
        hard: If True, use --hard (discards changes)

    Raises:
        GitError: If rollback fails
    """
    mode = "--hard" if hard else "--soft"
    try:
        subprocess.run(
            ["git", "reset", mode, commit_sha],
            cwd=path,
            capture_output=True,
            text=True,
            check=True,
        )
    except subprocess.CalledProcessError as e:
        raise GitError(f"Failed to rollback to {commit_sha}: {e.stderr}") from e


def stash_changes(path: Path, message: str | None = None) -> bool:
    """Stash uncommitted changes.

    Args:
        path: Path to git repository
        message: Optional stash message

    Returns:
        True if changes were stashed, False if nothing to stash

    Raises:
        GitError: If stash fails
    """
    status = get_git_status(path)
    if status.is_clean:
        return False

    cmd = ["git", "stash", "push"]
    if message:
        cmd.extend(["-m", message])

    try:
        subprocess.run(
            cmd,
            cwd=path,
            capture_output=True,
            text=True,
            check=True,
        )
        return True
    except subprocess.CalledProcessError as e:
        raise GitError(f"Failed to stash changes: {e.stderr}") from e


def pop_stash(path: Path) -> None:
    """Pop the most recent stash.

    Args:
        path: Path to git repository

    Raises:
        GitError: If stash pop fails
    """
    try:
        subprocess.run(
            ["git", "stash", "pop"],
            cwd=path,
            capture_output=True,
            text=True,
            check=True,
        )
    except subprocess.CalledProcessError as e:
        raise GitError(f"Failed to pop stash: {e.stderr}") from e


def check_gh_installed() -> bool:
    """Check if GitHub CLI (gh) is available."""
    try:
        subprocess.run(
            ["gh", "--version"],
            capture_output=True,
            check=True,
        )
        return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        return False


def create_pr(
    path: Path,
    title: str,
    body: str,
    base: str = "main",
    draft: bool = False,
) -> str | None:
    """Create a pull request using GitHub CLI.

    Args:
        path: Path to git repository
        title: PR title
        body: PR body/description
        base: Base branch (default: main)
        draft: Create as draft PR

    Returns:
        PR URL if successful, None if failed

    Raises:
        GitError: If gh is not installed
    """
    if not check_gh_installed():
        raise GitError("GitHub CLI (gh) is not installed. Install from https://cli.github.com/")

    cmd = [
        "gh",
        "pr",
        "create",
        "--title",
        title,
        "--body",
        body,
        "--base",
        base,
    ]
    if draft:
        cmd.append("--draft")

    try:
        result = subprocess.run(
            cmd,
            cwd=path,
            capture_output=True,
            text=True,
            check=True,
        )
        # gh pr create outputs the PR URL
        return result.stdout.strip()
    except subprocess.CalledProcessError as e:
        print(f"{Colors.RED}Failed to create PR: {e.stderr}{Colors.RESET}", file=sys.stderr)
        return None


def print_git_status(status: GitStatus) -> None:
    """Print formatted git status."""
    if status.is_clean:
        print(f"{Colors.GREEN}  Git status: Clean{Colors.RESET}")
    else:
        print(
            f"{Colors.YELLOW}  Git status: {len(status.uncommitted_files)} uncommitted files{Colors.RESET}"
        )
        for f in status.uncommitted_files[:5]:
            print(f"    {f}")
        if len(status.uncommitted_files) > 5:
            print(f"    ... and {len(status.uncommitted_files) - 5} more")
    print(f"  Branch: {status.branch}")
    print(f"  Commit: {status.commit_sha[:8]}")
