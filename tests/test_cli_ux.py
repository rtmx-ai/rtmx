"""Tests for CLI UX enhancements (Phase 5).

REQ-UX-001: Rich progress bars
REQ-UX-006: Graceful fallback without rich
"""

from __future__ import annotations

import io
import sys
from pathlib import Path
from unittest.mock import patch

import pytest
from click.testing import CliRunner


class TestRichProgressBars:
    """Tests for REQ-UX-001: Rich progress bars."""

    @pytest.fixture
    def sample_rtm_csv(self, tmp_path: Path) -> Path:
        """Create a sample RTM CSV for testing."""
        csv_content = """req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file
REQ-001,CORE,API,Core requirement 1,Target,tests/test.py,test_func,Unit Test,COMPLETE,HIGH,1,Note,1.0,,,,,,,
REQ-002,CORE,API,Core requirement 2,Target,tests/test.py,test_func,Unit Test,COMPLETE,HIGH,1,Note,1.0,,,,,,,
REQ-003,CORE,API,Core requirement 3,Target,tests/test.py,test_func,Unit Test,MISSING,MEDIUM,2,Note,1.0,,,,,,,
REQ-004,FEATURE,UI,Feature requirement,Target,tests/test.py,test_func,Unit Test,MISSING,LOW,2,Note,1.0,,,,,,,
"""
        csv_path = tmp_path / "rtm_database.csv"
        csv_path.write_text(csv_content)
        return csv_path

    @pytest.mark.req("REQ-UX-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_rich_progress_bars_render(self, sample_rtm_csv: Path) -> None:
        """Test that rich progress bars render when rich is available."""
        from rtmx.formatting import is_rich_available, render_rich_status

        assert is_rich_available(), "rich library must be installed for this test"

        # Test render_rich_status directly with captured output
        output = io.StringIO()
        render_rich_status(
            complete=2,
            partial=0,
            missing=2,
            total=4,
            completion_pct=50.0,
            phase_stats=[
                (1, 2, 0, 0, 100.0),  # Phase 1: 2 complete
                (2, 0, 0, 2, 0.0),  # Phase 2: 2 missing
            ],
            file=output,
        )

        result = output.getvalue()

        # Rich output should have box-drawing characters
        assert "╭" in result or "┌" in result, "Rich output should have box-drawing panels"
        # Should show phase progress
        assert "Phase" in result
        # Should show RTMX Status
        assert "RTMX Status" in result

    @pytest.mark.req("REQ-UX-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_rich_shows_overall_progress(self) -> None:
        """Test that rich output shows overall progress with percentage."""
        from rtmx.formatting import is_rich_available, render_rich_status

        assert is_rich_available()

        output = io.StringIO()
        render_rich_status(
            complete=5,
            partial=1,
            missing=4,
            total=10,
            completion_pct=55.0,
            phase_stats=[],
            file=output,
        )

        result = output.getvalue()

        # Should show completion stats
        assert "complete" in result.lower()
        assert "55.0%" in result

    @pytest.mark.req("REQ-UX-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_rich_shows_per_phase_progress(self) -> None:
        """Test that rich output shows progress bar for each phase."""
        from rtmx.formatting import is_rich_available, render_rich_status

        assert is_rich_available()

        output = io.StringIO()
        render_rich_status(
            complete=2,
            partial=0,
            missing=2,
            total=4,
            completion_pct=50.0,
            phase_stats=[
                (1, 2, 0, 0, 100.0),  # Phase 1: 100% complete
                (2, 0, 0, 2, 0.0),  # Phase 2: 0% complete
            ],
            file=output,
        )

        result = output.getvalue()

        # Should show both phases with percentages
        assert "Phase  1:" in result or "Phase 1:" in result
        assert "Phase  2:" in result or "Phase 2:" in result
        assert "100.0%" in result  # Phase 1 is 100% complete
        assert "0.0%" in result  # Phase 2 is 0% complete


class TestRichFallback:
    """Tests for REQ-UX-006: Graceful fallback without rich."""

    @pytest.fixture
    def sample_rtm_csv(self, tmp_path: Path) -> Path:
        """Create a sample RTM CSV for testing."""
        csv_content = """req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file
REQ-001,CORE,API,Core requirement 1,Target,tests/test.py,test_func,Unit Test,COMPLETE,HIGH,1,Note,1.0,,,,,,,
REQ-002,CORE,API,Core requirement 2,Target,tests/test.py,test_func,Unit Test,MISSING,MEDIUM,2,Note,1.0,,,,,,,
"""
        csv_path = tmp_path / "rtm_database.csv"
        csv_path.write_text(csv_content)
        return csv_path

    @pytest.mark.req("REQ-UX-006")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_fallback_without_rich_library(
        self, sample_rtm_csv: Path, capsys: pytest.CaptureFixture
    ) -> None:
        """Test that status command works when rich is not installed."""
        from rtmx.cli.status import run_status

        # Force plain mode (simulates rich not being available)
        with patch.object(sys, "exit"):
            run_status(sample_rtm_csv, verbosity=0, json_output=None, use_rich=False)

        captured = capsys.readouterr()
        output = captured.out

        # Should still show status information
        assert "RTM Status" in output
        assert "complete" in output.lower()
        # Should NOT have rich box-drawing characters
        assert "╭" not in output

    @pytest.mark.req("REQ-UX-006")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_no_rich_flag_forces_plain_output(self) -> None:
        """Test that --no-rich flag forces plain output."""
        from rtmx.cli.main import main

        runner = CliRunner()
        result = runner.invoke(main, ["status", "--no-rich"])

        # Should not fail
        assert result.exit_code in (0, 1)  # 1 is ok for incomplete requirements
        # Should NOT have rich box-drawing characters
        assert "╭" not in result.output

    @pytest.mark.req("REQ-UX-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_rich_flag_forces_rich_output(self, sample_rtm_csv: Path) -> None:
        """Test that --rich flag forces rich output."""
        from rtmx.cli.main import main

        runner = CliRunner()
        # --rtm-csv is a global option, must come before subcommand
        result = runner.invoke(main, ["--rtm-csv", str(sample_rtm_csv), "status", "--rich"])

        # Should not fail (assuming rich is installed in dev)
        assert result.exit_code in (0, 1), f"Exit code {result.exit_code}: {result.output}"
        # Should have rich box-drawing characters
        assert "╭" in result.output or "┌" in result.output


class TestRichDetection:
    """Tests for rich library detection."""

    @pytest.mark.req("REQ-UX-006")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_rich_available_detection(self) -> None:
        """Test that we can detect if rich is available."""
        from rtmx.formatting import is_rich_available

        # Since rich is installed in dev, it should be available
        assert is_rich_available() is True

    @pytest.mark.req("REQ-UX-006")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_rich_not_available_detection(self) -> None:
        """Test detection when rich is not available."""
        from rtmx import formatting

        # Mock rich not being available
        with patch.object(formatting, "_RICH_AVAILABLE", False):
            assert formatting.is_rich_available() is False


class TestBacklogListView:
    """Tests for REQ-UX-007: Backlog list view."""

    @pytest.fixture
    def sample_rtm_csv(self, tmp_path: Path) -> Path:
        """Create a sample RTM CSV for testing."""
        csv_content = """req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file
REQ-001,CORE,API,Core requirement 1,Target,tests/test.py,test_func,Unit Test,COMPLETE,HIGH,1,Note,1.0,,,,,,,
REQ-002,CORE,API,Core requirement 2,Target,tests/test.py,test_func,Unit Test,MISSING,HIGH,1,Note,1.0,REQ-001,,,,,,
REQ-003,CORE,API,Core requirement 3,Target,tests/test.py,test_func,Unit Test,MISSING,MEDIUM,2,Note,1.0,,,,,,,
"""
        csv_path = tmp_path / "rtm_database.csv"
        csv_path.write_text(csv_content)
        return csv_path

    @pytest.mark.req("REQ-UX-007")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_backlog_list_view_shows_all_phase_requirements(
        self, sample_rtm_csv: Path, capsys: pytest.CaptureFixture
    ) -> None:
        """Test that list view shows ALL requirements for a phase."""
        from rtmx.cli.backlog import BacklogView, run_backlog

        with patch.object(sys, "exit"):
            run_backlog(sample_rtm_csv, phase=1, view=BacklogView.LIST, limit=10)

        captured = capsys.readouterr()
        output = captured.out

        # Should show phase in title
        assert "Phase 1" in output
        # Should show both complete and incomplete requirements
        assert "REQ-001" in output  # Complete
        assert "REQ-002" in output  # Incomplete
        # Should NOT show phase 2 requirements
        assert "REQ-003" not in output
        # Should show completion stats
        assert "complete" in output.lower()

    @pytest.mark.req("REQ-UX-007")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_backlog_list_view_cli(self, sample_rtm_csv: Path) -> None:
        """Test that --view list works via CLI."""
        from rtmx.cli.main import main

        runner = CliRunner()
        result = runner.invoke(
            main, ["--rtm-csv", str(sample_rtm_csv), "backlog", "--view", "list", "--phase", "1"]
        )

        # Should show all phase 1 requirements
        assert "REQ-001" in result.output
        assert "REQ-002" in result.output
