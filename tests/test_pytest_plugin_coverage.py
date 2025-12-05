"""Comprehensive tests for rtmx.pytest.plugin module.

This module provides extensive test coverage for the pytest plugin:
- RequirementCoverage dataclass and properties
- RTMXPlugin class and methods
- pytest_configure hook for marker registration
- pytest_runtest_makereport hook for test outcome recording
- pytest_terminal_summary for coverage reporting
- Marker registration and usage
"""

from unittest.mock import MagicMock, Mock

import pytest

from rtmx.pytest.plugin import (
    RequirementCoverage,
    RTMXPlugin,
    get_plugin,
    pytest_configure,
    pytest_report_header,
    pytest_terminal_summary,
)


class TestRequirementCoverage:
    """Tests for RequirementCoverage dataclass."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_create_empty_coverage(self):
        """Test creating RequirementCoverage with defaults."""
        cov = RequirementCoverage(req_id="REQ-TEST-001")

        assert cov.req_id == "REQ-TEST-001"
        assert cov.tests == []
        assert cov.passed == 0
        assert cov.failed == 0
        assert cov.skipped == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_total_property_calculation(self):
        """Test that total property sums all outcomes."""
        cov = RequirementCoverage(
            req_id="REQ-TEST-001",
            passed=5,
            failed=2,
            skipped=3,
        )
        assert cov.total == 10

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_total_property_zero(self):
        """Test total property when no tests."""
        cov = RequirementCoverage(req_id="REQ-TEST-001")
        assert cov.total == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_parametric
    @pytest.mark.env_simulation
    def test_status_missing(self):
        """Test status is MISSING when no tests."""
        cov = RequirementCoverage(req_id="REQ-TEST-001")
        assert cov.status == "MISSING"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_parametric
    @pytest.mark.env_simulation
    def test_status_passing_all(self):
        """Test status is PASSING when all tests pass."""
        cov = RequirementCoverage(
            req_id="REQ-TEST-001",
            passed=10,
            failed=0,
            skipped=0,
        )
        assert cov.status == "PASSING"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_parametric
    @pytest.mark.env_simulation
    def test_status_passing_with_skipped(self):
        """Test status is PASSING when some pass and others skipped."""
        cov = RequirementCoverage(
            req_id="REQ-TEST-001",
            passed=5,
            failed=0,
            skipped=3,
        )
        assert cov.status == "PASSING"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_parametric
    @pytest.mark.env_simulation
    def test_status_failing_any_failure(self):
        """Test status is FAILING when any test fails."""
        cov = RequirementCoverage(
            req_id="REQ-TEST-001",
            passed=10,
            failed=1,
            skipped=0,
        )
        assert cov.status == "FAILING"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_parametric
    @pytest.mark.env_simulation
    def test_status_failing_priority_over_passed(self):
        """Test that FAILING status has priority over passed tests."""
        cov = RequirementCoverage(
            req_id="REQ-TEST-001",
            passed=100,
            failed=1,
            skipped=0,
        )
        assert cov.status == "FAILING"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_parametric
    @pytest.mark.env_simulation
    def test_status_skipped_only(self):
        """Test status is SKIPPED when only skipped tests."""
        cov = RequirementCoverage(
            req_id="REQ-TEST-001",
            passed=0,
            failed=0,
            skipped=5,
        )
        assert cov.status == "SKIPPED"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_tests_list_populated(self):
        """Test that tests list can be populated."""
        cov = RequirementCoverage(
            req_id="REQ-TEST-001",
            tests=["test_one.py::test_func", "test_two.py::test_other"],
        )
        assert len(cov.tests) == 2
        assert "test_one.py::test_func" in cov.tests


class TestRTMXPlugin:
    """Tests for RTMXPlugin class."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_plugin_initialization(self):
        """Test RTMXPlugin initializes with empty coverage."""
        plugin = RTMXPlugin()

        assert len(plugin.coverage) == 0
        assert plugin._current_req_ids == []

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_record_test_passed(self):
        """Test recording a passed test outcome."""
        plugin = RTMXPlugin()

        # Create mock test item with req marker
        item = Mock()
        item.nodeid = "tests/test_example.py::test_function"

        marker = Mock()
        marker.args = ("REQ-TEST-001",)
        item.iter_markers = Mock(return_value=[marker])

        plugin.record_test(item, "passed")

        assert "REQ-TEST-001" in plugin.coverage
        assert plugin.coverage["REQ-TEST-001"].passed == 1
        assert plugin.coverage["REQ-TEST-001"].failed == 0
        assert plugin.coverage["REQ-TEST-001"].skipped == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_record_test_failed(self):
        """Test recording a failed test outcome."""
        plugin = RTMXPlugin()

        item = Mock()
        item.nodeid = "tests/test_example.py::test_function"

        marker = Mock()
        marker.args = ("REQ-TEST-001",)
        item.iter_markers = Mock(return_value=[marker])

        plugin.record_test(item, "failed")

        assert "REQ-TEST-001" in plugin.coverage
        assert plugin.coverage["REQ-TEST-001"].passed == 0
        assert plugin.coverage["REQ-TEST-001"].failed == 1
        assert plugin.coverage["REQ-TEST-001"].skipped == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_record_test_skipped(self):
        """Test recording a skipped test outcome."""
        plugin = RTMXPlugin()

        item = Mock()
        item.nodeid = "tests/test_example.py::test_function"

        marker = Mock()
        marker.args = ("REQ-TEST-001",)
        item.iter_markers = Mock(return_value=[marker])

        plugin.record_test(item, "skipped")

        assert "REQ-TEST-001" in plugin.coverage
        assert plugin.coverage["REQ-TEST-001"].passed == 0
        assert plugin.coverage["REQ-TEST-001"].failed == 0
        assert plugin.coverage["REQ-TEST-001"].skipped == 1

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_record_multiple_tests_same_requirement(self):
        """Test recording multiple tests for same requirement."""
        plugin = RTMXPlugin()

        # First test
        item1 = Mock()
        item1.nodeid = "tests/test_example.py::test_one"
        marker1 = Mock()
        marker1.args = ("REQ-TEST-001",)
        item1.iter_markers = Mock(return_value=[marker1])

        # Second test
        item2 = Mock()
        item2.nodeid = "tests/test_example.py::test_two"
        marker2 = Mock()
        marker2.args = ("REQ-TEST-001",)
        item2.iter_markers = Mock(return_value=[marker2])

        plugin.record_test(item1, "passed")
        plugin.record_test(item2, "passed")

        cov = plugin.coverage["REQ-TEST-001"]
        assert cov.passed == 2
        assert len(cov.tests) == 2

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_record_test_multiple_requirements(self):
        """Test recording test with multiple requirement markers."""
        plugin = RTMXPlugin()

        item = Mock()
        item.nodeid = "tests/test_example.py::test_function"

        # Multiple req markers
        marker1 = Mock()
        marker1.args = ("REQ-TEST-001",)
        marker2 = Mock()
        marker2.args = ("REQ-TEST-002",)
        item.iter_markers = Mock(return_value=[marker1, marker2])

        plugin.record_test(item, "passed")

        assert "REQ-TEST-001" in plugin.coverage
        assert "REQ-TEST-002" in plugin.coverage
        assert plugin.coverage["REQ-TEST-001"].passed == 1
        assert plugin.coverage["REQ-TEST-002"].passed == 1

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_record_test_no_markers(self):
        """Test recording test with no req markers does nothing."""
        plugin = RTMXPlugin()

        item = Mock()
        item.nodeid = "tests/test_example.py::test_function"
        item.iter_markers = Mock(return_value=[])

        plugin.record_test(item, "passed")

        assert len(plugin.coverage) == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_record_test_marker_without_args(self):
        """Test recording test with marker but no args."""
        plugin = RTMXPlugin()

        item = Mock()
        item.nodeid = "tests/test_example.py::test_function"

        marker = Mock()
        marker.args = ()  # Empty args
        item.iter_markers = Mock(return_value=[marker])

        plugin.record_test(item, "passed")

        # Should not crash, should skip marker with no args
        assert len(plugin.coverage) == 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_get_coverage_report_empty(self):
        """Test coverage report when no tests recorded."""
        plugin = RTMXPlugin()
        report = plugin.get_coverage_report()

        assert report["summary"]["total_requirements"] == 0
        assert report["summary"]["passing"] == 0
        assert report["summary"]["failing"] == 0
        assert report["summary"]["skipped"] == 0
        assert report["requirements"] == {}

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_get_coverage_report_with_data(self):
        """Test coverage report with recorded tests."""
        plugin = RTMXPlugin()

        # Record some tests
        item1 = Mock()
        item1.nodeid = "tests/test_example.py::test_one"
        marker1 = Mock()
        marker1.args = ("REQ-TEST-001",)
        item1.iter_markers = Mock(return_value=[marker1])

        item2 = Mock()
        item2.nodeid = "tests/test_example.py::test_two"
        marker2 = Mock()
        marker2.args = ("REQ-TEST-002",)
        item2.iter_markers = Mock(return_value=[marker2])

        plugin.record_test(item1, "passed")
        plugin.record_test(item2, "failed")

        report = plugin.get_coverage_report()

        assert report["summary"]["total_requirements"] == 2
        assert report["summary"]["passing"] == 1
        assert report["summary"]["failing"] == 1
        assert report["summary"]["skipped"] == 0

        assert "REQ-TEST-001" in report["requirements"]
        assert report["requirements"]["REQ-TEST-001"]["status"] == "PASSING"
        assert "REQ-TEST-002" in report["requirements"]
        assert report["requirements"]["REQ-TEST-002"]["status"] == "FAILING"

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_get_coverage_report_structure(self):
        """Test that coverage report has expected structure."""
        plugin = RTMXPlugin()

        item = Mock()
        item.nodeid = "tests/test_example.py::test_function"
        marker = Mock()
        marker.args = ("REQ-TEST-001",)
        item.iter_markers = Mock(return_value=[marker])

        plugin.record_test(item, "passed")

        report = plugin.get_coverage_report()

        # Check top-level structure
        assert "requirements" in report
        assert "summary" in report

        # Check requirement details
        req_data = report["requirements"]["REQ-TEST-001"]
        assert "tests" in req_data
        assert "passed" in req_data
        assert "failed" in req_data
        assert "skipped" in req_data
        assert "status" in req_data


class TestGetPlugin:
    """Tests for get_plugin() function."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_get_plugin_returns_instance(self):
        """Test that get_plugin returns the global plugin instance."""
        plugin = get_plugin()

        # Should be initialized by pytest during test run
        assert plugin is not None
        assert isinstance(plugin, RTMXPlugin)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_get_plugin_same_instance(self):
        """Test that get_plugin returns the same instance."""
        plugin1 = get_plugin()
        plugin2 = get_plugin()

        assert plugin1 is plugin2


class TestPytestConfigureHook:
    """Tests for pytest_configure hook."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_pytest_configure_registers_markers(self):
        """Test that pytest_configure registers all markers."""
        config = MagicMock()

        pytest_configure(config)

        # Check that addinivalue_line was called for markers
        calls = config.addinivalue_line.call_args_list
        assert len(calls) > 0

        # Check that req marker is registered
        marker_lines = [str(call) for call in calls]
        assert any("req" in line for line in marker_lines)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_pytest_configure_registers_scope_markers(self):
        """Test that pytest_configure registers scope markers."""
        config = MagicMock()

        pytest_configure(config)

        calls = config.addinivalue_line.call_args_list
        marker_lines = [str(call) for call in calls]

        assert any("scope_unit" in line for line in marker_lines)
        assert any("scope_integration" in line for line in marker_lines)
        assert any("scope_system" in line for line in marker_lines)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_pytest_configure_registers_technique_markers(self):
        """Test that pytest_configure registers technique markers."""
        config = MagicMock()

        pytest_configure(config)

        calls = config.addinivalue_line.call_args_list
        marker_lines = [str(call) for call in calls]

        assert any("technique_nominal" in line for line in marker_lines)
        assert any("technique_parametric" in line for line in marker_lines)
        assert any("technique_monte_carlo" in line for line in marker_lines)
        assert any("technique_stress" in line for line in marker_lines)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_pytest_configure_registers_env_markers(self):
        """Test that pytest_configure registers environment markers."""
        config = MagicMock()

        pytest_configure(config)

        calls = config.addinivalue_line.call_args_list
        marker_lines = [str(call) for call in calls]

        assert any("env_simulation" in line for line in marker_lines)
        assert any("env_hil" in line for line in marker_lines)
        assert any("env_anechoic" in line for line in marker_lines)
        assert any("env_static_field" in line for line in marker_lines)
        assert any("env_dynamic_field" in line for line in marker_lines)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_pytest_configure_creates_plugin_instance(self):
        """Test that pytest_configure creates global plugin instance."""
        config = MagicMock()

        pytest_configure(config)

        plugin = get_plugin()
        assert plugin is not None
        assert isinstance(plugin, RTMXPlugin)


class TestPytestTerminalSummary:
    """Tests for pytest_terminal_summary hook."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_terminal_summary_no_plugin(self):
        """Test terminal summary when plugin is None."""
        # This should not crash
        terminalreporter = MagicMock()
        pytest_terminal_summary(terminalreporter, 0, MagicMock())

        # Should not write anything if plugin is None or has no coverage
        # (actual behavior depends on global state)

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_terminal_summary_empty_coverage(self):
        """Test terminal summary with empty coverage."""
        # Configure plugin with empty coverage
        config = MagicMock()
        pytest_configure(config)

        plugin = get_plugin()
        if plugin:
            plugin.coverage.clear()

        terminalreporter = MagicMock()
        pytest_terminal_summary(terminalreporter, 0, MagicMock())

        # Should not write if no coverage data
        # (exact behavior depends on implementation)


class TestPytestReportHeader:
    """Tests for pytest_report_header hook."""

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_report_header_returns_list(self):
        """Test that pytest_report_header returns a list."""
        config = MagicMock()
        result = pytest_report_header(config)

        assert isinstance(result, list)
        assert len(result) > 0

    @pytest.mark.req("REQ-CORE-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_report_header_contains_rtmx(self):
        """Test that pytest_report_header mentions RTMX."""
        config = MagicMock()
        result = pytest_report_header(config)

        header_text = " ".join(result)
        assert "rtmx" in header_text.lower()


class TestMarkerIntegration:
    """Integration tests for marker usage in actual tests."""

    @pytest.mark.req("REQ-PLUGIN-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_single_req_marker(self):
        """Test using single req marker on test."""
        assert True

    @pytest.mark.req("REQ-PLUGIN-002")
    @pytest.mark.req("REQ-PLUGIN-003")
    @pytest.mark.scope_integration
    @pytest.mark.technique_parametric
    @pytest.mark.env_simulation
    def test_multiple_req_markers(self):
        """Test using multiple req markers on test."""
        assert True

    @pytest.mark.req("REQ-PLUGIN-004")
    @pytest.mark.scope_system
    @pytest.mark.technique_stress
    @pytest.mark.env_hil
    def test_all_marker_types(self):
        """Test using all types of markers together."""
        assert True


@pytest.mark.req("REQ-PLUGIN-005")
class TestMarkerOnClass:
    """Test that req marker works on test class."""

    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_method_one(self):
        """First test method in class."""
        assert True

    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_method_two(self):
        """Second test method in class."""
        assert True
