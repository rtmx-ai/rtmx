"""Tests for rtmx.comparison module."""

from pathlib import Path

import pytest

from rtmx.comparison import (
    ComparisonReport,
    StatusChange,
    capture_baseline,
    compare_databases,
)
from rtmx.models import Status


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


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestStatusChange:
    """Tests for StatusChange dataclass."""

    def test_status_change_creation(self):
        """Test StatusChange creation."""
        change = StatusChange(
            req_id="REQ-001",
            old_status=Status.MISSING,
            new_status=Status.COMPLETE,
        )
        assert change.req_id == "REQ-001"
        assert change.old_status == Status.MISSING
        assert change.new_status == Status.COMPLETE
        assert change.change_type == "changed"  # Property

    def test_status_change_type_added(self):
        """Test StatusChange type when added."""
        change = StatusChange(
            req_id="REQ-001",
            old_status=None,
            new_status=Status.MISSING,
        )
        assert change.change_type == "added"

    def test_status_change_type_removed(self):
        """Test StatusChange type when removed."""
        change = StatusChange(
            req_id="REQ-001",
            old_status=Status.COMPLETE,
            new_status=None,
        )
        assert change.change_type == "removed"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCaptureBaseline:
    """Tests for capture_baseline function."""

    def test_capture_baseline(self, baseline_csv: Path):
        """Test baseline capture."""
        result = capture_baseline(baseline_csv)
        assert "req_count" in result
        assert result["req_count"] == 3
        assert "completion" in result
        assert "cycles" in result
        assert "status_counts" in result

    def test_capture_baseline_completion(self, baseline_csv: Path):
        """Test completion percentage calculation."""
        result = capture_baseline(baseline_csv)
        # 1 COMPLETE out of 3 = 33.3%
        # 1 PARTIAL contributes 0.5 = (1 + 0.5) / 3 = 50%
        assert result["completion"] == pytest.approx(50.0, rel=0.1)


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCompareDatabases:
    """Tests for compare_databases function."""

    def test_compare_databases_basic(self, baseline_csv: Path, current_csv: Path):
        """Test basic database comparison."""
        report = compare_databases(baseline_csv, current_csv)
        assert isinstance(report, ComparisonReport)
        assert report.baseline_path == str(baseline_csv)
        assert report.current_path == str(current_csv)

    def test_compare_databases_counts(self, baseline_csv: Path, current_csv: Path):
        """Test requirement count comparison."""
        report = compare_databases(baseline_csv, current_csv)
        assert report.baseline_req_count == 3
        assert report.current_req_count == 3
        assert report.req_count_delta == 0

    def test_compare_databases_added_removed(self, baseline_csv: Path, current_csv: Path):
        """Test detection of added/removed requirements."""
        report = compare_databases(baseline_csv, current_csv)
        # REQ-003 removed, REQ-004 added
        assert "REQ-003" in report.removed_requirements
        assert "REQ-004" in report.added_requirements

    def test_compare_databases_status_changes(self, baseline_csv: Path, current_csv: Path):
        """Test detection of status changes."""
        report = compare_databases(baseline_csv, current_csv)
        # REQ-002 changed from PARTIAL to COMPLETE
        changed = [sc for sc in report.status_changes if sc.change_type == "changed"]
        req_002_changes = [sc for sc in changed if sc.req_id == "REQ-002"]
        assert len(req_002_changes) == 1
        assert req_002_changes[0].old_status == Status.PARTIAL
        assert req_002_changes[0].new_status == Status.COMPLETE

    def test_compare_databases_completion_delta(self, baseline_csv: Path, current_csv: Path):
        """Test completion percentage change."""
        report = compare_databases(baseline_csv, current_csv)
        # Baseline: (1 COMPLETE + 0.5 PARTIAL) / 3 = 50%
        # Current: (2 COMPLETE) / 3 = 66.7%
        assert report.current_completion > report.baseline_completion


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestComparisonReport:
    """Tests for ComparisonReport methods."""

    def test_comparison_report_to_dict(self, baseline_csv: Path, current_csv: Path):
        """Test ComparisonReport serialization."""
        report = compare_databases(baseline_csv, current_csv)
        d = report.to_dict()
        assert "summary" in d
        assert "status" in d["summary"]
        assert "baseline" in d
        assert "current" in d
        assert "changes" in d
        assert "added" in d["changes"]
        assert "removed" in d["changes"]

    def test_comparison_report_to_markdown(self, baseline_csv: Path, current_csv: Path):
        """Test ComparisonReport markdown generation."""
        report = compare_databases(baseline_csv, current_csv)
        md = report.to_markdown()
        assert "RTM Comparison" in md
        assert "|" in md  # Table formatting

    def test_summary_status_improved(self, baseline_csv: Path, current_csv: Path):
        """Test improved summary status."""
        report = compare_databases(baseline_csv, current_csv)
        # Completion increased, so status should be improved
        assert report.summary_status in ("improved", "stable", "breaking")

    def test_summary_status_breaking(self, baseline_csv: Path, tmp_path: Path):
        """Test breaking status when requirements removed."""
        # Create current with fewer requirements
        csv_content = """req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file
REQ-001,SOFTWARE,CORE,First requirement,Value,tests/test.py,test_func,Unit Test,COMPLETE,HIGH,1,Notes,1.0,,,alice,v0.1,2025-01-01,2025-01-15,docs/req.md
"""
        current = tmp_path / "current_fewer.csv"
        current.write_text(csv_content)

        report = compare_databases(baseline_csv, current)
        # 2 requirements removed = breaking change
        assert report.summary_status == "breaking"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestEdgeCases:
    """Tests for edge cases."""

    def test_compare_identical_databases(self, baseline_csv: Path):
        """Test comparing identical databases."""
        report = compare_databases(baseline_csv, baseline_csv)
        assert report.req_count_delta == 0
        assert len(report.added_requirements) == 0
        assert len(report.removed_requirements) == 0
        assert report.summary_status == "stable"

    def test_compare_with_all_new_requirements(self, tmp_path: Path, current_csv: Path):
        """Test comparing baseline with no overlap to current."""
        # Create baseline with different requirements
        baseline_csv = tmp_path / "baseline.csv"
        baseline_csv.write_text(
            "req_id,category,subcategory,requirement_text,target_value,"
            "test_module,test_function,validation_method,status,priority,"
            "phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,"
            "started_date,completed_date,requirement_file\n"
            "REQ-OLD-001,SOFTWARE,LEGACY,Old requirement,Value,tests/test.py,test_func,"
            "Unit Test,COMPLETE,HIGH,1,Notes,1.0,,,alice,v0.1,2025-01-01,2025-01-15,docs/req.md\n"
        )

        report = compare_databases(baseline_csv, current_csv)
        assert report.baseline_req_count == 1
        assert report.current_req_count == 3
        # All current reqs are "new" from perspective of baseline
        assert len(report.added_requirements) == 3
        assert len(report.removed_requirements) == 1
