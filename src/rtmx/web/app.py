"""FastAPI application factory for RTMX Web UI."""

from __future__ import annotations

import asyncio
import contextlib
from contextlib import asynccontextmanager
from pathlib import Path
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from collections.abc import AsyncGenerator

    from fastapi import FastAPI as FastAPIType


@asynccontextmanager
async def lifespan(app: FastAPIType) -> AsyncGenerator[None, None]:
    """Lifespan context manager for app startup/shutdown.

    Starts the file watcher on startup and cancels it on shutdown.
    """
    from rtmx.web.routes.websocket import notify_update
    from rtmx.web.watcher import watch_rtm_file

    # Start file watcher if rtm_csv is configured
    watcher_task = None
    if app.state.rtm_csv is not None:
        watcher_task = asyncio.create_task(watch_rtm_file(app.state.rtm_csv, notify_update))
        app.state.watcher_task = watcher_task

    yield

    # Cancel watcher on shutdown
    if watcher_task is not None:
        watcher_task.cancel()
        with contextlib.suppress(asyncio.CancelledError):
            await watcher_task


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
        lifespan=lifespan,
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
