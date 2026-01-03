"""FastAPI application factory for RTMX Web UI."""

from __future__ import annotations

from pathlib import Path


def create_app(rtm_csv: Path | None = None):
    """Create and configure the FastAPI application.

    Args:
        rtm_csv: Path to RTM database CSV file

    Returns:
        Configured FastAPI application
    """
    from fastapi import FastAPI
    from fastapi.staticfiles import StaticFiles
    from fastapi.templating import Jinja2Templates

    from rtmx.web.routes import api, pages, websocket

    app = FastAPI(
        title="RTMX Dashboard",
        description="Requirements Traceability Matrix Dashboard",
        version="0.0.4",
    )

    # Store rtm_csv path in app state
    app.state.rtm_csv = rtm_csv

    # Set up templates
    templates_dir = Path(__file__).parent / "templates"
    app.state.templates = Jinja2Templates(directory=str(templates_dir))

    # Mount static files
    static_dir = Path(__file__).parent / "static"
    if static_dir.exists():
        app.mount("/static", StaticFiles(directory=str(static_dir)), name="static")

    # Include routers
    app.include_router(api.router, prefix="/api", tags=["api"])
    app.include_router(pages.router, tags=["pages"])
    app.include_router(websocket.router, tags=["websocket"])

    return app
