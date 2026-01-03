"""RTMX Web UI - FastAPI-based dashboard.

Provides a read-only web dashboard for viewing RTMX status.
Requires the web dependencies: pip install rtmx[web]
"""

from __future__ import annotations

# Detect if web dependencies are available
_WEB_AVAILABLE = False
try:
    import fastapi  # noqa: F401
    import uvicorn  # noqa: F401

    _WEB_AVAILABLE = True
except ImportError:
    pass


def is_web_available() -> bool:
    """Check if web dependencies are available.

    Returns:
        True if fastapi and uvicorn are installed
    """
    return _WEB_AVAILABLE
