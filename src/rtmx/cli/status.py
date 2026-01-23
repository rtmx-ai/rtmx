"""RTMX status command.

Displays RTM completion status with pytest-style verbosity levels.
"""

from __future__ import annotations

import contextlib
import json
import sys
import time
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


def get_file_mtime(path: Path) -> float | None:
    """Get file modification time.

    Args:
        path: Path to file

    Returns:
        Modification time as float, or None if file doesn't exist
    """
    try:
        return path.stat().st_mtime
    except (OSError, FileNotFoundError):
        return None


def clear_screen() -> str:
    """Return ANSI escape codes to clear screen and move cursor to top.

    Returns:
        ANSI escape sequence string
    """
    return "\033[2J\033[H"


def run_status(
    rtm_csv: Path | None,
    verbosity: int,
    json_output: Path | None,
    use_rich: bool | None = None,
    live: bool = False,
) -> None:
    """Run status command.

    Args:
        rtm_csv: Path to RTM CSV or None to auto-discover
        verbosity: Verbosity level (0-3)
        json_output: Optional path for JSON export
        use_rich: Force rich output (True), plain output (False), or auto-detect (None)
        live: Watch file and auto-refresh on changes
    """
    from rtmx.formatting import is_rich_available, render_rich_status

    # Handle live mode
    if live:
        _run_live_status(rtm_csv, verbosity, use_rich)
        return

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

    # Determine if we should use rich output
    should_use_rich = (
        use_rich if use_rich is not None else (is_rich_available() and sys.stdout.isatty())
    )

    if should_use_rich and use_rich is True and not is_rich_available():
        print(
            f"{Colors.RED}Error: --rich requires the 'rich' library. Install with: pip install rtmx[rich]{Colors.RESET}",
            file=sys.stderr,
        )
        sys.exit(1)
        return

    # Use rich output if available and requested
    if should_use_rich and is_rich_available() and verbosity == 0:
        from rtmx.config import load_config

        config = load_config()

        # Calculate phase stats
        by_phase: dict[int, dict[Status, int]] = {}
        for req in db:
            phase = req.phase or 0
            if phase == 0:
                continue
            if phase not in by_phase:
                by_phase[phase] = {}
            by_phase[phase][req.status] = by_phase[phase].get(req.status, 0) + 1

        phase_stats = []
        for phase in sorted(by_phase.keys()):
            phase_counts = by_phase[phase]
            p_complete = phase_counts.get(Status.COMPLETE, 0)
            p_partial = phase_counts.get(Status.PARTIAL, 0)
            p_missing = phase_counts.get(Status.MISSING, 0) + phase_counts.get(
                Status.NOT_STARTED, 0
            )
            p_total = p_complete + p_partial + p_missing
            p_pct = ((p_complete + p_partial * 0.5) / p_total * 100) if p_total else 0
            phase_display = config.get_phase_display(phase)
            phase_stats.append((phase_display, p_complete, p_partial, p_missing, p_pct))

        render_rich_status(complete, partial, missing, total, completion_pct, phase_stats)

        # Export JSON if requested
        if json_output:
            _export_json(db, json_output, completion_pct)

        sys.exit(0 if completion_pct >= 99 else 1)
        return

    # Plain output (original implementation)
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
    from rtmx.config import load_config

    config = load_config()

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

        # Use phase display with name if available
        phase_display = config.get_phase_display(phase)
        print(
            f"{phase_display}:  {format_percentage(phase_pct)}  {status}  "
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


def _run_live_status(
    rtm_csv: Path | None,
    verbosity: int,
    use_rich: bool | None,
) -> None:
    """Run status in live mode, watching for file changes.

    Args:
        rtm_csv: Path to RTM CSV or None to auto-discover
        verbosity: Verbosity level (0-3)
        use_rich: Force rich output (True), plain output (False), or auto-detect (None)
    """
    from rtmx.config import load_config

    # Resolve the RTM CSV path
    if rtm_csv is None:
        config = load_config()
        resolved_path = config.database
        if resolved_path is None or not resolved_path.exists():
            print(
                f"{Colors.RED}Error: Could not find RTM database. "
                f"Run 'rtmx init' or specify --rtm-csv{Colors.RESET}",
                file=sys.stderr,
            )
            sys.exit(1)
            return
    else:
        resolved_path = rtm_csv

    print(f"{Colors.CYAN}Watching {resolved_path} for changes...{Colors.RESET}")
    print(f"{Colors.DIM}Press Ctrl+C to exit{Colors.RESET}")
    print()

    last_mtime: float | None = None

    try:
        while True:
            current_mtime = get_file_mtime(resolved_path)

            # Render if file changed or first run
            if current_mtime != last_mtime:
                # Clear screen
                print(clear_screen(), end="")

                # Run regular status (non-live, suppress exit)
                with contextlib.suppress(SystemExit):
                    run_status(
                        rtm_csv=resolved_path,
                        verbosity=verbosity,
                        json_output=None,
                        use_rich=use_rich,
                        live=False,
                    )

                # Show timestamp and watch message
                now = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
                print()
                print(f"{Colors.DIM}Last updated: {now}{Colors.RESET}")
                print(f"{Colors.CYAN}Watching for changes... (Ctrl+C to exit){Colors.RESET}")

                last_mtime = current_mtime

            # Poll interval - check every 0.5 seconds for responsive updates
            time.sleep(0.5)

    except KeyboardInterrupt:
        print()
        print(f"{Colors.GREEN}✓ Stopped watching{Colors.RESET}")
        sys.exit(0)
