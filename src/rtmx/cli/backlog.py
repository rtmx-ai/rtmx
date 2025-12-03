"""RTMX backlog command.

Displays prioritized backlog with critical path analysis.
"""

from __future__ import annotations

import sys
from pathlib import Path

from rtmx.formatting import Colors, format_phase, format_percentage, header
from rtmx.models import Priority, RTMDatabase, RTMError, Status


def run_backlog(
    rtm_csv: Path | None,
    phase: int | None,
    critical: bool,
) -> None:
    """Run backlog command.

    Args:
        rtm_csv: Path to RTM CSV or None to auto-discover
        phase: Optional phase filter
        critical: Show only critical path
    """
    try:
        db = RTMDatabase.load(rtm_csv)
    except RTMError as e:
        print(f"{Colors.RED}Error: {e}{Colors.RESET}", file=sys.stderr)
        sys.exit(1)

    # Get incomplete requirements
    incomplete = [
        req for req in db
        if req.status in (Status.MISSING, Status.PARTIAL, Status.NOT_STARTED)
    ]

    # Filter by phase if specified
    if phase is not None:
        incomplete = [req for req in incomplete if req.phase == phase]

    if not incomplete:
        print(f"{Colors.GREEN}✓ No incomplete requirements!{Colors.RESET}")
        sys.exit(0)

    # Calculate blocking counts
    blocking_counts: dict[str, int] = {}
    for req in incomplete:
        blocked = db.transitive_blocks(req.req_id)
        # Count only incomplete blocked requirements
        incomplete_blocked = sum(1 for b in blocked if b in {r.req_id for r in incomplete})
        blocking_counts[req.req_id] = incomplete_blocked

    # Sort by: priority (P0 > HIGH > MEDIUM > LOW), then blocking count
    priority_order = {Priority.P0: 0, Priority.HIGH: 1, Priority.MEDIUM: 2, Priority.LOW: 3}

    def sort_key(req):
        return (
            priority_order.get(req.priority, 4),
            -blocking_counts.get(req.req_id, 0),
            req.phase or 99,
        )

    incomplete.sort(key=sort_key)

    # Filter to critical path if requested
    if critical:
        incomplete = [req for req in incomplete if blocking_counts.get(req.req_id, 0) > 0]

    # Print header
    title = "Critical Path" if critical else "Backlog"
    if phase:
        title += f" (Phase {phase})"
    print(header(title, "="))
    print()

    # Print requirements
    print(f"{'Priority':<10} {'ID':<18} {'Blocks':<8} {'Phase':<8} {'Description':<40}")
    print("-" * 90)

    for req in incomplete:
        priority_color = _priority_color(req.priority)
        blocks = blocking_counts.get(req.req_id, 0)

        # Truncate description
        desc = req.requirement_text[:40] + "..." if len(req.requirement_text) > 40 else req.requirement_text

        print(
            f"{priority_color}{req.priority.value:<10}{Colors.RESET} "
            f"{req.req_id:<18} "
            f"{blocks:<8} "
            f"P{req.phase or '-':<7} "
            f"{desc:<40}"
        )

    # Summary
    print()
    print(f"{Colors.BOLD}Total: {len(incomplete)} incomplete requirements{Colors.RESET}")

    high_priority = sum(1 for req in incomplete if req.priority in (Priority.P0, Priority.HIGH))
    if high_priority:
        print(f"{Colors.RED}  {high_priority} are HIGH/P0 priority{Colors.RESET}")

    blockers = sum(1 for req in incomplete if blocking_counts.get(req.req_id, 0) > 0)
    if blockers:
        print(f"{Colors.YELLOW}  {blockers} are blocking other requirements{Colors.RESET}")


def _priority_color(priority: Priority) -> str:
    """Get color for priority."""
    color_map = {
        Priority.P0: Colors.RED,
        Priority.HIGH: Colors.YELLOW,
        Priority.MEDIUM: Colors.BLUE,
        Priority.LOW: Colors.DIM,
    }
    return color_map.get(priority, Colors.RESET)
