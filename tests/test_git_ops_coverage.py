"""Comprehensive tests for rtmx.cli.git_ops module.

This file provides extensive coverage of git operations including
error conditions, edge cases, and mocked subprocess calls.
"""

from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest

from rtmx.cli.git_ops import (
    GitError,
    GitStatus,
    check_gh_installed,
    check_git_installed,
    checkout_branch,
    create_branch,
    create_pr,
    create_rollback_point,
    create_worktree,
    delete_branch,
    generate_branch_name,
    get_git_status,
    is_git_repo,
    list_worktrees,
    pop_stash,
    print_git_status,
    remove_worktree,
    rollback_to_commit,
    stash_changes,
)


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCheckGitInstalled:
    """Tests for check_git_installed function."""

    def test_check_git_installed_success(self):
        """Test git detection when installed."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)
            assert check_git_installed() is True
            mock_run.assert_called_once()
            args = mock_run.call_args[0][0]
            assert args[0] == "git"
            assert "--version" in args

    def test_check_git_installed_not_found(self):
        """Test git detection when not found."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = FileNotFoundError("git not found")
            assert check_git_installed() is False

    def test_check_git_installed_failed_process(self):
        """Test git detection when command fails."""
        import subprocess

        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.CalledProcessError(1, "git")
            assert check_git_installed() is False


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestIsGitRepo:
    """Tests for is_git_repo function."""

    def test_is_git_repo_mocked_true(self):
        """Test is_git_repo with mocked subprocess returning success."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)
            result = is_git_repo(Path("/some/path"))
            assert result is True
            mock_run.assert_called_once()
            args = mock_run.call_args[1]
            assert args["cwd"] == Path("/some/path")

    def test_is_git_repo_mocked_false(self):
        """Test is_git_repo with mocked subprocess returning error."""
        import subprocess

        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.CalledProcessError(128, "git")
            result = is_git_repo(Path("/some/path"))
            assert result is False


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGetGitStatus:
    """Tests for get_git_status function."""

    def test_get_git_status_not_a_repo(self):
        """Test get_git_status raises error for non-repo."""
        import subprocess

        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.CalledProcessError(128, "git")
            with pytest.raises(GitError, match="Not a git repository"):
                get_git_status(Path("/not/a/repo"))

    def test_get_git_status_clean_repo(self):
        """Test get_git_status for clean repository."""
        with patch("rtmx.cli.git_ops.is_git_repo", return_value=True):
            with patch("subprocess.run") as mock_run:
                # Mock status --porcelain (clean)
                status_result = MagicMock()
                status_result.stdout = ""

                # Mock branch --show-current
                branch_result = MagicMock()
                branch_result.stdout = "main\n"

                # Mock rev-parse HEAD
                sha_result = MagicMock()
                sha_result.stdout = "abc123def456\n"

                mock_run.side_effect = [status_result, branch_result, sha_result]

                status = get_git_status(Path("/repo"))

                assert status.is_clean is True
                assert status.branch == "main"
                assert status.commit_sha == "abc123def456"
                assert len(status.uncommitted_files) == 0

    def test_get_git_status_with_uncommitted_files(self):
        """Test get_git_status with uncommitted changes."""
        with patch("rtmx.cli.git_ops.is_git_repo", return_value=True):
            with patch("subprocess.run") as mock_run:
                # Mock status --porcelain (dirty)
                status_result = MagicMock()
                status_result.stdout = " M file1.py\nA  file2.py\n"

                # Mock branch --show-current
                branch_result = MagicMock()
                branch_result.stdout = "feature-branch\n"

                # Mock rev-parse HEAD
                sha_result = MagicMock()
                sha_result.stdout = "deadbeef1234\n"

                mock_run.side_effect = [status_result, branch_result, sha_result]

                status = get_git_status(Path("/repo"))

                assert status.is_clean is False
                assert status.branch == "feature-branch"
                assert status.commit_sha == "deadbeef1234"
                assert len(status.uncommitted_files) == 2
                assert "M file1.py" in status.uncommitted_files
                assert "A  file2.py" in status.uncommitted_files

    def test_get_git_status_detached_head(self):
        """Test get_git_status with detached HEAD."""
        with patch("rtmx.cli.git_ops.is_git_repo", return_value=True):
            with patch("subprocess.run") as mock_run:
                # Mock status --porcelain (clean)
                status_result = MagicMock()
                status_result.stdout = ""

                # Mock branch --show-current (empty for detached)
                branch_result = MagicMock()
                branch_result.stdout = "\n"

                # Mock rev-parse HEAD
                sha_result = MagicMock()
                sha_result.stdout = "1234567890abcdef\n"

                mock_run.side_effect = [status_result, branch_result, sha_result]

                status = get_git_status(Path("/repo"))

                assert status.branch == "(detached HEAD)"
                assert status.commit_sha == "1234567890abcdef"


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCreateRollbackPoint:
    """Tests for create_rollback_point function."""

    def test_create_rollback_point_returns_sha(self):
        """Test create_rollback_point returns current SHA."""
        with patch("rtmx.cli.git_ops.get_git_status") as mock_status:
            mock_status.return_value = GitStatus(
                is_clean=True, branch="main", uncommitted_files=[], commit_sha="abcdef123456"
            )

            sha = create_rollback_point(Path("/repo"))

            assert sha == "abcdef123456"
            mock_status.assert_called_once_with(Path("/repo"))


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGenerateBranchName:
    """Tests for generate_branch_name function."""

    def test_generate_branch_name_default_prefix(self):
        """Test branch name generation with default prefix."""
        name = generate_branch_name()
        assert name.startswith("integration/rtmx-")
        # Should have format: prefix-YYYYMMDD-HHMMSS
        # The prefix is "integration/rtmx" so split gives 3 parts: prefix, date, time
        parts = name.split("-")
        assert len(parts) == 3
        assert parts[0] == "integration/rtmx"

    def test_generate_branch_name_custom_prefix(self):
        """Test branch name generation with custom prefix."""
        name = generate_branch_name(prefix="feature/test")
        assert name.startswith("feature/test-")

    def test_generate_branch_name_format_valid(self):
        """Test branch name has valid timestamp format."""
        with patch("rtmx.cli.git_ops.datetime") as mock_dt:
            mock_now = Mock()
            mock_now.strftime.return_value = "20251205-143022"
            mock_dt.now.return_value = mock_now

            name = generate_branch_name()
            assert name == "integration/rtmx-20251205-143022"


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCreateBranch:
    """Tests for create_branch function."""

    def test_create_branch_success(self):
        """Test successful branch creation."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)

            create_branch(Path("/repo"), "new-branch")

            mock_run.assert_called_once()
            args = mock_run.call_args[0][0]
            assert args == ["git", "checkout", "-b", "new-branch"]
            assert mock_run.call_args[1]["cwd"] == Path("/repo")

    def test_create_branch_failure(self):
        """Test branch creation failure raises GitError."""
        import subprocess

        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.CalledProcessError(
                1, "git", stderr="fatal: branch exists"
            )

            with pytest.raises(GitError, match="Failed to create branch"):
                create_branch(Path("/repo"), "existing-branch")


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCheckoutBranch:
    """Tests for checkout_branch function."""

    def test_checkout_branch_success(self):
        """Test successful branch checkout."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)

            checkout_branch(Path("/repo"), "existing-branch")

            mock_run.assert_called_once()
            args = mock_run.call_args[0][0]
            assert args == ["git", "checkout", "existing-branch"]

    def test_checkout_branch_failure(self):
        """Test checkout failure raises GitError."""
        import subprocess

        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.CalledProcessError(
                1, "git", stderr="error: pathspec 'nonexistent' did not match"
            )

            with pytest.raises(GitError, match="Failed to checkout branch"):
                checkout_branch(Path("/repo"), "nonexistent")


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestDeleteBranch:
    """Tests for delete_branch function."""

    def test_delete_branch_normal(self):
        """Test normal branch deletion."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)

            delete_branch(Path("/repo"), "old-branch", force=False)

            args = mock_run.call_args[0][0]
            assert args == ["git", "branch", "-d", "old-branch"]

    def test_delete_branch_force(self):
        """Test force branch deletion."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)

            delete_branch(Path("/repo"), "old-branch", force=True)

            args = mock_run.call_args[0][0]
            assert args == ["git", "branch", "-D", "old-branch"]

    def test_delete_branch_failure(self):
        """Test delete branch failure raises GitError."""
        import subprocess

        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.CalledProcessError(
                1, "git", stderr="error: branch not found"
            )

            with pytest.raises(GitError, match="Failed to delete branch"):
                delete_branch(Path("/repo"), "missing")


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCreateWorktree:
    """Tests for create_worktree function."""

    def test_create_worktree_with_branch_name(self):
        """Test creating worktree with explicit branch name."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)

            result = create_worktree(
                Path("/repo"), Path("/repo/worktree"), branch_name="test-branch"
            )

            assert result == Path("/repo/worktree")
            args = mock_run.call_args[0][0]
            assert args == ["git", "worktree", "add", "/repo/worktree", "-b", "test-branch"]

    def test_create_worktree_auto_branch_name(self):
        """Test creating worktree with auto-generated branch name."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)
            with patch("rtmx.cli.git_ops.generate_branch_name") as mock_gen:
                mock_gen.return_value = "integration/rtmx-20251205-120000"

                result = create_worktree(Path("/repo"), Path("/repo/worktree"))

                assert result == Path("/repo/worktree")
                mock_gen.assert_called_once()

    def test_create_worktree_failure(self):
        """Test worktree creation failure raises GitError."""
        import subprocess

        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.CalledProcessError(
                1, "git", stderr="fatal: worktree add failed"
            )

            with pytest.raises(GitError, match="Failed to create worktree"):
                create_worktree(Path("/repo"), Path("/repo/worktree"), "test")


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRemoveWorktree:
    """Tests for remove_worktree function."""

    def test_remove_worktree_normal(self):
        """Test normal worktree removal."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)

            remove_worktree(Path("/repo"), Path("/repo/worktree"), force=False)

            args = mock_run.call_args[0][0]
            assert "git" in args
            assert "worktree" in args
            assert "remove" in args
            assert "--force" not in args

    def test_remove_worktree_force(self):
        """Test force worktree removal."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)

            remove_worktree(Path("/repo"), Path("/repo/worktree"), force=True)

            args = mock_run.call_args[0][0]
            assert "--force" in args

    def test_remove_worktree_failure(self):
        """Test worktree removal failure raises GitError."""
        import subprocess

        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.CalledProcessError(
                1, "git", stderr="fatal: worktree not found"
            )

            with pytest.raises(GitError, match="Failed to remove worktree"):
                remove_worktree(Path("/repo"), Path("/repo/worktree"))


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestListWorktrees:
    """Tests for list_worktrees function."""

    def test_list_worktrees_single(self):
        """Test listing single worktree."""
        with patch("subprocess.run") as mock_run:
            result = MagicMock()
            result.stdout = "worktree /repo\nHEAD abc123\nbranch refs/heads/main\n\n"
            mock_run.return_value = result

            worktrees = list_worktrees(Path("/repo"))

            assert len(worktrees) == 1
            assert worktrees[0]["path"] == "/repo"
            assert worktrees[0]["commit"] == "abc123"
            assert worktrees[0]["branch"] == "refs/heads/main"

    def test_list_worktrees_multiple(self):
        """Test listing multiple worktrees."""
        with patch("subprocess.run") as mock_run:
            result = MagicMock()
            result.stdout = """worktree /repo
HEAD abc123
branch refs/heads/main

worktree /repo/wt1
HEAD def456
branch refs/heads/feature

"""
            mock_run.return_value = result

            worktrees = list_worktrees(Path("/repo"))

            assert len(worktrees) == 2
            assert worktrees[0]["path"] == "/repo"
            assert worktrees[1]["path"] == "/repo/wt1"

    def test_list_worktrees_detached(self):
        """Test listing worktree with detached HEAD."""
        with patch("subprocess.run") as mock_run:
            result = MagicMock()
            result.stdout = "worktree /repo\nHEAD abc123\ndetached\n\n"
            mock_run.return_value = result

            worktrees = list_worktrees(Path("/repo"))

            assert len(worktrees) == 1
            assert worktrees[0]["branch"] == "(detached)"

    def test_list_worktrees_failure(self):
        """Test list worktrees failure raises GitError."""
        import subprocess

        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.CalledProcessError(
                1, "git", stderr="fatal: not a git repository"
            )

            with pytest.raises(GitError, match="Failed to list worktrees"):
                list_worktrees(Path("/repo"))


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRollbackToCommit:
    """Tests for rollback_to_commit function."""

    def test_rollback_hard(self):
        """Test hard rollback to commit."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)

            rollback_to_commit(Path("/repo"), "abc123", hard=True)

            args = mock_run.call_args[0][0]
            assert args == ["git", "reset", "--hard", "abc123"]

    def test_rollback_soft(self):
        """Test soft rollback to commit."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)

            rollback_to_commit(Path("/repo"), "abc123", hard=False)

            args = mock_run.call_args[0][0]
            assert args == ["git", "reset", "--soft", "abc123"]

    def test_rollback_failure(self):
        """Test rollback failure raises GitError."""
        import subprocess

        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.CalledProcessError(
                1, "git", stderr="fatal: ambiguous argument"
            )

            with pytest.raises(GitError, match="Failed to rollback"):
                rollback_to_commit(Path("/repo"), "invalid")


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestStashChanges:
    """Tests for stash_changes function."""

    def test_stash_changes_with_changes(self):
        """Test stashing when there are uncommitted changes."""
        with patch("rtmx.cli.git_ops.get_git_status") as mock_status:
            mock_status.return_value = GitStatus(
                is_clean=False, branch="main", uncommitted_files=["file.py"], commit_sha="abc123"
            )
            with patch("subprocess.run") as mock_run:
                mock_run.return_value = MagicMock(returncode=0)

                result = stash_changes(Path("/repo"), message="test stash")

                assert result is True
                args = mock_run.call_args[0][0]
                assert "git" in args
                assert "stash" in args
                assert "-m" in args
                assert "test stash" in args

    def test_stash_changes_clean_repo(self):
        """Test stashing returns False when repo is clean."""
        with patch("rtmx.cli.git_ops.get_git_status") as mock_status:
            mock_status.return_value = GitStatus(
                is_clean=True, branch="main", uncommitted_files=[], commit_sha="abc123"
            )

            result = stash_changes(Path("/repo"))

            assert result is False

    def test_stash_changes_without_message(self):
        """Test stashing without custom message."""
        with patch("rtmx.cli.git_ops.get_git_status") as mock_status:
            mock_status.return_value = GitStatus(
                is_clean=False, branch="main", uncommitted_files=["file.py"], commit_sha="abc123"
            )
            with patch("subprocess.run") as mock_run:
                mock_run.return_value = MagicMock(returncode=0)

                result = stash_changes(Path("/repo"))

                assert result is True
                args = mock_run.call_args[0][0]
                assert "-m" not in args

    def test_stash_changes_failure(self):
        """Test stash failure raises GitError."""
        import subprocess

        with patch("rtmx.cli.git_ops.get_git_status") as mock_status:
            mock_status.return_value = GitStatus(
                is_clean=False, branch="main", uncommitted_files=["file.py"], commit_sha="abc123"
            )
            with patch("subprocess.run") as mock_run:
                mock_run.side_effect = subprocess.CalledProcessError(
                    1, "git", stderr="fatal: stash failed"
                )

                with pytest.raises(GitError, match="Failed to stash changes"):
                    stash_changes(Path("/repo"))


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestPopStash:
    """Tests for pop_stash function."""

    def test_pop_stash_success(self):
        """Test successful stash pop."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)

            pop_stash(Path("/repo"))

            args = mock_run.call_args[0][0]
            assert args == ["git", "stash", "pop"]

    def test_pop_stash_failure(self):
        """Test stash pop failure raises GitError."""
        import subprocess

        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.CalledProcessError(
                1, "git", stderr="error: no stash entries"
            )

            with pytest.raises(GitError, match="Failed to pop stash"):
                pop_stash(Path("/repo"))


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCheckGhInstalled:
    """Tests for check_gh_installed function."""

    def test_check_gh_installed_success(self):
        """Test gh detection when installed."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)

            assert check_gh_installed() is True
            args = mock_run.call_args[0][0]
            assert args[0] == "gh"
            assert "--version" in args

    def test_check_gh_installed_not_found(self):
        """Test gh detection when not found."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = FileNotFoundError("gh not found")

            assert check_gh_installed() is False

    def test_check_gh_installed_failed_process(self):
        """Test gh detection when command fails."""
        import subprocess

        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.CalledProcessError(1, "gh")

            assert check_gh_installed() is False


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCreatePr:
    """Tests for create_pr function."""

    def test_create_pr_gh_not_installed(self):
        """Test create_pr raises error when gh not installed."""
        with patch("rtmx.cli.git_ops.check_gh_installed", return_value=False):
            with pytest.raises(GitError, match="GitHub CLI.*not installed"):
                create_pr(Path("/repo"), "title", "body")

    def test_create_pr_success(self):
        """Test successful PR creation."""
        with patch("rtmx.cli.git_ops.check_gh_installed", return_value=True):
            with patch("subprocess.run") as mock_run:
                result = MagicMock()
                result.stdout = "https://github.com/user/repo/pull/123\n"
                mock_run.return_value = result

                pr_url = create_pr(Path("/repo"), "Test PR", "Description")

                assert pr_url == "https://github.com/user/repo/pull/123"
                args = mock_run.call_args[0][0]
                assert "gh" in args
                assert "pr" in args
                assert "create" in args
                assert "--title" in args
                assert "Test PR" in args
                assert "--body" in args
                assert "Description" in args

    def test_create_pr_with_custom_base(self):
        """Test PR creation with custom base branch."""
        with patch("rtmx.cli.git_ops.check_gh_installed", return_value=True):
            with patch("subprocess.run") as mock_run:
                result = MagicMock()
                result.stdout = "https://github.com/user/repo/pull/124\n"
                mock_run.return_value = result

                create_pr(Path("/repo"), "Title", "Body", base="develop")

                args = mock_run.call_args[0][0]
                assert "--base" in args
                assert "develop" in args

    def test_create_pr_draft(self):
        """Test draft PR creation."""
        with patch("rtmx.cli.git_ops.check_gh_installed", return_value=True):
            with patch("subprocess.run") as mock_run:
                result = MagicMock()
                result.stdout = "https://github.com/user/repo/pull/125\n"
                mock_run.return_value = result

                create_pr(Path("/repo"), "Title", "Body", draft=True)

                args = mock_run.call_args[0][0]
                assert "--draft" in args

    def test_create_pr_failure(self):
        """Test PR creation failure returns None."""
        import subprocess

        with patch("rtmx.cli.git_ops.check_gh_installed", return_value=True):
            with patch("subprocess.run") as mock_run:
                mock_run.side_effect = subprocess.CalledProcessError(
                    1, "gh", stderr="fatal: not a git repository"
                )

                pr_url = create_pr(Path("/repo"), "Title", "Body")

                assert pr_url is None


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestPrintGitStatus:
    """Tests for print_git_status function."""

    def test_print_git_status_clean(self, capsys):
        """Test printing clean git status."""
        status = GitStatus(
            is_clean=True, branch="main", uncommitted_files=[], commit_sha="abcdef1234567890"
        )

        print_git_status(status)

        captured = capsys.readouterr()
        assert "Clean" in captured.out
        assert "main" in captured.out
        assert "abcdef12" in captured.out

    def test_print_git_status_dirty_few_files(self, capsys):
        """Test printing dirty status with few files."""
        status = GitStatus(
            is_clean=False,
            branch="feature",
            uncommitted_files=["M file1.py", "A file2.py"],
            commit_sha="1234567890abcdef",
        )

        print_git_status(status)

        captured = capsys.readouterr()
        assert "2 uncommitted files" in captured.out
        assert "file1.py" in captured.out
        assert "file2.py" in captured.out

    def test_print_git_status_many_files(self, capsys):
        """Test printing status with many uncommitted files."""
        files = [f"M file{i}.py" for i in range(10)]
        status = GitStatus(
            is_clean=False, branch="develop", uncommitted_files=files, commit_sha="fedcba0987654321"
        )

        print_git_status(status)

        captured = capsys.readouterr()
        assert "10 uncommitted files" in captured.out
        # Should show first 5 files
        assert "file0.py" in captured.out
        assert "file4.py" in captured.out
        # Should indicate there are more
        assert "... and 5 more" in captured.out
