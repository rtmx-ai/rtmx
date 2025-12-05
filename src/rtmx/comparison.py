"""RTMX comparison utilities.

Compare RTM databases before and after integration or changes.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

from rtmx.models import RTMDatabase, Status


@dataclass
class StatusChange:
    """Track a status change for a requirement."""

    req_id: str
    old_status: Status | None
    new_status: Status | None

    @property
    def change_type(self) -> str:
        """Describe the type of change."""
        if self.old_status is None:
            return "added"
        if self.new_status is None:
            return "removed"
        if self.old_status != self.new_status:
            return "changed"
        return "unchanged"


@dataclass
class ComparisonReport:
    """Before/after comparison of RTM databases."""

    baseline_path: str
    current_path: str
    baseline_req_count: int = 0
    current_req_count: int = 0
    baseline_completion: float = 0.0
    current_completion: float = 0.0
    baseline_status_counts: dict[str, int] = field(default_factory=dict)
    current_status_counts: dict[str, int] = field(default_factory=dict)
    added_requirements: list[str] = field(default_factory=list)
    removed_requirements: list[str] = field(default_factory=list)
    status_changes: list[StatusChange] = field(default_factory=list)
    baseline_cycles: int = 0
    current_cycles: int = 0
    baseline_reciprocity_violations: int = 0
    current_reciprocity_violations: int = 0

    @property
    def req_count_delta(self) -> int:
        """Change in requirement count."""
        return self.current_req_count - self.baseline_req_count

    @property
    def completion_delta(self) -> float:
        """Change in completion percentage."""
        return self.current_completion - self.baseline_completion

    @property
    def has_breaking_changes(self) -> bool:
        """Check if there are breaking changes."""
        return len(self.removed_requirements) > 0

    @property
    def summary_status(self) -> str:
        """Overall comparison status."""
        if self.has_breaking_changes:
            return "breaking"
        if self.completion_delta < -1.0:
            return "regressed"
        if self.current_cycles > self.baseline_cycles:
            return "degraded"
        if self.completion_delta > 0:
            return "improved"
        return "stable"

    def to_dict(self) -> dict[str, Any]:
        """Convert to dictionary for JSON export."""
        return {
            "baseline_path": self.baseline_path,
            "current_path": self.current_path,
            "summary": {
                "status": self.summary_status,
                "req_count_delta": self.req_count_delta,
                "completion_delta": self.completion_delta,
                "breaking_changes": self.has_breaking_changes,
            },
            "baseline": {
                "req_count": self.baseline_req_count,
                "completion": self.baseline_completion,
                "status_counts": self.baseline_status_counts,
                "cycles": self.baseline_cycles,
                "reciprocity_violations": self.baseline_reciprocity_violations,
            },
            "current": {
                "req_count": self.current_req_count,
                "completion": self.current_completion,
                "status_counts": self.current_status_counts,
                "cycles": self.current_cycles,
                "reciprocity_violations": self.current_reciprocity_violations,
            },
            "changes": {
                "added": self.added_requirements,
                "removed": self.removed_requirements,
                "status_changes": [
                    {
                        "req_id": sc.req_id,
                        "old_status": sc.old_status.value if sc.old_status else None,
                        "new_status": sc.new_status.value if sc.new_status else None,
                        "change_type": sc.change_type,
                    }
                    for sc in self.status_changes
                    if sc.change_type != "unchanged"
                ],
            },
        }

    def to_markdown(self) -> str:
        """Convert to markdown format for PR comments."""
        lines = []

        # Header with status
        status_emoji = {
            "breaking": "X",
            "regressed": "!",
            "degraded": "?",
            "improved": "^",
            "stable": "=",
        }
        emoji = status_emoji.get(self.summary_status, "?")
        lines.append(f"### RTM Comparison: {self.summary_status.upper()} [{emoji}]")
        lines.append("")

        # Summary table
        lines.append("| Metric | Baseline | Current | Delta |")
        lines.append("|--------|----------|---------|-------|")
        lines.append(
            f"| Requirements | {self.baseline_req_count} | {self.current_req_count} | {self.req_count_delta:+d} |"
        )
        lines.append(
            f"| Completion | {self.baseline_completion:.1f}% | {self.current_completion:.1f}% | {self.completion_delta:+.1f}% |"
        )
        lines.append(
            f"| Cycles | {self.baseline_cycles} | {self.current_cycles} | {self.current_cycles - self.baseline_cycles:+d} |"
        )
        lines.append(
            f"| Reciprocity | {self.baseline_reciprocity_violations} | {self.current_reciprocity_violations} | {self.current_reciprocity_violations - self.baseline_reciprocity_violations:+d} |"
        )
        lines.append("")

        # Added requirements
        if self.added_requirements:
            lines.append(f"**Added ({len(self.added_requirements)}):**")
            for req_id in self.added_requirements[:10]:
                lines.append(f"- `{req_id}`")
            if len(self.added_requirements) > 10:
                lines.append(f"- ... and {len(self.added_requirements) - 10} more")
            lines.append("")

        # Removed requirements
        if self.removed_requirements:
            lines.append(f"**Removed ({len(self.removed_requirements)}):** [BREAKING]")
            for req_id in self.removed_requirements[:10]:
                lines.append(f"- `{req_id}`")
            if len(self.removed_requirements) > 10:
                lines.append(f"- ... and {len(self.removed_requirements) - 10} more")
            lines.append("")

        # Status changes
        actual_changes = [sc for sc in self.status_changes if sc.change_type == "changed"]
        if actual_changes:
            lines.append(f"**Status Changes ({len(actual_changes)}):**")
            for sc in actual_changes[:10]:
                old = sc.old_status.value if sc.old_status else "N/A"
                new = sc.new_status.value if sc.new_status else "N/A"
                lines.append(f"- `{sc.req_id}`: {old} -> {new}")
            if len(actual_changes) > 10:
                lines.append(f"- ... and {len(actual_changes) - 10} more")
            lines.append("")

        return "\n".join(lines)


def compare_databases(
    baseline: RTMDatabase | Path | str,
    current: RTMDatabase | Path | str,
) -> ComparisonReport:
    """Compare two RTM databases.

    Args:
        baseline: Baseline database (path or loaded database)
        current: Current database (path or loaded database)

    Returns:
        ComparisonReport with comparison results
    """
    from rtmx.validation import check_reciprocity

    # Load databases if paths provided
    if isinstance(baseline, str | Path):
        baseline_path = str(baseline)
        baseline = RTMDatabase.load(Path(baseline))
    else:
        baseline_path = str(baseline._path) if baseline._path else "memory"

    if isinstance(current, str | Path):
        current_path = str(current)
        current = RTMDatabase.load(Path(current))
    else:
        current_path = str(current._path) if current._path else "memory"

    # Get requirement IDs
    baseline_ids = {req.req_id for req in baseline}
    current_ids = {req.req_id for req in current}

    # Calculate added/removed
    added = sorted(current_ids - baseline_ids)
    removed = sorted(baseline_ids - current_ids)

    # Calculate status changes
    status_changes: list[StatusChange] = []
    for req_id in baseline_ids | current_ids:
        baseline_req = baseline.get(req_id) if req_id in baseline_ids else None
        current_req = current.get(req_id) if req_id in current_ids else None

        old_status = baseline_req.status if baseline_req else None
        new_status = current_req.status if current_req else None

        status_changes.append(
            StatusChange(
                req_id=req_id,
                old_status=old_status,
                new_status=new_status,
            )
        )

    # Get status counts
    baseline_counts = baseline.status_counts()
    current_counts = current.status_counts()

    # Get cycle counts
    baseline_cycles = len(baseline.find_cycles())
    current_cycles = len(current.find_cycles())

    # Get reciprocity violations
    baseline_violations = len(check_reciprocity(baseline))
    current_violations = len(check_reciprocity(current))

    return ComparisonReport(
        baseline_path=baseline_path,
        current_path=current_path,
        baseline_req_count=len(baseline),
        current_req_count=len(current),
        baseline_completion=baseline.completion_percentage(),
        current_completion=current.completion_percentage(),
        baseline_status_counts={s.value: c for s, c in baseline_counts.items()},
        current_status_counts={s.value: c for s, c in current_counts.items()},
        added_requirements=added,
        removed_requirements=removed,
        status_changes=status_changes,
        baseline_cycles=baseline_cycles,
        current_cycles=current_cycles,
        baseline_reciprocity_violations=baseline_violations,
        current_reciprocity_violations=current_violations,
    )


def capture_baseline(db_path: Path | str) -> dict[str, Any]:
    """Capture baseline state for later comparison.

    Args:
        db_path: Path to RTM database

    Returns:
        Dictionary with baseline state (can be serialized to JSON)
    """
    from rtmx.validation import check_reciprocity

    db = RTMDatabase.load(Path(db_path))

    status_counts = db.status_counts()
    cycles = db.find_cycles()
    violations = check_reciprocity(db)

    return {
        "path": str(db_path),
        "req_count": len(db),
        "completion": db.completion_percentage(),
        "status_counts": {s.value: c for s, c in status_counts.items()},
        "cycles": len(cycles),
        "reciprocity_violations": len(violations),
        "requirements": {req.req_id: req.status.value for req in db},
    }
