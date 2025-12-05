"""Comprehensive tests for rtmx.formatting module."""

import pytest

from rtmx.formatting import (
    Colors,
    colorized_status,
    format_count,
    format_percentage,
    format_phase,
    header,
    percentage_color,
    priority_color,
    progress_bar,
    section,
    status_color,
    status_icon,
    truncate,
)
from rtmx.models import Priority, Status


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_colors_has_ansi_codes():
    """Test Colors class has proper ANSI escape codes."""
    Colors.enable()  # Ensure colors are enabled (may be disabled by other tests)
    assert Colors.GREEN == "\033[92m"
    assert Colors.YELLOW == "\033[93m"
    assert Colors.RED == "\033[91m"
    assert Colors.BLUE == "\033[94m"
    assert Colors.CYAN == "\033[96m"
    assert Colors.MAGENTA == "\033[95m"
    assert Colors.BOLD == "\033[1m"
    assert Colors.DIM == "\033[2m"
    assert Colors.RESET == "\033[0m"
    assert Colors.UNDERLINE == "\033[4m"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_colors_disable():
    """Test Colors.disable removes all ANSI codes."""
    Colors.disable()

    assert Colors.GREEN == ""
    assert Colors.YELLOW == ""
    assert Colors.RED == ""
    assert Colors.BLUE == ""
    assert Colors.CYAN == ""
    assert Colors.MAGENTA == ""
    assert Colors.BOLD == ""
    assert Colors.DIM == ""
    assert Colors.RESET == ""
    assert Colors.UNDERLINE == ""
    assert Colors._enabled is False

    # Clean up
    Colors.enable()


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_colors_enable():
    """Test Colors.enable restores ANSI codes."""
    Colors.disable()
    Colors.enable()

    assert Colors.GREEN == "\033[92m"
    assert Colors.YELLOW == "\033[93m"
    assert Colors.RED == "\033[91m"
    assert Colors.BOLD == "\033[1m"
    assert Colors.RESET == "\033[0m"
    assert Colors._enabled is True


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_status_color_complete():
    """Test status_color returns green for COMPLETE."""
    color = status_color(Status.COMPLETE)
    assert color == Colors.GREEN


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_status_color_partial():
    """Test status_color returns yellow for PARTIAL."""
    color = status_color(Status.PARTIAL)
    assert color == Colors.YELLOW


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_status_color_missing():
    """Test status_color returns red for MISSING."""
    color = status_color(Status.MISSING)
    assert color == Colors.RED


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_status_color_not_started():
    """Test status_color returns red for NOT_STARTED."""
    color = status_color(Status.NOT_STARTED)
    assert color == Colors.RED


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_priority_color_p0():
    """Test priority_color returns red for P0."""
    color = priority_color(Priority.P0)
    assert color == Colors.RED


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_priority_color_high():
    """Test priority_color returns yellow for HIGH."""
    color = priority_color(Priority.HIGH)
    assert color == Colors.YELLOW


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_priority_color_medium():
    """Test priority_color returns blue for MEDIUM."""
    color = priority_color(Priority.MEDIUM)
    assert color == Colors.BLUE


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_priority_color_low():
    """Test priority_color returns dim for LOW."""
    color = priority_color(Priority.LOW)
    assert color == Colors.DIM


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_status_icon_complete():
    """Test status_icon returns checkmark for COMPLETE."""
    icon = status_icon(Status.COMPLETE)
    assert icon == "✓"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_status_icon_partial():
    """Test status_icon returns warning for PARTIAL."""
    icon = status_icon(Status.PARTIAL)
    assert icon == "⚠"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_status_icon_missing():
    """Test status_icon returns X for MISSING."""
    icon = status_icon(Status.MISSING)
    assert icon == "✗"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_status_icon_not_started():
    """Test status_icon returns circle for NOT_STARTED."""
    icon = status_icon(Status.NOT_STARTED)
    assert icon == "○"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_colorized_status_complete():
    """Test colorized_status formats COMPLETE with color and icon."""
    result = colorized_status(Status.COMPLETE)
    assert Colors.GREEN in result
    assert "✓" in result
    assert "COMPLETE" in result
    assert Colors.RESET in result


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_colorized_status_partial():
    """Test colorized_status formats PARTIAL with color and icon."""
    result = colorized_status(Status.PARTIAL)
    assert Colors.YELLOW in result
    assert "⚠" in result
    assert "PARTIAL" in result


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_progress_bar_empty():
    """Test progress_bar with zero items returns empty bar."""
    result = progress_bar(0, 0, 0, width=10)
    assert result == "[          ]"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_progress_bar_all_complete():
    """Test progress_bar with all complete shows green bar."""
    result = progress_bar(10, 0, 0, width=10)
    assert Colors.GREEN in result
    assert "█" in result


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_progress_bar_mixed():
    """Test progress_bar with mixed statuses."""
    result = progress_bar(5, 3, 2, width=10)
    assert Colors.GREEN in result
    assert Colors.YELLOW in result
    assert Colors.RED in result
    assert result.startswith("[")
    assert result.endswith("]")


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_progress_bar_custom_width():
    """Test progress_bar respects custom width."""
    result = progress_bar(1, 1, 1, width=20)
    # Should have brackets plus 20 characters of content (plus ANSI codes)
    assert "[" in result
    assert "]" in result


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_percentage_color_high():
    """Test percentage_color returns green for >= 80%."""
    assert percentage_color(80.0) == Colors.GREEN
    assert percentage_color(90.0) == Colors.GREEN
    assert percentage_color(100.0) == Colors.GREEN


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_percentage_color_medium():
    """Test percentage_color returns yellow for 50-79%."""
    assert percentage_color(50.0) == Colors.YELLOW
    assert percentage_color(60.0) == Colors.YELLOW
    assert percentage_color(79.9) == Colors.YELLOW


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_percentage_color_low():
    """Test percentage_color returns red for < 50%."""
    assert percentage_color(0.0) == Colors.RED
    assert percentage_color(25.0) == Colors.RED
    assert percentage_color(49.9) == Colors.RED


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_format_percentage_high():
    """Test format_percentage formats high percentage in green."""
    result = format_percentage(85.5)
    assert Colors.GREEN in result
    assert "85.5%" in result
    assert Colors.RESET in result


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_format_percentage_medium():
    """Test format_percentage formats medium percentage in yellow."""
    result = format_percentage(65.0)
    assert Colors.YELLOW in result
    assert "65.0%" in result


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_format_percentage_low():
    """Test format_percentage formats low percentage in red."""
    result = format_percentage(25.0)
    assert Colors.RED in result
    assert "25.0%" in result


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_format_percentage_decimal_places():
    """Test format_percentage formats to one decimal place."""
    result = format_percentage(33.333)
    assert "33.3%" in result


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_header_default():
    """Test header creates centered header with default params."""
    result = header("Test Header")
    assert "Test Header" in result
    assert Colors.BOLD in result
    assert Colors.RESET in result
    assert "=" in result


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_header_custom_char():
    """Test header uses custom character for border."""
    result = header("Test", char="-")
    assert "Test" in result
    assert "-" in result
    assert "=" not in result


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_header_custom_width():
    """Test header respects custom width."""
    result = header("Test", width=40)
    # Strip ANSI codes to check actual length
    clean = result.replace(Colors.BOLD, "").replace(Colors.RESET, "")
    assert len(clean) <= 40


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_section():
    """Test section creates formatted section header."""
    result = section("My Section")
    assert "My Section:" in result
    assert Colors.BOLD in result
    assert Colors.CYAN in result
    assert Colors.RESET in result
    assert result.startswith("\n")
    assert result.endswith("\n")


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_truncate_no_truncation_needed():
    """Test truncate returns original text when under max length."""
    text = "Short text"
    result = truncate(text, max_len=20)
    assert result == text


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_truncate_exact_length():
    """Test truncate returns original when exactly at max length."""
    text = "x" * 10
    result = truncate(text, max_len=10)
    assert result == text


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_truncate_with_default_suffix():
    """Test truncate adds '...' suffix when text too long."""
    text = "This is a very long text that needs truncation"
    result = truncate(text, max_len=20)
    assert result.endswith("...")
    assert len(result) == 20


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_truncate_with_custom_suffix():
    """Test truncate uses custom suffix."""
    text = "Long text here"
    result = truncate(text, max_len=10, suffix=">>")
    assert result.endswith(">>")
    assert len(result) == 10


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_format_count():
    """Test format_count formats status counts with colors and icons."""
    result = format_count(5, 3, 2)
    assert Colors.GREEN in result
    assert Colors.YELLOW in result
    assert Colors.RED in result
    assert "✓ 5 complete" in result
    assert "⚠ 3 partial" in result
    assert "✗ 2 missing" in result


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_format_count_zeros():
    """Test format_count handles zero counts."""
    result = format_count(0, 0, 0)
    assert "✓ 0 complete" in result
    assert "⚠ 0 partial" in result
    assert "✗ 0 missing" in result


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_format_phase_none():
    """Test format_phase handles None with dim placeholder."""
    result = format_phase(None)
    assert Colors.DIM in result
    assert "[--]" in result
    assert Colors.RESET in result


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_format_phase_one():
    """Test format_phase shows phase 1 in green."""
    result = format_phase(1)
    assert Colors.GREEN in result
    assert "[P1]" in result
    assert Colors.RESET in result


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_format_phase_two():
    """Test format_phase shows phase 2 in yellow."""
    result = format_phase(2)
    assert Colors.YELLOW in result
    assert "[P2]" in result


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_format_phase_higher():
    """Test format_phase shows phase 3+ in red."""
    result = format_phase(3)
    assert Colors.RED in result
    assert "[P3]" in result

    result = format_phase(10)
    assert Colors.RED in result
    assert "[P10]" in result


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_colors_auto_detect_non_tty(monkeypatch):
    """Test Colors.auto_detect disables colors for non-TTY output."""
    # Mock isatty to return False
    monkeypatch.setattr("sys.stdout.isatty", lambda: False)

    Colors.enable()  # Ensure starting state
    Colors.auto_detect()

    assert Colors._enabled is False
    assert Colors.GREEN == ""

    # Clean up
    Colors.enable()


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_colors_disable_and_enable_idempotent():
    """Test Colors disable/enable can be called multiple times."""
    Colors.enable()
    Colors.enable()
    assert Colors.GREEN == "\033[92m"

    Colors.disable()
    Colors.disable()
    assert Colors.GREEN == ""

    # Clean up
    Colors.enable()
