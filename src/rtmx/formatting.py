"""Terminal formatting utilities for RTMX.

This module provides consistent terminal output formatting:
- ANSI color codes
- Progress bars
- Status indicators
- Table formatting
- Rich library integration (optional)
"""

from __future__ import annotations

import sys
from typing import IO, TYPE_CHECKING

if TYPE_CHECKING:
    from rtmx.models import Priority, Status

# Detect if rich library is available
try:
    from rich.console import Console
    from rich.panel import Panel
    from rich.text import Text

    _RICH_AVAILABLE = True
except ImportError:
    _RICH_AVAILABLE = False


def is_rich_available() -> bool:
    """Check if the rich library is available.

    Returns:
        True if rich is installed and importable
    """
    return _RICH_AVAILABLE


def format_table(
    data: list[list],
    headers: list[str],
    use_rich: bool | None = None,  # noqa: ARG001
) -> str:
    """Format tabular data with aligned columns using tabulate.

    Always uses tabulate with grid format for consistent display.
    The use_rich parameter is kept for API compatibility but ignored.

    Args:
        data: List of rows, each row is a list of cell values
        headers: List of column header strings
        use_rich: Ignored (kept for API compatibility)

    Returns:
        Formatted table string
    """
    return _format_tabulate_table(data, headers)


def _format_tabulate_table(data: list[list], headers: list[str]) -> str:
    """Format table using tabulate library."""
    from tabulate import tabulate

    # Strip any ANSI codes and Text objects for tabulate
    clean_data = []
    for row in data:
        clean_row = []
        for cell in row:
            if isinstance(cell, str):
                clean_row.append(cell)
            elif hasattr(cell, "plain"):
                # Rich Text object
                clean_row.append(cell.plain)
            else:
                clean_row.append(str(cell) if cell is not None else "")
        clean_data.append(clean_row)

    return tabulate(clean_data, headers=headers, tablefmt="grid")


class Colors:
    """ANSI color codes for terminal output."""

    GREEN = "\033[92m"
    YELLOW = "\033[93m"
    RED = "\033[91m"
    BLUE = "\033[94m"
    CYAN = "\033[96m"
    MAGENTA = "\033[95m"
    BOLD = "\033[1m"
    DIM = "\033[2m"
    RESET = "\033[0m"
    UNDERLINE = "\033[4m"

    _enabled = True

    @classmethod
    def disable(cls) -> None:
        """Disable all colors (for non-TTY or when requested)."""
        cls._enabled = False
        cls.GREEN = ""
        cls.YELLOW = ""
        cls.RED = ""
        cls.BLUE = ""
        cls.CYAN = ""
        cls.MAGENTA = ""
        cls.BOLD = ""
        cls.DIM = ""
        cls.RESET = ""
        cls.UNDERLINE = ""

    @classmethod
    def enable(cls) -> None:
        """Re-enable colors."""
        cls._enabled = True
        cls.GREEN = "\033[92m"
        cls.YELLOW = "\033[93m"
        cls.RED = "\033[91m"
        cls.BLUE = "\033[94m"
        cls.CYAN = "\033[96m"
        cls.MAGENTA = "\033[95m"
        cls.BOLD = "\033[1m"
        cls.DIM = "\033[2m"
        cls.RESET = "\033[0m"
        cls.UNDERLINE = "\033[4m"

    @classmethod
    def auto_detect(cls) -> None:
        """Enable colors only if stdout is a TTY."""
        if not sys.stdout.isatty():
            cls.disable()


def status_color(status: Status) -> str:
    """Get color code for a status.

    Args:
        status: Requirement status

    Returns:
        ANSI color code
    """
    from rtmx.models import Status

    color_map = {
        Status.COMPLETE: Colors.GREEN,
        Status.PARTIAL: Colors.YELLOW,
        Status.MISSING: Colors.RED,
        Status.NOT_STARTED: Colors.RED,
    }
    return color_map.get(status, Colors.RESET)


def priority_color(priority: Priority) -> str:
    """Get color code for a priority.

    Args:
        priority: Requirement priority

    Returns:
        ANSI color code
    """
    from rtmx.models import Priority

    color_map = {
        Priority.P0: Colors.RED,
        Priority.HIGH: Colors.YELLOW,
        Priority.MEDIUM: Colors.BLUE,
        Priority.LOW: Colors.DIM,
    }
    return color_map.get(priority, Colors.RESET)


def status_icon(status: Status) -> str:
    """Get icon for a status.

    Args:
        status: Requirement status

    Returns:
        Status icon character
    """
    from rtmx.models import Status

    icon_map = {
        Status.COMPLETE: "✓",
        Status.PARTIAL: "⚠",
        Status.MISSING: "✗",
        Status.NOT_STARTED: "○",
    }
    return icon_map.get(status, "?")


def colorized_status(status: Status) -> str:
    """Get colorized status string with icon.

    Args:
        status: Requirement status

    Returns:
        Colorized status string
    """
    color = status_color(status)
    icon = status_icon(status)
    return f"{color}{icon} {status.value}{Colors.RESET}"


def progress_bar(
    complete: int,
    partial: int,
    missing: int,
    width: int = 50,
) -> str:
    """Create a colored progress bar.

    Args:
        complete: Number of complete items
        partial: Number of partial items
        missing: Number of missing items
        width: Bar width in characters

    Returns:
        Colored progress bar string
    """
    total = complete + partial + missing
    if total == 0:
        return f"[{' ' * width}]"

    complete_width = int((complete / total) * width)
    partial_width = int((partial / total) * width)
    missing_width = width - complete_width - partial_width

    bar = (
        f"{Colors.GREEN}{'█' * complete_width}{Colors.RESET}"
        f"{Colors.YELLOW}{'█' * partial_width}{Colors.RESET}"
        f"{Colors.RED}{'█' * missing_width}{Colors.RESET}"
    )

    return f"[{bar}]"


def percentage_color(pct: float) -> str:
    """Get color based on percentage.

    Args:
        pct: Percentage (0-100)

    Returns:
        ANSI color code
    """
    if pct >= 80:
        return Colors.GREEN
    elif pct >= 50:
        return Colors.YELLOW
    else:
        return Colors.RED


def format_percentage(pct: float) -> str:
    """Format percentage with appropriate color.

    Args:
        pct: Percentage (0-100)

    Returns:
        Colorized percentage string
    """
    color = percentage_color(pct)
    return f"{color}{pct:5.1f}%{Colors.RESET}"


def header(text: str, char: str = "=", width: int = 80) -> str:
    """Create a header line.

    Args:
        text: Header text
        char: Character for border
        width: Total width

    Returns:
        Formatted header string
    """
    padding = (width - len(text) - 2) // 2
    return f"{Colors.BOLD}{char * padding} {text} {char * padding}{Colors.RESET}"


def section(text: str) -> str:
    """Create a section header.

    Args:
        text: Section text

    Returns:
        Formatted section string
    """
    return f"\n{Colors.BOLD}{Colors.CYAN}{text}:{Colors.RESET}\n"


def truncate(text: str, max_len: int = 60, suffix: str = "...") -> str:
    """Truncate text to maximum length.

    Args:
        text: Text to truncate
        max_len: Maximum length
        suffix: Suffix to add if truncated

    Returns:
        Truncated text
    """
    if len(text) <= max_len:
        return text
    return text[: max_len - len(suffix)] + suffix


def format_count(
    complete: int,
    partial: int,
    missing: int,
) -> str:
    """Format status counts with colors.

    Args:
        complete: Complete count
        partial: Partial count
        missing: Missing count

    Returns:
        Formatted count string
    """
    return (
        f"{Colors.GREEN}✓ {complete} complete{Colors.RESET}  "
        f"{Colors.YELLOW}⚠ {partial} partial{Colors.RESET}  "
        f"{Colors.RED}✗ {missing} missing{Colors.RESET}"
    )


def format_phase(phase: int | None) -> str:
    """Format phase number with color.

    Args:
        phase: Phase number

    Returns:
        Colorized phase string
    """
    if phase is None:
        return f"{Colors.DIM}[--]{Colors.RESET}"

    if phase == 1:
        color = Colors.GREEN
    elif phase == 2:
        color = Colors.YELLOW
    else:
        color = Colors.RED

    return f"{color}[P{phase}]{Colors.RESET}"


# Rich output functions (only available when rich is installed)


def render_rich_status(
    complete: int,
    partial: int,
    missing: int,
    total: int,
    completion_pct: float,
    phase_stats: list[tuple[int, int, int, int, float]],
    file: IO[str] | None = None,
    width: int = 70,
) -> None:
    """Render status using rich library.

    Args:
        complete: Number of complete requirements
        partial: Number of partial requirements
        missing: Number of missing requirements
        total: Total requirements
        completion_pct: Overall completion percentage
        phase_stats: List of (phase, complete, partial, missing, pct) tuples
        file: Optional file to write to (for testing)
        width: Panel width in characters (default 70 for 80-column terminals)
    """
    if not _RICH_AVAILABLE:
        raise RuntimeError("rich library is not available")

    console = Console(file=file, force_terminal=file is None, width=width + 4)

    # Bar width leaves room for: "Overall: " (9) + " XX.X%" (6) = 15 chars + padding
    bar_width = width - 20

    # Overall status panel
    overall_bar = _rich_progress_bar(complete, partial, missing, total, bar_width)
    overall_text = Text()
    overall_text.append("Overall: ")
    overall_text.append(overall_bar)
    overall_text.append(f" {completion_pct:5.1f}%\n\n")
    overall_text.append(f"✓ {complete} complete  ", style="green")
    overall_text.append(f"⚠ {partial} partial  ", style="yellow")
    overall_text.append(f"✗ {missing} missing", style="red")

    status_panel = Panel(
        overall_text,
        title="RTMX Status",
        border_style="blue",
        width=width,
    )
    console.print(status_panel)
    console.print()

    # Phase progress panel
    if phase_stats:
        # Bar width for phase: leaves room for "Phase XX: " (10) + " XXX.X% X  (Xc Xp Xm)" (~25)
        phase_bar_width = width - 40

        phase_lines = []
        for phase_num, p_complete, p_partial, p_missing, p_pct in phase_stats:
            p_total = p_complete + p_partial + p_missing
            bar = _rich_progress_bar(p_complete, p_partial, p_missing, p_total, phase_bar_width)

            # Status indicator
            if p_missing == 0 and p_partial == 0 and p_total > 0:
                status = "✓"
                style = "green"
            elif p_pct > 0:
                status = "⚠"
                style = "yellow"
            else:
                status = "✗"
                style = "red"

            line = Text()
            line.append(f"Phase {phase_num:2d}: ", style="bold")
            line.append(bar)
            line.append(f" {p_pct:5.1f}% ")
            line.append(status, style=style)
            line.append(f"  ({p_complete}✓ {p_partial}⚠ {p_missing}✗)", style="dim")
            phase_lines.append(line)

        phase_text = Text("\n").join(phase_lines)
        phase_panel = Panel(
            phase_text,
            title="Phase Progress",
            border_style="cyan",
            width=width,
        )
        console.print(phase_panel)


def _rich_progress_bar(
    complete: int, partial: int, _missing: int, total: int, width: int = 40
) -> Text:
    """Create a rich progress bar with colored segments.

    Colors match the legend: green (complete), yellow (partial), red (missing).
    The _missing parameter is unused but kept for API consistency with callers.

    Args:
        complete: Number of complete items
        partial: Number of partial items
        _missing: Number of missing items (unused, calculated from remainder)
        total: Total items
        width: Bar width in characters

    Returns:
        Rich Text object with colored progress bar
    """
    if total == 0:
        return Text("░" * width, style="dim red")

    complete_width = int((complete / total) * width)
    partial_width = int((partial / total) * width)
    remaining_width = width - complete_width - partial_width

    bar = Text()
    bar.append("█" * complete_width, style="green")
    bar.append("█" * partial_width, style="yellow")
    bar.append("░" * remaining_width, style="dim red")

    return bar
