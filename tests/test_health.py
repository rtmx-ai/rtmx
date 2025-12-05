"""Tests for rtmx.cli.health module."""

from pathlib import Path

import pytest

from rtmx.cli.health import (
    Check,
    CheckResult,
    HealthReport,
    HealthStatus,
    check_config_valid,
    check_cycles,
    check_reciprocity,
    check_rtm_exists,
    check_rtm_loads,
    check_schema_valid,
    run_health_checks,
)
from rtmx.config import RTMXConfig


@pytest.fixture
def test_config(core_rtm_path: Path) -> RTMXConfig:
    """Create test config pointing to core RTM fixture."""
    return RTMXConfig(database=str(core_rtm_path))


class TestCheckResult:
    """Tests for CheckResult enum."""

    def test_check_result_values(self):
        """Test CheckResult enum has expected values."""
        assert CheckResult.PASS.value == "pass"
        assert CheckResult.WARN.value == "warn"
        assert CheckResult.FAIL.value == "fail"
        assert CheckResult.SKIP.value == "skip"


class TestHealthStatus:
    """Tests for HealthStatus enum."""

    def test_health_status_values(self):
        """Test HealthStatus enum has expected values."""
        assert HealthStatus.HEALTHY.value == "healthy"
        assert HealthStatus.DEGRADED.value == "degraded"
        assert HealthStatus.UNHEALTHY.value == "unhealthy"


class TestCheck:
    """Tests for Check dataclass."""

    def test_check_creation(self):
        """Test Check dataclass creation."""
        check = Check(
            name="test_check",
            result=CheckResult.PASS,
            message="Test passed",
            blocking=True,
        )
        assert check.name == "test_check"
        assert check.result == CheckResult.PASS
        assert check.message == "Test passed"
        assert check.blocking is True

    def test_check_to_dict(self):
        """Test Check serialization."""
        check = Check(
            name="test_check",
            result=CheckResult.WARN,
            message="Test warning",
            blocking=False,
            details={"count": 5},
        )
        d = check.to_dict()
        assert d["name"] == "test_check"
        assert d["result"] == "warn"
        assert d["message"] == "Test warning"
        assert d["blocking"] is False
        assert d["details"] == {"count": 5}


class TestHealthReport:
    """Tests for HealthReport dataclass."""

    def test_health_report_creation(self):
        """Test HealthReport creation."""
        report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[],
            summary={"total": 0, "passed": 0, "warnings": 0, "failed": 0},
        )
        assert report.status == HealthStatus.HEALTHY

    def test_health_report_to_dict(self):
        """Test HealthReport serialization."""
        check = Check(
            name="test",
            result=CheckResult.PASS,
            message="OK",
            blocking=True,
        )
        report = HealthReport(
            status=HealthStatus.HEALTHY,
            checks=[check],
            summary={"total": 1, "passed": 1, "warnings": 0, "failed": 0},
        )
        d = report.to_dict()
        assert d["status"] == "healthy"
        assert len(d["checks"]) == 1
        assert d["summary"]["passed"] == 1


class TestCheckRtmExists:
    """Tests for check_rtm_exists function."""

    def test_rtm_exists_pass(self, test_config: RTMXConfig):
        """Test passing when RTM exists."""
        result = check_rtm_exists(test_config)
        assert result.name == "rtm_exists"
        assert result.result == CheckResult.PASS

    def test_rtm_exists_fail_missing(self, tmp_path: Path):
        """Test failing when RTM doesn't exist."""
        config = RTMXConfig(database=str(tmp_path / "nonexistent.csv"))
        result = check_rtm_exists(config)
        assert result.result == CheckResult.FAIL
        assert "not found" in result.message

    def test_rtm_exists_fail_empty(self, tmp_path: Path):
        """Test failing when RTM is empty."""
        empty_file = tmp_path / "empty.csv"
        empty_file.write_text("")
        config = RTMXConfig(database=str(empty_file))
        result = check_rtm_exists(config)
        assert result.result == CheckResult.FAIL
        assert "empty" in result.message.lower()


class TestCheckRtmLoads:
    """Tests for check_rtm_loads function."""

    def test_rtm_loads_pass(self, test_config: RTMXConfig):
        """Test passing when RTM loads successfully."""
        result = check_rtm_loads(test_config)
        assert result.name == "rtm_loads"
        assert result.result == CheckResult.PASS
        assert result.details is not None
        assert result.details["requirement_count"] > 0

    def test_rtm_loads_fail_invalid(self, tmp_path: Path):
        """Test failing when RTM has invalid format."""
        invalid_file = tmp_path / "invalid.csv"
        invalid_file.write_text("not,valid,csv,format\n")
        config = RTMXConfig(database=str(invalid_file))
        result = check_rtm_loads(config)
        assert result.result == CheckResult.FAIL


class TestCheckSchemaValid:
    """Tests for check_schema_valid function."""

    def test_schema_valid_pass(self, test_config: RTMXConfig):
        """Test passing when schema is valid."""
        result = check_schema_valid(test_config)
        assert result.name == "schema_valid"
        assert result.result == CheckResult.PASS


class TestCheckReciprocity:
    """Tests for check_reciprocity function."""

    def test_reciprocity_check_runs(self, test_config: RTMXConfig):
        """Test reciprocity check executes."""
        result = check_reciprocity(test_config)
        assert result.name == "reciprocity"
        assert result.result in (CheckResult.PASS, CheckResult.WARN)
        assert result.blocking is False


class TestCheckCycles:
    """Tests for check_cycles function."""

    def test_cycles_check_runs(self, test_config: RTMXConfig):
        """Test cycles check executes."""
        result = check_cycles(test_config)
        assert result.name == "cycles"
        assert result.result in (CheckResult.PASS, CheckResult.WARN)
        assert result.blocking is False


class TestCheckConfigValid:
    """Tests for check_config_valid function."""

    def test_config_valid_default(self, monkeypatch: pytest.MonkeyPatch, tmp_path: Path):
        """Test config validation with no config file."""
        monkeypatch.chdir(tmp_path)
        result = check_config_valid()
        assert result.name == "config_valid"
        assert result.result == CheckResult.PASS
        assert "default" in result.message.lower()


class TestRunHealthChecks:
    """Tests for run_health_checks function."""

    def test_run_all_checks(self, test_config: RTMXConfig):
        """Test running all health checks."""
        report = run_health_checks(test_config)
        assert isinstance(report, HealthReport)
        assert report.status in (HealthStatus.HEALTHY, HealthStatus.DEGRADED, HealthStatus.UNHEALTHY)
        assert len(report.checks) > 0
        assert report.summary["total"] == len(report.checks)

    def test_run_specific_checks(self, test_config: RTMXConfig):
        """Test running specific checks only."""
        report = run_health_checks(test_config, checks_to_run=["rtm_exists", "rtm_loads"])
        assert len(report.checks) == 2
        check_names = {c.name for c in report.checks}
        assert "rtm_exists" in check_names
        assert "rtm_loads" in check_names

    def test_healthy_status(self, test_config: RTMXConfig):
        """Test healthy status when all blocking checks pass."""
        report = run_health_checks(test_config, checks_to_run=["rtm_exists", "rtm_loads"])
        # Core RTM fixture should pass these basic checks
        assert report.status in (HealthStatus.HEALTHY, HealthStatus.DEGRADED)

    def test_summary_counts(self, test_config: RTMXConfig):
        """Test summary counts are correct."""
        report = run_health_checks(test_config)
        total = (
            report.summary["passed"]
            + report.summary["warnings"]
            + report.summary["failed"]
            + report.summary["skipped"]
        )
        assert total == report.summary["total"]
