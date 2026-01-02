"""RTMX status command.

Displays RTM completion status with pytest-style verbosity levels.
"""

from __future__ import annotations

import json
import sys
from collections import defaultdict
from datetime import datetime
from pathlib import Path

from rtmx.formatting import (
    Colors,
    colorized_status,
    format_count,
    format_percentage,
    format_phase,
    header,
    progress_bar,
    truncate,
)
from rtmx.models import RTMDatabase, RTMError, Status


def run_status(
    rtm_csv: Path | None,
    verbosity: int,
    json_output: Path | None,
) -> None:
    """Run status command.

    Args:
        rtm_csv: Path to RTM CSV or None to auto-discover
        verbosity: Verbosity level (0-3)
        json_output: Optional path for JSON export
    """
    try:
        db = RTMDatabase.load(rtm_csv)
    except RTMError as e:
        print(f"{Colors.RED}Error: {e}{Colors.RESET}", file=sys.stderr)
        sys.exit(1)
        return  # Unreachable, but needed for mocked sys.exit in tests

    # Calculate statistics
    counts = db.status_counts()
    complete = counts[Status.COMPLETE]
    partial = counts[Status.PARTIAL]
    missing = counts[Status.MISSING] + counts[Status.NOT_STARTED]
    total = len(db)
    completion_pct = db.completion_percentage()

    # Print header
    print(header("RTM Status Check", "="))
    print()

    # Print based on verbosity
    if verbosity == 0:
        _print_summary(db, complete, partial, missing, total, completion_pct)
    elif verbosity == 1:
        _print_by_category(db, complete, partial, missing, total, completion_pct)
    elif verbosity == 2:
        _print_by_subcategory(db, complete, partial, missing, total, completion_pct)
    else:
        _print_all_requirements(db, complete, partial, missing, total, completion_pct)

    # Print footer
    print()
    _print_footer(db, complete, partial, missing, total, completion_pct)

    # Export JSON if requested
    if json_output:
        _export_json(db, json_output, completion_pct)

    # Exit code based on completion
    sys.exit(0 if completion_pct >= 99 else 1)


def _print_summary(
    _db: RTMDatabase,
    complete: int,
    partial: int,
    missing: int,
    _total: int,
    completion_pct: float,
) -> None:
    """Print summary statistics only."""
    # Progress bar
    bar = progress_bar(complete, partial, missing)
    print(f"Requirements: {bar} {format_percentage(completion_pct)}")
    print()

    # Counts
    total_count = complete + partial + missing
    print(format_count(complete, partial, missing))
    print(f"{Colors.DIM}({total_count} total){Colors.RESET}")


def _print_by_category(
    db: RTMDatabase,
    complete: int,
    partial: int,
    missing: int,
    total: int,
    completion_pct: float,
) -> None:
    """Print breakdown by category."""
    _print_summary(db, complete, partial, missing, total, completion_pct)
    print()
    print(f"{Colors.BOLD}Requirements by Category:{Colors.RESET}")
    print()

    # Group by category
    by_category: dict[str, dict[Status, int]] = defaultdict(lambda: defaultdict(int))
    for req in db:
        by_category[req.category][req.status] += 1

    for category in sorted(by_category.keys()):
        cat_counts = by_category[category]
        cat_complete = cat_counts[Status.COMPLETE]
        cat_partial = cat_counts[Status.PARTIAL]
        cat_missing = cat_counts[Status.MISSING] + cat_counts[Status.NOT_STARTED]
        cat_total = cat_complete + cat_partial + cat_missing

        cat_pct = ((cat_complete + cat_partial * 0.5) / cat_total * 100) if cat_total else 0

        # Status indicator
        if cat_pct >= 80:
            status_color = Colors.GREEN
            status_icon = "✓"
        elif cat_pct >= 50:
            status_color = Colors.YELLOW
            status_icon = "⚠"
        else:
            status_color = Colors.RED
            status_icon = "✗"

        print(
            f"  {status_color}{status_icon} {category:15s}{Colors.RESET}  "
            f"{format_percentage(cat_pct)}  "
            f"{Colors.GREEN}{cat_complete:2d} complete{Colors.RESET}  "
            f"{Colors.YELLOW}{cat_partial:2d} partial{Colors.RESET}  "
            f"{Colors.RED}{cat_missing:2d} missing{Colors.RESET}"
        )


def _print_by_subcategory(
    db: RTMDatabase,
    complete: int,
    partial: int,
    missing: int,
    total: int,
    completion_pct: float,
) -> None:
    """Print breakdown by category and subcategory."""
    _print_summary(db, complete, partial, missing, total, completion_pct)
    print()

    # Group by category and subcategory
    by_category: dict[str, dict[str, dict[Status, int]]] = defaultdict(
        lambda: defaultdict(lambda: defaultdict(int))
    )
    for req in db:
        by_category[req.category][req.subcategory][req.status] += 1

    for category in sorted(by_category.keys()):
        print(f"{Colors.BOLD}{category}:{Colors.RESET}")

        for subcategory in sorted(by_category[category].keys()):
            sub_counts = by_category[category][subcategory]
            sub_complete = sub_counts[Status.COMPLETE]
            sub_partial = sub_counts[Status.PARTIAL]
            sub_missing = sub_counts[Status.MISSING] + sub_counts[Status.NOT_STARTED]
            sub_total = sub_complete + sub_partial + sub_missing

            sub_pct = ((sub_complete + sub_partial * 0.5) / sub_total * 100) if sub_total else 0

            # Status indicator
            if sub_pct >= 80:
                status_color = Colors.GREEN
                status_icon = "✓"
            elif sub_pct >= 50:
                status_color = Colors.YELLOW
                status_icon = "⚠"
            else:
                status_color = Colors.RED
                status_icon = "✗"

            print(
                f"  {status_color}{status_icon} {subcategory:18s}{Colors.RESET}  "
                f"{format_percentage(sub_pct)}  "
                f"{Colors.DIM}({sub_complete}✓ {sub_partial}⚠ {sub_missing}✗){Colors.RESET}"
            )

        print()


def _print_all_requirements(
    db: RTMDatabase,
    complete: int,
    partial: int,
    missing: int,
    total: int,
    completion_pct: float,
) -> None:
    """Print all individual requirements."""
    _print_summary(db, complete, partial, missing, total, completion_pct)
    print()

    # Group by category and subcategory
    by_category: dict[str, dict[str, list]] = defaultdict(lambda: defaultdict(list))
    for req in db:
        by_category[req.category][req.subcategory].append(req)

    for category in sorted(by_category.keys()):
        print(f"{Colors.BOLD}{category}:{Colors.RESET}")

        for subcategory in sorted(by_category[category].keys()):
            print(f"  {Colors.CYAN}{subcategory}:{Colors.RESET}")

            for req in by_category[category][subcategory]:
                # Status indicator
                status_str = colorized_status(req.status)

                # Test indicator
                test_indicator = (
                    f"{Colors.DIM}[T]{Colors.RESET}"
                    if req.has_test()
                    else f"{Colors.DIM}[ ]{Colors.RESET}"
                )

                # Phase
                phase_str = format_phase(req.phase)

                # Requirement text (truncated)
                req_text = truncate(req.requirement_text, 50)

                print(
                    f"    {status_str} {test_indicator} "
                    f"{Colors.BOLD}{req.req_id:15s}{Colors.RESET}  "
                    f"{req_text}  {phase_str}"
                )

            print()
        print()


def _print_footer(
    db: RTMDatabase,
    complete: int,
    partial: int,
    missing: int,
    _total: int,
    completion_pct: float,
) -> None:
    """Print footer with phase breakdown."""
    print(header("Phase Status", "="))
    print()

    # Group by phase
    by_phase: dict[int, dict[Status, int]] = defaultdict(lambda: defaultdict(int))
    for req in db:
        phase = req.phase or 0
        by_phase[phase][req.status] += 1

    for phase in sorted(by_phase.keys()):
        if phase == 0:
            continue  # Skip requirements without phase

        phase_counts = by_phase[phase]
        phase_complete = phase_counts[Status.COMPLETE]
        phase_partial = phase_counts[Status.PARTIAL]
        phase_missing = phase_counts[Status.MISSING] + phase_counts[Status.NOT_STARTED]
        phase_total = phase_complete + phase_partial + phase_missing

        phase_pct = (
            ((phase_complete + phase_partial * 0.5) / phase_total * 100) if phase_total else 0
        )

        # Status indicator
        if phase_missing == 0 and phase_partial == 0 and phase_total > 0:
            status = f"{Colors.GREEN}✓ Complete{Colors.RESET}"
        elif phase_pct > 0:
            status = f"{Colors.YELLOW}⚠ In Progress{Colors.RESET}"
        else:
            status = f"{Colors.RED}✗ Not Started{Colors.RESET}"

        print(
            f"Phase {phase}:  {format_percentage(phase_pct)}  {status}  "
            f"{Colors.DIM}({phase_complete}✓ {phase_partial}⚠ {phase_missing}✗){Colors.RESET}"
        )

    # Find critical blockers
    blockers = [
        req
        for req in db
        if req.status in (Status.MISSING, Status.NOT_STARTED)
        and req.priority.value in ("P0", "HIGH")
        and req.phase == 1
    ]

    if blockers:
        print()
        print(
            f"{Colors.RED}{Colors.BOLD}⚠  {len(blockers)} CRITICAL BLOCKERS "
            f"for Phase 1{Colors.RESET}"
        )
        print(f"{Colors.DIM}   (Run with -vvv to see all requirements){Colors.RESET}")

    # Final summary line
    print()
    from rtmx.formatting import percentage_color

    color = percentage_color(completion_pct)
    print(
        f"{color}{Colors.BOLD}{'=' * 20} "
        f"{complete} complete, {partial} partial, {missing} missing "
        f"({completion_pct:.1f}%) "
        f"{'=' * 20}{Colors.RESET}"
    )


def _export_json(db: RTMDatabase, path: Path, completion_pct: float) -> None:
    """Export status data as JSON."""
    counts = db.status_counts()

    # Build category breakdown
    by_category: dict[str, int] = {}
    for req in db:
        key = f"{req.category}_{req.status.value}"
        by_category[key] = by_category.get(key, 0) + 1

    # Build phase breakdown
    by_phase: dict[str, int] = {}
    for req in db:
        if req.phase:
            key = f"phase{req.phase}_{req.status.value}"
            by_phase[key] = by_phase.get(key, 0) + 1

    data = {
        "generated_at": datetime.now().isoformat(),
        "summary": {
            "total_requirements": len(db),
            "complete": counts[Status.COMPLETE],
            "partial": counts[Status.PARTIAL],
            "missing": counts[Status.MISSING] + counts[Status.NOT_STARTED],
            "completion_pct": completion_pct,
        },
        "by_category": by_category,
        "by_phase": by_phase,
        "all_requirements": [req.to_dict() for req in db],
    }

    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w") as f:
        json.dump(data, f, indent=2)

    print(f"{Colors.GREEN}✅ JSON exported: {path}{Colors.RESET}")
