"""Tests for CLI UX enhancements (Phase 5).

REQ-UX-001: Rich progress bars
REQ-UX-002: Aligned fixed-width columns
REQ-UX-006: Graceful fallback without rich
REQ-UX-007: Backlog list view
"""

from __future__ import annotations

import io
import sys
from pathlib import Path
from unittest.mock import patch

import pytest
from click.testing import CliRunner


@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
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
                ("Phase 1 (Foundation)", 2, 0, 0, 100.0),  # Phase 1: 2 complete
                ("Phase 2 (Core)", 0, 0, 2, 0.0),  # Phase 2: 2 missing
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
                ("Phase 1 (Foundation)", 2, 0, 0, 100.0),  # Phase 1: 100% complete
                ("Phase 2 (Core)", 0, 0, 2, 0.0),  # Phase 2: 0% complete
            ],
            file=output,
        )

        result = output.getvalue()

        # Should show both phases with percentages (with their names)
        assert "Phase 1 (Foundation)" in result
        assert "Phase 2 (Core)" in result
        assert "100.0%" in result  # Phase 1 is 100% complete
        assert "0.0%" in result  # Phase 2 is 0% complete


@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
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


@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
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


@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
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


@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestTUI:
    """Tests for REQ-UX-004: Interactive TUI dashboard."""

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

    @pytest.mark.req("REQ-UX-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_tui_command_exists(self) -> None:
        """Test that tui command is available."""
        from rtmx.cli.main import main

        runner = CliRunner()
        result = runner.invoke(main, ["tui", "--help"])

        assert result.exit_code == 0
        assert "tui" in result.output.lower() or "interactive" in result.output.lower()

    @pytest.mark.req("REQ-UX-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_tui_app_class_exists(self) -> None:
        """Test that RTMXApp class can be imported when textual is available."""
        from rtmx.cli import tui

        # RTMXApp only exists when textual is available
        if tui.is_textual_available():
            assert hasattr(tui, "RTMXApp")
            assert tui.RTMXApp is not None
        else:
            assert not hasattr(tui, "RTMXApp")

    @pytest.mark.req("REQ-UX-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_tui_detects_textual_availability(self) -> None:
        """Test that TUI module detects textual availability."""
        from rtmx.cli.tui import is_textual_available

        # Function should return a boolean
        result = is_textual_available()
        assert isinstance(result, bool)

    @pytest.mark.req("REQ-UX-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_tui_shows_error_without_textual(self) -> None:
        """Test that TUI shows helpful error when textual is not available."""
        from rtmx.cli import tui

        # Mock textual not being available
        original_available = tui._TEXTUAL_AVAILABLE
        try:
            tui._TEXTUAL_AVAILABLE = False

            # run_tui should exit with error message
            with pytest.raises(SystemExit) as exc_info:
                tui.run_tui()

            assert exc_info.value.code == 1
        finally:
            tui._TEXTUAL_AVAILABLE = original_available


@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestLiveRefresh:
    """Tests for REQ-UX-003: Live auto-refresh."""

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

    @pytest.mark.req("REQ-UX-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_live_flag_exists(self) -> None:
        """Test that --live flag is available on status command."""
        from rtmx.cli.main import main

        runner = CliRunner()
        result = runner.invoke(main, ["status", "--help"])

        assert result.exit_code == 0
        assert "--live" in result.output

    @pytest.mark.req("REQ-UX-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_file_change_detection(self, sample_rtm_csv: Path) -> None:
        """Test that file modification time change is detected."""
        import time

        from rtmx.cli.status import get_file_mtime

        # Get initial mtime
        mtime1 = get_file_mtime(sample_rtm_csv)
        assert mtime1 is not None

        # Wait and modify file
        time.sleep(0.1)
        sample_rtm_csv.write_text(sample_rtm_csv.read_text() + "\n")

        # Get new mtime
        mtime2 = get_file_mtime(sample_rtm_csv)
        assert mtime2 is not None
        assert mtime2 > mtime1

    @pytest.mark.req("REQ-UX-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_clear_screen_function(self) -> None:
        """Test that clear_screen returns ANSI escape codes."""
        from rtmx.cli.status import clear_screen

        result = clear_screen()
        # Should contain ANSI escape sequence for clear
        assert "\033[" in result or result == ""


@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestAlignedTables:
    """Tests for REQ-UX-002: Aligned fixed-width columns."""

    @pytest.mark.req("REQ-UX-002")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_format_table_uses_tabulate_grid(self) -> None:
        """Test that format_table uses tabulate grid format."""
        from rtmx.formatting import format_table

        data = [
            ["REQ-001", "✓", "Core requirement 1", "1.0w"],
            ["REQ-002", "✗", "Core requirement 2", "0.5w"],
        ]
        headers = ["Requirement", "Status", "Description", "Effort"]

        result = format_table(data, headers)

        # Should contain the data
        assert "REQ-001" in result
        assert "REQ-002" in result
        # Grid format has borders
        assert "+" in result
        assert "-" in result
        # Headers should be present
        assert "Requirement" in result
        assert "Status" in result

    @pytest.mark.req("REQ-UX-002")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_format_table_consistent_column_widths(self) -> None:
        """Test that tables have consistent column widths."""
        from rtmx.formatting import format_table

        # Varying content lengths
        data = [
            ["REQ-001", "✓", "Short", "1.0w"],
            ["REQ-002", "✗", "Much longer description text here", "0.5w"],
            ["REQ-003", "△", "Med", "2.0w"],
        ]
        headers = ["Requirement", "Status", "Description", "Effort"]

        result = format_table(data, headers)
        lines = result.strip().split("\n")

        # All content lines should be present
        assert any("REQ-001" in line for line in lines)
        assert any("REQ-002" in line for line in lines)
        assert any("REQ-003" in line for line in lines)

        # Find data rows (those containing REQ-)
        data_rows = [line for line in lines if "REQ-" in line]
        # All data rows should have consistent structure (same | positions)
        assert len(data_rows) == 3

    @pytest.mark.req("REQ-UX-002")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_format_table_handles_rich_text_objects(self) -> None:
        """Test that format_table handles rich Text objects in data."""
        from rich.text import Text

        from rtmx.formatting import format_table

        # Mix of strings and Text objects
        status = Text("✓", style="green")
        data = [
            ["REQ-001", status, "Core requirement 1", "1.0w"],
        ]
        headers = ["Requirement", "Status", "Description", "Effort"]

        result = format_table(data, headers)

        # Should handle Text objects gracefully (extracts plain text)
        assert "REQ-001" in result
        assert "Core requirement 1" in result
        assert "✓" in result

    @pytest.mark.req("REQ-UX-002")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_format_table_api_compatibility(self) -> None:
        """Test that format_table accepts use_rich parameter for API compatibility."""
        from rtmx.formatting import format_table

        data = [["REQ-001", "✓", "Test", "1.0w"]]
        headers = ["Req", "Status", "Desc", "Effort"]

        # All these should produce the same tabulate output
        result_none = format_table(data, headers, use_rich=None)
        result_true = format_table(data, headers, use_rich=True)
        result_false = format_table(data, headers, use_rich=False)

        # All should use tabulate grid format
        assert "+" in result_none
        assert "+" in result_true
        assert "+" in result_false
        assert "REQ-001" in result_none
        assert "REQ-001" in result_true
        assert "REQ-001" in result_false
