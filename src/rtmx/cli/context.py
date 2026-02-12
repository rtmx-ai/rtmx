"""CLI command for generating RTM context for AI assistants.

REQ-CLAUDE-001: Claude Code Hooks Integration

Provides token-efficient requirements context for AI coding sessions.
"""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any


def generate_context(
    project_path: Path | str | None = None,
    compact: bool = False,
    phase: int | None = None,
    files: list[str] | None = None,
    verbose: bool = False,
) -> dict[str, Any]:
    """Generate RTM context for AI assistants.

    Args:
        project_path: Project root directory (default: cwd)
        compact: Generate minimal token-efficient output
        phase: Filter to specific phase
        files: Focus on requirements related to specific files
        verbose: Include full requirement descriptions

    Returns:
        Context dictionary suitable for JSON serialization.
    """
    from rtmx.config import find_config_file, load_config
    from rtmx.models import RTMDatabase

    # TODO: Implement file-based filtering
    _ = files  # Reserved for future use

    project_path = Path.cwd() if project_path is None else Path(project_path)

    # Try to load config from project_path
    try:
        config_path = find_config_file(project_path)
        if config_path is None:
            return {
                "error": "No RTMX configuration found",
                "project": None,
            }
        config = load_config(config_path)
        db_path = config.database
        # Resolve relative database path relative to config file location
        if db_path and not Path(db_path).is_absolute():
            config_dir = config_path.parent
            # If config is in .rtmx/, go up one level for project root
            if config_dir.name == ".rtmx":
                config_dir = config_dir.parent
            db_path = config_dir / db_path
    except Exception:
        # No config found
        return {
            "error": "No RTMX configuration found",
            "project": None,
        }

    # Try to load database
    try:
        db = RTMDatabase.load(str(db_path))
    except Exception as e:
        return {
            "error": f"Failed to load RTM database: {e}",
            "project": str(project_path.name),
        }

    # Calculate statistics
    all_reqs = db.all()
    total = len(all_reqs)
    complete = sum(1 for r in all_reqs if r.status == "COMPLETE")
    partial = sum(1 for r in all_reqs if r.status == "PARTIAL")
    missing = total - complete - partial

    completion = (complete / total * 100) if total > 0 else 0

    # Build context
    context: dict[str, Any] = {
        "project": str(project_path.name),
        "completion": round(completion, 1),
        "requirements_count": {
            "total": total,
            "complete": complete,
            "partial": partial,
            "missing": missing,
        },
    }

    # Filter requirements
    filtered_reqs = list(all_reqs)

    if phase is not None:
        context["active_phase"] = phase
        filtered_reqs = [r for r in filtered_reqs if getattr(r, "phase", None) == phase]

    # Get relevant incomplete requirements
    incomplete = [r for r in filtered_reqs if r.status != "COMPLETE"]

    if compact:
        # Minimal output for token efficiency
        context["incomplete_count"] = len(incomplete)

        # Only include top 5 most important
        top_reqs = incomplete[:5]
        context["top_requirements"] = [
            {
                "id": r.req_id,
                "status": r.status,
            }
            for r in top_reqs
        ]
    else:
        # Standard output with more detail
        context["incomplete_requirements"] = [
            {
                "id": r.req_id,
                "category": getattr(r, "category", None),
                "status": r.status,
                "description": r.requirement_text[:100] + "..."
                if len(r.requirement_text) > 100
                else r.requirement_text
                if verbose
                else None,
            }
            for r in incomplete[:20]  # Limit to 20
        ]

        # Include blockers (requirements that block others)
        blockers = []
        for r in incomplete:
            blocks = getattr(r, "blocks", None)
            if blocks:
                blockers.append(r.req_id)
        if blockers:
            context["blockers"] = blockers[:10]

    return context


def run_context(
    format_type: str = "text",
    compact: bool = False,
    phase: int | None = None,
    files: list[str] | None = None,
    verbose: bool = False,
) -> int:
    """Run the context command.

    Args:
        format_type: Output format (text, json, markdown)
        compact: Minimal token-efficient output
        phase: Filter to specific phase
        files: Focus on files
        verbose: Include full descriptions

    Returns:
        Exit code (0 for success)
    """
    from rtmx.formatting import Colors

    context = generate_context(
        compact=compact,
        phase=phase,
        files=files,
        verbose=verbose,
    )

    if "error" in context and context.get("project") is None:
        if format_type == "json":
            print(json.dumps(context, indent=2))
        else:
            print(f"{Colors.YELLOW}No RTMX project found{Colors.RESET}")
        return 1

    if format_type == "json":
        print(json.dumps(context, indent=2))
    elif format_type == "markdown":
        _print_markdown(context)
    else:
        _print_text(context)

    return 0


def _print_text(context: dict[str, Any]) -> None:
    """Print context in plain text format."""
    from rtmx.formatting import Colors

    print(f"{Colors.BOLD}RTMX Context: {context.get('project', 'Unknown')}{Colors.RESET}")
    print(f"Completion: {context.get('completion', 0):.1f}%")

    counts = context.get("requirements_count", {})
    print(f"Requirements: {counts.get('total', 0)} total")
    print(f"  {Colors.GREEN}Complete: {counts.get('complete', 0)}{Colors.RESET}")
    print(f"  {Colors.YELLOW}Partial: {counts.get('partial', 0)}{Colors.RESET}")
    print(f"  {Colors.RED}Missing: {counts.get('missing', 0)}{Colors.RESET}")

    if "active_phase" in context:
        print(f"\nActive Phase: {context['active_phase']}")

    if "top_requirements" in context:
        print("\nTop Requirements:")
        for r in context["top_requirements"]:
            print(f"  - {r['id']} ({r['status']})")

    if "blockers" in context:
        print(f"\nBlockers: {', '.join(context['blockers'])}")


def _print_markdown(context: dict[str, Any]) -> None:
    """Print context in markdown format."""
    print(f"# RTMX Context: {context.get('project', 'Unknown')}")
    print(f"\n**Completion:** {context.get('completion', 0):.1f}%")

    counts = context.get("requirements_count", {})
    print(f"\n## Requirements ({counts.get('total', 0)} total)")
    print(f"- Complete: {counts.get('complete', 0)}")
    print(f"- Partial: {counts.get('partial', 0)}")
    print(f"- Missing: {counts.get('missing', 0)}")

    if "active_phase" in context:
        print(f"\n**Active Phase:** {context['active_phase']}")

    if "incomplete_requirements" in context:
        print("\n## Incomplete Requirements")
        for r in context["incomplete_requirements"]:
            desc = r.get("description") or ""
            print(f"- **{r['id']}** ({r['status']}): {desc}")

    if "blockers" in context:
        print("\n## Blockers")
        for b in context["blockers"]:
            print(f"- {b}")
