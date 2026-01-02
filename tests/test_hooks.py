"""Tests for git hook integration (REQ-DX-005).

This module tests the git hook installation and removal functionality.
"""

from __future__ import annotations

import os
import stat
from pathlib import Path

import pytest


@pytest.mark.req("REQ-DX-005")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestHookTemplates:
    """Test hook template generation."""

    def test_precommit_hook_template_exists(self) -> None:
        """Pre-commit hook template should be defined."""
        from rtmx.cli.install import PRE_COMMIT_HOOK_TEMPLATE

        assert PRE_COMMIT_HOOK_TEMPLATE is not None
        assert "#!/bin/sh" in PRE_COMMIT_HOOK_TEMPLATE
        assert "rtmx health --strict" in PRE_COMMIT_HOOK_TEMPLATE

    def test_precommit_hook_template_contains_rtmx_marker(self) -> None:
        """Pre-commit hook should contain RTMX marker for identification."""
        from rtmx.cli.install import PRE_COMMIT_HOOK_TEMPLATE

        assert "RTMX" in PRE_COMMIT_HOOK_TEMPLATE

    def test_precommit_hook_template_fails_on_error(self) -> None:
        """Pre-commit hook should exit 1 on health check failure."""
        from rtmx.cli.install import PRE_COMMIT_HOOK_TEMPLATE

        assert "exit 1" in PRE_COMMIT_HOOK_TEMPLATE

    def test_prepush_hook_template_exists(self) -> None:
        """Pre-push hook template should be defined."""
        from rtmx.cli.install import PRE_PUSH_HOOK_TEMPLATE

        assert PRE_PUSH_HOOK_TEMPLATE is not None
        assert "#!/bin/sh" in PRE_PUSH_HOOK_TEMPLATE

    def test_prepush_hook_template_checks_markers(self) -> None:
        """Pre-push hook should check for test markers."""
        from rtmx.cli.install import PRE_PUSH_HOOK_TEMPLATE

        # Should reference pytest or marker checking
        assert "pytest" in PRE_PUSH_HOOK_TEMPLATE or "marker" in PRE_PUSH_HOOK_TEMPLATE


@pytest.mark.req("REQ-DX-005")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestFindGitDir:
    """Test git directory detection."""

    def test_find_git_dir_in_repo(self, tmp_path: Path) -> None:
        """Should find .git directory in a git repository."""
        from rtmx.cli.install import find_git_dir

        git_dir = tmp_path / ".git"
        git_dir.mkdir()

        result = find_git_dir(tmp_path)
        assert result == git_dir

    def test_find_git_dir_not_found(self, tmp_path: Path) -> None:
        """Should return None when not in a git repository."""
        from rtmx.cli.install import find_git_dir

        result = find_git_dir(tmp_path)
        assert result is None

    def test_find_git_dir_with_hooks_subdir(self, tmp_path: Path) -> None:
        """Should handle .git/hooks subdirectory."""
        from rtmx.cli.install import find_git_dir

        git_dir = tmp_path / ".git"
        git_dir.mkdir()
        hooks_dir = git_dir / "hooks"
        hooks_dir.mkdir()

        result = find_git_dir(tmp_path)
        assert result == git_dir


@pytest.mark.req("REQ-DX-005")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestInstallHooks:
    """Test hook installation functionality."""

    def test_install_precommit_hook(self, tmp_path: Path) -> None:
        """rtmx install --hooks should install pre-commit hook."""
        from rtmx.cli.install import install_hooks

        # Setup git directory structure
        git_dir = tmp_path / ".git"
        git_dir.mkdir()
        hooks_dir = git_dir / "hooks"
        hooks_dir.mkdir()

        os.chdir(tmp_path)
        result = install_hooks(dry_run=False, pre_push=False, remove=False)

        assert result is True
        precommit = hooks_dir / "pre-commit"
        assert precommit.exists()
        assert precommit.stat().st_mode & stat.S_IXUSR  # Executable

    def test_install_precommit_hook_content(self, tmp_path: Path) -> None:
        """Pre-commit hook should contain health check command."""
        from rtmx.cli.install import install_hooks

        # Setup git directory structure
        git_dir = tmp_path / ".git"
        git_dir.mkdir()
        hooks_dir = git_dir / "hooks"
        hooks_dir.mkdir()

        os.chdir(tmp_path)
        install_hooks(dry_run=False, pre_push=False, remove=False)

        precommit = hooks_dir / "pre-commit"
        content = precommit.read_text()
        assert "rtmx health --strict" in content
        assert "#!/bin/sh" in content

    def test_install_prepush_hook(self, tmp_path: Path) -> None:
        """rtmx install --hooks --pre-push should install pre-push hook."""
        from rtmx.cli.install import install_hooks

        # Setup git directory structure
        git_dir = tmp_path / ".git"
        git_dir.mkdir()
        hooks_dir = git_dir / "hooks"
        hooks_dir.mkdir()

        os.chdir(tmp_path)
        result = install_hooks(dry_run=False, pre_push=True, remove=False)

        assert result is True
        prepush = hooks_dir / "pre-push"
        assert prepush.exists()
        assert prepush.stat().st_mode & stat.S_IXUSR  # Executable

    def test_install_both_hooks(self, tmp_path: Path) -> None:
        """Installing with --pre-push should install both hooks."""
        from rtmx.cli.install import install_hooks

        # Setup git directory structure
        git_dir = tmp_path / ".git"
        git_dir.mkdir()
        hooks_dir = git_dir / "hooks"
        hooks_dir.mkdir()

        os.chdir(tmp_path)
        result = install_hooks(dry_run=False, pre_push=True, remove=False)

        assert result is True
        assert (hooks_dir / "pre-commit").exists()
        assert (hooks_dir / "pre-push").exists()

    def test_install_hooks_creates_hooks_dir(self, tmp_path: Path) -> None:
        """Should create hooks directory if it doesn't exist."""
        from rtmx.cli.install import install_hooks

        # Setup git directory without hooks subdirectory
        git_dir = tmp_path / ".git"
        git_dir.mkdir()

        os.chdir(tmp_path)
        result = install_hooks(dry_run=False, pre_push=False, remove=False)

        assert result is True
        assert (git_dir / "hooks").exists()
        assert (git_dir / "hooks" / "pre-commit").exists()

    def test_install_hooks_not_in_git_repo(self, tmp_path: Path) -> None:
        """Should return False when not in a git repository."""
        from rtmx.cli.install import install_hooks

        os.chdir(tmp_path)
        result = install_hooks(dry_run=False, pre_push=False, remove=False)

        assert result is False


@pytest.mark.req("REQ-DX-005")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRemoveHooks:
    """Test hook removal functionality."""

    def test_remove_precommit_hook(self, tmp_path: Path) -> None:
        """rtmx install --hooks --remove should remove pre-commit hook."""
        from rtmx.cli.install import install_hooks

        # Setup and install hook first
        git_dir = tmp_path / ".git"
        git_dir.mkdir()
        hooks_dir = git_dir / "hooks"
        hooks_dir.mkdir()

        os.chdir(tmp_path)
        install_hooks(dry_run=False, pre_push=False, remove=False)
        assert (hooks_dir / "pre-commit").exists()

        # Now remove
        result = install_hooks(dry_run=False, pre_push=False, remove=True)
        assert result is True
        assert not (hooks_dir / "pre-commit").exists()

    def test_remove_all_hooks(self, tmp_path: Path) -> None:
        """--remove should remove both pre-commit and pre-push hooks."""
        from rtmx.cli.install import install_hooks

        # Setup and install both hooks
        git_dir = tmp_path / ".git"
        git_dir.mkdir()
        hooks_dir = git_dir / "hooks"
        hooks_dir.mkdir()

        os.chdir(tmp_path)
        install_hooks(dry_run=False, pre_push=True, remove=False)
        assert (hooks_dir / "pre-commit").exists()
        assert (hooks_dir / "pre-push").exists()

        # Remove all rtmx hooks
        result = install_hooks(dry_run=False, pre_push=True, remove=True)
        assert result is True
        assert not (hooks_dir / "pre-commit").exists()
        assert not (hooks_dir / "pre-push").exists()

    def test_remove_only_rtmx_hooks(self, tmp_path: Path) -> None:
        """--remove should only remove RTMX-installed hooks, not others."""
        from rtmx.cli.install import install_hooks

        # Setup git directory
        git_dir = tmp_path / ".git"
        git_dir.mkdir()
        hooks_dir = git_dir / "hooks"
        hooks_dir.mkdir()

        # Create a non-rtmx hook
        other_hook = hooks_dir / "pre-commit"
        other_hook.write_text("#!/bin/sh\necho 'Not RTMX'\n")
        other_hook.chmod(0o755)

        os.chdir(tmp_path)
        result = install_hooks(dry_run=False, pre_push=False, remove=True)

        # Should return True but leave non-RTMX hook intact
        assert result is True
        assert other_hook.exists()  # Non-RTMX hook preserved

    def test_remove_hooks_not_in_git_repo(self, tmp_path: Path) -> None:
        """Remove should return False when not in a git repository."""
        from rtmx.cli.install import install_hooks

        os.chdir(tmp_path)
        result = install_hooks(dry_run=False, pre_push=False, remove=True)

        assert result is False


@pytest.mark.req("REQ-DX-005")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestDryRun:
    """Test dry run mode for hook operations."""

    def test_dry_run_does_not_create_hooks(self, tmp_path: Path) -> None:
        """--dry-run should not actually create hooks."""
        from rtmx.cli.install import install_hooks

        # Setup git directory structure
        git_dir = tmp_path / ".git"
        git_dir.mkdir()
        hooks_dir = git_dir / "hooks"
        hooks_dir.mkdir()

        os.chdir(tmp_path)
        result = install_hooks(dry_run=True, pre_push=False, remove=False)

        assert result is True  # Operation would succeed
        assert not (hooks_dir / "pre-commit").exists()

    def test_dry_run_does_not_remove_hooks(self, tmp_path: Path) -> None:
        """--dry-run --remove should not actually remove hooks."""
        from rtmx.cli.install import install_hooks

        # Setup and install hook first
        git_dir = tmp_path / ".git"
        git_dir.mkdir()
        hooks_dir = git_dir / "hooks"
        hooks_dir.mkdir()

        os.chdir(tmp_path)
        install_hooks(dry_run=False, pre_push=False, remove=False)
        assert (hooks_dir / "pre-commit").exists()

        # Dry run remove
        result = install_hooks(dry_run=True, pre_push=False, remove=True)
        assert result is True  # Operation would succeed
        assert (hooks_dir / "pre-commit").exists()  # Still there


@pytest.mark.req("REQ-DX-005")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestHookBackup:
    """Test hook backup functionality."""

    def test_install_backs_up_existing_hook(self, tmp_path: Path) -> None:
        """Installing should back up existing non-RTMX hook."""
        from rtmx.cli.install import install_hooks

        # Setup git directory
        git_dir = tmp_path / ".git"
        git_dir.mkdir()
        hooks_dir = git_dir / "hooks"
        hooks_dir.mkdir()

        # Create existing non-rtmx hook
        existing_hook = hooks_dir / "pre-commit"
        existing_hook.write_text("#!/bin/sh\necho 'Original'\n")
        existing_hook.chmod(0o755)

        os.chdir(tmp_path)
        result = install_hooks(dry_run=False, pre_push=False, remove=False)

        assert result is True
        # Check backup exists
        backups = list(hooks_dir.glob("pre-commit.rtmx-backup-*"))
        assert len(backups) == 1

    def test_install_overwrites_existing_rtmx_hook(self, tmp_path: Path) -> None:
        """Installing should overwrite existing RTMX hook without backup."""
        from rtmx.cli.install import install_hooks

        # Setup and install hook first
        git_dir = tmp_path / ".git"
        git_dir.mkdir()
        hooks_dir = git_dir / "hooks"
        hooks_dir.mkdir()

        os.chdir(tmp_path)
        install_hooks(dry_run=False, pre_push=False, remove=False)

        # Install again
        result = install_hooks(dry_run=False, pre_push=False, remove=False)

        assert result is True
        # No backup should be created for RTMX hooks
        backups = list(hooks_dir.glob("pre-commit.rtmx-backup-*"))
        assert len(backups) == 0


@pytest.mark.req("REQ-DX-005")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestIsRtmxHook:
    """Test RTMX hook detection."""

    def test_is_rtmx_hook_true(self, tmp_path: Path) -> None:
        """Should identify RTMX-installed hooks."""
        from rtmx.cli.install import is_rtmx_hook

        hook_file = tmp_path / "pre-commit"
        hook_file.write_text("#!/bin/sh\n# RTMX pre-commit hook\nrtmx health --strict\n")

        assert is_rtmx_hook(hook_file) is True

    def test_is_rtmx_hook_false(self, tmp_path: Path) -> None:
        """Should not identify non-RTMX hooks."""
        from rtmx.cli.install import is_rtmx_hook

        hook_file = tmp_path / "pre-commit"
        hook_file.write_text("#!/bin/sh\necho 'Not RTMX'\n")

        assert is_rtmx_hook(hook_file) is False

    def test_is_rtmx_hook_missing_file(self, tmp_path: Path) -> None:
        """Should return False for missing files."""
        from rtmx.cli.install import is_rtmx_hook

        hook_file = tmp_path / "nonexistent"
        assert is_rtmx_hook(hook_file) is False


@pytest.mark.req("REQ-DX-005")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestHookExecution:
    """Test that installed hooks are executable."""

    def test_precommit_hook_is_executable(self, tmp_path: Path) -> None:
        """Installed pre-commit hook should be executable."""
        from rtmx.cli.install import install_hooks

        # Setup git directory
        git_dir = tmp_path / ".git"
        git_dir.mkdir()
        hooks_dir = git_dir / "hooks"
        hooks_dir.mkdir()

        os.chdir(tmp_path)
        install_hooks(dry_run=False, pre_push=False, remove=False)

        precommit = hooks_dir / "pre-commit"
        mode = precommit.stat().st_mode
        assert mode & stat.S_IXUSR  # Owner execute
        assert mode & stat.S_IXGRP  # Group execute
        assert mode & stat.S_IXOTH  # Other execute

    def test_prepush_hook_is_executable(self, tmp_path: Path) -> None:
        """Installed pre-push hook should be executable."""
        from rtmx.cli.install import install_hooks

        # Setup git directory
        git_dir = tmp_path / ".git"
        git_dir.mkdir()
        hooks_dir = git_dir / "hooks"
        hooks_dir.mkdir()

        os.chdir(tmp_path)
        install_hooks(dry_run=False, pre_push=True, remove=False)

        prepush = hooks_dir / "pre-push"
        mode = prepush.stat().st_mode
        assert mode & stat.S_IXUSR  # Owner execute


@pytest.mark.req("REQ-DX-005")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCLIIntegration:
    """Test CLI command integration for hooks."""

    def test_run_hooks_function_exists(self) -> None:
        """run_hooks function should exist in install module."""
        from rtmx.cli.install import run_hooks

        assert callable(run_hooks)

    def test_run_hooks_with_dry_run(self, tmp_path: Path, capsys) -> None:
        """run_hooks should support dry-run mode."""
        from rtmx.cli.install import run_hooks

        # Setup git directory
        git_dir = tmp_path / ".git"
        git_dir.mkdir()
        hooks_dir = git_dir / "hooks"
        hooks_dir.mkdir()

        os.chdir(tmp_path)
        run_hooks(dry_run=True, pre_push=False, remove=False)

        captured = capsys.readouterr()
        assert "DRY RUN" in captured.out or "dry" in captured.out.lower()

    def test_run_hooks_install_message(self, tmp_path: Path, capsys) -> None:
        """run_hooks should print success message on install."""
        from rtmx.cli.install import run_hooks

        # Setup git directory
        git_dir = tmp_path / ".git"
        git_dir.mkdir()
        hooks_dir = git_dir / "hooks"
        hooks_dir.mkdir()

        os.chdir(tmp_path)
        run_hooks(dry_run=False, pre_push=False, remove=False)

        captured = capsys.readouterr()
        assert "pre-commit" in captured.out.lower()

    def test_run_hooks_remove_message(self, tmp_path: Path, capsys) -> None:
        """run_hooks should print remove message."""
        from rtmx.cli.install import run_hooks

        # Setup and install hook first
        git_dir = tmp_path / ".git"
        git_dir.mkdir()
        hooks_dir = git_dir / "hooks"
        hooks_dir.mkdir()

        os.chdir(tmp_path)
        run_hooks(dry_run=False, pre_push=False, remove=False)

        # Now remove
        run_hooks(dry_run=False, pre_push=False, remove=True)

        captured = capsys.readouterr()
        assert "removed" in captured.out.lower() or "remove" in captured.out.lower()
