"""RTMX diff command.

Compare RTM databases before and after changes.
"""

from __future__ import annotations

import json
import sys
from pathlib import Path

from rtmx.comparison import compare_databases
from rtmx.formatting import Colors, header


def run_diff(
    baseline_path: Path,
    current_path: Path | None = None,
    format_type: str = "terminal",
    output_path: Path | None = None,
) -> None:
    """Run diff command.

    Args:
        baseline_path: Path to baseline RTM CSV
        current_path: Path to current RTM CSV (default: docs/rtm_database.csv)
        format_type: Output format (terminal, markdown, json)
        output_path: Optional output file path
    """
    # Default current path
    if current_path is None:
        current_path = Path("docs/rtm_database.csv")

    # Validate paths
    if not baseline_path.exists():
        print(
            f"{Colors.RED}Error: Baseline not found: {baseline_path}{Colors.RESET}", file=sys.stderr
        )
        sys.exit(1)
        return  # Unreachable, but needed for mocked sys.exit in tests

    if not current_path.exists():
        print(
            f"{Colors.RED}Error: Current RTM not found: {current_path}{Colors.RESET}",
            file=sys.stderr,
        )
        sys.exit(1)
        return  # Unreachable, but needed for mocked sys.exit in tests

    # Compare databases
    try:
        report = compare_databases(baseline_path, current_path)
    except Exception as e:
        print(f"{Colors.RED}Error comparing databases: {e}{Colors.RESET}", file=sys.stderr)
        sys.exit(1)

    # Generate output
    if format_type == "json":
        output = json.dumps(report.to_dict(), indent=2)
    elif format_type == "markdown":
        output = report.to_markdown()
    else:
        output = format_terminal_report(report)

    # Write or print output
    if output_path:
        output_path.write_text(output)
        print(f"Diff written to: {output_path}")
    else:
        print(output)

    # Exit code based on status
    if report.summary_status == "breaking":
        sys.exit(2)
    elif report.summary_status in ("regressed", "degraded"):
        sys.exit(1)
    else:
        sys.exit(0)


def format_terminal_report(report) -> str:
    """Format comparison report for terminal output."""
    lines = []

    # Header
    lines.append(header("RTM Comparison", "="))
    lines.append("")

    # Status with color
    status_colors = {
        "breaking": Colors.RED,
        "regressed": Colors.RED,
        "degraded": Colors.YELLOW,
        "improved": Colors.GREEN,
        "stable": Colors.GREEN,
    }
    color = status_colors.get(report.summary_status, Colors.RESET)
    lines.append(f"Status: {color}{report.summary_status.upper()}{Colors.RESET}")
    lines.append("")

    # File paths
    lines.append(f"Baseline: {report.baseline_path}")
    lines.append(f"Current:  {report.current_path}")
    lines.append("")

    # Summary table
    lines.append("-" * 60)
    lines.append(f"{'Metric':<25} {'Baseline':>12} {'Current':>12} {'Delta':>10}")
    lines.append("-" * 60)

    # Requirement count
    delta_color = Colors.GREEN if report.req_count_delta >= 0 else Colors.RED
    lines.append(
        f"{'Requirements':<25} {report.baseline_req_count:>12} {report.current_req_count:>12} "
        f"{delta_color}{report.req_count_delta:>+10}{Colors.RESET}"
    )

    # Completion
    delta_color = Colors.GREEN if report.completion_delta >= 0 else Colors.RED
    lines.append(
        f"{'Completion %':<25} {report.baseline_completion:>11.1f}% {report.current_completion:>11.1f}% "
        f"{delta_color}{report.completion_delta:>+9.1f}%{Colors.RESET}"
    )

    # Cycles
    cycle_delta = report.current_cycles - report.baseline_cycles
    delta_color = (
        Colors.RED if cycle_delta > 0 else (Colors.GREEN if cycle_delta < 0 else Colors.RESET)
    )
    lines.append(
        f"{'Circular Dependencies':<25} {report.baseline_cycles:>12} {report.current_cycles:>12} "
        f"{delta_color}{cycle_delta:>+10}{Colors.RESET}"
    )

    # Reciprocity
    recip_delta = report.current_reciprocity_violations - report.baseline_reciprocity_violations
    delta_color = (
        Colors.RED if recip_delta > 0 else (Colors.GREEN if recip_delta < 0 else Colors.RESET)
    )
    lines.append(
        f"{'Reciprocity Violations':<25} {report.baseline_reciprocity_violations:>12} "
        f"{report.current_reciprocity_violations:>12} {delta_color}{recip_delta:>+10}{Colors.RESET}"
    )

    lines.append("-" * 60)
    lines.append("")

    # Status distribution
    lines.append("Status Distribution:")
    all_statuses = set(report.baseline_status_counts.keys()) | set(
        report.current_status_counts.keys()
    )
    for status in sorted(all_statuses):
        baseline_count = report.baseline_status_counts.get(status, 0)
        current_count = report.current_status_counts.get(status, 0)
        delta = current_count - baseline_count
        delta_str = f"({delta:+d})" if delta != 0 else ""
        lines.append(f"  {status:<12} {baseline_count:>5} -> {current_count:>5} {delta_str}")
    lines.append("")

    # Added requirements
    if report.added_requirements:
        lines.append(
            f"{Colors.GREEN}Added Requirements ({len(report.added_requirements)}):{Colors.RESET}"
        )
        for req_id in report.added_requirements[:15]:
            lines.append(f"  + {req_id}")
        if len(report.added_requirements) > 15:
            lines.append(f"  ... and {len(report.added_requirements) - 15} more")
        lines.append("")

    # Removed requirements
    if report.removed_requirements:
        lines.append(
            f"{Colors.RED}Removed Requirements ({len(report.removed_requirements)}) [BREAKING]:{Colors.RESET}"
        )
        for req_id in report.removed_requirements[:15]:
            lines.append(f"  - {req_id}")
        if len(report.removed_requirements) > 15:
            lines.append(f"  ... and {len(report.removed_requirements) - 15} more")
        lines.append("")

    # Status changes
    actual_changes = [sc for sc in report.status_changes if sc.change_type == "changed"]
    if actual_changes:
        lines.append(f"Status Changes ({len(actual_changes)}):")
        for sc in actual_changes[:15]:
            old = sc.old_status.value if sc.old_status else "N/A"
            new = sc.new_status.value if sc.new_status else "N/A"

            # Color based on direction
            if new == "COMPLETE":
                color = Colors.GREEN
            elif new == "MISSING" and old in ("COMPLETE", "PARTIAL"):
                color = Colors.RED
            else:
                color = Colors.YELLOW

            lines.append(f"  {sc.req_id}: {old} -> {color}{new}{Colors.RESET}")
        if len(actual_changes) > 15:
            lines.append(f"  ... and {len(actual_changes) - 15} more")
        lines.append("")

    return "\n".join(lines)
