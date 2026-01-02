"""Tests for additional RTMX CLI commands.

This module tests the core logic functions of additional CLI commands:
- run_diff: Compare RTM databases before and after changes
- run_deps: Show dependency graph visualization
- run_analyze: Analyze project for requirements artifacts
- run_from_tests: Scan test files for requirement markers
"""

from __future__ import annotations

import csv
import json
import sys
from io import StringIO
from pathlib import Path

import pytest

from rtmx import RTMDatabase

# =============================================================================
# Fixtures
# =============================================================================


@pytest.fixture
def sample_rtm_csv(tmp_path: Path) -> Path:
    """Create a sample RTM CSV for testing."""
    csv_path = tmp_path / "rtm_database.csv"

    headers = [
        "req_id",
        "category",
        "subcategory",
        "requirement_text",
        "target_value",
        "test_module",
        "test_function",
        "validation_method",
        "status",
        "priority",
        "phase",
        "notes",
        "effort_weeks",
        "dependencies",
        "blocks",
        "assignee",
        "sprint",
        "started_date",
        "completed_date",
        "requirement_file",
    ]

    rows = [
        {
            "req_id": "REQ-CORE-001",
            "category": "CORE",
            "subcategory": "Foundation",
            "requirement_text": "System shall initialize",
            "target_value": "Success rate 100%",
            "test_module": "tests/test_core.py",
            "test_function": "test_init",
            "validation_method": "Unit Test",
            "status": "COMPLETE",
            "priority": "P0",
            "phase": "1",
            "notes": "Critical foundation",
            "effort_weeks": "2.0",
            "dependencies": "",
            "blocks": "REQ-CORE-002|REQ-CORE-003",
            "assignee": "alice",
            "sprint": "v0.1",
            "started_date": "2025-01-01",
            "completed_date": "2025-01-15",
            "requirement_file": "docs/requirements/CORE/REQ-CORE-001.md",
        },
        {
            "req_id": "REQ-CORE-002",
            "category": "CORE",
            "subcategory": "Foundation",
            "requirement_text": "System shall handle config",
            "target_value": "All configs loaded",
            "test_module": "tests/test_core.py",
            "test_function": "test_config",
            "validation_method": "Unit Test",
            "status": "PARTIAL",
            "priority": "HIGH",
            "phase": "1",
            "notes": "Config management",
            "effort_weeks": "1.5",
            "dependencies": "REQ-CORE-001",
            "blocks": "REQ-FEAT-001",
            "assignee": "bob",
            "sprint": "v0.1",
            "started_date": "2025-01-10",
            "completed_date": "",
            "requirement_file": "docs/requirements/CORE/REQ-CORE-002.md",
        },
        {
            "req_id": "REQ-CORE-003",
            "category": "CORE",
            "subcategory": "Data",
            "requirement_text": "System shall persist data",
            "target_value": "No data loss",
            "test_module": "",
            "test_function": "",
            "validation_method": "Integration Test",
            "status": "MISSING",
            "priority": "HIGH",
            "phase": "1",
            "notes": "Database layer",
            "effort_weeks": "3.0",
            "dependencies": "REQ-CORE-001",
            "blocks": "",
            "assignee": "charlie",
            "sprint": "v0.2",
            "started_date": "",
            "completed_date": "",
            "requirement_file": "docs/requirements/CORE/REQ-CORE-003.md",
        },
        {
            "req_id": "REQ-FEAT-001",
            "category": "FEATURES",
            "subcategory": "UI",
            "requirement_text": "UI shall be responsive",
            "target_value": "Response < 100ms",
            "test_module": "",
            "test_function": "",
            "validation_method": "System Test",
            "status": "MISSING",
            "priority": "MEDIUM",
            "phase": "2",
            "notes": "User interface",
            "effort_weeks": "2.0",
            "dependencies": "REQ-CORE-002",
            "blocks": "",
            "assignee": "alice",
            "sprint": "v0.3",
            "started_date": "",
            "completed_date": "",
            "requirement_file": "docs/requirements/FEATURES/REQ-FEAT-001.md",
        },
    ]

    with open(csv_path, "w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=headers)
        writer.writeheader()
        for row in rows:
            writer.writerow(row)

    return csv_path


@pytest.fixture
def baseline_rtm_csv(tmp_path: Path) -> Path:
    """Create a baseline RTM CSV for diff testing."""
    csv_path = tmp_path / "baseline.csv"

    headers = [
        "req_id",
        "category",
        "subcategory",
        "requirement_text",
        "target_value",
        "test_module",
        "test_function",
        "validation_method",
        "status",
        "priority",
        "phase",
        "notes",
        "effort_weeks",
        "dependencies",
        "blocks",
        "assignee",
        "sprint",
        "started_date",
        "completed_date",
        "requirement_file",
    ]

    rows = [
        {
            "req_id": "REQ-CORE-001",
            "category": "CORE",
            "subcategory": "Foundation",
            "requirement_text": "System shall initialize",
            "status": "PARTIAL",  # Changed from COMPLETE
            "priority": "P0",
            "phase": "1",
        },
        {
            "req_id": "REQ-CORE-002",
            "category": "CORE",
            "subcategory": "Foundation",
            "requirement_text": "System shall handle config",
            "status": "MISSING",  # Changed from PARTIAL
            "priority": "HIGH",
            "phase": "1",
        },
        {
            "req_id": "REQ-CORE-003",
            "category": "CORE",
            "subcategory": "Data",
            "requirement_text": "System shall persist data",
            "status": "MISSING",
            "priority": "HIGH",
            "phase": "1",
        },
        # REQ-FEAT-001 will be "added" in current
    ]

    with open(csv_path, "w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=headers)
        writer.writeheader()
        for row in rows:
            full_row = {field: row.get(field, "") for field in headers}
            writer.writerow(full_row)

    return csv_path


@pytest.fixture
def sample_test_file(tmp_path: Path) -> Path:
    """Create a sample test file with requirement markers."""
    test_path = tmp_path / "test_sample.py"

    test_content = '''"""Sample test file with requirement markers."""

import pytest


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_initialization():
    """Test system initialization."""
    assert True


@pytest.mark.req("REQ-CORE-002")
@pytest.mark.scope_unit
@pytest.mark.technique_parametric
@pytest.mark.env_simulation
def test_configuration():
    """Test configuration handling."""
    assert True


@pytest.mark.req("REQ-FEAT-001")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestFeatures:
    """Test feature functionality."""

    def test_ui_responsiveness(self):
        """Test UI response time."""
        assert True

    def test_ui_rendering(self):
        """Test UI rendering."""
        assert True


@pytest.mark.req("REQ-MISSING-001")
@pytest.mark.scope_integration
@pytest.mark.technique_stress
@pytest.mark.env_hil
def test_missing_requirement():
    """Test requirement not in database."""
    assert True


def test_unmarked():
    """Test without requirement markers."""
    assert True
'''

    test_path.write_text(test_content)
    return test_path


@pytest.fixture
def sample_test_directory(tmp_path: Path) -> Path:
    """Create a test directory with multiple test files."""
    test_dir = tmp_path / "tests"
    test_dir.mkdir()

    # Create first test file
    test1 = test_dir / "test_core.py"
    test1.write_text('''"""Core tests."""
import pytest

@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_init():
    assert True

@pytest.mark.req("REQ-CORE-002")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_config():
    assert True
''')

    # Create second test file
    test2 = test_dir / "test_features.py"
    test2.write_text('''"""Feature tests."""
import pytest

@pytest.mark.req("REQ-FEAT-001")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_feature():
    assert True

def test_no_marker():
    assert True
''')

    return test_dir


@pytest.fixture
def rtmx_config_basic(tmp_path: Path) -> Path:
    """Create a basic rtmx.yaml config file."""
    config_path = tmp_path / "rtmx.yaml"
    config_content = """
database: docs/rtm_database.csv
adapters:
  github:
    enabled: false
  jira:
    enabled: false
"""
    config_path.write_text(config_content)
    return config_path


@pytest.fixture
def rtmx_config_with_github(tmp_path: Path) -> Path:
    """Create rtmx.yaml config with GitHub adapter enabled."""
    config_path = tmp_path / "rtmx.yaml"
    config_content = """
database: docs/rtm_database.csv
adapters:
  github:
    enabled: true
    repo: org/repo
    token: test_token
  jira:
    enabled: false
"""
    config_path.write_text(config_content)
    return config_path


# =============================================================================
# Tests for run_diff
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRunDiff:
    """Tests for run_diff CLI function."""

    def test_diff_terminal_format(
        self,
        baseline_rtm_csv: Path,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test diff command with terminal format output."""
        from rtmx.cli.diff import run_diff

        # Mock sys.exit to prevent actual exit
        exit_code = None

        def mock_exit(code: int) -> None:
            nonlocal exit_code
            exit_code = code

        monkeypatch.setattr(sys, "exit", mock_exit)

        run_diff(baseline_rtm_csv, sample_rtm_csv, format_type="terminal")

        captured = capsys.readouterr()
        output = captured.out
        assert "RTM Comparison" in output
        assert "Baseline:" in output
        assert "Current:" in output
        assert "Requirements" in output
        assert "Completion" in output

    def test_diff_markdown_format(
        self,
        baseline_rtm_csv: Path,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test diff command with markdown format output."""
        from rtmx.cli.diff import run_diff

        monkeypatch.setattr(sys, "exit", lambda x: None)

        run_diff(baseline_rtm_csv, sample_rtm_csv, format_type="markdown")

        captured = capsys.readouterr()
        output = captured.out
        assert "###" in output  # Markdown header
        assert "|" in output  # Markdown table
        assert "Metric" in output
        assert "Baseline" in output
        assert "Current" in output

    def test_diff_json_format(
        self,
        baseline_rtm_csv: Path,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test diff command with JSON format output."""
        from rtmx.cli.diff import run_diff

        monkeypatch.setattr(sys, "exit", lambda x: None)

        run_diff(baseline_rtm_csv, sample_rtm_csv, format_type="json")

        captured = capsys.readouterr()
        output = captured.out
        data = json.loads(output)

        assert "baseline_path" in data
        assert "current_path" in data
        assert "summary" in data
        assert "baseline" in data
        assert "current" in data
        assert "changes" in data

    def test_diff_to_output_file(
        self,
        baseline_rtm_csv: Path,
        sample_rtm_csv: Path,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test diff command writing to output file."""
        from rtmx.cli.diff import run_diff

        monkeypatch.setattr(sys, "exit", lambda x: None)

        output_file = tmp_path / "diff_output.md"
        run_diff(baseline_rtm_csv, sample_rtm_csv, format_type="markdown", output_path=output_file)

        assert output_file.exists()
        content = output_file.read_text()
        assert "###" in content
        assert "Metric" in content

    def test_diff_shows_added_requirements(
        self,
        baseline_rtm_csv: Path,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test diff command shows added requirements."""
        from rtmx.cli.diff import run_diff

        monkeypatch.setattr(sys, "exit", lambda x: None)

        run_diff(baseline_rtm_csv, sample_rtm_csv, format_type="terminal")

        captured = capsys.readouterr()
        output = captured.out
        assert "Added Requirements" in output
        assert "REQ-FEAT-001" in output

    def test_diff_shows_status_changes(
        self,
        baseline_rtm_csv: Path,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test diff command shows status changes."""
        from rtmx.cli.diff import run_diff

        monkeypatch.setattr(sys, "exit", lambda x: None)

        run_diff(baseline_rtm_csv, sample_rtm_csv, format_type="terminal")

        captured = capsys.readouterr()
        output = captured.out
        assert "Status Changes" in output
        # REQ-CORE-001: PARTIAL -> COMPLETE
        # REQ-CORE-002: MISSING -> PARTIAL

    def test_diff_completion_delta(
        self,
        baseline_rtm_csv: Path,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test diff command calculates completion delta."""
        from rtmx.cli.diff import run_diff

        monkeypatch.setattr(sys, "exit", lambda x: None)

        run_diff(baseline_rtm_csv, sample_rtm_csv, format_type="json")

        captured = capsys.readouterr()
        output = captured.out
        data = json.loads(output)

        # Should have completion delta (positive = improved)
        assert "completion_delta" in data["summary"]

    def test_diff_exit_code_improved(
        self, baseline_rtm_csv: Path, sample_rtm_csv: Path, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        """Test diff exits with code 0 when status is improved."""
        from rtmx.cli.diff import run_diff

        exit_code = None

        def mock_exit(code: int) -> None:
            nonlocal exit_code
            exit_code = code

        monkeypatch.setattr(sys, "exit", mock_exit)

        # Redirect stdout to suppress output
        old_stdout = sys.stdout
        sys.stdout = StringIO()

        run_diff(baseline_rtm_csv, sample_rtm_csv, format_type="terminal")

        sys.stdout = old_stdout

        # Should exit 0 for improved/stable
        assert exit_code == 0

    def test_diff_baseline_not_found(
        self,
        sample_rtm_csv: Path,
        tmp_path: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test diff with non-existent baseline file."""
        from rtmx.cli.diff import run_diff

        exit_code = None

        def mock_exit(code: int) -> None:
            nonlocal exit_code
            exit_code = code

        monkeypatch.setattr(sys, "exit", mock_exit)

        nonexistent = tmp_path / "missing.csv"
        run_diff(nonexistent, sample_rtm_csv, format_type="terminal")

        captured = capsys.readouterr()
        assert "Error" in captured.err or "Error" in captured.out
        assert exit_code == 1

    def test_diff_current_not_found(
        self,
        baseline_rtm_csv: Path,
        tmp_path: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test diff with non-existent current file."""
        from rtmx.cli.diff import run_diff

        exit_code = None

        def mock_exit(code: int) -> None:
            nonlocal exit_code
            exit_code = code

        monkeypatch.setattr(sys, "exit", mock_exit)

        nonexistent = tmp_path / "missing.csv"
        run_diff(baseline_rtm_csv, nonexistent, format_type="terminal")

        captured = capsys.readouterr()
        assert "Error" in captured.err or "Error" in captured.out
        assert exit_code == 1


# =============================================================================
# Tests for run_deps
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRunDeps:
    """Tests for run_deps CLI function."""

    def test_deps_all_requirements(
        self, sample_rtm_csv: Path, capsys: pytest.CaptureFixture, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        """Test deps command shows all requirements."""
        from rtmx.cli.deps import run_deps

        monkeypatch.setattr(sys, "exit", lambda x: None)

        run_deps(sample_rtm_csv, category=None, phase=None, req_id=None)

        captured = capsys.readouterr()
        output = captured.out
        assert "Dependencies" in output
        assert "REQ-CORE-001" in output
        assert "REQ-CORE-002" in output

    def test_deps_filter_by_category(
        self, sample_rtm_csv: Path, capsys: pytest.CaptureFixture, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        """Test deps command filtered by category."""
        from rtmx.cli.deps import run_deps

        monkeypatch.setattr(sys, "exit", lambda x: None)

        run_deps(sample_rtm_csv, category="CORE", phase=None, req_id=None)

        captured = capsys.readouterr()
        output = captured.out
        assert "Dependencies (CORE)" in output
        assert "REQ-CORE-001" in output
        # REQ-FEAT-001 should not appear (different category)

    def test_deps_filter_by_phase(
        self, sample_rtm_csv: Path, capsys: pytest.CaptureFixture, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        """Test deps command filtered by phase."""
        from rtmx.cli.deps import run_deps

        monkeypatch.setattr(sys, "exit", lambda x: None)

        run_deps(sample_rtm_csv, category=None, phase=1, req_id=None)

        captured = capsys.readouterr()
        output = captured.out
        assert "Dependencies (Phase 1)" in output
        assert "REQ-CORE-001" in output
        assert "REQ-CORE-002" in output
        assert "REQ-CORE-003" in output

    def test_deps_single_requirement_detail(
        self, sample_rtm_csv: Path, capsys: pytest.CaptureFixture, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        """Test deps command for single requirement shows details."""
        from rtmx.cli.deps import run_deps

        monkeypatch.setattr(sys, "exit", lambda x: None)

        run_deps(sample_rtm_csv, category=None, phase=None, req_id="REQ-CORE-001")

        captured = capsys.readouterr()
        output = captured.out
        assert "Dependencies for REQ-CORE-001" in output
        assert "System shall initialize" in output
        assert "Blocks:" in output
        assert "REQ-CORE-002" in output
        assert "REQ-CORE-003" in output

    def test_deps_shows_blocking_counts(
        self, sample_rtm_csv: Path, capsys: pytest.CaptureFixture, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        """Test deps command shows blocking counts."""
        from rtmx.cli.deps import run_deps

        monkeypatch.setattr(sys, "exit", lambda x: None)

        run_deps(sample_rtm_csv, category=None, phase=None, req_id=None)

        captured = capsys.readouterr()
        output = captured.out
        assert "Blocks" in output
        # Should show numeric blocking counts

    def test_deps_requirement_not_found(
        self, sample_rtm_csv: Path, capsys: pytest.CaptureFixture, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        """Test deps command with non-existent requirement ID."""
        from rtmx.cli.deps import run_deps

        exit_code = None

        def mock_exit(code: int) -> None:
            nonlocal exit_code
            exit_code = code

        monkeypatch.setattr(sys, "exit", mock_exit)

        run_deps(sample_rtm_csv, category=None, phase=None, req_id="REQ-INVALID-999")

        captured = capsys.readouterr()
        assert "not found" in captured.out
        assert exit_code == 1

    def test_deps_summary_statistics(
        self, sample_rtm_csv: Path, capsys: pytest.CaptureFixture, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        """Test deps command shows summary statistics."""
        from rtmx.cli.deps import run_deps

        monkeypatch.setattr(sys, "exit", lambda x: None)

        run_deps(sample_rtm_csv, category=None, phase=None, req_id=None)

        captured = capsys.readouterr()
        output = captured.out
        assert "Summary:" in output
        assert "Total requirements:" in output
        assert "Requirements with dependencies:" in output
        assert "Requirements blocking others:" in output


# =============================================================================
# Tests for run_analyze
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRunAnalyze:
    """Tests for run_analyze CLI function."""

    def test_analyze_finds_test_files(
        self,
        sample_test_directory: Path,
        rtmx_config_basic: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test analyze command finds test files."""
        from rtmx.cli.analyze import run_analyze
        from rtmx.config import load_config

        monkeypatch.setattr(sys, "exit", lambda x: None)

        # Create parent directory for test files
        project_dir = sample_test_directory.parent
        config = load_config(rtmx_config_basic)

        run_analyze(project_dir, None, "terminal", False, config)

        captured = capsys.readouterr()
        output = captured.out
        assert "Test Files:" in output

    def test_analyze_shows_unmarked_tests(
        self,
        sample_test_directory: Path,
        rtmx_config_basic: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test analyze command identifies unmarked tests."""
        from rtmx.cli.analyze import run_analyze
        from rtmx.config import load_config

        monkeypatch.setattr(sys, "exit", lambda x: None)

        project_dir = sample_test_directory.parent
        config = load_config(rtmx_config_basic)

        run_analyze(project_dir, None, "terminal", False, config)

        captured = capsys.readouterr()
        output = captured.out
        assert "Tests without req markers:" in output

    def test_analyze_github_not_configured(
        self,
        tmp_path: Path,
        rtmx_config_basic: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test analyze command shows GitHub not configured."""
        from rtmx.cli.analyze import run_analyze
        from rtmx.config import load_config

        monkeypatch.setattr(sys, "exit", lambda x: None)

        config = load_config(rtmx_config_basic)

        run_analyze(tmp_path, None, "terminal", False, config)

        captured = capsys.readouterr()
        output = captured.out
        assert "GitHub Issues:" in output
        assert "Not configured" in output

    def test_analyze_github_configured(
        self,
        tmp_path: Path,
        rtmx_config_with_github: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test analyze command shows GitHub configuration."""
        from rtmx.cli.analyze import run_analyze
        from rtmx.config import load_config

        monkeypatch.setattr(sys, "exit", lambda x: None)

        config = load_config(rtmx_config_with_github)

        run_analyze(tmp_path, None, "terminal", False, config)

        captured = capsys.readouterr()
        output = captured.out
        assert "GitHub Issues:" in output
        assert "Repository:" in output
        assert "org/repo" in output

    def test_analyze_shows_recommendations(
        self,
        tmp_path: Path,
        rtmx_config_basic: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test analyze command provides recommendations."""
        from rtmx.cli.analyze import run_analyze
        from rtmx.config import load_config

        monkeypatch.setattr(sys, "exit", lambda x: None)

        config = load_config(rtmx_config_basic)

        run_analyze(tmp_path, None, "terminal", False, config)

        captured = capsys.readouterr()
        output = captured.out
        assert "Recommendations:" in output

    def test_analyze_checks_existing_rtm(
        self,
        sample_rtm_csv: Path,
        rtmx_config_basic: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test analyze command detects existing RTM database."""
        from rtmx.cli.analyze import run_analyze
        from rtmx.config import load_config

        monkeypatch.setattr(sys, "exit", lambda x: None)

        # Create docs directory with RTM
        project_dir = sample_rtm_csv.parent
        docs_dir = project_dir / "docs"
        docs_dir.mkdir(exist_ok=True)
        rtm_path = docs_dir / "rtm_database.csv"
        rtm_path.write_text(sample_rtm_csv.read_text())

        config = load_config(rtmx_config_basic)

        run_analyze(project_dir, None, "terminal", False, config)

        captured = capsys.readouterr()
        output = captured.out
        assert "Existing RTM:" in output


# =============================================================================
# Tests for run_from_tests
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRunFromTests:
    """Tests for run_from_tests CLI function."""

    def test_from_tests_single_file(
        self,
        sample_test_file: Path,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test from_tests scans single test file."""
        from rtmx.cli.from_tests import run_from_tests

        monkeypatch.setattr(sys, "exit", lambda x: None)

        # Provide rtm_csv path since it's required when not None
        run_from_tests(str(sample_test_file), str(sample_rtm_csv), False, False, False)

        captured = capsys.readouterr()
        output = captured.out
        assert "Scanning" in output
        assert "requirement markers" in output
        assert "Found" in output

    def test_from_tests_directory(
        self,
        sample_test_directory: Path,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test from_tests scans test directory."""
        from rtmx.cli.from_tests import run_from_tests

        monkeypatch.setattr(sys, "exit", lambda x: None)

        # Provide rtm_csv path since it's required when not None
        run_from_tests(str(sample_test_directory), str(sample_rtm_csv), False, False, False)

        captured = capsys.readouterr()
        output = captured.out
        assert "Found" in output
        assert "test(s) linked to" in output
        assert "requirement(s)" in output

    def test_from_tests_show_all(
        self,
        sample_test_file: Path,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test from_tests with show_all flag."""
        from rtmx.cli.from_tests import run_from_tests

        monkeypatch.setattr(sys, "exit", lambda x: None)

        run_from_tests(str(sample_test_file), str(sample_rtm_csv), True, False, False)

        captured = capsys.readouterr()
        output = captured.out
        assert "All Requirements with Tests:" in output
        assert "REQ-CORE-001" in output
        assert "REQ-CORE-002" in output

    def test_from_tests_show_missing(
        self,
        sample_test_file: Path,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test from_tests shows missing requirements."""
        from rtmx.cli.from_tests import run_from_tests

        monkeypatch.setattr(sys, "exit", lambda x: None)

        run_from_tests(str(sample_test_file), str(sample_rtm_csv), False, True, False)

        captured = capsys.readouterr()
        output = captured.out
        assert "Requirements in tests but not in RTM database:" in output
        assert "REQ-MISSING-001" in output

    def test_from_tests_requirements_without_tests(
        self,
        sample_test_file: Path,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test from_tests shows requirements without tests."""
        from rtmx.cli.from_tests import run_from_tests

        monkeypatch.setattr(sys, "exit", lambda x: None)

        run_from_tests(str(sample_test_file), str(sample_rtm_csv), False, True, False)

        captured = capsys.readouterr()
        output = captured.out
        assert "Requirements in RTM database without tests:" in output
        assert "REQ-CORE-003" in output  # Has no test in sample_test_file

    def test_from_tests_summary(
        self,
        sample_test_file: Path,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test from_tests displays summary."""
        from rtmx.cli.from_tests import run_from_tests

        monkeypatch.setattr(sys, "exit", lambda x: None)

        run_from_tests(str(sample_test_file), str(sample_rtm_csv), False, False, False)

        captured = capsys.readouterr()
        output = captured.out
        assert "Summary:" in output
        assert "Requirements with tests:" in output
        assert "Tests linked to requirements:" in output

    def test_from_tests_update_database(
        self,
        sample_test_file: Path,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test from_tests updates RTM database."""
        from rtmx.cli.from_tests import run_from_tests

        monkeypatch.setattr(sys, "exit", lambda x: None)

        # Make a copy of the RTM CSV to avoid modifying fixture
        import shutil

        rtm_copy = sample_rtm_csv.parent / "rtm_copy.csv"
        shutil.copy(sample_rtm_csv, rtm_copy)

        run_from_tests(str(sample_test_file), str(rtm_copy), False, False, True)

        captured = capsys.readouterr()
        output = captured.out
        assert "Updated" in output
        assert "requirement(s) in RTM database" in output

        # Verify database was updated
        db = RTMDatabase.load(rtm_copy)
        req = db.get("REQ-CORE-001")
        assert "test_sample.py" in req.test_module or req.test_function == "test_initialization"

    def test_from_tests_no_markers_found(
        self, tmp_path: Path, capsys: pytest.CaptureFixture, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        """Test from_tests when no markers found."""
        from rtmx.cli.from_tests import run_from_tests

        monkeypatch.setattr(sys, "exit", lambda x: None)

        # Create test file without markers
        test_file = tmp_path / "test_no_markers.py"
        test_file.write_text("""
def test_something():
    assert True
""")

        run_from_tests(str(test_file), None, False, False, False)

        captured = capsys.readouterr()
        output = captured.out
        assert "No requirement markers found" in output

    def test_from_tests_path_not_found(
        self, tmp_path: Path, capsys: pytest.CaptureFixture, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        """Test from_tests with non-existent path."""
        from rtmx.cli.from_tests import run_from_tests

        exit_code = None

        def mock_exit(code: int) -> None:
            nonlocal exit_code
            exit_code = code

        monkeypatch.setattr(sys, "exit", mock_exit)

        nonexistent = tmp_path / "nonexistent"
        run_from_tests(str(nonexistent), None, False, False, False)

        captured = capsys.readouterr()
        output = captured.out
        assert "Error" in output
        assert exit_code == 1

    def test_from_tests_extract_markers_from_file(self, sample_test_file: Path) -> None:
        """Test extract_markers_from_file function directly."""
        from rtmx.cli.from_tests import extract_markers_from_file

        markers = extract_markers_from_file(sample_test_file)

        assert len(markers) > 0
        req_ids = {m.req_id for m in markers}
        assert "REQ-CORE-001" in req_ids
        assert "REQ-CORE-002" in req_ids
        assert "REQ-FEAT-001" in req_ids
        assert "REQ-MISSING-001" in req_ids

    def test_from_tests_extract_markers_class_level(self, sample_test_file: Path) -> None:
        """Test extraction of class-level requirement markers."""
        from rtmx.cli.from_tests import extract_markers_from_file

        markers = extract_markers_from_file(sample_test_file)

        # Should find REQ-FEAT-001 applied to both test methods in TestFeatures class
        feat_markers = [m for m in markers if m.req_id == "REQ-FEAT-001"]
        assert len(feat_markers) == 2  # Two test methods
        assert any("test_ui_responsiveness" in m.test_function for m in feat_markers)
        assert any("test_ui_rendering" in m.test_function for m in feat_markers)

    def test_from_tests_scan_test_directory(self, sample_test_directory: Path) -> None:
        """Test scan_test_directory function directly."""
        from rtmx.cli.from_tests import scan_test_directory

        markers = scan_test_directory(sample_test_directory)

        assert len(markers) >= 3  # At least REQ-CORE-001, REQ-CORE-002, REQ-FEAT-001
        req_ids = {m.req_id for m in markers}
        assert "REQ-CORE-001" in req_ids
        assert "REQ-CORE-002" in req_ids
        assert "REQ-FEAT-001" in req_ids

    def test_from_tests_marker_attributes(self, sample_test_file: Path) -> None:
        """Test that markers capture all attributes correctly."""
        from rtmx.cli.from_tests import extract_markers_from_file

        markers = extract_markers_from_file(sample_test_file)

        # Find REQ-CORE-001 marker
        core_marker = next(m for m in markers if m.req_id == "REQ-CORE-001")

        assert core_marker.req_id == "REQ-CORE-001"
        assert "test_sample.py" in core_marker.test_file
        assert core_marker.test_function == "test_initialization"
        assert core_marker.line_number > 0
        assert "scope_unit" in core_marker.markers
        assert "technique_nominal" in core_marker.markers
        assert "env_simulation" in core_marker.markers


# =============================================================================
# Integration tests for CLI commands
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCLIAdditionalIntegration:
    """Integration tests for additional CLI commands working together."""

    def test_from_tests_to_diff_workflow(
        self,
        sample_test_directory: Path,
        sample_rtm_csv: Path,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Test workflow: from_tests updates DB, then diff shows changes."""
        from rtmx.cli.diff import run_diff
        from rtmx.cli.from_tests import run_from_tests

        monkeypatch.setattr(sys, "exit", lambda x: None)

        # Make copies of RTM for baseline and current
        import shutil

        baseline = tmp_path / "baseline.csv"
        current = tmp_path / "current.csv"
        shutil.copy(sample_rtm_csv, baseline)
        shutil.copy(sample_rtm_csv, current)

        # Update current with test info
        old_stdout = sys.stdout
        sys.stdout = StringIO()
        run_from_tests(str(sample_test_directory), str(current), False, False, True)
        sys.stdout = old_stdout

        # Run diff to see changes
        sys.stdout = StringIO()
        run_diff(baseline, current, format_type="terminal")
        diff_output = sys.stdout.getvalue()
        sys.stdout = old_stdout

        assert "RTM Comparison" in diff_output

    def test_deps_with_analyze_workflow(
        self, sample_rtm_csv: Path, rtmx_config_basic: Path, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        """Test workflow: analyze project, then view deps."""
        from rtmx.cli.analyze import run_analyze
        from rtmx.cli.deps import run_deps
        from rtmx.config import load_config

        monkeypatch.setattr(sys, "exit", lambda x: None)

        config = load_config(rtmx_config_basic)

        # Analyze project
        old_stdout = sys.stdout
        sys.stdout = StringIO()
        run_analyze(sample_rtm_csv.parent, None, "terminal", False, config)
        analyze_output = sys.stdout.getvalue()
        sys.stdout = old_stdout

        assert "RTMX Project Analysis" in analyze_output

        # View dependencies
        sys.stdout = StringIO()
        run_deps(sample_rtm_csv, None, None, None)
        deps_output = sys.stdout.getvalue()
        sys.stdout = old_stdout

        assert "Dependencies" in deps_output
