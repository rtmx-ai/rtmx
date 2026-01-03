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


class TestFileWatcher:
    """Tests for file watcher functionality."""

    @pytest.mark.req("REQ-WEB-007")
    @pytest.mark.scope_unit
    def test_poll_watch_detects_changes(self, tmp_path):
        """Polling watcher detects file changes."""

        from rtmx.web.watcher import _poll_watch

        test_file = tmp_path / "test.csv"
        test_file.write_text("initial")

        changes_detected = []

        async def on_change():
            changes_detected.append(True)

        async def run_test():
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

        asyncio.run(run_test())

        # Should have detected the change
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
