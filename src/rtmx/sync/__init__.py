"""RTMX Sync - CRDT-based real-time collaboration for requirements management.

This module provides:
- CRDT operations via pycrdt (Y.Doc, Y.Map, Y.Text)
- Requirement <-> Y.Map conversion
- CSV <-> CRDT serialization
- WebSocket sync client (Phase 10)

Installation:
    pip install rtmx[sync]

Usage:
    from rtmx.sync import is_sync_available
    if is_sync_available():
        from rtmx.sync.crdt import RTMDocument
"""

from __future__ import annotations

# Optional dependency check (follows RTMX pattern from web/__init__.py)
_SYNC_AVAILABLE = False
_SYNC_IMPORT_ERROR: str | None = None

try:
    import pycrdt  # noqa: F401

    _SYNC_AVAILABLE = True
except ImportError as e:
    _SYNC_IMPORT_ERROR = str(e)


def is_sync_available() -> bool:
    """Check if sync dependencies are installed.

    Returns:
        True if pycrdt is available, False otherwise.
    """
    return _SYNC_AVAILABLE


def get_sync_import_error() -> str | None:
    """Get the import error message if sync is not available.

    Returns:
        Error message string, or None if sync is available.
    """
    return _SYNC_IMPORT_ERROR


def require_sync() -> None:
    """Raise ImportError if sync dependencies are not available.

    Raises:
        ImportError: If pycrdt is not installed.
    """
    if not _SYNC_AVAILABLE:
        raise ImportError(
            f"RTMX sync requires pycrdt. Install with: pip install rtmx[sync]\n"
            f"Original error: {_SYNC_IMPORT_ERROR}"
        )


__all__ = [
    "is_sync_available",
    "get_sync_import_error",
    "require_sync",
]
