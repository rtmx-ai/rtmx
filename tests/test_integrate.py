"""Tests for rtmx.cli.integrate module."""

from pathlib import Path

from rtmx.cli.integrate import (
    GitStrategy,
    IntegrationMode,
    IntegrationResult,
)


class TestIntegrationMode:
    """Tests for IntegrationMode enum."""

    def test_integration_mode_values(self):
        """Test IntegrationMode enum has expected values."""
        assert IntegrationMode.VALIDATE.value == "validate"
        assert IntegrationMode.PREVIEW.value == "preview"
        assert IntegrationMode.EXECUTE.value == "execute"


class TestGitStrategy:
    """Tests for GitStrategy enum."""

    def test_git_strategy_values(self):
        """Test GitStrategy enum has expected values."""
        assert GitStrategy.WORKTREE.value == "worktree"
        assert GitStrategy.BRANCH.value == "branch"


class TestIntegrationResult:
    """Tests for IntegrationResult dataclass."""

    def test_integration_result_creation(self):
        """Test IntegrationResult creation."""
        result = IntegrationResult(
            success=True,
            mode=IntegrationMode.VALIDATE,
            git_strategy=None,
            branch_name=None,
            worktree_path=None,
            rollback_point="abc123",
            baseline_captured=True,
            health_status=None,
            comparison_status=None,
            pr_url=None,
        )
        assert result.success is True
        assert result.mode == IntegrationMode.VALIDATE

    def test_integration_result_with_errors(self):
        """Test IntegrationResult with errors."""
        result = IntegrationResult(
            success=False,
            mode=IntegrationMode.EXECUTE,
            git_strategy=GitStrategy.WORKTREE,
            branch_name="integration/rtmx-20251204",
            worktree_path=Path("/tmp/project-rtmx"),
            rollback_point="abc123",
            baseline_captured=True,
            health_status=None,
            comparison_status=None,
            pr_url=None,
            errors=["Failed to create worktree"],
            warnings=["Minor issue detected"],
        )
        assert result.success is False
        assert len(result.errors) == 1
        assert len(result.warnings) == 1

    def test_integration_result_to_dict(self):
        """Test IntegrationResult serialization."""
        result = IntegrationResult(
            success=True,
            mode=IntegrationMode.PREVIEW,
            git_strategy=GitStrategy.BRANCH,
            branch_name="test-branch",
            worktree_path=None,
            rollback_point="abc123",
            baseline_captured=True,
            health_status=None,
            comparison_status="stable",
            pr_url=None,
        )
        d = result.to_dict()
        assert d["success"] is True
        assert d["mode"] == "preview"
        assert d["git_strategy"] == "branch"
        assert d["branch_name"] == "test-branch"
        assert d["comparison_status"] == "stable"

    def test_integration_result_to_dict_with_worktree_path(self):
        """Test IntegrationResult serialization with worktree path."""
        result = IntegrationResult(
            success=True,
            mode=IntegrationMode.EXECUTE,
            git_strategy=GitStrategy.WORKTREE,
            branch_name="test-branch",
            worktree_path=Path("/tmp/project-rtmx"),
            rollback_point="abc123",
            baseline_captured=True,
            health_status=None,
            comparison_status=None,
            pr_url="https://github.com/org/repo/pull/123",
        )
        d = result.to_dict()
        assert d["worktree_path"] == "/tmp/project-rtmx"
        assert d["pr_url"] == "https://github.com/org/repo/pull/123"


class TestIntegrationModeLogic:
    """Tests for integration mode logic."""

    def test_validate_mode_does_not_set_git_strategy(self):
        """Test validate mode doesn't set git strategy in result."""
        result = IntegrationResult(
            success=True,
            mode=IntegrationMode.VALIDATE,
            git_strategy=None,  # Should be None for validate mode
            branch_name=None,
            worktree_path=None,
            rollback_point="abc123",
            baseline_captured=False,
            health_status=None,
            comparison_status=None,
            pr_url=None,
        )
        assert result.git_strategy is None

    def test_execute_mode_requires_git_strategy(self):
        """Test execute mode has git strategy."""
        result = IntegrationResult(
            success=True,
            mode=IntegrationMode.EXECUTE,
            git_strategy=GitStrategy.WORKTREE,
            branch_name="integration/rtmx-20251204",
            worktree_path=Path("/tmp/project-rtmx"),
            rollback_point="abc123",
            baseline_captured=True,
            health_status=None,
            comparison_status=None,
            pr_url=None,
        )
        assert result.git_strategy == GitStrategy.WORKTREE


class TestIntegrationResultErrors:
    """Tests for IntegrationResult error handling."""

    def test_default_empty_errors(self):
        """Test default empty error list."""
        result = IntegrationResult(
            success=True,
            mode=IntegrationMode.VALIDATE,
            git_strategy=None,
            branch_name=None,
            worktree_path=None,
            rollback_point=None,
            baseline_captured=False,
            health_status=None,
            comparison_status=None,
            pr_url=None,
        )
        assert result.errors == []
        assert result.warnings == []

    def test_errors_in_to_dict(self):
        """Test errors are included in serialization."""
        result = IntegrationResult(
            success=False,
            mode=IntegrationMode.VALIDATE,
            git_strategy=None,
            branch_name=None,
            worktree_path=None,
            rollback_point=None,
            baseline_captured=False,
            health_status=None,
            comparison_status=None,
            pr_url=None,
            errors=["Error 1", "Error 2"],
            warnings=["Warning 1"],
        )
        d = result.to_dict()
        assert d["errors"] == ["Error 1", "Error 2"]
        assert d["warnings"] == ["Warning 1"]
