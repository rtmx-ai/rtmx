"""RTMX service adapters.

Adapters for external services like GitHub Issues and Jira.
"""

from __future__ import annotations

from rtmx.adapters.base import (
    ConflictResolution,
    ExternalItem,
    ServiceAdapter,
    SyncDirection,
    SyncResult,
)

__all__ = [
    "ConflictResolution",
    "ExternalItem",
    "ServiceAdapter",
    "SyncDirection",
    "SyncResult",
]
