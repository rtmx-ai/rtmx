"""RTMX deps command.

Show dependency graph visualization.
"""

from __future__ import annotations

import sys
from pathlib import Path

from rtmx.formatting import Colors, header
from rtmx.models import RTMDatabase, RTMError


def run_deps(
    rtm_csv: Path | None,
    category: str | None,
    phase: int | None,
    req_id: str | None,
) -> None:
    """Run deps command.

    Args:
        rtm_csv: Path to RTM CSV or None to auto-discover
        category: Filter by category
        phase: Filter by phase
        req_id: Show deps for specific requirement
    """
    try:
        db = RTMDatabase.load(rtm_csv)
    except RTMError as e:
        print(f"{Colors.RED}Error: {e}{Colors.RESET}", file=sys.stderr)
        sys.exit(1)

    # Build title
    title = "Dependencies"
    if req_id:
        title = f"Dependencies for {req_id}"
    elif category:
        title = f"Dependencies ({category})"
    elif phase:
        title = f"Dependencies (Phase {phase})"

    print(header(title, "="))
    print()

    # Filter requirements
    reqs = list(db)
    if category:
        reqs = [r for r in reqs if r.category == category]
    if phase:
        reqs = [r for r in reqs if r.phase == phase]

    if req_id:
        # Show detailed deps for single requirement
        if not db.exists(req_id):
            print(f"{Colors.RED}Error: Requirement {req_id} not found{Colors.RESET}")
            sys.exit(1)

        req = db.get(req_id)
        _print_requirement_deps(db, req)
    else:
        # Show summary for filtered requirements
        _print_deps_summary(db, reqs)


def _print_requirement_deps(db: RTMDatabase, req) -> None:
    """Print dependencies for a single requirement."""
    print(f"{Colors.BOLD}{req.req_id}{Colors.RESET}: {req.requirement_text}")
    print()

    # Direct dependencies
    if req.dependencies:
        print(f"{Colors.CYAN}Depends on:{Colors.RESET}")
        for dep_id in sorted(req.dependencies):
            if db.exists(dep_id):
                dep = db.get(dep_id)
                status_icon = "✓" if dep.status.value == "COMPLETE" else "✗"
                print(f"  {status_icon} {dep_id}: {dep.requirement_text[:50]}")
            else:
                print(f"  ? {dep_id} (not found)")
        print()

    # Direct blocks
    if req.blocks:
        print(f"{Colors.CYAN}Blocks:{Colors.RESET}")
        for block_id in sorted(req.blocks):
            if db.exists(block_id):
                blocked = db.get(block_id)
                print(f"  → {block_id}: {blocked.requirement_text[:50]}")
            else:
                print(f"  → {block_id} (not found)")
        print()

    # Transitive analysis
    transitive = db.transitive_blocks(req.req_id)
    if transitive:
        print(f"{Colors.CYAN}Transitively blocks {len(transitive)} requirement(s){Colors.RESET}")


def _print_deps_summary(db: RTMDatabase, reqs: list) -> None:
    """Print dependency summary for a set of requirements."""
    # Count dependencies
    dep_counts: dict[str, int] = {}
    block_counts: dict[str, int] = {}

    for req in reqs:
        dep_counts[req.req_id] = len(req.dependencies)
        block_counts[req.req_id] = len(db.transitive_blocks(req.req_id))

    # Sort by blocking count
    sorted_reqs = sorted(reqs, key=lambda r: block_counts.get(r.req_id, 0), reverse=True)

    print(f"{'ID':<18} {'Deps':<6} {'Blocks':<8} {'Description':<45}")
    print("-" * 85)

    for req in sorted_reqs[:30]:
        deps = dep_counts.get(req.req_id, 0)
        blocks = block_counts.get(req.req_id, 0)
        desc = req.requirement_text[:45] + "..." if len(req.requirement_text) > 45 else req.requirement_text

        # Highlight high-blocking requirements
        if blocks > 5:
            color = Colors.RED
        elif blocks > 0:
            color = Colors.YELLOW
        else:
            color = Colors.RESET

        print(f"{req.req_id:<18} {deps:<6} {color}{blocks:<8}{Colors.RESET} {desc}")

    if len(sorted_reqs) > 30:
        print(f"... and {len(sorted_reqs) - 30} more")

    print()
    print(f"{Colors.BOLD}Summary:{Colors.RESET}")
    print(f"  Total requirements: {len(reqs)}")
    print(f"  Requirements with dependencies: {sum(1 for r in reqs if r.dependencies)}")
    print(f"  Requirements blocking others: {sum(1 for r in reqs if block_counts.get(r.req_id, 0) > 0)}")
