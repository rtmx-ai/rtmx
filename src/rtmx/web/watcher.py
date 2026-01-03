"""File watcher for RTM database changes."""

from __future__ import annotations

import asyncio
from pathlib import Path
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from collections.abc import Callable, Coroutine
    from typing import Any


async def watch_rtm_file(
    path: Path,
    on_change: Callable[[], Coroutine[Any, Any, None]],
    debounce_ms: int = 100,
) -> None:
    """Watch an RTM file for changes and trigger callback.

    Args:
        path: Path to the RTM CSV file to watch
        on_change: Async callback to run when file changes
        debounce_ms: Debounce interval in milliseconds
    """
    try:
        from watchfiles import awatch
    except ImportError:
        # Fallback to polling if watchfiles not available
        await _poll_watch(path, on_change, debounce_ms)
        return

    debounce_s = debounce_ms / 1000.0

    async for _changes in awatch(path, debounce=int(debounce_ms)):
        await on_change()
        # Small delay to prevent rapid-fire updates
        await asyncio.sleep(debounce_s)


async def _poll_watch(
    path: Path,
    on_change: Callable[[], Coroutine[Any, Any, None]],
    interval_ms: int = 1000,
) -> None:
    """Fallback polling-based file watcher.

    Args:
        path: Path to the file to watch
        on_change: Async callback to run when file changes
        interval_ms: Polling interval in milliseconds
    """
    interval_s = interval_ms / 1000.0
    last_mtime: float | None = None

    while True:
        try:
            current_mtime = path.stat().st_mtime
            if last_mtime is not None and current_mtime != last_mtime:
                await on_change()
            last_mtime = current_mtime
        except FileNotFoundError:
            # File doesn't exist yet, wait for it
            last_mtime = None
        except Exception:
            # Ignore other errors, keep polling
            pass

        await asyncio.sleep(interval_s)
