"""HTML page routes for RTMX Web UI."""

from __future__ import annotations

from fastapi import APIRouter, Request
from fastapi.responses import HTMLResponse

from rtmx.config import load_config
from rtmx.models import RTMDatabase, Status

router = APIRouter()


def _load_database(request: Request) -> RTMDatabase:
    """Load the RTM database from the configured path."""
    rtm_csv = request.app.state.rtm_csv
    return RTMDatabase.load(rtm_csv)


def _get_status_context(db: RTMDatabase) -> dict:
    """Build context data for status display."""
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

        # Determine status
        if phase_data["missing"] == 0 and phase_data["partial"] == 0:
            status_class = "complete"
            status_icon = "✓"
            status_text = "Complete"
        elif phase_pct > 0:
            status_class = "in-progress"
            status_icon = "⚠"
            status_text = "In Progress"
        else:
            status_class = "not-started"
            status_icon = "✗"
            status_text = "Not Started"

        phases.append(
            {
                "number": phase_num,
                "name": config.get_phase_name(phase_num),
                "display": config.get_phase_display(phase_num),
                "complete": phase_data["complete"],
                "partial": phase_data["partial"],
                "missing": phase_data["missing"],
                "total": phase_total,
                "percentage": round(phase_pct, 1),
                "status_class": status_class,
                "status_icon": status_icon,
                "status_text": status_text,
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


@router.get("/", response_class=HTMLResponse)
@router.get("/dashboard", response_class=HTMLResponse)
async def dashboard(request: Request) -> HTMLResponse:
    """Render the status dashboard page."""
    db = _load_database(request)
    context = _get_status_context(db)

    templates = request.app.state.templates
    return templates.TemplateResponse(request, "dashboard.html", context)


@router.get("/backlog", response_class=HTMLResponse)
async def backlog_page(request: Request) -> HTMLResponse:
    """Render the backlog page."""
    db = _load_database(request)
    config = load_config()

    # Get all requirements grouped by status
    incomplete = []
    for req in db:
        if req.status in (Status.MISSING, Status.NOT_STARTED):
            incomplete.append(
                {
                    "req_id": req.req_id,
                    "description": req.requirement_text,
                    "category": req.category,
                    "priority": req.priority.value,
                    "phase": req.phase,
                    "phase_display": config.get_phase_display(req.phase),
                    "effort_weeks": req.effort_weeks,
                    "status_class": "missing",
                }
            )

    # Sort by priority and phase
    priority_order = {"P0": 0, "HIGH": 1, "MEDIUM": 2, "LOW": 3}
    incomplete.sort(
        key=lambda r: (
            priority_order.get(str(r["priority"]), 4),
            r["phase"] or 99,
        )
    )

    context = {
        "requirements": incomplete,
        "total": len(incomplete),
    }

    templates = request.app.state.templates
    return templates.TemplateResponse(request, "backlog.html", context)


@router.get("/partials/status", response_class=HTMLResponse)
async def partial_status(request: Request) -> HTMLResponse:
    """Render just the status card partial (for HTMX updates)."""
    db = _load_database(request)
    context = _get_status_context(db)

    templates = request.app.state.templates
    return templates.TemplateResponse(request, "partials/status_card.html", context)


@router.get("/partials/phases", response_class=HTMLResponse)
async def partial_phases(request: Request) -> HTMLResponse:
    """Render just the phase progress partial (for HTMX updates)."""
    db = _load_database(request)
    context = _get_status_context(db)

    templates = request.app.state.templates
    return templates.TemplateResponse(request, "partials/phase_progress.html", context)
