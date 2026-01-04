"""WebSocket handler for real-time RTM updates."""

from __future__ import annotations

from typing import TYPE_CHECKING

from fastapi import APIRouter, WebSocket, WebSocketDisconnect

if TYPE_CHECKING:
    from typing import Any

router = APIRouter()


class ConnectionManager:
    """Manage WebSocket connections and broadcasts."""

    def __init__(self) -> None:
        """Initialize the connection manager."""
        self.active_connections: list[WebSocket] = []

    async def connect(self, websocket: WebSocket) -> None:
        """Accept a new WebSocket connection."""
        await websocket.accept()
        self.active_connections.append(websocket)

    def disconnect(self, websocket: WebSocket) -> None:
        """Remove a WebSocket connection."""
        if websocket in self.active_connections:
            self.active_connections.remove(websocket)

    async def broadcast(self, message: dict[str, Any]) -> None:
        """Send a message to all connected clients."""
        disconnected: list[WebSocket] = []
        for connection in self.active_connections:
            try:
                await connection.send_json(message)
            except Exception:
                disconnected.append(connection)
        # Clean up disconnected clients
        for conn in disconnected:
            self.disconnect(conn)


# Global connection manager instance
manager = ConnectionManager()


@router.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket) -> None:
    """WebSocket endpoint for real-time updates.

    Clients connect here to receive push notifications when the RTM
    database changes. The watcher module triggers broadcasts through
    the connection manager.
    """
    await manager.connect(websocket)
    try:
        while True:
            # Keep connection alive, wait for client messages
            # (mainly just pings/pongs to keep connection open)
            data = await websocket.receive_text()
            # Echo back any received message (for ping/pong)
            if data == "ping":
                await websocket.send_text("pong")
    except WebSocketDisconnect:
        manager.disconnect(websocket)


async def notify_update() -> None:
    """Notify all connected clients of an RTM update.

    Called by the file watcher when the CSV changes.
    """
    await manager.broadcast({"event": "rtm-update"})


async def notify_update_with_delta(delta: dict) -> None:
    """Notify all connected clients with delta information.

    Called by the file watcher when the CSV changes, providing
    details about what specifically changed.

    Args:
        delta: Dictionary with 'changed', 'added', 'removed' lists
    """
    await manager.broadcast({"event": "rtm-update", "delta": delta})
