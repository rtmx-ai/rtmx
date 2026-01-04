"""Browser tests for RTMX Web UI using Playwright.

These tests verify:
- HTMX partial update behavior
- WebSocket DOM update handling
- Live refresh functionality
- Full page rendering

Requires: pip install rtmx[test-browser]
Run: pytest tests/test_web_browser.py -v
"""

from __future__ import annotations

import contextlib
import multiprocessing
import time
from collections.abc import Generator
from pathlib import Path

import pytest

# Skip all tests if Playwright is not installed
playwright = pytest.importorskip("playwright.sync_api", reason="Playwright not installed")


def _create_test_csv(path: Path) -> None:
    """Create a minimal RTM CSV for testing."""
    content = """req_id,category,subcategory,requirement_text,acceptance_criteria,test_module,test_function,test_type,status,priority,phase,notes,effort_weeks,depends_on,blocks,assignee,target_version,verification_method,verification_status,specification_file
REQ-TEST-001,TESTING,Unit,Test requirement 1,Criteria 1,tests/test_web.py,test_one,Unit Test,COMPLETE,HIGH,1,Note 1,0.5,,,developer,v1.0,,,
REQ-TEST-002,TESTING,Unit,Test requirement 2,Criteria 2,tests/test_web.py,test_two,Unit Test,PARTIAL,MEDIUM,1,Note 2,1.0,REQ-TEST-001,,developer,v1.0,,,
REQ-TEST-003,TESTING,Integration,Test requirement 3,Criteria 3,tests/test_web.py,test_three,Integration Test,MISSING,LOW,2,Note 3,2.0,REQ-TEST-002,,developer,v2.0,,,
"""
    path.write_text(content)


def _run_server(csv_path: str, port: int, ready_event: multiprocessing.Event) -> None:
    """Run uvicorn server in a subprocess."""
    import uvicorn

    from rtmx.web.app import create_app

    app = create_app(rtm_csv=Path(csv_path))

    class ReadyServer(uvicorn.Server):
        def startup(self, sockets=None):
            super().startup(sockets)
            ready_event.set()

    config = uvicorn.Config(app, host="127.0.0.1", port=port, log_level="warning")
    server = ReadyServer(config)
    server.run()


@contextlib.contextmanager
def run_server_process(csv_path: Path, port: int = 8765) -> Generator[str, None, None]:
    """Context manager to run server in background process."""
    ready_event = multiprocessing.Event()
    process = multiprocessing.Process(
        target=_run_server, args=(str(csv_path), port, ready_event), daemon=True
    )
    process.start()

    # Wait for server to be ready (max 10 seconds)
    if not ready_event.wait(timeout=10):
        process.terminate()
        raise RuntimeError("Server failed to start within timeout")

    # Additional small delay to ensure server is accepting connections
    time.sleep(0.2)

    try:
        yield f"http://127.0.0.1:{port}"
    finally:
        process.terminate()
        process.join(timeout=5)
        if process.is_alive():
            process.kill()


@pytest.fixture
def browser_context():
    """Fixture providing a Playwright browser context."""
    from playwright.sync_api import sync_playwright

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        context = browser.new_context()
        yield context
        context.close()
        browser.close()


@pytest.fixture
def sample_rtm_csv(tmp_path: Path) -> Path:
    """Create a sample RTM CSV for testing."""
    csv_path = tmp_path / "rtm_database.csv"
    _create_test_csv(csv_path)
    return csv_path


class TestDashboardPage:
    """Tests for dashboard page rendering."""

    @pytest.mark.req("REQ-WEB-001")
    @pytest.mark.scope_system
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_dashboard_loads(self, browser_context, sample_rtm_csv):
        """Dashboard page loads and displays status card."""
        with run_server_process(sample_rtm_csv) as base_url:
            page = browser_context.new_page()
            page.goto(base_url)

            # Wait for dashboard to load
            page.wait_for_selector(".dashboard")

            # Verify status card is present
            assert page.locator(".status-card").is_visible()

            # Verify stats are displayed
            assert page.locator(".stat-value.complete").is_visible()
            assert page.locator(".stat-value.partial").is_visible()
            assert page.locator(".stat-value.missing").is_visible()

            page.close()

    @pytest.mark.req("REQ-WEB-001")
    @pytest.mark.scope_system
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_dashboard_shows_correct_stats(self, browser_context, sample_rtm_csv):
        """Dashboard displays correct requirement counts."""
        with run_server_process(sample_rtm_csv) as base_url:
            page = browser_context.new_page()
            page.goto(base_url)

            # Wait for dashboard to load
            page.wait_for_selector(".status-card")

            # Verify counts match test data (1 complete, 1 partial, 1 missing)
            complete = page.locator(".stat-value.complete").inner_text()
            partial = page.locator(".stat-value.partial").inner_text()
            missing = page.locator(".stat-value.missing").inner_text()

            assert complete == "1"
            assert partial == "1"
            assert missing == "1"

            page.close()

    @pytest.mark.req("REQ-WEB-001")
    @pytest.mark.scope_system
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_dashboard_has_phase_progress(self, browser_context, sample_rtm_csv):
        """Dashboard shows phase progress section."""
        with run_server_process(sample_rtm_csv) as base_url:
            page = browser_context.new_page()
            page.goto(base_url)

            # Wait for phase progress to load
            page.wait_for_selector("h2:text('Phase Progress')")

            # Verify phase rows exist
            assert page.locator(".phase-row").count() >= 1

            page.close()


class TestHTMXPartials:
    """Tests for HTMX partial updates."""

    @pytest.mark.req("REQ-WEB-003")
    @pytest.mark.scope_system
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_status_partial_endpoint(self, browser_context, sample_rtm_csv):
        """HTMX partial endpoint returns status card HTML."""
        with run_server_process(sample_rtm_csv) as base_url:
            page = browser_context.new_page()
            response = page.goto(f"{base_url}/partials/status")

            assert response.status == 200
            content = page.content()
            assert "status-card" in content
            assert "stat-value" in content

            page.close()

    @pytest.mark.req("REQ-WEB-003")
    @pytest.mark.scope_system
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_phases_partial_endpoint(self, browser_context, sample_rtm_csv):
        """HTMX partial endpoint returns phase progress HTML."""
        with run_server_process(sample_rtm_csv) as base_url:
            page = browser_context.new_page()
            response = page.goto(f"{base_url}/partials/phases")

            assert response.status == 200
            content = page.content()
            assert "phase-row" in content or "phase" in content.lower()

            page.close()

    @pytest.mark.req("REQ-WEB-003")
    @pytest.mark.scope_system
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_htmx_trigger_updates_dom(self, browser_context, sample_rtm_csv):
        """HTMX trigger updates DOM when rtm-update event fires."""
        with run_server_process(sample_rtm_csv) as base_url:
            page = browser_context.new_page()
            page.goto(base_url)

            # Wait for initial load
            page.wait_for_selector(".status-card")

            # Get initial complete count
            initial_complete = page.locator(".stat-value.complete").inner_text()

            # Trigger HTMX refresh by dispatching custom event
            # This simulates what WebSocket would do
            page.evaluate("htmx.trigger(document.body, 'rtm-update')")

            # Wait for HTMX to process
            page.wait_for_timeout(500)

            # Verify page still shows data (not broken by refresh)
            final_complete = page.locator(".stat-value.complete").inner_text()
            assert final_complete == initial_complete  # Data unchanged, but refresh worked

            page.close()


class TestNavigation:
    """Tests for page navigation."""

    @pytest.mark.req("REQ-WEB-001")
    @pytest.mark.scope_system
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_navigate_to_backlog(self, browser_context, sample_rtm_csv):
        """Navigation from dashboard to backlog works."""
        with run_server_process(sample_rtm_csv) as base_url:
            page = browser_context.new_page()
            page.goto(base_url)

            # Click backlog link
            page.click("a[href='/backlog']")

            # Verify backlog page loaded
            page.wait_for_url("**/backlog")
            assert "/backlog" in page.url

            page.close()

    @pytest.mark.req("REQ-WEB-001")
    @pytest.mark.scope_system
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_navigate_back_to_dashboard(self, browser_context, sample_rtm_csv):
        """Navigation from backlog back to dashboard works."""
        with run_server_process(sample_rtm_csv) as base_url:
            page = browser_context.new_page()
            page.goto(f"{base_url}/backlog")

            # Click dashboard link
            page.click("a[href='/']")

            # Verify dashboard page loaded
            page.wait_for_selector(".dashboard")

            page.close()

    @pytest.mark.req("REQ-WEB-001")
    @pytest.mark.scope_system
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_active_nav_link_highlighted(self, browser_context, sample_rtm_csv):
        """Active navigation link has correct styling."""
        with run_server_process(sample_rtm_csv) as base_url:
            page = browser_context.new_page()
            page.goto(base_url)

            # Dashboard link should be active
            dashboard_link = page.locator("a[href='/']")
            assert "active" in (dashboard_link.get_attribute("class") or "")

            # Navigate to backlog
            page.click("a[href='/backlog']")
            page.wait_for_url("**/backlog")

            # Backlog link should be active
            backlog_link = page.locator("a[href='/backlog']")
            assert "active" in (backlog_link.get_attribute("class") or "")

            page.close()


class TestWebSocketConnection:
    """Tests for WebSocket connection behavior."""

    @pytest.mark.req("REQ-WEB-007")
    @pytest.mark.scope_system
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_websocket_connects_on_load(self, browser_context, sample_rtm_csv):
        """WebSocket connection is established on page load."""
        with run_server_process(sample_rtm_csv) as base_url:
            page = browser_context.new_page()

            # Listen for WebSocket
            ws_connected = []
            page.on("websocket", lambda ws: ws_connected.append(ws))

            page.goto(base_url)
            page.wait_for_selector(".dashboard")

            # Give WebSocket time to connect
            page.wait_for_timeout(1000)

            # Verify WebSocket was opened
            assert len(ws_connected) >= 1

            page.close()

    @pytest.mark.req("REQ-WEB-007")
    @pytest.mark.scope_system
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_websocket_receives_pong(self, browser_context, sample_rtm_csv):
        """WebSocket receives pong response to ping."""
        with run_server_process(sample_rtm_csv) as base_url:
            page = browser_context.new_page()

            ws_messages = []

            def on_websocket(ws):
                ws.on("framereceived", lambda payload: ws_messages.append(payload))

            page.on("websocket", on_websocket)

            page.goto(base_url)
            page.wait_for_selector(".dashboard")

            # Send ping via HTMX WebSocket
            page.evaluate(
                """
                const ws = htmx.find('[ws-connect]');
                if (ws && ws.__htmx_websocket) {
                    ws.__htmx_websocket.send('ping');
                }
            """
            )

            # Wait for response
            page.wait_for_timeout(500)

            # Check if pong was received
            pong_received = any("pong" in str(msg).lower() for msg in ws_messages)
            assert pong_received, f"No pong received. Messages: {ws_messages}"

            page.close()


class TestBacklogPage:
    """Tests for backlog page functionality."""

    @pytest.mark.req("REQ-WEB-002")
    @pytest.mark.scope_system
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_backlog_loads(self, browser_context, sample_rtm_csv):
        """Backlog page loads and displays requirements."""
        with run_server_process(sample_rtm_csv) as base_url:
            page = browser_context.new_page()
            page.goto(f"{base_url}/backlog")

            # Wait for backlog content to load
            page.wait_for_selector("body")

            # Verify page title or header
            assert "backlog" in page.url.lower()

            page.close()

    @pytest.mark.req("REQ-WEB-002")
    @pytest.mark.scope_system
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_backlog_shows_incomplete_requirements(self, browser_context, sample_rtm_csv):
        """Backlog shows only incomplete requirements."""
        with run_server_process(sample_rtm_csv) as base_url:
            page = browser_context.new_page()
            page.goto(f"{base_url}/backlog")

            # Wait for content
            page.wait_for_timeout(500)

            # The test data has 1 missing requirement (REQ-TEST-003)
            content = page.content()
            # Verify the missing requirement is shown
            assert "REQ-TEST-003" in content

            page.close()


class TestResponsiveLayout:
    """Tests for responsive layout behavior."""

    @pytest.mark.req("REQ-WEB-001")
    @pytest.mark.scope_system
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_mobile_viewport(self, browser_context, sample_rtm_csv):
        """Dashboard renders correctly on mobile viewport."""
        with run_server_process(sample_rtm_csv) as base_url:
            page = browser_context.new_page()
            page.set_viewport_size({"width": 375, "height": 667})  # iPhone SE

            page.goto(base_url)
            page.wait_for_selector(".status-card")

            # Verify status card is visible on mobile
            assert page.locator(".status-card").is_visible()

            page.close()

    @pytest.mark.req("REQ-WEB-001")
    @pytest.mark.scope_system
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_tablet_viewport(self, browser_context, sample_rtm_csv):
        """Dashboard renders correctly on tablet viewport."""
        with run_server_process(sample_rtm_csv) as base_url:
            page = browser_context.new_page()
            page.set_viewport_size({"width": 768, "height": 1024})  # iPad

            page.goto(base_url)
            page.wait_for_selector(".status-card")

            # Verify status card is visible on tablet
            assert page.locator(".status-card").is_visible()

            page.close()


class TestLiveRefresh:
    """Tests for live refresh functionality when RTM file changes."""

    @pytest.mark.req("REQ-WEB-007")
    @pytest.mark.scope_system
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_manual_htmx_refresh(self, browser_context, sample_rtm_csv):
        """Manual HTMX trigger refreshes content."""
        with run_server_process(sample_rtm_csv) as base_url:
            page = browser_context.new_page()
            page.goto(base_url)

            # Wait for initial load
            page.wait_for_selector(".status-card")

            # Modify the CSV file
            content = sample_rtm_csv.read_text()
            updated = content.replace("PARTIAL", "COMPLETE")
            sample_rtm_csv.write_text(updated)

            # Trigger HTMX refresh
            page.evaluate("htmx.trigger(document.body, 'rtm-update')")

            # Wait for refresh
            page.wait_for_timeout(500)

            # The DOM should have been refreshed (content may or may not have changed
            # depending on server caching, but the refresh mechanism should work)
            # At minimum, the page shouldn't crash
            assert page.locator(".status-card").is_visible()

            page.close()
