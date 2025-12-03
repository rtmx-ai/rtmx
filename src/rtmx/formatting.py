"""Terminal formatting utilities for RTMX.

This module provides consistent terminal output formatting:
- ANSI color codes
- Progress bars
- Status indicators
- Table formatting
"""

from __future__ import annotations

import sys
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from rtmx.models import Priority, Status


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
