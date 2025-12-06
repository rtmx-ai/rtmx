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

    # Get incomplete requirements
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
        _show_all(incomplete, blocking_counts, phase)


def _format_blocks(transitive: int, direct: int) -> str:
    """Format blocking count display."""
    if transitive == 0:
        return f"{Colors.DIM}-{Colors.RESET}"
    if transitive == direct:
        return f"{Colors.YELLOW}{transitive}{Colors.RESET}"
    return f"{Colors.YELLOW}{transitive}{Colors.RESET} {Colors.DIM}({direct}){Colors.RESET}"


def _format_status(status: Status) -> str:
    """Format status icon."""
    icons = {
        Status.COMPLETE: f"{Colors.GREEN}✓{Colors.RESET}",
        Status.PARTIAL: f"{Colors.YELLOW}⚠{Colors.RESET}",
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
    return str(phase)


def _truncate(text: str, max_len: int = 35) -> str:
    """Truncate text with ellipsis."""
    if len(text) <= max_len:
        return text
    return text[: max_len - 3] + "..."


def _sort_by_priority(incomplete: list, blocking_counts: dict[str, tuple[int, int]]) -> list:
    """Sort requirements by priority and blocking count."""
    priority_order = {Priority.P0: 0, Priority.HIGH: 1, Priority.MEDIUM: 2, Priority.LOW: 3}

    def sort_key(req):
        trans, _ = blocking_counts.get(req.req_id, (0, 0))
        return (
            priority_order.get(req.priority, 4),
            -trans,  # More blocks = higher priority
            req.phase or 99,
        )

    return sorted(incomplete, key=sort_key)


def _show_all(
    incomplete: list,
    blocking_counts: dict[str, tuple[int, int]],
    phase: int | None,
) -> None:
    """Show all incomplete requirements in grid format."""
    title = "Backlog"
    if phase:
        title += f" (Phase {phase})"
    print(header(title, "="))
    print()

    # Sort by priority
    sorted_reqs = _sort_by_priority(incomplete, blocking_counts)

    # Build table data
    table_data = []
    for i, req in enumerate(sorted_reqs, 1):
        trans, direct = blocking_counts.get(req.req_id, (0, 0))
        table_data.append(
            [
                i,
                _format_status(req.status),
                req.req_id,
                _truncate(req.requirement_text),
                _format_priority(req.priority),
                _format_blocks(trans, direct),
                _format_phase(req.phase),
            ]
        )

    headers = ["#", "", "Requirement", "Description", "Priority", "Blocks", "Phase"]
    print(tabulate(table_data, headers=headers, tablefmt="grid"))

    # Summary
    _print_summary(incomplete, blocking_counts)


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

    # Filter to items that block others and sort by blocking count
    blockers = [(req, blocking_counts.get(req.req_id, (0, 0))) for req in incomplete]
    blockers = [(req, counts) for req, counts in blockers if counts[0] > 0]
    blockers.sort(key=lambda x: (-x[1][0], -x[1][1]))  # Sort by transitive, then direct

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
                _format_blocks(trans, direct),
                _format_phase(req.phase),
            ]
        )

    headers = ["#", "", "Requirement", "Description", "Blocks", "Phase"]
    print(tabulate(table_data, headers=headers, tablefmt="grid"))

    print()
    print(f"{Colors.DIM}Showing top {len(blockers)} critical path items{Colors.RESET}")


def _show_quick_wins(
    incomplete: list,
    _blocking_counts: dict[str, tuple[int, int]],
    phase: int | None,
    limit: int,
) -> None:
    """Show quick wins (low effort, high priority, not blocked)."""
    title = "Quick Wins"
    if phase:
        title += f" (Phase {phase})"
    print(header(title, "="))
    print()

    # Filter: HIGH/P0 priority, effort <= 1 week, not blocked by others
    quick_wins = []
    for req in incomplete:
        if req.priority not in (Priority.P0, Priority.HIGH):
            continue
        if req.effort_weeks is not None and req.effort_weeks > 1.0:
            continue
        # Check if blocked by incomplete items
        blocked = any(dep in {r.req_id for r in incomplete} for dep in (req.dependencies or []))
        if blocked:
            continue
        quick_wins.append(req)

    if not quick_wins:
        print(
            f"{Colors.DIM}No quick wins found (HIGH/P0 priority, ≤1 week, unblocked){Colors.RESET}"
        )
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

    headers = ["#", "", "Requirement", "Description", "Effort", "Phase"]
    print(tabulate(table_data, headers=headers, tablefmt="grid"))

    print()
    print(
        f"{Colors.DIM}Showing {len(quick_wins)} quick wins (HIGH/P0, ≤1 week, unblocked){Colors.RESET}"
    )


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
                _format_priority(req.priority),
                _format_blocks(trans, direct),
            ]
        )

    headers = ["#", "", "Requirement", "Description", "Priority", "Blocks"]
    print(tabulate(table_data, headers=headers, tablefmt="grid"))

    # Total blocked
    total_blocked = sum(t for t, _ in blocking_counts.values())
    print()
    print(
        f"{Colors.YELLOW}{len(blockers)} requirements blocking {total_blocked} others{Colors.RESET}"
    )


def _print_summary(incomplete: list, blocking_counts: dict[str, tuple[int, int]]) -> None:
    """Print backlog summary."""
    print()
    print(f"{Colors.BOLD}Total: {len(incomplete)} incomplete requirements{Colors.RESET}")

    high_priority = sum(1 for req in incomplete if req.priority in (Priority.P0, Priority.HIGH))
    if high_priority:
        print(f"{Colors.RED}  {high_priority} HIGH/P0 priority{Colors.RESET}")

    blockers = sum(1 for req in incomplete if blocking_counts.get(req.req_id, (0, 0))[0] > 0)
    if blockers:
        print(f"{Colors.YELLOW}  {blockers} blocking other requirements{Colors.RESET}")

    # Effort estimate
    total_effort = sum(req.effort_weeks or 0 for req in incomplete)
    if total_effort > 0:
        print(f"{Colors.DIM}  {total_effort:.1f} weeks estimated effort{Colors.RESET}")
