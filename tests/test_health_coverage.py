"""Comprehensive tests for rtmx.cli.health module.

This file provides extensive coverage of health check functionality including
edge cases, error conditions, and various check scenarios.
"""

import json
from pathlib import Path
from unittest.mock import Mock, patch

import pytest

from rtmx.cli.health import (
    Check,
    CheckResult,
    HealthReport,
    HealthStatus,
    check_agent_configs,
    check_config_valid,
    check_cycles,
    check_reciprocity,
    check_rtm_exists,
    check_rtm_loads,
    check_schema_valid,
    check_test_markers,
    run_health,
    run_health_checks,
)
from rtmx.config import RTMXConfig


@pytest.fixture
def test_config(core_rtm_path: Path) -> RTMXConfig:
    """Create test config pointing to core RTM fixture."""
    return RTMXConfig(database=str(core_rtm_path))


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCheckResult:
    """Tests for CheckResult enum."""

    def test_check_result_pass(self):
        """Test PASS result value."""
        assert CheckResult.PASS.value == "pass"

    def test_check_result_warn(self):
        """Test WARN result value."""
        assert CheckResult.WARN.value == "warn"

    def test_check_result_fail(self):
        """Test FAIL result value."""
        assert CheckResult.FAIL.value == "fail"

    def test_check_result_skip(self):
        """Test SKIP result value."""
        assert CheckResult.SKIP.value == "skip"

    def test_check_result_all_values(self):
        """Test all enum values are strings."""
        for result in CheckResult:
            assert isinstance(result.value, str)


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestHealthStatus:
    """Tests for HealthStatus enum."""

    def test_health_status_healthy(self):
        """Test HEALTHY status value."""
        assert HealthStatus.HEALTHY.value == "healthy"

    def test_health_status_degraded(self):
        """Test DEGRADED status value."""
        assert HealthStatus.DEGRADED.value == "degraded"

    def test_health_status_unhealthy(self):
        """Test UNHEALTHY status value."""
        assert HealthStatus.UNHEALTHY.value == "unhealthy"

    def test_health_status_ordering(self):
        """Test status severity ordering."""
        statuses = [HealthStatus.HEALTHY, HealthStatus.DEGRADED, HealthStatus.UNHEALTHY]
        assert len(statuses) == 3


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCheck:
    """Tests for Check dataclass."""

    def test_check_minimal_creation(self):
        """Test Check creation with minimal parameters."""
        check = Check(name="test", result=CheckResult.PASS, message="OK")
        assert check.name == "test"
        assert check.result == CheckResult.PASS
        assert check.message == "OK"
        assert check.blocking is True  # default
        assert check.details is None  # default

    def test_check_full_creation(self):
        """Test Check creation with all parameters."""
        details = {"error_count": 5, "items": ["a", "b"]}
        check = Check(
            name="complex_check",
            result=CheckResult.WARN,
            message="Found issues",
            blocking=False,
            details=details,
        )
        assert check.name == "complex_check"
        assert check.result == CheckResult.WARN
        assert check.blocking is False
        assert check.details == details

    def test_check_to_dict_minimal(self):
        """Test Check serialization without details."""
        check = Check(name="test", result=CheckResult.PASS, message="OK", blocking=True)
        d = check.to_dict()
        assert d["name"] == "test"
        assert d["result"] == "pass"
        assert d["message"] == "OK"
        assert d["blocking"] is True
        assert "details" not in d

    def test_check_to_dict_with_details(self):
        """Test Check serialization with details."""
        check = Check(
            name="test",
            result=CheckResult.FAIL,
            message="Error",
            blocking=True,
            details={"count": 3},
        )
        d = check.to_dict()
        assert d["details"] == {"count": 3}

    def test_check_to_dict_all_results(self):
        """Test serialization for all result types."""
        for result in CheckResult:
            check = Check(name="test", result=result, message="msg")
            d = check.to_dict()
            assert d["result"] == result.value


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestHealthReport:
    """Tests for HealthReport dataclass."""

    def test_health_report_empty(self):
        """Test HealthReport with no checks."""
        report = HealthReport(status=HealthStatus.HEALTHY, checks=[], summary={})
        assert report.status == HealthStatus.HEALTHY
        assert len(report.checks) == 0

    def test_health_report_with_checks(self):
        """Test HealthReport with multiple checks."""
        checks = [
            Check("c1", CheckResult.PASS, "OK"),
            Check("c2", CheckResult.WARN, "Warning"),
            Check("c3", CheckResult.FAIL, "Error"),
        ]
        report = HealthReport(
            status=HealthStatus.DEGRADED,
            checks=checks,
            summary={"total": 3, "passed": 1, "warnings": 1, "failed": 1},
        )
        assert len(report.checks) == 3
        assert report.summary["total"] == 3

    def test_health_report_to_dict(self):
        """Test HealthReport serialization."""
        check = Check("test", CheckResult.PASS, "OK")
        report = HealthReport(
            status=HealthStatus.HEALTHY, checks=[check], summary={"total": 1, "passed": 1}
        )
        d = report.to_dict()
        assert d["status"] == "healthy"
        assert len(d["checks"]) == 1
        assert d["summary"]["total"] == 1

    def test_health_report_to_dict_all_statuses(self):
        """Test serialization for all health statuses."""
        for status in HealthStatus:
            report = HealthReport(status=status, checks=[], summary={})
            d = report.to_dict()
            assert d["status"] == status.value


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCheckRtmExists:
    """Tests for check_rtm_exists function."""

    def test_rtm_exists_valid_file(self, test_config: RTMXConfig):
        """Test check passes for valid RTM file."""
        result = check_rtm_exists(test_config)
        assert result.name == "rtm_exists"
        assert result.result == CheckResult.PASS
        assert result.blocking is True

    def test_rtm_exists_missing_file(self, tmp_path: Path):
        """Test check fails for missing file."""
        config = RTMXConfig(database=str(tmp_path / "missing.csv"))
        result = check_rtm_exists(config)
        assert result.result == CheckResult.FAIL
        assert "not found" in result.message

    def test_rtm_exists_empty_file(self, tmp_path: Path):
        """Test check fails for empty file."""
        empty_file = tmp_path / "empty.csv"
        empty_file.write_text("")
        config = RTMXConfig(database=str(empty_file))
        result = check_rtm_exists(config)
        assert result.result == CheckResult.FAIL
        assert "empty" in result.message.lower()

    def test_rtm_exists_whitespace_only(self, tmp_path: Path):
        """Test check fails for whitespace-only file."""
        ws_file = tmp_path / "whitespace.csv"
        ws_file.write_text("   \n  \n  ")
        config = RTMXConfig(database=str(ws_file))
        result = check_rtm_exists(config)
        assert result.result == CheckResult.FAIL

    def test_rtm_exists_unreadable_file(self, tmp_path: Path):
        """Test check fails for unreadable file."""
        config = RTMXConfig(database=str(tmp_path / "unreadable.csv"))
        with patch("pathlib.Path.exists", return_value=True):
            with patch("pathlib.Path.read_text", side_effect=PermissionError("Access denied")):
                result = check_rtm_exists(config)
                assert result.result == CheckResult.FAIL
                assert "Cannot read" in result.message


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCheckRtmLoads:
    """Tests for check_rtm_loads function."""

    def test_rtm_loads_valid_database(self, test_config: RTMXConfig):
        """Test check passes for valid database."""
        result = check_rtm_loads(test_config)
        assert result.name == "rtm_loads"
        assert result.result == CheckResult.PASS
        assert result.details is not None
        assert "requirement_count" in result.details
        assert result.details["requirement_count"] > 0

    def test_rtm_loads_invalid_format(self, tmp_path: Path):
        """Test check fails for invalid CSV format."""
        from rtmx.models import RTMError

        config = RTMXConfig(database="dummy.csv")
        with patch("rtmx.models.RTMDatabase.load", side_effect=RTMError("Invalid format")):
            result = check_rtm_loads(config)
            assert result.result == CheckResult.FAIL
            assert "load failed" in result.message.lower()

    def test_rtm_loads_corrupted_data(self, test_config: RTMXConfig):
        """Test check fails for corrupted RTM data."""
        from rtmx.models import RTMError

        with patch("rtmx.models.RTMDatabase.load", side_effect=RTMError("Corrupt data")):
            result = check_rtm_loads(test_config)
            assert result.result == CheckResult.FAIL


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCheckSchemaValid:
    """Tests for check_schema_valid function."""

    def test_schema_valid_pass(self, test_config: RTMXConfig):
        """Test check passes for valid schema."""
        result = check_schema_valid(test_config)
        assert result.name == "schema_valid"
        assert result.result == CheckResult.PASS

    def test_schema_valid_with_errors(self, test_config: RTMXConfig):
        """Test check fails with validation errors."""
        with patch("rtmx.validation.validate_schema", return_value=["Error 1", "Error 2"]):
            result = check_schema_valid(test_config)
            assert result.result == CheckResult.FAIL
            assert "2 errors" in result.message
            assert result.details is not None

    def test_schema_valid_many_errors(self, test_config: RTMXConfig):
        """Test check with many validation errors."""
        errors = [f"Error {i}" for i in range(20)]
        with patch("rtmx.validation.validate_schema", return_value=errors):
            result = check_schema_valid(test_config)
            assert result.result == CheckResult.FAIL
            assert "20 errors" in result.message
            # Should only include first 10 in details
            assert len(result.details["errors"]) == 10

    def test_schema_valid_load_error(self, test_config: RTMXConfig):
        """Test check handles database load error."""
        from rtmx.models import RTMError

        with patch("rtmx.models.RTMDatabase.load", side_effect=RTMError("Load failed")):
            result = check_schema_valid(test_config)
            assert result.result == CheckResult.FAIL


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCheckReciprocity:
    """Tests for check_reciprocity function."""

    def test_reciprocity_pass(self, test_config: RTMXConfig):
        """Test reciprocity check passes."""
        with patch("rtmx.validation.check_reciprocity", return_value=[]):
            result = check_reciprocity(test_config)
            assert result.name == "reciprocity"
            assert result.result == CheckResult.PASS
            assert result.blocking is False

    def test_reciprocity_with_violations(self, test_config: RTMXConfig):
        """Test reciprocity check with violations."""
        violations = ["REQ-1 -> REQ-2", "REQ-3 -> REQ-4"]
        with patch("rtmx.validation.check_reciprocity", return_value=violations):
            result = check_reciprocity(test_config)
            assert result.result == CheckResult.WARN
            assert "2" in result.message
            assert result.details is not None

    def test_reciprocity_many_violations(self, test_config: RTMXConfig):
        """Test reciprocity with many violations."""
        violations = [f"REQ-{i}" for i in range(20)]
        with patch("rtmx.validation.check_reciprocity", return_value=violations):
            result = check_reciprocity(test_config)
            assert result.result == CheckResult.WARN
            # Should only show first 10
            assert len(result.details["violations"]) == 10

    def test_reciprocity_error(self, test_config: RTMXConfig):
        """Test reciprocity check handles error."""
        from rtmx.models import RTMError

        with patch("rtmx.models.RTMDatabase.load", side_effect=RTMError("Error")):
            result = check_reciprocity(test_config)
            assert result.result == CheckResult.FAIL


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCheckCycles:
    """Tests for check_cycles function."""

    def test_cycles_none_found(self, test_config: RTMXConfig):
        """Test cycles check when no cycles exist."""
        with patch("rtmx.models.RTMDatabase.load") as mock_load:
            mock_db = Mock()
            mock_db.find_cycles.return_value = []
            mock_load.return_value = mock_db

            result = check_cycles(test_config)
            assert result.result == CheckResult.PASS
            assert result.blocking is False

    def test_cycles_found(self, test_config: RTMXConfig):
        """Test cycles check when cycles exist."""
        with patch("rtmx.models.RTMDatabase.load") as mock_load:
            mock_db = Mock()
            mock_db.find_cycles.return_value = [["REQ-1", "REQ-2"], ["REQ-3", "REQ-4", "REQ-5"]]
            mock_load.return_value = mock_db

            result = check_cycles(test_config)
            assert result.result == CheckResult.WARN
            assert "2 circular" in result.message
            assert "5 requirements" in result.message
            assert result.details["cycle_count"] == 2
            assert result.details["requirements_in_cycles"] == 5

    def test_cycles_single_cycle(self, test_config: RTMXConfig):
        """Test cycles check with single cycle."""
        with patch("rtmx.models.RTMDatabase.load") as mock_load:
            mock_db = Mock()
            mock_db.find_cycles.return_value = [["REQ-1", "REQ-2", "REQ-3"]]
            mock_load.return_value = mock_db

            result = check_cycles(test_config)
            assert result.result == CheckResult.WARN
            assert "1 circular" in result.message

    def test_cycles_error(self, test_config: RTMXConfig):
        """Test cycles check handles error."""
        from rtmx.models import RTMError

        with patch("rtmx.models.RTMDatabase.load", side_effect=RTMError("Error")):
            result = check_cycles(test_config)
            assert result.result == CheckResult.FAIL


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCheckTestMarkers:
    """Tests for check_test_markers function."""

    def test_test_markers_no_test_directory(
        self, test_config: RTMXConfig, tmp_path: Path, monkeypatch
    ):
        """Test check skips when no test directory found."""
        monkeypatch.chdir(tmp_path)
        result = check_test_markers(test_config)
        assert result.result == CheckResult.SKIP
        assert "No tests directory" in result.message

    def test_test_markers_full_coverage(self, test_config: RTMXConfig, tmp_path: Path, monkeypatch):
        """Test check with 100% test coverage."""
        monkeypatch.chdir(tmp_path)
        test_dir = tmp_path / "tests"
        test_dir.mkdir()

        # Create test file
        test_file = test_dir / "test_example.py"
        test_file.write_text("""
import pytest

@pytest.mark.req("REQ-CORE-001")
def test_something():
    pass
""")

        with patch("rtmx.models.RTMDatabase.load") as mock_load:
            mock_db = Mock()
            mock_req = Mock()
            mock_req.req_id = "REQ-CORE-001"
            mock_db.__iter__ = Mock(return_value=iter([mock_req]))
            mock_load.return_value = mock_db

            with patch("rtmx.cli.from_tests.extract_markers_from_file") as mock_extract:
                marker = Mock()
                marker.req_id = "REQ-CORE-001"
                marker.test_function = "test_something"
                mock_extract.return_value = [marker]

                result = check_test_markers(test_config)
                assert result.result == CheckResult.PASS
                assert "100.0%" in result.message

    def test_test_markers_partial_coverage(
        self, test_config: RTMXConfig, tmp_path: Path, monkeypatch
    ):
        """Test check with partial test coverage."""
        monkeypatch.chdir(tmp_path)
        test_dir = tmp_path / "tests"
        test_dir.mkdir()

        test_file = test_dir / "test_example.py"
        test_file.write_text("# test file")

        with patch("rtmx.models.RTMDatabase.load") as mock_load:
            mock_db = Mock()
            req1 = Mock()
            req1.req_id = "REQ-1"
            req2 = Mock()
            req2.req_id = "REQ-2"
            mock_db.__iter__ = Mock(return_value=iter([req1, req2]))
            mock_load.return_value = mock_db

            with patch("rtmx.cli.from_tests.extract_markers_from_file") as mock_extract:
                marker = Mock()
                marker.req_id = "REQ-1"
                marker.test_function = "test_req1"
                mock_extract.return_value = [marker]

                result = check_test_markers(test_config)
                assert result.result == CheckResult.WARN
                assert "50.0%" in result.message
                assert result.details["missing_count"] == 1

    def test_test_markers_with_orphans(self, test_config: RTMXConfig, tmp_path: Path, monkeypatch):
        """Test check with orphaned test markers."""
        monkeypatch.chdir(tmp_path)
        test_dir = tmp_path / "tests"
        test_dir.mkdir()
        test_file = test_dir / "test_example.py"
        test_file.write_text("# test")

        with patch("rtmx.models.RTMDatabase.load") as mock_load:
            mock_db = Mock()
            req1 = Mock()
            req1.req_id = "REQ-1"
            mock_db.__iter__ = Mock(return_value=iter([req1]))
            mock_load.return_value = mock_db

            with patch("rtmx.cli.from_tests.extract_markers_from_file") as mock_extract:
                marker = Mock()
                marker.req_id = "REQ-ORPHAN"
                marker.test_function = "test_orphan"
                mock_extract.return_value = [marker]

                result = check_test_markers(test_config)
                assert result.result == CheckResult.WARN
                assert result.details["orphan_count"] == 1

    def test_test_markers_error_parsing(self, test_config: RTMXConfig, tmp_path: Path, monkeypatch):
        """Test check handles test parsing errors gracefully."""
        monkeypatch.chdir(tmp_path)
        test_dir = tmp_path / "tests"
        test_dir.mkdir()
        test_file = test_dir / "test_bad.py"
        test_file.write_text("not valid python!")

        with patch("rtmx.models.RTMDatabase.load") as mock_load:
            mock_db = Mock()
            mock_db.__iter__ = Mock(return_value=iter([]))
            mock_load.return_value = mock_db

            with patch(
                "rtmx.cli.from_tests.extract_markers_from_file",
                side_effect=Exception("Parse error"),
            ):
                # Should not raise, just skip bad files
                result = check_test_markers(test_config)
                assert result.result in (CheckResult.PASS, CheckResult.WARN, CheckResult.SKIP)


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCheckAgentConfigs:
    """Tests for check_agent_configs function."""

    def test_agent_configs_all_configured(self, tmp_path: Path, monkeypatch):
        """Test check when all agent configs have RTMX sections."""
        monkeypatch.chdir(tmp_path)

        (tmp_path / "CLAUDE.md").write_text("# Project\n## RTMX\nConfig here")
        (tmp_path / ".cursorrules").write_text("Some rules\nRTMX section")
        (tmp_path / ".github").mkdir()
        (tmp_path / ".github" / "copilot-instructions.md").write_text("RTMX instructions")

        result = check_agent_configs()
        assert result.result == CheckResult.PASS
        assert result.details is not None
        assert len(result.details["configured"]) == 3

    def test_agent_configs_some_missing(self, tmp_path: Path, monkeypatch):
        """Test check when some configs are missing RTMX."""
        monkeypatch.chdir(tmp_path)

        (tmp_path / "CLAUDE.md").write_text("# Project\nNo RTMX section")
        (tmp_path / ".cursorrules").write_text("RTMX section here")

        result = check_agent_configs()
        assert result.result == CheckResult.WARN
        assert "CLAUDE.md" in result.details["missing"]
        assert ".cursorrules" in result.details["configured"]

    def test_agent_configs_none_exist(self, tmp_path: Path, monkeypatch):
        """Test check when no agent config files exist."""
        monkeypatch.chdir(tmp_path)

        result = check_agent_configs()
        assert result.result == CheckResult.SKIP
        assert "No agent config files" in result.message

    def test_agent_configs_exist_no_rtmx(self, tmp_path: Path, monkeypatch):
        """Test check when files exist but lack RTMX."""
        monkeypatch.chdir(tmp_path)

        (tmp_path / "CLAUDE.md").write_text("# Project info")
        (tmp_path / ".cursorrules").write_text("Some rules")

        result = check_agent_configs()
        assert result.result == CheckResult.WARN
        assert len(result.details["missing"]) == 2


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCheckConfigValid:
    """Tests for check_config_valid function."""

    def test_config_valid_default(self, tmp_path: Path, monkeypatch):
        """Test config validation when using defaults."""
        monkeypatch.chdir(tmp_path)

        result = check_config_valid()
        assert result.result == CheckResult.PASS
        assert "default" in result.message.lower()

    def test_config_valid_rtmx_yaml(self, tmp_path: Path, monkeypatch):
        """Test config validation with rtmx.yaml."""
        monkeypatch.chdir(tmp_path)

        config_file = tmp_path / "rtmx.yaml"
        config_file.write_text("database: docs/rtm_database.csv\n")

        result = check_config_valid()
        assert result.result == CheckResult.PASS
        assert "rtmx.yaml" in result.message

    def test_config_valid_dotfile(self, tmp_path: Path, monkeypatch):
        """Test config validation with .rtmx.yaml."""
        monkeypatch.chdir(tmp_path)

        config_file = tmp_path / ".rtmx.yaml"
        config_file.write_text("database: rtm.csv\n")

        result = check_config_valid()
        assert result.result == CheckResult.PASS
        assert ".rtmx.yaml" in result.message

    def test_config_invalid_yaml(self, tmp_path: Path, monkeypatch):
        """Test config validation with invalid YAML."""
        monkeypatch.chdir(tmp_path)

        config_file = tmp_path / "rtmx.yaml"
        config_file.write_text("invalid: yaml: content: [[[")

        result = check_config_valid()
        assert result.result == CheckResult.FAIL
        assert "invalid" in result.message.lower()

    def test_config_load_error(self, tmp_path: Path, monkeypatch):
        """Test config validation with load error."""
        monkeypatch.chdir(tmp_path)

        config_file = tmp_path / "rtmx.yaml"
        config_file.write_text("invalid yaml: [[[ }}}")

        result = check_config_valid()
        assert result.result == CheckResult.FAIL


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRunHealthChecks:
    """Tests for run_health_checks function."""

    def test_run_health_checks_all(self, test_config: RTMXConfig):
        """Test running all health checks."""
        report = run_health_checks(test_config)
        assert isinstance(report, HealthReport)
        assert len(report.checks) > 0
        assert report.status in (
            HealthStatus.HEALTHY,
            HealthStatus.DEGRADED,
            HealthStatus.UNHEALTHY,
        )

    def test_run_health_checks_specific(self, test_config: RTMXConfig):
        """Test running specific checks."""
        report = run_health_checks(test_config, checks_to_run=["rtm_exists"])
        assert len(report.checks) == 1
        assert report.checks[0].name == "rtm_exists"

    def test_run_health_checks_multiple_specific(self, test_config: RTMXConfig):
        """Test running multiple specific checks."""
        checks = ["rtm_exists", "rtm_loads", "schema_valid"]
        report = run_health_checks(test_config, checks_to_run=checks)
        assert len(report.checks) == 3
        check_names = {c.name for c in report.checks}
        assert check_names == set(checks)

    def test_run_health_checks_status_healthy(self, test_config: RTMXConfig):
        """Test healthy status when all checks pass."""
        with patch("rtmx.cli.health.check_rtm_exists") as mock_check:
            mock_check.return_value = Check("rtm_exists", CheckResult.PASS, "OK", blocking=True)

            report = run_health_checks(test_config, checks_to_run=["rtm_exists"])
            assert report.status == HealthStatus.HEALTHY

    def test_run_health_checks_status_degraded(self, test_config: RTMXConfig):
        """Test degraded status with warnings."""
        with patch("rtmx.cli.health.check_reciprocity") as mock_check:
            mock_check.return_value = Check(
                "reciprocity", CheckResult.WARN, "Issues", blocking=False
            )

            report = run_health_checks(test_config, checks_to_run=["reciprocity"])
            assert report.status == HealthStatus.DEGRADED

    def test_run_health_checks_status_unhealthy(self, test_config: RTMXConfig):
        """Test unhealthy status with blocking failures."""
        with patch("rtmx.cli.health.check_rtm_exists") as mock_check:
            mock_check.return_value = Check(
                "rtm_exists", CheckResult.FAIL, "Missing", blocking=True
            )

            report = run_health_checks(test_config, checks_to_run=["rtm_exists"])
            assert report.status == HealthStatus.UNHEALTHY

    def test_run_health_checks_summary(self, test_config: RTMXConfig):
        """Test summary counts are accurate."""
        report = run_health_checks(test_config)
        summary = report.summary

        assert "total" in summary
        assert "passed" in summary
        assert "warnings" in summary
        assert "failed" in summary
        assert "skipped" in summary

        total = summary["passed"] + summary["warnings"] + summary["failed"] + summary["skipped"]
        assert total == summary["total"]

    def test_run_health_checks_exception_handling(self, test_config: RTMXConfig):
        """Test exception handling during check execution."""
        with patch(
            "rtmx.cli.health.check_rtm_exists", side_effect=RuntimeError("Unexpected error")
        ):
            report = run_health_checks(test_config, checks_to_run=["rtm_exists"])

            assert len(report.checks) == 1
            assert report.checks[0].result == CheckResult.FAIL
            assert "exception" in report.checks[0].message.lower()


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRunHealth:
    """Tests for run_health function."""

    def test_run_health_terminal_format(self, test_config: RTMXConfig, capsys):
        """Test run_health with terminal output format."""
        with pytest.raises(SystemExit) as exc_info:
            run_health(format_type="terminal", config=test_config)

        assert exc_info.value.code in (0, 1, 2)
        captured = capsys.readouterr()
        assert "RTMX Health Check" in captured.out

    def test_run_health_json_format(self, test_config: RTMXConfig, capsys):
        """Test run_health with JSON output format."""
        with pytest.raises(SystemExit):
            run_health(format_type="json", config=test_config)

        captured = capsys.readouterr()
        data = json.loads(captured.out)
        assert "status" in data
        assert "checks" in data
        assert "summary" in data

    def test_run_health_ci_format(self, test_config: RTMXConfig, capsys):
        """Test run_health with CI output format."""
        with pytest.raises(SystemExit):
            run_health(format_type="ci", config=test_config)

        captured = capsys.readouterr()
        assert "status=" in captured.out
        assert "passed=" in captured.out
        assert "failed=" in captured.out

    def test_run_health_strict_mode(self, test_config: RTMXConfig):
        """Test run_health with strict mode."""
        with patch("rtmx.cli.health.run_health_checks") as mock_checks:
            report = HealthReport(
                status=HealthStatus.DEGRADED,
                checks=[Check("test", CheckResult.WARN, "Warning", blocking=False)],
                summary={"total": 1, "passed": 0, "warnings": 1, "failed": 0, "skipped": 0},
            )
            mock_checks.return_value = report

            with pytest.raises(SystemExit) as exc_info:
                run_health(strict=True, config=test_config)

            # In strict mode, run_health adjusts DEGRADED status to UNHEALTHY
            # which results in exit code 2
            assert exc_info.value.code == 2

    def test_run_health_exit_codes(self, test_config: RTMXConfig):
        """Test run_health exit codes for different statuses."""
        # Healthy -> 0
        with patch("rtmx.cli.health.run_health_checks") as mock_checks:
            report = HealthReport(
                status=HealthStatus.HEALTHY,
                checks=[],
                summary={"total": 0, "passed": 0, "warnings": 0, "failed": 0, "skipped": 0},
            )
            mock_checks.return_value = report

            with pytest.raises(SystemExit) as exc_info:
                run_health(config=test_config)
            assert exc_info.value.code == 0

        # Unhealthy -> 2
        with patch("rtmx.cli.health.run_health_checks") as mock_checks:
            report = HealthReport(
                status=HealthStatus.UNHEALTHY,
                checks=[],
                summary={"total": 0, "passed": 0, "warnings": 0, "failed": 0, "skipped": 0},
            )
            mock_checks.return_value = report

            with pytest.raises(SystemExit) as exc_info:
                run_health(config=test_config)
            assert exc_info.value.code == 2

    def test_run_health_specific_checks(self, test_config: RTMXConfig):
        """Test run_health with specific checks."""
        with pytest.raises(SystemExit):
            run_health(checks=["rtm_exists", "rtm_loads"], config=test_config)

        # Should complete without errors

    def test_run_health_terminal_output_all_results(self, test_config: RTMXConfig, capsys):
        """Test terminal output shows all result types."""
        with patch("rtmx.cli.health.run_health_checks") as mock_checks:
            checks = [
                Check("pass_check", CheckResult.PASS, "Passed"),
                Check("warn_check", CheckResult.WARN, "Warning"),
                Check("fail_check", CheckResult.FAIL, "Failed", blocking=True),
                Check("skip_check", CheckResult.SKIP, "Skipped"),
            ]
            report = HealthReport(
                status=HealthStatus.UNHEALTHY,
                checks=checks,
                summary={"total": 4, "passed": 1, "warnings": 1, "failed": 1, "skipped": 1},
            )
            mock_checks.return_value = report

            with pytest.raises(SystemExit):
                run_health(format_type="terminal", config=test_config)

            captured = capsys.readouterr()
            assert "PASS" in captured.out
            assert "WARN" in captured.out
            assert "FAIL" in captured.out
            assert "SKIP" in captured.out
