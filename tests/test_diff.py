"""Tests for rtmx.cli.diff module."""

from pathlib import Path

import pytest

from rtmx.cli.diff import format_terminal_report
from rtmx.comparison import compare_databases


@pytest.fixture
def baseline_csv(tmp_path: Path) -> Path:
    """Create a baseline RTM CSV for testing."""
    csv_content = """req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file
REQ-001,SOFTWARE,CORE,First requirement,Value,tests/test.py,test_func,Unit Test,COMPLETE,HIGH,1,Notes,1.0,,,alice,v0.1,2025-01-01,2025-01-15,docs/req.md
REQ-002,SOFTWARE,CORE,Second requirement,Value,tests/test.py,test_func2,Unit Test,PARTIAL,MEDIUM,1,Notes,0.5,REQ-001,,bob,v0.1,2025-01-10,,docs/req2.md
REQ-003,SOFTWARE,CORE,Third requirement,Value,tests/test.py,test_func3,Unit Test,MISSING,LOW,2,Notes,2.0,REQ-002,,charlie,v0.2,,,docs/req3.md
"""
    path = tmp_path / "baseline.csv"
    path.write_text(csv_content)
    return path


@pytest.fixture
def current_csv(tmp_path: Path) -> Path:
    """Create a current RTM CSV for testing (with changes from baseline)."""
    csv_content = """req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file
REQ-001,SOFTWARE,CORE,First requirement,Value,tests/test.py,test_func,Unit Test,COMPLETE,HIGH,1,Notes,1.0,,,alice,v0.1,2025-01-01,2025-01-15,docs/req.md
REQ-002,SOFTWARE,CORE,Second requirement,Value,tests/test.py,test_func2,Unit Test,COMPLETE,MEDIUM,1,Notes,0.5,REQ-001,,bob,v0.1,2025-01-10,2025-01-20,docs/req2.md
REQ-004,SOFTWARE,NEW,New requirement,Value,tests/test.py,test_func4,Unit Test,MISSING,HIGH,1,Notes,1.0,,,dave,v0.2,,,docs/req4.md
"""
    path = tmp_path / "current.csv"
    path.write_text(csv_content)
    return path


class TestFormatTerminalReport:
    """Tests for format_terminal_report function."""

    def test_format_terminal_report_contains_header(self, baseline_csv: Path, current_csv: Path):
        """Test terminal report contains header."""
        report = compare_databases(baseline_csv, current_csv)
        output = format_terminal_report(report)
        assert "RTM Comparison" in output

    def test_format_terminal_report_contains_status(self, baseline_csv: Path, current_csv: Path):
        """Test terminal report contains status."""
        report = compare_databases(baseline_csv, current_csv)
        output = format_terminal_report(report)
        # Should contain one of the status types
        assert any(
            status in output.upper()
            for status in ["BREAKING", "REGRESSED", "DEGRADED", "IMPROVED", "STABLE"]
        )

    def test_format_terminal_report_contains_paths(self, baseline_csv: Path, current_csv: Path):
        """Test terminal report contains file paths."""
        report = compare_databases(baseline_csv, current_csv)
        output = format_terminal_report(report)
        assert "Baseline" in output
        assert "Current" in output

    def test_format_terminal_report_contains_metrics(self, baseline_csv: Path, current_csv: Path):
        """Test terminal report contains key metrics."""
        report = compare_databases(baseline_csv, current_csv)
        output = format_terminal_report(report)
        assert "Requirements" in output
        assert "Completion" in output

    def test_format_terminal_report_shows_added(self, baseline_csv: Path, current_csv: Path):
        """Test terminal report shows added requirements."""
        report = compare_databases(baseline_csv, current_csv)
        output = format_terminal_report(report)
        if report.added_requirements:
            assert "Added Requirements" in output
            assert "REQ-004" in output

    def test_format_terminal_report_shows_removed(self, baseline_csv: Path, current_csv: Path):
        """Test terminal report shows removed requirements."""
        report = compare_databases(baseline_csv, current_csv)
        output = format_terminal_report(report)
        if report.removed_requirements:
            assert "Removed Requirements" in output
            assert "REQ-003" in output

    def test_format_terminal_report_shows_status_changes(
        self, baseline_csv: Path, current_csv: Path
    ):
        """Test terminal report shows status changes."""
        report = compare_databases(baseline_csv, current_csv)
        output = format_terminal_report(report)
        # REQ-002 changed from PARTIAL to COMPLETE
        if any(sc.change_type == "changed" for sc in report.status_changes):
            assert "Status Changes" in output


class TestDiffOutputFormats:
    """Tests for different output formats."""

    def test_json_output_format(self, baseline_csv: Path, current_csv: Path):
        """Test JSON output format."""
        report = compare_databases(baseline_csv, current_csv)
        d = report.to_dict()

        # Verify JSON structure (nested format)
        assert "summary" in d
        assert isinstance(d["summary"]["status"], str)
        assert "baseline" in d
        assert isinstance(d["baseline"]["req_count"], int)
        assert "current" in d
        assert isinstance(d["current"]["req_count"], int)
        assert "changes" in d
        assert isinstance(d["changes"]["added"], list)
        assert isinstance(d["changes"]["removed"], list)

    def test_markdown_output_format(self, baseline_csv: Path, current_csv: Path):
        """Test Markdown output format."""
        report = compare_databases(baseline_csv, current_csv)
        md = report.to_markdown()

        # Verify Markdown structure
        assert md.startswith("#")
        assert "|" in md  # Table formatting


class TestDiffExitCodes:
    """Tests for diff exit code logic."""

    def test_breaking_change_status(self, baseline_csv: Path, tmp_path: Path):
        """Test breaking status when requirements removed."""
        # Create current with removed requirements
        csv_content = """req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file
REQ-001,SOFTWARE,CORE,First requirement,Value,tests/test.py,test_func,Unit Test,COMPLETE,HIGH,1,Notes,1.0,,,alice,v0.1,2025-01-01,2025-01-15,docs/req.md
"""
        current = tmp_path / "fewer.csv"
        current.write_text(csv_content)

        report = compare_databases(baseline_csv, current)
        assert report.summary_status == "breaking"

    def test_stable_status(self, baseline_csv: Path):
        """Test stable status when no changes."""
        report = compare_databases(baseline_csv, baseline_csv)
        assert report.summary_status == "stable"

    def test_improved_status(self, baseline_csv: Path, tmp_path: Path):
        """Test improved status when completion increases."""
        # Create current with better completion
        csv_content = """req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file
REQ-001,SOFTWARE,CORE,First requirement,Value,tests/test.py,test_func,Unit Test,COMPLETE,HIGH,1,Notes,1.0,,,alice,v0.1,2025-01-01,2025-01-15,docs/req.md
REQ-002,SOFTWARE,CORE,Second requirement,Value,tests/test.py,test_func2,Unit Test,COMPLETE,MEDIUM,1,Notes,0.5,REQ-001,,bob,v0.1,2025-01-10,2025-01-20,docs/req2.md
REQ-003,SOFTWARE,CORE,Third requirement,Value,tests/test.py,test_func3,Unit Test,COMPLETE,LOW,2,Notes,2.0,REQ-002,,charlie,v0.2,2025-01-15,2025-01-30,docs/req3.md
"""
        current = tmp_path / "improved.csv"
        current.write_text(csv_content)

        report = compare_databases(baseline_csv, current)
        assert report.summary_status == "improved"
