"""Tests for RTMX Web UI.

Tests REQ-WEB-001 through REQ-WEB-008.
"""

from __future__ import annotations

import asyncio
import contextlib

import pytest

# Mark all tests in this module
pytestmark = [
    pytest.mark.env_simulation,
    pytest.mark.technique_nominal,
]


@pytest.fixture
def sample_rtm_csv(tmp_path):
    """Create a sample RTM CSV for testing."""
    csv_content = """\
req_id,category,subcategory,requirement_text,acceptance_criteria,test_module,test_function,verification_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,owner,target_release,github_issue,jira_key,spec_file
REQ-WEB-001,WEB,CLI,rtmx serve command,CLI starts server,tests/test_web.py,test_serve,Unit Test,COMPLETE,HIGH,6,Done,0.5,,,developer,v0.1.0,,,
REQ-WEB-002,WEB,API,Status endpoint,Returns JSON,tests/test_web.py,test_status,Unit Test,COMPLETE,HIGH,6,Done,0.5,REQ-WEB-001,,developer,v0.1.0,,,
REQ-WEB-003,WEB,API,Backlog endpoint,Returns JSON,tests/test_web.py,test_backlog,Unit Test,MISSING,MEDIUM,6,,0.5,REQ-WEB-001,,developer,v0.1.0,,,
REQ-TEST-001,TEST,UNIT,Test coverage,80% coverage,tests/test_models.py,test_coverage,Unit Test,COMPLETE,HIGH,1,Done,1,,,developer,v0.1.0,,,
REQ-TEST-002,TEST,UNIT,Unit tests pass,All pass,tests/test_models.py,test_pass,Unit Test,PARTIAL,HIGH,1,WIP,0.5,REQ-TEST-001,,developer,v0.1.0,,,
"""
    csv_file = tmp_path / "rtm_database.csv"
    csv_file.write_text(csv_content)
    return csv_file


class TestWebAvailability:
    """Tests for REQ-WEB-008: Optional web dependencies."""

    @pytest.mark.req("REQ-WEB-008")
    @pytest.mark.scope_unit
    def test_is_web_available_returns_bool(self):
        """is_web_available() returns a boolean."""
        from rtmx.web import is_web_available

        result = is_web_available()
        assert isinstance(result, bool)

    @pytest.mark.req("REQ-WEB-008")
    @pytest.mark.scope_unit
    def test_web_available_when_installed(self):
        """Web is available when fastapi and uvicorn are installed."""
        from rtmx.web import is_web_available

        # We have web deps installed for testing
        assert is_web_available() is True


class TestAppFactory:
    """Tests for FastAPI app creation."""

    @pytest.mark.req("REQ-WEB-001")
    @pytest.mark.scope_unit
    def test_create_app_returns_fastapi(self, sample_rtm_csv):
        """create_app() returns a FastAPI application."""
        from fastapi import FastAPI

        from rtmx.web.app import create_app

        app = create_app(sample_rtm_csv)
        assert isinstance(app, FastAPI)

    @pytest.mark.req("REQ-WEB-001")
    @pytest.mark.scope_unit
    def test_create_app_stores_rtm_csv(self, sample_rtm_csv):
        """App stores RTM CSV path in state."""
        from rtmx.web.app import create_app

        app = create_app(sample_rtm_csv)
        assert app.state.rtm_csv == sample_rtm_csv

    @pytest.mark.req("REQ-WEB-001")
    @pytest.mark.scope_unit
    def test_create_app_has_templates(self, sample_rtm_csv):
        """App has templates configured."""
        from rtmx.web.app import create_app

        app = create_app(sample_rtm_csv)
        assert hasattr(app.state, "templates")


class TestStatusAPI:
    """Tests for REQ-WEB-002: Status API endpoint."""

    @pytest.mark.req("REQ-WEB-002")
    @pytest.mark.scope_integration
    def test_api_status_endpoint(self, sample_rtm_csv):
        """GET /api/status returns status data."""
        from fastapi.testclient import TestClient

        from rtmx.web.app import create_app

        app = create_app(sample_rtm_csv)
        client = TestClient(app)

        response = client.get("/api/status")
        assert response.status_code == 200

        data = response.json()
        assert "total" in data
        assert "complete" in data
        assert "partial" in data
        assert "missing" in data
        assert "percentage" in data
        assert "phases" in data

    @pytest.mark.req("REQ-WEB-002")
    @pytest.mark.scope_integration
    def test_api_status_phase_breakdown(self, sample_rtm_csv):
        """Status API returns phase breakdown."""
        from fastapi.testclient import TestClient

        from rtmx.web.app import create_app

        app = create_app(sample_rtm_csv)
        client = TestClient(app)

        response = client.get("/api/status")
        data = response.json()

        # Should have phases
        assert len(data["phases"]) > 0

        # Each phase should have expected fields
        phase = data["phases"][0]
        assert "phase" in phase
        assert "complete" in phase
        assert "partial" in phase
        assert "missing" in phase
        assert "percentage" in phase


class TestBacklogAPI:
    """Tests for REQ-WEB-003: Backlog API endpoint."""

    @pytest.mark.req("REQ-WEB-003")
    @pytest.mark.scope_integration
    def test_api_backlog_endpoint(self, sample_rtm_csv):
        """GET /api/backlog returns backlog data."""
        from fastapi.testclient import TestClient

        from rtmx.web.app import create_app

        app = create_app(sample_rtm_csv)
        client = TestClient(app)

        response = client.get("/api/backlog")
        assert response.status_code == 200

        data = response.json()
        assert "total_incomplete" in data
        assert "critical_path" in data
        assert "quick_wins" in data
        assert "all" in data

    @pytest.mark.req("REQ-WEB-003")
    @pytest.mark.scope_integration
    def test_api_backlog_phase_filter(self, sample_rtm_csv):
        """Backlog API accepts phase filter."""
        from fastapi.testclient import TestClient

        from rtmx.web.app import create_app

        app = create_app(sample_rtm_csv)
        client = TestClient(app)

        response = client.get("/api/backlog?phase=6")
        assert response.status_code == 200

        data = response.json()
        assert data["phase_filter"] == 6


class TestRequirementsAPI:
    """Tests for REQ-WEB-004: Requirements API endpoint."""

    @pytest.mark.req("REQ-WEB-004")
    @pytest.mark.scope_integration
    def test_api_requirements_endpoint(self, sample_rtm_csv):
        """GET /api/requirements returns requirements list."""
        from fastapi.testclient import TestClient

        from rtmx.web.app import create_app

        app = create_app(sample_rtm_csv)
        client = TestClient(app)

        response = client.get("/api/requirements")
        assert response.status_code == 200

        data = response.json()
        assert "total" in data
        assert "requirements" in data
        assert "limit" in data
        assert "offset" in data

    @pytest.mark.req("REQ-WEB-004")
    @pytest.mark.scope_integration
    def test_api_requirements_filtering(self, sample_rtm_csv):
        """Requirements API supports filtering."""
        from fastapi.testclient import TestClient

        from rtmx.web.app import create_app

        app = create_app(sample_rtm_csv)
        client = TestClient(app)

        # Filter by status
        response = client.get("/api/requirements?status=COMPLETE")
        assert response.status_code == 200
        data = response.json()
        for req in data["requirements"]:
            assert req["status"] == "COMPLETE"

        # Filter by category
        response = client.get("/api/requirements?category=WEB")
        assert response.status_code == 200
        data = response.json()
        for req in data["requirements"]:
            assert req["category"] == "WEB"


class TestDashboardPage:
    """Tests for REQ-WEB-005: Dashboard page."""

    @pytest.mark.req("REQ-WEB-005")
    @pytest.mark.scope_integration
    def test_dashboard_renders(self, sample_rtm_csv):
        """GET / returns HTML dashboard."""
        from fastapi.testclient import TestClient

        from rtmx.web.app import create_app

        app = create_app(sample_rtm_csv)
        client = TestClient(app)

        response = client.get("/")
        assert response.status_code == 200
        assert "text/html" in response.headers["content-type"]

    @pytest.mark.req("REQ-WEB-005")
    @pytest.mark.scope_integration
    def test_dashboard_alias(self, sample_rtm_csv):
        """GET /dashboard also returns dashboard."""
        from fastapi.testclient import TestClient

        from rtmx.web.app import create_app

        app = create_app(sample_rtm_csv)
        client = TestClient(app)

        response = client.get("/dashboard")
        assert response.status_code == 200
        assert "text/html" in response.headers["content-type"]


class TestBacklogPage:
    """Tests for REQ-WEB-006: Backlog page."""

    @pytest.mark.req("REQ-WEB-006")
    @pytest.mark.scope_integration
    def test_backlog_page_renders(self, sample_rtm_csv):
        """GET /backlog returns HTML page."""
        from fastapi.testclient import TestClient

        from rtmx.web.app import create_app

        app = create_app(sample_rtm_csv)
        client = TestClient(app)

        response = client.get("/backlog")
        assert response.status_code == 200
        assert "text/html" in response.headers["content-type"]


class TestWebSocket:
    """Tests for REQ-WEB-007: WebSocket connection."""

    @pytest.mark.req("REQ-WEB-007")
    @pytest.mark.scope_unit
    def test_connection_manager_connect(self):
        """ConnectionManager can track connections."""
        from rtmx.web.routes.websocket import ConnectionManager

        manager = ConnectionManager()
        assert len(manager.active_connections) == 0

    @pytest.mark.req("REQ-WEB-007")
    @pytest.mark.scope_integration
    def test_websocket_endpoint(self, sample_rtm_csv):
        """WebSocket endpoint accepts connections."""
        from fastapi.testclient import TestClient

        from rtmx.web.app import create_app

        app = create_app(sample_rtm_csv)
        client = TestClient(app)

        with client.websocket_connect("/ws") as websocket:
            # Send ping
            websocket.send_text("ping")
            # Should receive pong
            data = websocket.receive_text()
            assert data == "pong"


class TestWebSocketConcurrent:
    """Tests for concurrent WebSocket connections - Phase 7 readiness."""

    @pytest.mark.req("REQ-WEB-007")
    @pytest.mark.scope_integration
    def test_multiple_connections(self, sample_rtm_csv):
        """Multiple WebSocket clients can connect simultaneously."""
        from fastapi.testclient import TestClient

        from rtmx.web.app import create_app
        from rtmx.web.routes.websocket import manager

        app = create_app(sample_rtm_csv)
        client = TestClient(app)

        # Clear any existing connections
        manager.active_connections.clear()

        # Connect first client
        with client.websocket_connect("/ws") as ws1:
            assert len(manager.active_connections) == 1

            # Connect second client
            with client.websocket_connect("/ws") as ws2:
                assert len(manager.active_connections) == 2

                # Both can ping/pong independently
                ws1.send_text("ping")
                assert ws1.receive_text() == "pong"

                ws2.send_text("ping")
                assert ws2.receive_text() == "pong"

            # Second disconnected
            assert len(manager.active_connections) == 1

        # Both disconnected
        assert len(manager.active_connections) == 0

    @pytest.mark.req("REQ-WEB-007")
    @pytest.mark.scope_unit
    async def test_broadcast_to_all_clients(self):
        """Broadcast sends message to all connected clients."""
        from unittest.mock import AsyncMock, MagicMock

        from rtmx.web.routes.websocket import ConnectionManager

        manager = ConnectionManager()

        # Create mock websockets
        ws1 = MagicMock()
        ws1.send_json = AsyncMock()
        ws2 = MagicMock()
        ws2.send_json = AsyncMock()

        # Manually add to connections (simulating accepted connections)
        manager.active_connections = [ws1, ws2]

        # Broadcast message
        await manager.broadcast({"event": "rtm-update"})

        # Both should receive the message
        ws1.send_json.assert_called_once_with({"event": "rtm-update"})
        ws2.send_json.assert_called_once_with({"event": "rtm-update"})

    @pytest.mark.req("REQ-WEB-007")
    @pytest.mark.scope_unit
    async def test_broadcast_removes_disconnected_clients(self):
        """Broadcast cleans up clients that fail to receive."""
        from unittest.mock import AsyncMock, MagicMock

        from rtmx.web.routes.websocket import ConnectionManager

        manager = ConnectionManager()

        # Create mock websockets - one fails
        ws_good = MagicMock()
        ws_good.send_json = AsyncMock()
        ws_bad = MagicMock()
        ws_bad.send_json = AsyncMock(side_effect=Exception("Connection closed"))

        manager.active_connections = [ws_good, ws_bad]

        # Broadcast - bad client should be removed
        await manager.broadcast({"event": "test"})

        # Good client received message
        ws_good.send_json.assert_called_once()
        # Bad client was removed
        assert ws_bad not in manager.active_connections
        assert len(manager.active_connections) == 1

    @pytest.mark.req("REQ-WEB-007")
    @pytest.mark.scope_unit
    async def test_notify_update_broadcasts(self):
        """notify_update() broadcasts rtm-update event."""
        from unittest.mock import AsyncMock, MagicMock, patch

        from rtmx.web.routes import websocket

        mock_manager = MagicMock()
        mock_manager.broadcast = AsyncMock()

        with patch.object(websocket, "manager", mock_manager):
            await websocket.notify_update()

        mock_manager.broadcast.assert_called_once_with({"event": "rtm-update"})


class TestFileWatcher:
    """Tests for file watcher functionality."""

    @pytest.mark.req("REQ-WEB-007")
    @pytest.mark.scope_unit
    async def test_poll_watch_detects_changes(self, tmp_path):
        """Polling watcher detects file changes."""
        from rtmx.web.watcher import _poll_watch

        test_file = tmp_path / "test.csv"
        test_file.write_text("initial")

        changes_detected = []

        async def on_change():
            changes_detected.append(True)

        # Start watcher
        watcher_task = asyncio.create_task(_poll_watch(test_file, on_change, 50))

        # Wait a bit for watcher to start
        await asyncio.sleep(0.1)

        # Modify file
        test_file.write_text("modified")

        # Wait for detection
        await asyncio.sleep(0.2)

        # Cancel watcher
        watcher_task.cancel()
        with contextlib.suppress(asyncio.CancelledError):
            await watcher_task

        # Should have detected the change
        assert len(changes_detected) >= 1

    @pytest.mark.req("REQ-WEB-007")
    @pytest.mark.scope_unit
    async def test_poll_watch_handles_missing_file(self, tmp_path):
        """Polling watcher handles missing file gracefully."""
        from rtmx.web.watcher import _poll_watch

        test_file = tmp_path / "nonexistent.csv"

        changes_detected = []

        async def on_change():
            changes_detected.append(True)

        # Start watcher on non-existent file
        watcher_task = asyncio.create_task(_poll_watch(test_file, on_change, 50))

        # Wait a bit
        await asyncio.sleep(0.1)

        # Create the file
        test_file.write_text("created")

        # Wait for first poll to register the file
        await asyncio.sleep(0.1)

        # Modify it - this should trigger change detection
        test_file.write_text("modified")

        # Wait for detection
        await asyncio.sleep(0.2)

        # Cancel watcher
        watcher_task.cancel()
        with contextlib.suppress(asyncio.CancelledError):
            await watcher_task

        # Should have detected the change
        assert len(changes_detected) >= 1

    @pytest.mark.req("REQ-WEB-007")
    @pytest.mark.scope_unit
    async def test_poll_watch_multiple_changes(self, tmp_path):
        """Polling watcher detects multiple sequential changes."""
        from rtmx.web.watcher import _poll_watch

        test_file = tmp_path / "test.csv"
        test_file.write_text("v1")

        changes_detected = []

        async def on_change():
            changes_detected.append(True)

        # Start watcher with fast polling
        watcher_task = asyncio.create_task(_poll_watch(test_file, on_change, 30))

        # Wait for watcher to start
        await asyncio.sleep(0.05)

        # Make multiple changes
        for i in range(3):
            test_file.write_text(f"v{i+2}")
            await asyncio.sleep(0.1)

        # Cancel watcher
        watcher_task.cancel()
        with contextlib.suppress(asyncio.CancelledError):
            await watcher_task

        # Should have detected multiple changes
        assert len(changes_detected) >= 2

    @pytest.mark.req("REQ-WEB-007")
    @pytest.mark.scope_integration
    async def test_watch_rtm_file_with_watchfiles(self, tmp_path):
        """watch_rtm_file uses watchfiles when available."""
        from unittest.mock import AsyncMock, MagicMock, patch

        test_file = tmp_path / "test.csv"
        test_file.write_text("initial")

        on_change = AsyncMock()

        # Mock watchfiles.awatch to yield one change then stop
        async def mock_awatch(*args, **kwargs):
            yield [("modified", str(test_file))]

        mock_watchfiles = MagicMock()
        mock_watchfiles.awatch = mock_awatch

        # Patch the import inside watch_rtm_file
        with patch.dict("sys.modules", {"watchfiles": mock_watchfiles}):
            # Need to reload the module to pick up the mock
            import importlib

            import rtmx.web.watcher

            importlib.reload(rtmx.web.watcher)

            # Run with timeout
            with contextlib.suppress(asyncio.TimeoutError):
                await asyncio.wait_for(
                    rtmx.web.watcher.watch_rtm_file(test_file, on_change, debounce_ms=10),
                    timeout=0.5,
                )

            # Reload to restore original
            importlib.reload(rtmx.web.watcher)

        # on_change should have been called
        on_change.assert_called()

    @pytest.mark.req("REQ-WEB-007")
    @pytest.mark.scope_integration
    async def test_watch_rtm_file_falls_back_to_polling(self, tmp_path):
        """watch_rtm_file falls back to polling when watchfiles unavailable."""
        import importlib

        test_file = tmp_path / "test.csv"
        test_file.write_text("initial")

        changes_detected = []

        async def on_change():
            changes_detected.append(True)

        # Make watchfiles import fail by removing it from sys.modules
        import sys

        original_watchfiles = sys.modules.get("watchfiles")
        sys.modules["watchfiles"] = None  # This will cause ImportError

        try:
            # Reload to pick up the "missing" watchfiles
            import rtmx.web.watcher

            importlib.reload(rtmx.web.watcher)

            watcher_task = asyncio.create_task(
                rtmx.web.watcher.watch_rtm_file(test_file, on_change, debounce_ms=50)
            )

            # Wait for watcher
            await asyncio.sleep(0.1)

            # Modify file
            test_file.write_text("modified")

            # Wait for detection
            await asyncio.sleep(0.2)

            # Cancel
            watcher_task.cancel()
            with contextlib.suppress(asyncio.CancelledError):
                await watcher_task
        finally:
            # Restore original
            if original_watchfiles is not None:
                sys.modules["watchfiles"] = original_watchfiles
            elif "watchfiles" in sys.modules:
                del sys.modules["watchfiles"]

            import rtmx.web.watcher

            importlib.reload(rtmx.web.watcher)

        # Fallback should have worked
        assert len(changes_detected) >= 1


class TestPartials:
    """Tests for HTMX partial endpoints."""

    @pytest.mark.req("REQ-WEB-005")
    @pytest.mark.scope_integration
    def test_status_partial(self, sample_rtm_csv):
        """GET /partials/status returns HTML fragment."""
        from fastapi.testclient import TestClient

        from rtmx.web.app import create_app

        app = create_app(sample_rtm_csv)
        client = TestClient(app)

        response = client.get("/partials/status")
        assert response.status_code == 200
        assert "text/html" in response.headers["content-type"]

    @pytest.mark.req("REQ-WEB-005")
    @pytest.mark.scope_integration
    def test_phases_partial(self, sample_rtm_csv):
        """GET /partials/phases returns HTML fragment."""
        from fastapi.testclient import TestClient

        from rtmx.web.app import create_app

        app = create_app(sample_rtm_csv)
        client = TestClient(app)

        response = client.get("/partials/phases")
        assert response.status_code == 200
        assert "text/html" in response.headers["content-type"]
