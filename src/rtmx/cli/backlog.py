"""RTMX backlog command.

Displays prioritized backlog with critical path analysis.
Uses tabulated grid format matching Phoenix/Cyclone style.
"""

from __future__ import annotations

import sys
from enum import Enum
from pathlib import Path

from tabulate import tabulate

from rtmx.formatting import Colors, header
from rtmx.models import Priority, RTMDatabase, RTMError, Status


class BacklogView(str, Enum):
    """Backlog view modes."""

    ALL = "all"
    CRITICAL = "critical"
    QUICK_WINS = "quick-wins"
    BLOCKERS = "blockers"


def run_backlog(
    rtm_csv: Path | None,
    phase: int | None = None,
    view: BacklogView = BacklogView.ALL,
    limit: int = 10,
) -> None:
    """Run backlog command.

    Args:
        rtm_csv: Path to RTM CSV or None to auto-discover
        phase: Optional phase filter
        view: View mode (all, critical, quick-wins, blockers)
        limit: Max items to show in summary views
    """
    try:
        db = RTMDatabase.load(rtm_csv)
    except RTMError as e:
        print(f"{Colors.RED}Error: {e}{Colors.RESET}", file=sys.stderr)
        sys.exit(1)

    # Get all requirements and incomplete ones
    all_reqs = list(db)
    incomplete = [
        req for req in db if req.status in (Status.MISSING, Status.PARTIAL, Status.NOT_STARTED)
    ]

    # Filter by phase if specified
    if phase is not None:
        incomplete = [req for req in incomplete if req.phase == phase]

    if not incomplete:
        print(f"{Colors.GREEN}✓ No incomplete requirements!{Colors.RESET}")
        sys.exit(0)

    # Calculate blocking counts (transitive and direct)
    blocking_counts: dict[str, tuple[int, int]] = {}  # (transitive, direct)
    incomplete_ids = {r.req_id for r in incomplete}

    for req in incomplete:
        transitive = db.transitive_blocks(req.req_id)
        transitive_incomplete = sum(1 for b in transitive if b in incomplete_ids)
        direct_incomplete = sum(1 for b in req.blocks if b in incomplete_ids)
        blocking_counts[req.req_id] = (transitive_incomplete, direct_incomplete)

    # Dispatch to appropriate view
    if view == BacklogView.CRITICAL:
        _show_critical_path(incomplete, blocking_counts, phase, limit)
    elif view == BacklogView.QUICK_WINS:
        _show_quick_wins(incomplete, blocking_counts, phase, limit)
    elif view == BacklogView.BLOCKERS:
        _show_blockers(incomplete, blocking_counts, phase, limit)
    else:
        _show_all(all_reqs, incomplete, blocking_counts, phase, limit)


def _format_blocks(transitive: int, direct: int) -> str:
    """Format blocking count display."""
    if transitive == 0:
        return f"{Colors.DIM}-{Colors.RESET}"
    if transitive == direct:
        return f"{Colors.YELLOW}{transitive} reqs{Colors.RESET}"
    return f"{Colors.YELLOW}{transitive} ({direct}){Colors.RESET}"


def _format_status(status: Status) -> str:
    """Format status icon."""
    icons = {
        Status.COMPLETE: f"{Colors.GREEN}✓{Colors.RESET}",
        Status.PARTIAL: f"{Colors.YELLOW}△{Colors.RESET}",
        Status.MISSING: f"{Colors.RED}✗{Colors.RESET}",
        Status.NOT_STARTED: f"{Colors.DIM}○{Colors.RESET}",
    }
    return icons.get(status, "?")


def _format_priority(priority: Priority) -> str:
    """Format priority with color."""
    colors = {
        Priority.P0: Colors.RED,
        Priority.HIGH: Colors.RED,
        Priority.MEDIUM: Colors.YELLOW,
        Priority.LOW: Colors.DIM,
    }
    color = colors.get(priority, Colors.RESET)
    return f"{color}{priority.value}{Colors.RESET}"


def _format_effort(effort: float | None) -> str:
    """Format effort in weeks."""
    if effort is None:
        return f"{Colors.DIM}-{Colors.RESET}"
    return f"{effort:.1f}w"


def _format_phase(phase: int | None) -> str:
    """Format phase number."""
    if phase is None:
        return f"{Colors.DIM}-{Colors.RESET}"
    return f"Phase {phase}"


def _truncate(text: str, max_len: int = 35) -> str:
    """Truncate text with ellipsis."""
    if len(text) <= max_len:
        return text
    return text[: max_len - 3] + "..."


def _sort_by_blocking(incomplete: list, blocking_counts: dict[str, tuple[int, int]]) -> list:
    """Sort requirements by blocking count (highest first)."""

    def sort_key(req):
        trans, direct = blocking_counts.get(req.req_id, (0, 0))
        return (-trans, -direct, req.phase or 99)

    return sorted(incomplete, key=sort_key)


def _print_summary_header(all_reqs: list, incomplete: list, phase: int | None) -> None:
    """Print summary header at top of backlog."""
    # Count by status
    missing = sum(1 for r in incomplete if r.status == Status.MISSING)
    partial = sum(1 for r in incomplete if r.status == Status.PARTIAL)
    total = len(all_reqs)

    # Calculate percentages
    missing_pct = (missing / total * 100) if total > 0 else 0
    partial_pct = (partial / total * 100) if total > 0 else 0

    # Total effort
    total_effort = sum(req.effort_weeks or 0 for req in incomplete)

    # Print title
    title = "Prioritized Backlog"
    if phase:
        title += f" (Phase {phase})"
    print(header(title, "="))
    print()

    # Print summary stats
    print(f"Total Requirements: {total}")
    print(f"  {Colors.RED}✗ MISSING: {missing} ({missing_pct:.1f}%){Colors.RESET}")
    print(f"  {Colors.YELLOW}△ PARTIAL: {partial} ({partial_pct:.1f}%){Colors.RESET}")
    print(f"Estimated Effort: {total_effort:.1f} weeks")
    print()


def _show_all(
    all_reqs: list,
    incomplete: list,
    blocking_counts: dict[str, tuple[int, int]],
    phase: int | None,
    limit: int,
) -> None:
    """Show combined view with Critical Path and Quick Wins sections."""
    # Print summary header
    _print_summary_header(all_reqs, incomplete, phase)

    # Critical Path section
    _show_critical_path_section(incomplete, blocking_counts, limit)

    print()

    # Quick Wins section
    _show_quick_wins_section(incomplete, blocking_counts, limit)


def _show_critical_path_section(
    incomplete: list,
    blocking_counts: dict[str, tuple[int, int]],
    limit: int,
) -> None:
    """Show critical path items section."""
    print(f"{Colors.BOLD}CRITICAL PATH ITEMS (TOP {limit}){Colors.RESET}")
    print()

    # Filter to items that block others and sort by blocking count
    blockers = [(req, blocking_counts.get(req.req_id, (0, 0))) for req in incomplete]
    blockers = [(req, counts) for req, counts in blockers if counts[0] > 0]
    blockers.sort(key=lambda x: (-x[1][0], -x[1][1]))

    if not blockers:
        print(f"{Colors.GREEN}✓ No blocking requirements found!{Colors.RESET}")
        return

    # Limit results
    blockers = blockers[:limit]

    # Build table
    table_data = []
    for i, (req, (trans, direct)) in enumerate(blockers, 1):
        table_data.append(
            [
                i,
                _format_status(req.status),
                req.req_id,
                _truncate(req.requirement_text),
                _format_effort(req.effort_weeks),
                _format_blocks(trans, direct),
                _format_phase(req.phase),
            ]
        )

    headers = ["#", "Status", "Requirement", "Description", "Effort", "Blocks", "Phase"]
    print(tabulate(table_data, headers=headers, tablefmt="grid"))


def _show_quick_wins_section(
    incomplete: list,
    _blocking_counts: dict[str, tuple[int, int]],
    limit: int,
) -> None:
    """Show quick wins section."""
    print(f"{Colors.BOLD}QUICK WINS (<1 week, HIGH priority){Colors.RESET}")
    print()

    # Filter: HIGH/P0 priority, effort <= 1 week, not blocked by others
    incomplete_ids = {r.req_id for r in incomplete}
    quick_wins = []
    for req in incomplete:
        if req.priority not in (Priority.P0, Priority.HIGH):
            continue
        if req.effort_weeks is not None and req.effort_weeks > 1.0:
            continue
        # Check if blocked by incomplete items
        blocked = any(dep in incomplete_ids for dep in (req.dependencies or []))
        if blocked:
            continue
        quick_wins.append(req)

    if not quick_wins:
        print(f"{Colors.DIM}No quick wins found{Colors.RESET}")
        return

    # Sort by effort (ascending), then priority
    priority_order = {Priority.P0: 0, Priority.HIGH: 1}
    quick_wins.sort(key=lambda r: (r.effort_weeks or 0.5, priority_order.get(r.priority, 2)))
    quick_wins = quick_wins[:limit]

    # Build table
    table_data = []
    for i, req in enumerate(quick_wins, 1):
        table_data.append(
            [
                i,
                _format_status(req.status),
                req.req_id,
                _truncate(req.requirement_text),
                _format_effort(req.effort_weeks),
                _format_phase(req.phase),
            ]
        )

    headers = ["#", "Status", "Requirement", "Description", "Effort", "Phase"]
    print(tabulate(table_data, headers=headers, tablefmt="grid"))


def _show_critical_path(
    incomplete: list,
    blocking_counts: dict[str, tuple[int, int]],
    phase: int | None,
    limit: int,
) -> None:
    """Show critical path items (highest blocking impact)."""
    title = "Critical Path"
    if phase:
        title += f" (Phase {phase})"
    print(header(title, "="))
    print()

    _show_critical_path_section(incomplete, blocking_counts, limit)


def _show_quick_wins(
    incomplete: list,
    blocking_counts: dict[str, tuple[int, int]],
    phase: int | None,
    limit: int,
) -> None:
    """Show quick wins (low effort, high priority, not blocked)."""
    title = "Quick Wins"
    if phase:
        title += f" (Phase {phase})"
    print(header(title, "="))
    print()

    _show_quick_wins_section(incomplete, blocking_counts, limit)


def _show_blockers(
    incomplete: list,
    blocking_counts: dict[str, tuple[int, int]],
    phase: int | None,
    limit: int,
) -> None:
    """Show blocking requirements summary."""
    title = "Blocking Requirements"
    if phase:
        title += f" (Phase {phase})"
    print(header(title, "="))
    print()

    # Get all blockers sorted by impact
    blockers = [(req, blocking_counts.get(req.req_id, (0, 0))) for req in incomplete]
    blockers = [(req, counts) for req, counts in blockers if counts[0] > 0]
    blockers.sort(key=lambda x: (-x[1][0], -x[1][1]))

    if not blockers:
        print(f"{Colors.GREEN}✓ No blocking requirements!{Colors.RESET}")
        return

    blockers = blockers[:limit]

    # Build table
    table_data = []
    for i, (req, (trans, direct)) in enumerate(blockers, 1):
        table_data.append(
            [
                i,
                _format_status(req.status),
                req.req_id,
                _truncate(req.requirement_text),
                _format_effort(req.effort_weeks),
                _format_blocks(trans, direct),
                _format_phase(req.phase),
            ]
        )

    headers = ["#", "Status", "Requirement", "Description", "Effort", "Blocks", "Phase"]
    print(tabulate(table_data, headers=headers, tablefmt="grid"))

    # Total blocked
    total_blocked = sum(t for t, _ in blocking_counts.values())
    print()
    print(
        f"{Colors.YELLOW}{len(blockers)} requirements blocking {total_blocked} others{Colors.RESET}"
    )
