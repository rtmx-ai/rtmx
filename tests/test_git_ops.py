"""Tests for rtmx.cli.git_ops module."""

import subprocess
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from rtmx.cli.git_ops import (
    GitError,
    GitStatus,
    check_gh_installed,
    create_branch,
    create_rollback_point,
    generate_branch_name,
    get_git_status,
    is_git_repo,
)


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGitError:
    """Tests for GitError exception."""

    def test_git_error_message(self):
        """Test GitError carries message."""
        error = GitError("Test error message")
        assert str(error) == "Test error message"


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGitStatus:
    """Tests for GitStatus dataclass."""

    def test_git_status_clean(self):
        """Test clean status detection."""
        status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123",
        )
        assert status.is_clean is True

    def test_git_status_dirty(self):
        """Test dirty status detection."""
        status = GitStatus(
            is_clean=False,
            branch="main",
            uncommitted_files=["file.py"],
            commit_sha="abc123",
        )
        assert status.is_clean is False
        assert len(status.uncommitted_files) == 1


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestIsGitRepo:
    """Tests for is_git_repo function."""

    def test_is_git_repo_true(self, tmp_path: Path):
        """Test detection of git repository."""
        subprocess.run(["git", "init"], cwd=tmp_path, capture_output=True)
        assert is_git_repo(tmp_path) is True

    def test_is_git_repo_false(self, tmp_path: Path):
        """Test detection of non-git directory."""
        assert is_git_repo(tmp_path) is False


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGetGitStatus:
    """Tests for get_git_status function."""

    def test_get_git_status_clean(self, tmp_path: Path):
        """Test status of clean git repo."""
        # Create a git repo with initial commit
        subprocess.run(["git", "init"], cwd=tmp_path, capture_output=True)
        subprocess.run(
            ["git", "config", "user.email", "test@test.com"],
            cwd=tmp_path,
            capture_output=True,
        )
        subprocess.run(
            ["git", "config", "user.name", "Test"],
            cwd=tmp_path,
            capture_output=True,
        )
        (tmp_path / "file.txt").write_text("content")
        subprocess.run(["git", "add", "."], cwd=tmp_path, capture_output=True)
        subprocess.run(
            ["git", "commit", "-m", "Initial commit"],
            cwd=tmp_path,
            capture_output=True,
        )

        status = get_git_status(tmp_path)
        assert status.is_clean is True
        assert len(status.uncommitted_files) == 0

    def test_get_git_status_with_changes(self, tmp_path: Path):
        """Test status with uncommitted changes."""
        # Create a git repo with initial commit
        subprocess.run(["git", "init"], cwd=tmp_path, capture_output=True)
        subprocess.run(
            ["git", "config", "user.email", "test@test.com"],
            cwd=tmp_path,
            capture_output=True,
        )
        subprocess.run(
            ["git", "config", "user.name", "Test"],
            cwd=tmp_path,
            capture_output=True,
        )
        (tmp_path / "file.txt").write_text("content")
        subprocess.run(["git", "add", "."], cwd=tmp_path, capture_output=True)
        subprocess.run(
            ["git", "commit", "-m", "Initial commit"],
            cwd=tmp_path,
            capture_output=True,
        )

        # Modify file
        (tmp_path / "file.txt").write_text("modified content")

        status = get_git_status(tmp_path)
        assert status.is_clean is False
        # Uncommitted files include status prefix like "M file.txt"
        assert any("file.txt" in f for f in status.uncommitted_files)


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCreateRollbackPoint:
    """Tests for create_rollback_point function."""

    def test_create_rollback_point(self, tmp_path: Path):
        """Test rollback point creation."""
        # Create a git repo with initial commit
        subprocess.run(["git", "init"], cwd=tmp_path, capture_output=True)
        subprocess.run(
            ["git", "config", "user.email", "test@test.com"],
            cwd=tmp_path,
            capture_output=True,
        )
        subprocess.run(
            ["git", "config", "user.name", "Test"],
            cwd=tmp_path,
            capture_output=True,
        )
        (tmp_path / "file.txt").write_text("content")
        subprocess.run(["git", "add", "."], cwd=tmp_path, capture_output=True)
        subprocess.run(
            ["git", "commit", "-m", "Initial commit"],
            cwd=tmp_path,
            capture_output=True,
        )

        rollback_sha = create_rollback_point(tmp_path)
        assert len(rollback_sha) == 40  # Full SHA
        assert all(c in "0123456789abcdef" for c in rollback_sha)


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGenerateBranchName:
    """Tests for generate_branch_name function."""

    def test_generate_branch_name_format(self):
        """Test branch name format."""
        name = generate_branch_name()
        assert name.startswith("integration/rtmx-")
        # Format is prefix-YYYYMMDD-HHMMSS
        parts = name.split("-")
        assert len(parts) >= 3  # prefix-date-time
        # Date part should be 8 digits
        date_part = parts[-2]
        assert len(date_part) == 8
        assert date_part.isdigit()
        # Time part should be 6 digits
        time_part = parts[-1]
        assert len(time_part) == 6
        assert time_part.isdigit()

    def test_generate_branch_name_with_prefix(self):
        """Test branch name with custom prefix."""
        name = generate_branch_name(prefix="feature")
        assert name.startswith("feature-")


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCreateBranch:
    """Tests for create_branch function."""

    def test_create_branch_success(self, tmp_path: Path):
        """Test successful branch creation."""
        # Create a git repo with initial commit
        subprocess.run(["git", "init"], cwd=tmp_path, capture_output=True)
        subprocess.run(
            ["git", "config", "user.email", "test@test.com"],
            cwd=tmp_path,
            capture_output=True,
        )
        subprocess.run(
            ["git", "config", "user.name", "Test"],
            cwd=tmp_path,
            capture_output=True,
        )
        (tmp_path / "file.txt").write_text("content")
        subprocess.run(["git", "add", "."], cwd=tmp_path, capture_output=True)
        subprocess.run(
            ["git", "commit", "-m", "Initial commit"],
            cwd=tmp_path,
            capture_output=True,
        )

        create_branch(tmp_path, "test-branch")

        # Verify branch exists
        result = subprocess.run(
            ["git", "branch", "--list", "test-branch"],
            cwd=tmp_path,
            capture_output=True,
            text=True,
        )
        assert "test-branch" in result.stdout


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCheckGhInstalled:
    """Tests for check_gh_installed function."""

    def test_check_gh_installed_mock_true(self):
        """Test gh detection when installed (mocked)."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)
            assert check_gh_installed() is True

    def test_check_gh_installed_mock_false(self):
        """Test gh detection when not installed (mocked)."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = FileNotFoundError("gh not found")
            assert check_gh_installed() is False
