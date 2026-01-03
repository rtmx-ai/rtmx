"""Web server command for RTMX dashboard."""

from __future__ import annotations

import asyncio
import contextlib
import sys
from pathlib import Path


def run_serve(
    rtm_csv: Path | None,
    host: str,
    port: int,
    reload: bool,
) -> None:
    """Start the RTMX web dashboard server.

    Args:
        rtm_csv: Path to RTM database CSV
        host: Bind address
        port: Port number
        reload: Enable auto-reload on file changes
    """
    from rtmx.web import is_web_available

    if not is_web_available():
        print(
            "Web dependencies not installed. Install with:\n\n" "    pip install rtmx[web]\n",
            file=sys.stderr,
        )
        sys.exit(1)

    import uvicorn

    from rtmx.web.app import create_app
    from rtmx.web.routes.websocket import notify_update
    from rtmx.web.watcher import watch_rtm_file

    # Create the app
    app = create_app(rtm_csv)

    # Resolve the RTM CSV path
    csv_path = rtm_csv
    if csv_path is None:
        from rtmx.config import load_config

        config = load_config()
        csv_path = config.database

    print(f"Starting RTMX Dashboard at http://{host}:{port}")
    print(f"Watching: {csv_path}")
    print("Press Ctrl+C to stop\n")

    if reload:
        # Use uvicorn's reload functionality
        uvicorn.run(
            "rtmx.web.app:create_app",
            host=host,
            port=port,
            reload=True,
            factory=True,
        )
    else:
        # Run with file watcher for WebSocket updates
        async def run_with_watcher() -> None:
            """Run server with file watcher."""
            config = uvicorn.Config(app, host=host, port=port, log_level="info")
            server = uvicorn.Server(config)

            # Start file watcher if we have a valid path
            watcher_task = None
            if csv_path and csv_path.exists():
                watcher_task = asyncio.create_task(watch_rtm_file(csv_path, notify_update))

            try:
                await server.serve()
            finally:
                if watcher_task:
                    watcher_task.cancel()
                    with contextlib.suppress(asyncio.CancelledError):
                        await watcher_task

        asyncio.run(run_with_watcher())
