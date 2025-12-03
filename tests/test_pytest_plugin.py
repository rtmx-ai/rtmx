"""Tests for rtmx.pytest.plugin module."""

import pytest

from rtmx.pytest.plugin import RequirementCoverage, RTMXPlugin, get_plugin


class TestRequirementCoverage:
    """Tests for RequirementCoverage dataclass."""

    def test_total(self):
        """Test total property calculation."""
        cov = RequirementCoverage(
            req_id="REQ-TEST-001",
            passed=3,
            failed=1,
            skipped=2,
        )
        assert cov.total == 6

    def test_status_missing(self):
        """Test status when no tests run."""
        cov = RequirementCoverage(req_id="REQ-TEST-001")
        assert cov.status == "MISSING"

    def test_status_passing(self):
        """Test status when all tests pass."""
        cov = RequirementCoverage(
            req_id="REQ-TEST-001",
            passed=5,
            failed=0,
            skipped=0,
        )
        assert cov.status == "PASSING"

    def test_status_failing(self):
        """Test status when any test fails."""
        cov = RequirementCoverage(
            req_id="REQ-TEST-001",
            passed=4,
            failed=1,
            skipped=0,
        )
        assert cov.status == "FAILING"

    def test_status_skipped_only(self):
        """Test status when all tests skipped."""
        cov = RequirementCoverage(
            req_id="REQ-TEST-001",
            passed=0,
            failed=0,
            skipped=3,
        )
        assert cov.status == "SKIPPED"


class TestRTMXPlugin:
    """Tests for RTMXPlugin class."""

    def test_init(self):
        """Test plugin initialization."""
        plugin = RTMXPlugin()
        assert len(plugin.coverage) == 0

    def test_get_coverage_report_empty(self):
        """Test coverage report when empty."""
        plugin = RTMXPlugin()
        report = plugin.get_coverage_report()

        assert report["summary"]["total_requirements"] == 0
        assert report["summary"]["passing"] == 0
        assert report["summary"]["failing"] == 0
        assert report["summary"]["skipped"] == 0


class TestGetPlugin:
    """Tests for get_plugin function."""

    def test_get_plugin_returns_instance(self):
        """Test that get_plugin returns the global instance."""
        plugin = get_plugin()
        # Plugin should be initialized by pytest_configure
        assert plugin is not None
        assert isinstance(plugin, RTMXPlugin)


# Test with actual markers
@pytest.mark.req("REQ-PLUGIN-001")
def test_marker_registration():
    """Test that req marker is registered and usable."""
    assert True


@pytest.mark.req("REQ-PLUGIN-002")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_multiple_markers():
    """Test that all RTM markers can be used together."""
    assert True


@pytest.mark.req("REQ-PLUGIN-003")
class TestMarkerOnClass:
    """Test that markers work on test classes."""

    def test_method_one(self):
        """First test method."""
        assert True

    def test_method_two(self):
        """Second test method."""
        assert True
