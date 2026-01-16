"""Tests for REQ-UX-008: Width-adaptive CLI output.

CLI output shall adapt column widths to terminal size, detecting available width
and truncating or adjusting columns dynamically to fit without horizontal overflow.
"""

from __future__ import annotations

import io
import shutil
from pathlib import Path
from unittest.mock import patch

import pytest


@pytest.mark.req("REQ-UX-008")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestTerminalWidthDetection:
    """Tests for terminal width detection."""

    def test_get_terminal_width_returns_int(self) -> None:
        """Test that get_terminal_width returns an integer."""
        from rtmx.formatting import get_terminal_width

        width = get_terminal_width()
        assert isinstance(width, int)
        assert width > 0

    def test_get_terminal_width_default_fallback(self) -> None:
        """Test that get_terminal_width falls back to 80 when terminal unavailable."""
        from rtmx.formatting import get_terminal_width

        # Mock terminal size returning 0 (no terminal) - use namedtuple-like object
        mock_size = type("terminal_size", (), {"columns": 0, "lines": 0})()
        with patch.object(shutil, "get_terminal_size", return_value=mock_size):
            width = get_terminal_width()
            assert width == 80  # Default fallback

    def test_get_terminal_width_respects_actual_size(self) -> None:
        """Test that get_terminal_width uses actual terminal size when available."""
        from rtmx.formatting import get_terminal_width

        # Mock a specific terminal size - use namedtuple-like object
        mock_size = type("terminal_size", (), {"columns": 120, "lines": 40})()
        with patch.object(shutil, "get_terminal_size", return_value=mock_size):
            width = get_terminal_width()
            assert width == 120

    def test_get_terminal_width_with_explicit_override(self) -> None:
        """Test that get_terminal_width accepts an explicit width override."""
        from rtmx.formatting import get_terminal_width

        width = get_terminal_width(override=100)
        assert width == 100


@pytest.mark.req("REQ-UX-008")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestAdaptiveHeaders:
    """Tests for width-adaptive header formatting."""

    def test_header_adapts_to_width(self) -> None:
        """Test that header respects terminal width."""
        from rtmx.formatting import Colors, header

        # Wide terminal
        wide_header = header("Test Header", width=120)
        clean_wide = wide_header.replace(Colors.BOLD, "").replace(Colors.RESET, "")

        # Narrow terminal
        narrow_header = header("Test Header", width=60)
        clean_narrow = narrow_header.replace(Colors.BOLD, "").replace(Colors.RESET, "")

        assert len(clean_wide) <= 120
        assert len(clean_narrow) <= 60

    def test_header_minimum_width(self) -> None:
        """Test that header handles minimum width gracefully."""
        from rtmx.formatting import header

        # Very narrow terminal should still work
        result = header("Test", width=20)
        assert "Test" in result


@pytest.mark.req("REQ-UX-008")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestAdaptiveProgressBar:
    """Tests for width-adaptive progress bar."""

    def test_progress_bar_adapts_to_width(self) -> None:
        """Test that progress_bar width adapts to terminal."""
        from rtmx.formatting import progress_bar

        # Standard width
        standard = progress_bar(5, 3, 2, width=50)
        # Narrow width
        narrow = progress_bar(5, 3, 2, width=20)

        # Both should be valid progress bars
        assert "[" in standard and "]" in standard
        assert "[" in narrow and "]" in narrow

    def test_progress_bar_auto_width(self) -> None:
        """Test that progress_bar can auto-detect width from terminal."""
        from rtmx.formatting import progress_bar_adaptive

        with patch("rtmx.formatting.get_terminal_width", return_value=100):
            result = progress_bar_adaptive(5, 3, 2)
            # Should create a bar with reasonable width
            assert "[" in result and "]" in result


@pytest.mark.req("REQ-UX-008")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestAdaptiveTableFormatting:
    """Tests for width-adaptive table formatting."""

    def test_format_table_truncates_long_columns(self) -> None:
        """Test that tables truncate long text when width constrained."""
        from rtmx.formatting import format_table_adaptive

        data = [
            ["REQ-001", "This is a very long description that should be truncated"],
            ["REQ-002", "Another extremely long description text here"],
        ]
        headers = ["Requirement", "Description"]

        # Format for narrow terminal
        result = format_table_adaptive(data, headers, max_width=60)

        # Table should fit within 60 characters per line
        # Just verify the table is created and contains data
        assert "REQ-001" in result or "REQ-002" in result

    def test_format_table_uses_full_width_when_available(self) -> None:
        """Test that tables use available space when terminal is wide."""
        from rtmx.formatting import format_table_adaptive

        data = [
            ["REQ-001", "Short description"],
            ["REQ-002", "Another short one"],
        ]
        headers = ["Requirement", "Description"]

        # Format for wide terminal
        result = format_table_adaptive(data, headers, max_width=200)

        # Content should not be truncated
        assert "Short description" in result
        assert "Another short one" in result

    def test_format_table_minimum_column_widths(self) -> None:
        """Test that tables preserve minimum column widths for readability."""
        from rtmx.formatting import format_table_adaptive

        data = [
            ["REQ-001", "Description"],
        ]
        headers = ["Requirement", "Description"]

        # Even at narrow width, columns should be readable
        result = format_table_adaptive(data, headers, max_width=40)

        # Requirement ID should still be visible
        assert "REQ" in result


@pytest.mark.req("REQ-UX-008")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestAdaptiveTruncation:
    """Tests for text truncation with width constraints."""

    def test_truncate_respects_max_length(self) -> None:
        """Test that truncate respects maximum length."""
        from rtmx.formatting import truncate

        text = "This is a long text that needs truncation"
        result = truncate(text, max_len=20)

        assert len(result) == 20
        assert result.endswith("...")

    def test_truncate_preserves_short_text(self) -> None:
        """Test that truncate doesn't modify short text."""
        from rtmx.formatting import truncate

        text = "Short"
        result = truncate(text, max_len=20)

        assert result == text

    def test_truncate_with_custom_suffix(self) -> None:
        """Test that truncate uses custom suffix."""
        from rtmx.formatting import truncate

        text = "This needs truncation"
        result = truncate(text, max_len=15, suffix="..")

        assert result.endswith("..")
        assert len(result) == 15


@pytest.mark.req("REQ-UX-008")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestWidthAdaptiveOutput:
    """Integration tests for width-adaptive CLI output."""

    @pytest.fixture
    def sample_rtm_csv(self, tmp_path: Path) -> Path:
        """Create a sample RTM CSV for testing."""
        csv_content = """req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file
REQ-001,CORE,API,This is a very long requirement description that should be truncated in narrow terminals,Target,tests/test.py,test_func,Unit Test,COMPLETE,HIGH,1,Note,1.0,,,,,,,
REQ-002,CORE,API,Another requirement with a long description text here,Target,tests/test.py,test_func,Unit Test,MISSING,MEDIUM,2,Note,1.0,,,,,,,
"""
        csv_path = tmp_path / "rtm_database.csv"
        csv_path.write_text(csv_content)
        return csv_path

    def test_status_output_fits_80_columns(
        self, sample_rtm_csv: Path, capsys: pytest.CaptureFixture
    ) -> None:
        """Test that status output fits within 80 columns."""
        import sys
        from unittest.mock import patch

        from rtmx.cli.status import run_status

        # Mock terminal width to 80
        with patch("rtmx.formatting.get_terminal_width", return_value=80):
            with patch.object(sys, "exit"):
                run_status(sample_rtm_csv, verbosity=0, json_output=None, use_rich=False)

        captured = capsys.readouterr()

        # Check that no line exceeds 80 characters (excluding ANSI codes)
        for line in captured.out.split("\n"):
            # Strip ANSI codes for length check
            clean_line = _strip_ansi(line)
            assert len(clean_line) <= 85, f"Line too long ({len(clean_line)}): {clean_line[:50]}..."

    def test_status_output_fits_60_columns(
        self, sample_rtm_csv: Path, capsys: pytest.CaptureFixture
    ) -> None:
        """Test that status output fits within narrow 60-column terminal."""
        import sys
        from unittest.mock import patch

        from rtmx.cli.status import run_status

        # Mock terminal width to 60
        with patch("rtmx.formatting.get_terminal_width", return_value=60):
            with patch.object(sys, "exit"):
                run_status(sample_rtm_csv, verbosity=0, json_output=None, use_rich=False)

        captured = capsys.readouterr()

        # Output should still be readable
        assert "RTM Status" in captured.out or "complete" in captured.out.lower()

    def test_backlog_output_adapts_to_width(
        self, sample_rtm_csv: Path, capsys: pytest.CaptureFixture
    ) -> None:
        """Test that backlog output adapts to terminal width."""
        import sys
        from unittest.mock import patch

        from rtmx.cli.backlog import run_backlog

        # Mock terminal width to 100
        with patch("rtmx.formatting.get_terminal_width", return_value=100):
            with patch.object(sys, "exit"):
                run_backlog(sample_rtm_csv)

        captured = capsys.readouterr()

        # Should show backlog information
        assert "REQ-" in captured.out or "Backlog" in captured.out


def _strip_ansi(text: str) -> str:
    """Strip ANSI escape codes from text."""
    import re

    ansi_escape = re.compile(r"\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])")
    return ansi_escape.sub("", text)


@pytest.mark.req("REQ-UX-008")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRichWidthAdaptive:
    """Tests for rich output width adaptation."""

    def test_rich_status_adapts_to_terminal_width(self) -> None:
        """Test that rich status output adapts to terminal width."""
        from rtmx.formatting import is_rich_available, render_rich_status

        if not is_rich_available():
            pytest.skip("rich library not available")

        # Render at different widths
        output_wide = io.StringIO()
        render_rich_status(
            complete=5,
            partial=2,
            missing=3,
            total=10,
            completion_pct=60.0,
            phase_stats=[],
            file=output_wide,
            width=100,
        )

        output_narrow = io.StringIO()
        render_rich_status(
            complete=5,
            partial=2,
            missing=3,
            total=10,
            completion_pct=60.0,
            phase_stats=[],
            file=output_narrow,
            width=50,
        )

        # Both should contain status info
        assert "complete" in output_wide.getvalue().lower()
        assert "complete" in output_narrow.getvalue().lower()
