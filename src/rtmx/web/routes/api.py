"""REST API endpoints for RTMX Web UI."""

from __future__ import annotations

from typing import Any

from fastapi import APIRouter, Query, Request

from rtmx.config import load_config
from rtmx.models import RTMDatabase, Status

router = APIRouter()


def _load_database(request: Request) -> RTMDatabase:
    """Load the RTM database from the configured path."""
    rtm_csv = request.app.state.rtm_csv
    return RTMDatabase.load(rtm_csv)


@router.get("/status")
async def get_status(request: Request) -> dict[str, Any]:
    """Get overall status and phase breakdown.

    Returns:
        JSON with completion stats and phase breakdown
    """
    db = _load_database(request)
    config = load_config()

    # Calculate overall stats
    counts = db.status_counts()
    complete = counts.get(Status.COMPLETE, 0)
    partial = counts.get(Status.PARTIAL, 0)
    missing = counts.get(Status.MISSING, 0) + counts.get(Status.NOT_STARTED, 0)
    total = complete + partial + missing
    completion_pct = ((complete + partial * 0.5) / total * 100) if total else 0

    # Calculate phase stats
    by_phase: dict[int, dict[str, int]] = {}
    for req in db:
        phase = req.phase or 0
        if phase == 0:
            continue
        if phase not in by_phase:
            by_phase[phase] = {"complete": 0, "partial": 0, "missing": 0}
        if req.status == Status.COMPLETE:
            by_phase[phase]["complete"] += 1
        elif req.status == Status.PARTIAL:
            by_phase[phase]["partial"] += 1
        else:
            by_phase[phase]["missing"] += 1

    phases = []
    for phase_num in sorted(by_phase.keys()):
        phase_data = by_phase[phase_num]
        phase_total = sum(phase_data.values())
        phase_pct = (
            (phase_data["complete"] + phase_data["partial"] * 0.5) / phase_total * 100
            if phase_total
            else 0
        )
        phases.append(
            {
                "phase": phase_num,
                "name": config.get_phase_name(phase_num),
                "display": config.get_phase_display(phase_num),
                "complete": phase_data["complete"],
                "partial": phase_data["partial"],
                "missing": phase_data["missing"],
                "total": phase_total,
                "percentage": round(phase_pct, 1),
            }
        )

    return {
        "total": total,
        "complete": complete,
        "partial": partial,
        "missing": missing,
        "percentage": round(completion_pct, 1),
        "phases": phases,
    }


@router.get("/backlog")
async def get_backlog(
    request: Request,
    phase: int | None = Query(None, description="Filter by phase number"),
    limit: int = Query(10, description="Limit items per section"),
) -> dict[str, Any]:
    """Get prioritized backlog with critical path and quick wins.

    Args:
        phase: Optional phase filter
        limit: Max items per section

    Returns:
        JSON with critical path and quick wins
    """
    db = _load_database(request)
    config = load_config()

    # Get incomplete requirements
    incomplete = [
        req
        for req in db
        if req.status in (Status.MISSING, Status.NOT_STARTED)
        and (phase is None or req.phase == phase)
    ]

    # Sort by priority and phase
    priority_order = {"P0": 0, "HIGH": 1, "MEDIUM": 2, "LOW": 3}

    def sort_key(req):
        return (
            priority_order.get(req.priority.value, 4),
            req.phase or 99,
            req.req_id,
        )

    sorted_reqs = sorted(incomplete, key=sort_key)

    # Build response
    def req_to_dict(req):
        return {
            "req_id": req.req_id,
            "description": req.requirement_text,
            "category": req.category,
            "subcategory": req.subcategory,
            "priority": req.priority.value,
            "phase": req.phase,
            "phase_display": config.get_phase_display(req.phase),
            "effort_weeks": req.effort_weeks,
            "dependencies": list(req.dependencies) if req.dependencies else [],
            "blocks": list(req.blocks) if req.blocks else [],
        }

    # Quick wins: high priority, low effort
    quick_wins = [
        req
        for req in sorted_reqs
        if req.priority.value in ("P0", "HIGH") and (req.effort_weeks or 0) < 1.0
    ][:limit]

    # Critical path: requirements that block the most others
    blocking_counts = {}
    for req in db:
        if req.blocks:
            blocking_counts[req.req_id] = len(req.blocks)

    critical = sorted(
        [req for req in sorted_reqs if req.req_id in blocking_counts],
        key=lambda r: -blocking_counts.get(r.req_id, 0),
    )[:limit]

    return {
        "total_incomplete": len(incomplete),
        "phase_filter": phase,
        "critical_path": [req_to_dict(r) for r in critical],
        "quick_wins": [req_to_dict(r) for r in quick_wins],
        "all": [req_to_dict(r) for r in sorted_reqs[:limit]],
    }


@router.get("/requirements")
async def get_requirements(
    request: Request,
    phase: int | None = Query(None, description="Filter by phase number"),
    status: str | None = Query(None, description="Filter by status (COMPLETE, MISSING, etc)"),
    category: str | None = Query(None, description="Filter by category"),
    limit: int = Query(100, description="Max items to return"),
    offset: int = Query(0, description="Offset for pagination"),
) -> dict[str, Any]:
    """Get full requirements list with filtering and pagination.

    Args:
        phase: Optional phase filter
        status: Optional status filter
        category: Optional category filter
        limit: Max items to return
        offset: Pagination offset

    Returns:
        JSON with requirements list and pagination info
    """
    db = _load_database(request)
    config = load_config()

    # Filter requirements
    reqs = list(db)

    if phase is not None:
        reqs = [r for r in reqs if r.phase == phase]

    if status is not None:
        status_upper = status.upper()
        reqs = [r for r in reqs if r.status.value == status_upper]

    if category is not None:
        reqs = [r for r in reqs if r.category.upper() == category.upper()]

    # Sort by phase, then priority, then ID
    priority_order = {"P0": 0, "HIGH": 1, "MEDIUM": 2, "LOW": 3}
    reqs = sorted(
        reqs,
        key=lambda r: (
            r.phase or 99,
            priority_order.get(r.priority.value, 4),
            r.req_id,
        ),
    )

    # Pagination
    total = len(reqs)
    paginated = reqs[offset : offset + limit]

    def req_to_dict(req):
        return {
            "req_id": req.req_id,
            "description": req.requirement_text,
            "category": req.category,
            "subcategory": req.subcategory,
            "status": req.status.value,
            "priority": req.priority.value,
            "phase": req.phase,
            "phase_display": config.get_phase_display(req.phase),
            "effort_weeks": req.effort_weeks,
            "dependencies": list(req.dependencies) if req.dependencies else [],
            "blocks": list(req.blocks) if req.blocks else [],
            "notes": req.notes,
            "test_module": req.test_module,
            "test_function": req.test_function,
        }

    return {
        "total": total,
        "limit": limit,
        "offset": offset,
        "requirements": [req_to_dict(r) for r in paginated],
    }
