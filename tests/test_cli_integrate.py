"""Comprehensive tests for rtmx.cli.integrate module.

This file provides extensive coverage of the integration workflow including
validation mode, preview mode, execute mode, git strategies, and error handling.
"""

from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from rtmx.cli.git_ops import GitError, GitStatus
from rtmx.cli.health import Check, CheckResult, HealthReport, HealthStatus
from rtmx.cli.integrate import (
    GitStrategy,
    IntegrationMode,
    IntegrationResult,
    run_integrate,
)
from rtmx.config import RTMXConfig


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestIntegrationMode:
    """Tests for IntegrationMode enum."""

    def test_integration_mode_values(self):
        """Test IntegrationMode enum has expected values."""
        assert IntegrationMode.VALIDATE.value == "validate"
        assert IntegrationMode.PREVIEW.value == "preview"
        assert IntegrationMode.EXECUTE.value == "execute"

    def test_integration_mode_string_comparison(self):
        """Test IntegrationMode can be compared as strings."""
        assert IntegrationMode.VALIDATE == "validate"
        assert IntegrationMode.PREVIEW == "preview"
        assert IntegrationMode.EXECUTE == "execute"


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGitStrategy:
    """Tests for GitStrategy enum."""

    def test_git_strategy_values(self):
        """Test GitStrategy enum has expected values."""
        assert GitStrategy.WORKTREE.value == "worktree"
        assert GitStrategy.BRANCH.value == "branch"

    def test_git_strategy_string_comparison(self):
        """Test GitStrategy can be compared as strings."""
        assert GitStrategy.WORKTREE == "worktree"
        assert GitStrategy.BRANCH == "branch"


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestIntegrationResult:
    """Tests for IntegrationResult dataclass."""

    def test_integration_result_defaults(self):
        """Test IntegrationResult with default values."""
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
        )
        assert result.success is False
        assert result.mode == IntegrationMode.VALIDATE
        assert result.errors == []
        assert result.warnings == []

    def test_integration_result_with_errors(self):
        """Test IntegrationResult with errors and warnings."""
        result = IntegrationResult(
            success=False,
            mode=IntegrationMode.EXECUTE,
            git_strategy=GitStrategy.WORKTREE,
            branch_name="test-branch",
            worktree_path=Path("/test/worktree"),
            rollback_point="abc123",
            baseline_captured=True,
            health_status=HealthStatus.UNHEALTHY,
            comparison_status="degraded",
            pr_url=None,
            errors=["Error 1", "Error 2"],
            warnings=["Warning 1"],
        )
        assert len(result.errors) == 2
        assert len(result.warnings) == 1

    def test_integration_result_to_dict(self):
        """Test IntegrationResult serialization to dict."""
        result = IntegrationResult(
            success=True,
            mode=IntegrationMode.EXECUTE,
            git_strategy=GitStrategy.BRANCH,
            branch_name="test-branch",
            worktree_path=Path("/test/worktree"),
            rollback_point="abc123def456",
            baseline_captured=True,
            health_status=HealthStatus.HEALTHY,
            comparison_status="improved",
            pr_url="https://github.com/user/repo/pull/1",
        )

        result_dict = result.to_dict()

        assert result_dict["success"] is True
        assert result_dict["mode"] == "execute"
        assert result_dict["git_strategy"] == "branch"
        assert result_dict["branch_name"] == "test-branch"
        assert result_dict["worktree_path"] == "/test/worktree"
        assert result_dict["rollback_point"] == "abc123def456"
        assert result_dict["baseline_captured"] is True
        assert result_dict["health_status"] == "healthy"
        assert result_dict["comparison_status"] == "improved"
        assert result_dict["pr_url"] == "https://github.com/user/repo/pull/1"
        assert result_dict["errors"] == []
        assert result_dict["warnings"] == []

    def test_integration_result_to_dict_with_none_values(self):
        """Test IntegrationResult to_dict handles None values correctly."""
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
        )

        result_dict = result.to_dict()

        assert result_dict["git_strategy"] is None
        assert result_dict["branch_name"] is None
        assert result_dict["worktree_path"] is None
        assert result_dict["rollback_point"] is None
        assert result_dict["health_status"] is None
        assert result_dict["comparison_status"] is None
        assert result_dict["pr_url"] is None


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRunIntegrateValidateMode:
    """Tests for run_integrate in validate mode."""

    def test_validate_mode_not_git_repo(self, tmp_path):
        """Test validation fails when not a git repository."""
        with patch("rtmx.cli.integrate.is_git_repo") as mock_is_git:
            mock_is_git.return_value = False

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.VALIDATE,
            )

            assert result.success is False
            assert "Not a git repository" in result.errors
            assert result.mode == IntegrationMode.VALIDATE
            assert result.git_strategy is None

    def test_validate_mode_git_status_clean(self, tmp_path):
        """Test validation with clean git status."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[
                Check(
                    name="test_check",
                    result=CheckResult.PASS,
                    message="All good",
                )
            ],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.VALIDATE,
            )

            assert result.success is True
            assert result.mode == IntegrationMode.VALIDATE
            assert result.health_status == HealthStatus.HEALTHY
            assert result.rollback_point == "abc123def456"
            assert len(result.errors) == 0

    def test_validate_mode_git_status_dirty_warns(self, tmp_path):
        """Test validation with uncommitted changes shows warning."""
        git_status = GitStatus(
            is_clean=False,
            branch="main",
            uncommitted_files=["file1.txt", "file2.txt"],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.VALIDATE,
            )

            assert result.success is True
            assert "Uncommitted changes detected" in result.warnings
            assert len(result.errors) == 0

    def test_validate_mode_health_unhealthy_fails(self, tmp_path):
        """Test validation fails with unhealthy health status."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.UNHEALTHY,
            checks=[
                Check(
                    name="critical_check",
                    result=CheckResult.FAIL,
                    message="Critical failure",
                    blocking=True,
                )
            ],
            summary={"passed": 0, "warnings": 0, "failed": 1},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.VALIDATE,
            )

            assert result.success is False
            assert result.health_status == HealthStatus.UNHEALTHY
            assert "Critical failure" in result.errors

    def test_validate_mode_health_degraded_warns(self, tmp_path):
        """Test validation succeeds with degraded health status."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.DEGRADED,
            checks=[
                Check(
                    name="warning_check",
                    result=CheckResult.WARN,
                    message="Non-critical warning",
                    blocking=False,
                )
            ],
            summary={"passed": 4, "warnings": 1, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.VALIDATE,
            )

            assert result.success is True
            assert result.health_status == HealthStatus.DEGRADED
            assert "Non-critical warning" in result.warnings

    def test_validate_mode_git_error_fails(self, tmp_path):
        """Test validation fails when git status check errors."""
        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
        ):
            mock_is_git.return_value = True
            mock_status.side_effect = GitError("Git command failed")

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.VALIDATE,
            )

            assert result.success is False
            assert any("Git command failed" in error for error in result.errors)


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRunIntegratePreviewMode:
    """Tests for run_integrate in preview mode."""

    def test_preview_mode_worktree_strategy(self, tmp_path):
        """Test preview mode with worktree strategy shows expected steps."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.generate_branch_name") as mock_branch,
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_branch.return_value = "integration/rtmx-20251205-120000"

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.PREVIEW,
                git_strategy=GitStrategy.WORKTREE,
            )

            assert result.success is True
            assert result.mode == IntegrationMode.PREVIEW
            assert result.branch_name == "integration/rtmx-20251205-120000"
            assert result.git_strategy is None  # Preview doesn't set strategy
            assert len(result.errors) == 0

    def test_preview_mode_branch_strategy(self, tmp_path):
        """Test preview mode with branch strategy shows expected steps."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.generate_branch_name") as mock_branch,
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_branch.return_value = "integration/rtmx-20251205-120000"

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.PREVIEW,
                git_strategy=GitStrategy.BRANCH,
            )

            assert result.success is True
            assert result.mode == IntegrationMode.PREVIEW
            assert result.branch_name == "integration/rtmx-20251205-120000"

    def test_preview_mode_custom_branch_name(self, tmp_path):
        """Test preview mode with custom branch name."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.PREVIEW,
                branch_name="custom-rtmx-branch",
            )

            assert result.success is True
            assert result.branch_name == "custom-rtmx-branch"

    def test_preview_mode_with_pr_flag(self, tmp_path):
        """Test preview mode shows PR creation step when flag is set."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.generate_branch_name") as mock_branch,
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_branch.return_value = "integration/rtmx-test"

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.PREVIEW,
                create_pr_flag=True,
            )

            assert result.success is True
            assert result.pr_url is None  # Preview doesn't create PR


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRunIntegrateExecuteMode:
    """Tests for run_integrate in execute mode."""

    def test_execute_mode_dirty_repo_fails(self, tmp_path):
        """Test execute mode fails with uncommitted changes."""
        git_status = GitStatus(
            is_clean=False,
            branch="main",
            uncommitted_files=["file1.txt"],
            commit_sha="abc123def456",
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.EXECUTE,
            )

            assert result.success is False
            assert "Uncommitted changes detected" in result.errors

    def test_execute_mode_worktree_strategy_success(self, tmp_path):
        """Test execute mode with worktree strategy creates worktree."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.generate_branch_name") as mock_branch,
            patch("rtmx.cli.integrate.create_worktree") as mock_worktree,
            patch("rtmx.cli.init.run_init"),
            patch("rtmx.cli.install.run_install"),
            patch("os.chdir"),
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_branch.return_value = "integration/rtmx-test"

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.EXECUTE,
                git_strategy=GitStrategy.WORKTREE,
            )

            assert result.success is True
            assert result.mode == IntegrationMode.EXECUTE
            assert result.git_strategy == GitStrategy.WORKTREE
            assert result.branch_name == "integration/rtmx-test"
            mock_worktree.assert_called_once()

    def test_execute_mode_worktree_creation_fails(self, tmp_path):
        """Test execute mode handles worktree creation failure."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.generate_branch_name") as mock_branch,
            patch("rtmx.cli.integrate.create_worktree") as mock_worktree,
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_branch.return_value = "integration/rtmx-test"
            mock_worktree.side_effect = GitError("Worktree already exists")

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.EXECUTE,
                git_strategy=GitStrategy.WORKTREE,
            )

            assert result.success is False
            assert any("Failed to create worktree" in error for error in result.errors)

    def test_execute_mode_branch_strategy_success(self, tmp_path):
        """Test execute mode with branch strategy creates branch."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.generate_branch_name") as mock_branch,
            patch("rtmx.cli.integrate.create_branch") as mock_create_branch,
            patch("rtmx.cli.init.run_init"),
            patch("rtmx.cli.install.run_install"),
            patch("os.chdir"),
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_branch.return_value = "integration/rtmx-test"

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.EXECUTE,
                git_strategy=GitStrategy.BRANCH,
            )

            assert result.success is True
            assert result.mode == IntegrationMode.EXECUTE
            assert result.git_strategy == GitStrategy.BRANCH
            assert result.branch_name == "integration/rtmx-test"
            mock_create_branch.assert_called_once()

    def test_execute_mode_branch_creation_fails(self, tmp_path):
        """Test execute mode handles branch creation failure."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.generate_branch_name") as mock_branch,
            patch("rtmx.cli.integrate.create_branch") as mock_create_branch,
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_branch.return_value = "integration/rtmx-test"
            mock_create_branch.side_effect = GitError("Branch already exists")

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.EXECUTE,
                git_strategy=GitStrategy.BRANCH,
            )

            assert result.success is False
            assert any("Failed to create branch" in error for error in result.errors)


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRunIntegrateBaselineCapture:
    """Tests for baseline capture in run_integrate."""

    def test_baseline_capture_success(self, tmp_path):
        """Test successful baseline capture."""
        rtm_path = tmp_path / "docs" / "rtm_database.csv"
        rtm_path.parent.mkdir(parents=True)
        rtm_path.write_text("req_id,phase,subsystem\n")

        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        baseline_data = {
            "req_count": 10,
            "completion": 75.5,
            "cycles": 0,
        }

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.capture_baseline") as mock_baseline,
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_baseline.return_value = baseline_data

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.VALIDATE,
            )

            assert result.baseline_captured is True
            mock_baseline.assert_called_once()

    def test_baseline_capture_no_rtm_file(self, tmp_path):
        """Test baseline capture when RTM file doesn't exist."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.VALIDATE,
            )

            assert result.baseline_captured is False

    def test_baseline_capture_exception_warns(self, tmp_path):
        """Test baseline capture exception adds warning."""
        rtm_path = tmp_path / "docs" / "rtm_database.csv"
        rtm_path.parent.mkdir(parents=True)
        rtm_path.write_text("req_id,phase,subsystem\n")

        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.capture_baseline") as mock_baseline,
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_baseline.side_effect = Exception("Baseline error")

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.VALIDATE,
            )

            assert result.baseline_captured is False
            assert any("Baseline capture failed" in warning for warning in result.warnings)


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRunIntegratePRCreation:
    """Tests for PR creation in run_integrate."""

    def test_execute_mode_create_pr_success(self, tmp_path):
        """Test PR creation succeeds in execute mode."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.generate_branch_name") as mock_branch,
            patch("rtmx.cli.integrate.create_branch"),
            patch("rtmx.cli.init.run_init"),
            patch("rtmx.cli.install.run_install"),
            patch("rtmx.cli.integrate.check_gh_installed") as mock_gh_check,
            patch("rtmx.cli.integrate.create_pr") as mock_pr,
            patch("os.chdir"),
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_branch.return_value = "integration/rtmx-test"
            mock_gh_check.return_value = True
            mock_pr.return_value = "https://github.com/user/repo/pull/42"

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.EXECUTE,
                git_strategy=GitStrategy.BRANCH,
                create_pr_flag=True,
            )

            assert result.success is True
            assert result.pr_url == "https://github.com/user/repo/pull/42"
            mock_pr.assert_called_once()

    def test_execute_mode_create_pr_no_gh_cli(self, tmp_path):
        """Test PR creation skipped when gh CLI not installed."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.generate_branch_name") as mock_branch,
            patch("rtmx.cli.integrate.create_branch"),
            patch("rtmx.cli.init.run_init"),
            patch("rtmx.cli.install.run_install"),
            patch("rtmx.cli.integrate.check_gh_installed") as mock_gh_check,
            patch("os.chdir"),
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_branch.return_value = "integration/rtmx-test"
            mock_gh_check.return_value = False

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.EXECUTE,
                git_strategy=GitStrategy.BRANCH,
                create_pr_flag=True,
            )

            assert result.success is True
            assert result.pr_url is None
            assert any("GitHub CLI not installed" in warning for warning in result.warnings)

    def test_execute_mode_create_pr_fails(self, tmp_path):
        """Test PR creation failure adds warning."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.generate_branch_name") as mock_branch,
            patch("rtmx.cli.integrate.create_branch"),
            patch("rtmx.cli.init.run_init"),
            patch("rtmx.cli.install.run_install"),
            patch("rtmx.cli.integrate.check_gh_installed") as mock_gh_check,
            patch("rtmx.cli.integrate.create_pr") as mock_pr,
            patch("os.chdir"),
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_branch.return_value = "integration/rtmx-test"
            mock_gh_check.return_value = True
            mock_pr.side_effect = GitError("PR creation failed")

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.EXECUTE,
                git_strategy=GitStrategy.BRANCH,
                create_pr_flag=True,
            )

            assert result.success is True  # PR failure doesn't fail integration
            assert result.pr_url is None
            assert any("PR creation failed" in warning for warning in result.warnings)

    def test_execute_mode_create_pr_returns_none(self, tmp_path):
        """Test PR creation returning None adds warning."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.generate_branch_name") as mock_branch,
            patch("rtmx.cli.integrate.create_branch"),
            patch("rtmx.cli.init.run_init"),
            patch("rtmx.cli.install.run_install"),
            patch("rtmx.cli.integrate.check_gh_installed") as mock_gh_check,
            patch("rtmx.cli.integrate.create_pr") as mock_pr,
            patch("os.chdir"),
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_branch.return_value = "integration/rtmx-test"
            mock_gh_check.return_value = True
            mock_pr.return_value = None

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.EXECUTE,
                git_strategy=GitStrategy.BRANCH,
                create_pr_flag=True,
            )

            assert result.success is True
            assert result.pr_url is None
            assert any("PR creation returned no URL" in warning for warning in result.warnings)


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRunIntegratePostValidation:
    """Tests for post-integration validation."""

    def test_execute_mode_runs_init_when_needed(self, tmp_path):
        """Test rtmx init is called when rtmx.yaml doesn't exist."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.generate_branch_name") as mock_branch,
            patch("rtmx.cli.integrate.create_branch"),
            patch("rtmx.cli.init.run_init") as mock_init,
            patch("rtmx.cli.install.run_install"),
            patch("os.chdir"),
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_branch.return_value = "integration/rtmx-test"

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.EXECUTE,
                git_strategy=GitStrategy.BRANCH,
            )

            assert result.success is True
            mock_init.assert_called_once()

    def test_execute_mode_skips_init_when_exists(self, tmp_path):
        """Test rtmx init is skipped when rtmx.yaml exists."""
        # Create rtmx.yaml in tmp_path
        (tmp_path / "rtmx.yaml").write_text("database: docs/rtm_database.csv\n")

        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.generate_branch_name") as mock_branch,
            patch("rtmx.cli.integrate.create_branch"),
            patch("rtmx.cli.init.run_init") as mock_init,
            patch("rtmx.cli.install.run_install"),
            patch("os.chdir"),
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_branch.return_value = "integration/rtmx-test"

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.EXECUTE,
                git_strategy=GitStrategy.BRANCH,
            )

            assert result.success is True
            mock_init.assert_not_called()

    def test_execute_mode_init_warning_doesnt_fail(self, tmp_path):
        """Test rtmx init warning doesn't fail integration."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.generate_branch_name") as mock_branch,
            patch("rtmx.cli.integrate.create_branch"),
            patch("rtmx.cli.init.run_init") as mock_init,
            patch("rtmx.cli.install.run_install"),
            patch("os.chdir"),
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_branch.return_value = "integration/rtmx-test"
            mock_init.side_effect = Exception("Init failed")

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.EXECUTE,
                git_strategy=GitStrategy.BRANCH,
            )

            assert result.success is True
            assert any("rtmx init warning" in warning for warning in result.warnings)

    def test_execute_mode_install_warning_doesnt_fail(self, tmp_path):
        """Test rtmx install warning doesn't fail integration."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.generate_branch_name") as mock_branch,
            patch("rtmx.cli.integrate.create_branch"),
            patch("rtmx.cli.init.run_init"),
            patch("rtmx.cli.install.run_install") as mock_install,
            patch("os.chdir"),
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_branch.return_value = "integration/rtmx-test"
            mock_install.side_effect = Exception("Install failed")

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.EXECUTE,
                git_strategy=GitStrategy.BRANCH,
            )

            assert result.success is True
            assert any("rtmx install warning" in warning for warning in result.warnings)

    def test_execute_mode_post_health_unhealthy_adds_error(self, tmp_path):
        """Test post-integration unhealthy health adds error."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        pre_health = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        post_health = HealthReport(
            status=HealthStatus.UNHEALTHY,
            checks=[],
            summary={"passed": 0, "warnings": 0, "failed": 5},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.generate_branch_name") as mock_branch,
            patch("rtmx.cli.integrate.create_branch"),
            patch("rtmx.cli.init.run_init"),
            patch("rtmx.cli.install.run_install"),
            patch("os.chdir"),
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.side_effect = [pre_health, post_health]
            mock_config.return_value = RTMXConfig()
            mock_branch.return_value = "integration/rtmx-test"

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.EXECUTE,
                git_strategy=GitStrategy.BRANCH,
            )

            assert result.success is False
            assert "Post-integration health check failed" in result.errors


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRunIntegrateComparison:
    """Tests for database comparison in run_integrate."""

    def test_comparison_success_in_execute_mode(self, tmp_path):
        """Test database comparison runs successfully in execute mode."""
        # Create RTM database
        rtm_path = tmp_path / "docs" / "rtm_database.csv"
        rtm_path.parent.mkdir(parents=True)
        rtm_path.write_text("req_id,phase,subsystem\n")

        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        baseline_data = {
            "req_count": 10,
            "completion": 75.5,
            "cycles": 0,
        }

        comparison_data = MagicMock()
        comparison_data.summary_status = "improved"
        comparison_data.baseline_req_count = 10
        comparison_data.current_req_count = 15
        comparison_data.baseline_completion = 75.5
        comparison_data.current_completion = 85.0

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.generate_branch_name") as mock_branch,
            patch("rtmx.cli.integrate.create_branch"),
            patch("rtmx.cli.init.run_init"),
            patch("rtmx.cli.install.run_install"),
            patch("rtmx.cli.integrate.capture_baseline") as mock_baseline,
            patch("rtmx.cli.integrate.compare_databases") as mock_compare,
            patch("os.chdir"),
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_branch.return_value = "integration/rtmx-test"
            mock_baseline.return_value = baseline_data
            mock_compare.return_value = comparison_data

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.EXECUTE,
                git_strategy=GitStrategy.BRANCH,
            )

            assert result.success is True
            assert result.comparison_status == "improved"
            mock_compare.assert_called_once()

    def test_comparison_exception_adds_warning(self, tmp_path):
        """Test database comparison exception adds warning."""
        # Create RTM database
        rtm_path = tmp_path / "docs" / "rtm_database.csv"
        rtm_path.parent.mkdir(parents=True)
        rtm_path.write_text("req_id,phase,subsystem\n")

        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        baseline_data = {
            "req_count": 10,
            "completion": 75.5,
            "cycles": 0,
        }

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.generate_branch_name") as mock_branch,
            patch("rtmx.cli.integrate.create_branch"),
            patch("rtmx.cli.init.run_init"),
            patch("rtmx.cli.install.run_install"),
            patch("rtmx.cli.integrate.capture_baseline") as mock_baseline,
            patch("rtmx.cli.integrate.compare_databases") as mock_compare,
            patch("os.chdir"),
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()
            mock_branch.return_value = "integration/rtmx-test"
            mock_baseline.return_value = baseline_data
            mock_compare.side_effect = Exception("Comparison error")

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.EXECUTE,
                git_strategy=GitStrategy.BRANCH,
            )

            assert result.success is True
            assert any("Comparison failed" in warning for warning in result.warnings)


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRunIntegrateParameters:
    """Tests for run_integrate parameter handling."""

    def test_default_project_path_uses_cwd(self, tmp_path):
        """Test default project path is current working directory."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
            patch("rtmx.cli.integrate.Path.cwd") as mock_cwd,
        ):
            mock_cwd.return_value = tmp_path
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()

            result = run_integrate(mode=IntegrationMode.VALIDATE)

            assert result.success is True

    def test_preloaded_config_used(self, tmp_path):
        """Test pre-loaded config is used instead of loading."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        custom_config = RTMXConfig()

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report

            result = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.VALIDATE,
                config=custom_config,
            )

            assert result.success is True
            # load_config should not be called since config was provided
            mock_config.assert_not_called()

    def test_git_strategy_set_in_result_only_for_execute(self, tmp_path):
        """Test git_strategy is only set in result for execute mode."""
        git_status = GitStatus(
            is_clean=True,
            branch="main",
            uncommitted_files=[],
            commit_sha="abc123def456",
        )

        health_report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"passed": 5, "warnings": 0, "failed": 0},
        )

        with (
            patch("rtmx.cli.integrate.is_git_repo") as mock_is_git,
            patch("rtmx.cli.integrate.get_git_status") as mock_status,
            patch("rtmx.cli.integrate.create_rollback_point") as mock_rollback,
            patch("rtmx.cli.integrate.run_health_checks") as mock_health,
            patch("rtmx.cli.integrate.load_config") as mock_config,
        ):
            mock_is_git.return_value = True
            mock_status.return_value = git_status
            mock_rollback.return_value = "abc123def456"
            mock_health.return_value = health_report
            mock_config.return_value = RTMXConfig()

            # Validate mode
            result_validate = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.VALIDATE,
                git_strategy=GitStrategy.WORKTREE,
            )
            assert result_validate.git_strategy is None

            # Preview mode
            result_preview = run_integrate(
                project_path=tmp_path,
                mode=IntegrationMode.PREVIEW,
                git_strategy=GitStrategy.WORKTREE,
            )
            assert result_preview.git_strategy is None
