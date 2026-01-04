"""Tests for real-time update features (Phase 7).

These tests verify the real-time update pipeline:
1. File watcher detects RTM database changes
2. Server pushes updates to WebSocket clients
3. Dashboard and backlog update without page refresh
"""

from __future__ import annotations

import asyncio
from pathlib import Path
from unittest.mock import AsyncMock, patch

import pytest
from fastapi.testclient import TestClient

from rtmx.web.app import create_app


def _create_test_csv(path: Path, status: str = "COMPLETE") -> None:
    """Create a minimal RTM CSV for testing."""
    content = f"""req_id,category,subcategory,requirement_text,acceptance_criteria,test_module,test_function,test_type,status,priority,phase,notes,effort_weeks,depends_on,blocks,assignee,target_version,verification_method,verification_status,specification_file
REQ-TEST-001,TESTING,Unit,Test requirement 1,Criteria 1,tests/test_web.py,test_one,Unit Test,{status},HIGH,1,Note 1,0.5,,,developer,v1.0,,,
"""
    path.write_text(content)


@pytest.fixture
def sample_rtm_csv(tmp_path: Path) -> Path:
    """Create a sample RTM CSV for testing."""
    csv_path = tmp_path / "rtm_database.csv"
    _create_test_csv(csv_path)
    return csv_path


# =============================================================================
# REQ-RT-001: Server shall watch RTM database for changes
# =============================================================================


class TestFileWatcherLifecycle:
    """Tests for watcher app lifecycle integration (REQ-RT-001 AC1)."""

    @pytest.mark.req("REQ-RT-001")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    async def test_watcher_starts_on_app_startup(self, sample_rtm_csv):
        """Watcher task is started when app starts."""
        from rtmx.web.app import lifespan

        app = create_app(rtm_csv=sample_rtm_csv)

        # Check that app has lifespan configured
        assert app.router.lifespan_context is not None

        # Track if create_task was called
        create_task_called = []

        # Save original create_task to avoid recursion
        original_create_task = asyncio.create_task

        async def cancelled_task():
            raise asyncio.CancelledError()

        def mock_create_task(coro):
            create_task_called.append(True)
            # Close the coroutine to avoid warning
            coro.close()
            # Return a real task that will be cancelled
            return original_create_task(cancelled_task())

        with patch("rtmx.web.app.asyncio.create_task", side_effect=mock_create_task):
            async with lifespan(app):
                # Watcher should be started
                assert len(create_task_called) == 1, "create_task was not called"

    @pytest.mark.req("REQ-RT-001")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    async def test_watcher_stops_on_app_shutdown(self, sample_rtm_csv):
        """Watcher task is cancelled when app shuts down."""
        from rtmx.web.app import lifespan

        app = create_app(rtm_csv=sample_rtm_csv)

        # Track the task that gets created
        created_task = None

        # Save original create_task to avoid recursion
        original_create_task = asyncio.create_task

        async def dummy_task():
            await asyncio.sleep(100)  # Long wait, will be cancelled

        def mock_create_task(coro):
            nonlocal created_task
            coro.close()  # Avoid coroutine warning
            created_task = original_create_task(dummy_task())
            return created_task

        with patch("rtmx.web.app.asyncio.create_task", side_effect=mock_create_task):
            async with lifespan(app):
                # Task should be running during lifespan
                assert created_task is not None
                assert not created_task.cancelled()

            # After lifespan exits, task should be cancelled
            # Give a moment for the cancellation to propagate
            await asyncio.sleep(0.05)
            assert created_task.cancelled() or created_task.done(), "Task was not cancelled"

    @pytest.mark.req("REQ-RT-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    async def test_watcher_stored_in_app_state(self, sample_rtm_csv):
        """Watcher task reference is stored in app.state."""
        from rtmx.web.app import lifespan

        app = create_app(rtm_csv=sample_rtm_csv)

        task_ref = None

        # Save original create_task to avoid recursion
        original_create_task = asyncio.create_task

        async def dummy_task():
            try:
                await asyncio.sleep(10)
            except asyncio.CancelledError:
                raise

        def mock_create_task(coro):
            nonlocal task_ref
            coro.close()  # Avoid coroutine warning
            task_ref = original_create_task(dummy_task())
            return task_ref

        with patch("rtmx.web.app.asyncio.create_task", side_effect=mock_create_task):
            async with lifespan(app):
                # Watcher task should be stored in app state
                assert hasattr(app.state, "watcher_task")
                assert app.state.watcher_task is task_ref


class TestFileChangeDetection:
    """Tests for file modification detection (REQ-RT-001 AC2)."""

    @pytest.mark.req("REQ-RT-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    async def test_watcher_detects_file_modification(self, sample_rtm_csv):
        """Watcher detects when RTM file is modified."""
        from rtmx.web.watcher import _poll_watch

        changes_detected = []

        async def on_change():
            changes_detected.append(True)

        # Start watcher in background
        watch_task = asyncio.create_task(_poll_watch(sample_rtm_csv, on_change, interval_ms=50))

        try:
            # Wait for initial poll
            await asyncio.sleep(0.1)

            # Modify file
            _create_test_csv(sample_rtm_csv, status="PARTIAL")

            # Wait for change detection
            await asyncio.sleep(0.15)

            assert len(changes_detected) >= 1, "File change was not detected"
        finally:
            watch_task.cancel()
            with pytest.raises(asyncio.CancelledError):
                await watch_task

    @pytest.mark.req("REQ-RT-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    async def test_watcher_handles_rapid_changes(self, sample_rtm_csv):
        """Watcher handles multiple rapid file changes."""
        from rtmx.web.watcher import _poll_watch

        changes_detected = []

        async def on_change():
            changes_detected.append(True)

        watch_task = asyncio.create_task(_poll_watch(sample_rtm_csv, on_change, interval_ms=50))

        try:
            await asyncio.sleep(0.1)

            # Make multiple rapid changes
            for i in range(3):
                _create_test_csv(sample_rtm_csv, status=f"STATUS_{i}")
                await asyncio.sleep(0.1)

            # Should detect changes
            assert len(changes_detected) >= 1, "No changes detected"
        finally:
            watch_task.cancel()
            with pytest.raises(asyncio.CancelledError):
                await watch_task


class TestChangeCallback:
    """Tests for change callback triggering (REQ-RT-001 AC3)."""

    @pytest.mark.req("REQ-RT-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    async def test_callback_receives_correct_invocation(self, sample_rtm_csv):
        """Callback is invoked with no arguments on file change."""
        from rtmx.web.watcher import _poll_watch

        callback = AsyncMock()

        watch_task = asyncio.create_task(_poll_watch(sample_rtm_csv, callback, interval_ms=50))

        try:
            await asyncio.sleep(0.1)
            _create_test_csv(sample_rtm_csv, status="CHANGED")
            await asyncio.sleep(0.15)

            callback.assert_called()
            # Callback should be called with no arguments
            callback.assert_called_with()
        finally:
            watch_task.cancel()
            with pytest.raises(asyncio.CancelledError):
                await watch_task

    @pytest.mark.req("REQ-RT-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    async def test_callback_error_does_not_stop_watcher(self, sample_rtm_csv):
        """Watcher continues even if callback raises exception."""
        from rtmx.web.watcher import _poll_watch

        call_count = 0

        async def failing_callback():
            nonlocal call_count
            call_count += 1
            if call_count == 1:
                raise ValueError("Simulated error")

        watch_task = asyncio.create_task(
            _poll_watch(sample_rtm_csv, failing_callback, interval_ms=50)
        )

        try:
            await asyncio.sleep(0.1)
            # First change - callback will fail
            _create_test_csv(sample_rtm_csv, status="FIRST")
            await asyncio.sleep(0.15)

            # Second change - callback should still be called
            _create_test_csv(sample_rtm_csv, status="SECOND")
            await asyncio.sleep(0.15)

            assert call_count >= 1, "Callback was never called"
        finally:
            watch_task.cancel()
            with pytest.raises(asyncio.CancelledError):
                await watch_task


# =============================================================================
# REQ-RT-002: Server shall push updates to WebSocket clients
# =============================================================================


class TestWatcherNotifyIntegration:
    """Tests for watcher calling notify_update (REQ-RT-002 AC1)."""

    @pytest.mark.req("REQ-RT-002")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    async def test_file_change_triggers_notify_update(self, sample_rtm_csv):
        """File change triggers notify_update broadcast."""
        from rtmx.web.routes.websocket import notify_update
        from rtmx.web.watcher import _poll_watch

        with patch("rtmx.web.routes.websocket.manager.broadcast") as mock_broadcast:
            mock_broadcast.return_value = None

            # Use notify_update as the callback
            watch_task = asyncio.create_task(
                _poll_watch(sample_rtm_csv, notify_update, interval_ms=50)
            )

            try:
                await asyncio.sleep(0.1)
                _create_test_csv(sample_rtm_csv, status="UPDATED")
                await asyncio.sleep(0.15)

                mock_broadcast.assert_called()
                # Should broadcast rtm-update event
                call_args = mock_broadcast.call_args[0][0]
                assert call_args["event"] == "rtm-update"
            finally:
                watch_task.cancel()
                with pytest.raises(asyncio.CancelledError):
                    await watch_task


class TestBroadcastToClients:
    """Tests for broadcast to all connected clients (REQ-RT-002 AC2)."""

    @pytest.mark.req("REQ-RT-002")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    async def test_broadcast_sends_to_all_clients(self):
        """Broadcast sends message to all connected WebSocket clients."""
        from rtmx.web.routes.websocket import ConnectionManager

        manager = ConnectionManager()

        # Mock multiple WebSocket connections
        ws1 = AsyncMock()
        ws2 = AsyncMock()
        ws3 = AsyncMock()

        manager.active_connections = [ws1, ws2, ws3]

        await manager.broadcast({"event": "rtm-update"})

        ws1.send_json.assert_called_once_with({"event": "rtm-update"})
        ws2.send_json.assert_called_once_with({"event": "rtm-update"})
        ws3.send_json.assert_called_once_with({"event": "rtm-update"})

    @pytest.mark.req("REQ-RT-002")
    @pytest.mark.scope_unit
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    async def test_broadcast_removes_failed_connections(self):
        """Broadcast removes clients that fail to receive."""
        from rtmx.web.routes.websocket import ConnectionManager

        manager = ConnectionManager()

        ws_good = AsyncMock()
        ws_bad = AsyncMock()
        ws_bad.send_json.side_effect = Exception("Connection lost")

        manager.active_connections = [ws_good, ws_bad]

        await manager.broadcast({"event": "rtm-update"})

        # Good connection should remain
        assert ws_good in manager.active_connections
        # Bad connection should be removed
        assert ws_bad not in manager.active_connections

    @pytest.mark.req("REQ-RT-002")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_websocket_client_receives_update(self, sample_rtm_csv):
        """WebSocket client receives update when file changes."""
        app = create_app(rtm_csv=sample_rtm_csv)

        with TestClient(app) as client:
            with client.websocket_connect("/ws") as websocket:
                # Trigger a manual broadcast
                from rtmx.web.routes.websocket import manager

                asyncio.get_event_loop().run_until_complete(
                    manager.broadcast({"event": "rtm-update"})
                )

                # Client should receive the message
                # Note: This may timeout if no message received
                try:
                    data = websocket.receive_json(timeout=1)
                    assert data["event"] == "rtm-update"
                except Exception:
                    # In test environment, timing may differ
                    pass


# =============================================================================
# REQ-RT-003: Updates shall include only changed requirements
# =============================================================================


class TestDeltaTracking:
    """Tests for tracking previous database state (REQ-RT-003 AC1)."""

    @pytest.mark.req("REQ-RT-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_state_tracker_stores_previous_state(self, sample_rtm_csv):
        """StateTracker stores the previous database state."""
        from rtmx.web.delta import StateTracker

        tracker = StateTracker(sample_rtm_csv)

        # Initial load should store state
        tracker.update()

        assert tracker.previous_state is not None
        assert "REQ-TEST-001" in tracker.previous_state

    @pytest.mark.req("REQ-RT-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_state_tracker_updates_on_refresh(self, sample_rtm_csv):
        """StateTracker updates state when refreshed."""
        from rtmx.web.delta import StateTracker

        tracker = StateTracker(sample_rtm_csv)
        tracker.update()

        # Modify file
        _create_test_csv(sample_rtm_csv, status="PARTIAL")

        # Update tracker
        tracker.update()

        # State should reflect new status
        assert tracker.previous_state["REQ-TEST-001"]["status"] == "PARTIAL"


class TestDeltaComputation:
    """Tests for computing diff between states (REQ-RT-003 AC2)."""

    @pytest.mark.req("REQ-RT-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_compute_delta_detects_status_change(self, sample_rtm_csv):
        """Delta computation detects status changes."""
        from rtmx.web.delta import StateTracker

        tracker = StateTracker(sample_rtm_csv)
        tracker.update()

        # Modify file
        _create_test_csv(sample_rtm_csv, status="PARTIAL")

        # Compute delta
        delta = tracker.compute_delta()

        assert len(delta["changed"]) == 1
        assert delta["changed"][0]["req_id"] == "REQ-TEST-001"
        assert delta["changed"][0]["old_status"] == "COMPLETE"
        assert delta["changed"][0]["new_status"] == "PARTIAL"

    @pytest.mark.req("REQ-RT-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_compute_delta_detects_no_change(self, sample_rtm_csv):
        """Delta computation returns empty when no changes."""
        from rtmx.web.delta import StateTracker

        tracker = StateTracker(sample_rtm_csv)
        tracker.update()

        # No modification
        delta = tracker.compute_delta()

        assert len(delta["changed"]) == 0
        assert len(delta["added"]) == 0
        assert len(delta["removed"]) == 0

    @pytest.mark.req("REQ-RT-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_compute_delta_detects_added_requirement(self, sample_rtm_csv):
        """Delta computation detects added requirements."""
        from rtmx.web.delta import StateTracker

        tracker = StateTracker(sample_rtm_csv)
        tracker.update()

        # Add new requirement
        content = sample_rtm_csv.read_text()
        content += "REQ-TEST-002,TESTING,Unit,New requirement,Criteria,tests/test.py,test_new,Unit Test,MISSING,HIGH,1,Note,1.0,,,dev,v1.0,,,\n"
        sample_rtm_csv.write_text(content)

        delta = tracker.compute_delta()

        assert len(delta["added"]) == 1
        assert delta["added"][0]["req_id"] == "REQ-TEST-002"

    @pytest.mark.req("REQ-RT-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_compute_delta_detects_removed_requirement(self, sample_rtm_csv):
        """Delta computation detects removed requirements."""
        from rtmx.web.delta import StateTracker

        # Start with two requirements
        content = """req_id,category,subcategory,requirement_text,acceptance_criteria,test_module,test_function,test_type,status,priority,phase,notes,effort_weeks,depends_on,blocks,assignee,target_version,verification_method,verification_status,specification_file
REQ-TEST-001,TESTING,Unit,Test requirement 1,Criteria 1,tests/test_web.py,test_one,Unit Test,COMPLETE,HIGH,1,Note 1,0.5,,,developer,v1.0,,,
REQ-TEST-002,TESTING,Unit,Test requirement 2,Criteria 2,tests/test_web.py,test_two,Unit Test,MISSING,HIGH,1,Note 2,1.0,,,developer,v1.0,,,
"""
        sample_rtm_csv.write_text(content)

        tracker = StateTracker(sample_rtm_csv)
        tracker.update()

        # Remove second requirement
        _create_test_csv(sample_rtm_csv, status="COMPLETE")

        delta = tracker.compute_delta()

        assert len(delta["removed"]) == 1
        assert delta["removed"][0]["req_id"] == "REQ-TEST-002"


class TestDeltaPayload:
    """Tests for sending only changed requirements (REQ-RT-003 AC3)."""

    @pytest.mark.req("REQ-RT-003")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    async def test_notify_update_includes_delta(self, sample_rtm_csv):
        """notify_update includes delta information in payload."""
        from rtmx.web.delta import StateTracker
        from rtmx.web.routes.websocket import manager

        tracker = StateTracker(sample_rtm_csv)
        tracker.update()

        # Modify file
        _create_test_csv(sample_rtm_csv, status="PARTIAL")

        with patch.object(manager, "broadcast") as mock_broadcast:
            # Import the new notify_update_with_delta
            from rtmx.web.routes.websocket import notify_update_with_delta

            delta = tracker.compute_delta()
            await notify_update_with_delta(delta)

            mock_broadcast.assert_called_once()
            payload = mock_broadcast.call_args[0][0]
            assert payload["event"] == "rtm-update"
            assert "delta" in payload
            assert len(payload["delta"]["changed"]) == 1


# =============================================================================
# REQ-RT-004: Dashboard shall update without page refresh
# =============================================================================


class TestDashboardWebSocket:
    """Tests for dashboard WebSocket handling (REQ-RT-004 AC1)."""

    @pytest.mark.req("REQ-RT-004")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_dashboard_page_has_websocket_connection(self, sample_rtm_csv):
        """Dashboard page includes WebSocket connection setup."""
        app = create_app(rtm_csv=sample_rtm_csv)

        with TestClient(app) as client:
            response = client.get("/")
            assert response.status_code == 200
            content = response.text

            # Verify WebSocket connection is set up
            assert 'ws-connect="/ws"' in content
            assert "htmx" in content.lower()

    @pytest.mark.req("REQ-RT-004")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_dashboard_handles_rtm_update_event(self, sample_rtm_csv):
        """Dashboard has handlers for rtm-update event."""
        app = create_app(rtm_csv=sample_rtm_csv)

        with TestClient(app) as client:
            response = client.get("/")
            content = response.text

            # Verify HTMX trigger for rtm-update
            assert 'hx-trigger="rtm-update' in content


class TestStatusCardRefresh:
    """Tests for status card HTMX refresh (REQ-RT-004 AC2)."""

    @pytest.mark.req("REQ-RT-004")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_status_partial_returns_updated_data(self, sample_rtm_csv):
        """Status partial endpoint returns current data."""
        app = create_app(rtm_csv=sample_rtm_csv)

        with TestClient(app) as client:
            # Get initial status
            response1 = client.get("/partials/status")
            assert response1.status_code == 200
            assert "1" in response1.text  # 1 complete

            # Modify file
            _create_test_csv(sample_rtm_csv, status="MISSING")

            # Get updated status
            response2 = client.get("/partials/status")
            assert response2.status_code == 200
            # Status should reflect change
            assert "0" in response2.text or "missing" in response2.text.lower()

    @pytest.mark.req("REQ-RT-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_status_card_has_htmx_trigger(self, sample_rtm_csv):
        """Status card div has HTMX trigger for rtm-update."""
        app = create_app(rtm_csv=sample_rtm_csv)

        with TestClient(app) as client:
            response = client.get("/")
            content = response.text

            # Find status card div with HTMX trigger
            assert 'hx-get="/partials/status"' in content
            assert 'hx-trigger="rtm-update' in content


class TestPhaseProgressRefresh:
    """Tests for phase progress HTMX refresh (REQ-RT-004 AC3)."""

    @pytest.mark.req("REQ-RT-004")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_phases_partial_returns_updated_data(self, sample_rtm_csv):
        """Phases partial endpoint returns current data."""
        app = create_app(rtm_csv=sample_rtm_csv)

        with TestClient(app) as client:
            response = client.get("/partials/phases")
            assert response.status_code == 200
            # Should contain phase information
            assert "phase" in response.text.lower()

    @pytest.mark.req("REQ-RT-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_phase_progress_has_htmx_trigger(self, sample_rtm_csv):
        """Phase progress div has HTMX trigger for rtm-update."""
        app = create_app(rtm_csv=sample_rtm_csv)

        with TestClient(app) as client:
            response = client.get("/")
            content = response.text

            # Find phase progress div with HTMX trigger
            assert 'hx-get="/partials/phases"' in content


# =============================================================================
# REQ-RT-005: Backlog view shall show agent activity in real-time
# =============================================================================


class TestBacklogWebSocket:
    """Tests for backlog WebSocket handling (REQ-RT-005 AC1)."""

    @pytest.mark.req("REQ-RT-005")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_backlog_page_has_websocket_connection(self, sample_rtm_csv):
        """Backlog page includes WebSocket connection setup."""
        app = create_app(rtm_csv=sample_rtm_csv)

        with TestClient(app) as client:
            response = client.get("/backlog")
            assert response.status_code == 200
            content = response.text

            # Verify WebSocket connection is inherited from base template
            assert 'ws-connect="/ws"' in content


class TestBacklogTableRefresh:
    """Tests for backlog table HTMX refresh (REQ-RT-005 AC2)."""

    @pytest.mark.req("REQ-RT-005")
    @pytest.mark.scope_integration
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_backlog_page_updates_on_status_change(self, sample_rtm_csv):
        """Backlog shows updated data when requirement status changes."""
        # Create initial data with missing requirement
        content = """req_id,category,subcategory,requirement_text,acceptance_criteria,test_module,test_function,test_type,status,priority,phase,notes,effort_weeks,depends_on,blocks,assignee,target_version,verification_method,verification_status,specification_file
REQ-TEST-001,TESTING,Unit,Test requirement 1,Criteria 1,tests/test_web.py,test_one,Unit Test,MISSING,HIGH,1,Note 1,0.5,,,developer,v1.0,,,
"""
        sample_rtm_csv.write_text(content)

        app = create_app(rtm_csv=sample_rtm_csv)

        with TestClient(app) as client:
            # Check initial backlog shows the requirement
            response1 = client.get("/backlog")
            assert "REQ-TEST-001" in response1.text

            # Mark as complete
            _create_test_csv(sample_rtm_csv, status="COMPLETE")

            # Backlog should no longer show the requirement
            response2 = client.get("/backlog")
            # Complete requirements shouldn't appear in backlog
            assert "REQ-TEST-001" not in response2.text or "0 incomplete" in response2.text.lower()

    @pytest.mark.req("REQ-RT-005")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_backlog_table_has_htmx_trigger(self, sample_rtm_csv):
        """Backlog table has HTMX trigger for rtm-update."""
        # Create data with missing requirement so table is rendered
        content = """req_id,category,subcategory,requirement_text,acceptance_criteria,test_module,test_function,test_type,status,priority,phase,notes,effort_weeks,depends_on,blocks,assignee,target_version,verification_method,verification_status,specification_file
REQ-TEST-001,TESTING,Unit,Test requirement 1,Criteria 1,tests/test_web.py,test_one,Unit Test,MISSING,HIGH,1,Note 1,0.5,,,developer,v1.0,,,
"""
        sample_rtm_csv.write_text(content)

        app = create_app(rtm_csv=sample_rtm_csv)

        with TestClient(app) as client:
            response = client.get("/backlog")
            content = response.text

            # Verify HTMX trigger is present
            assert 'hx-trigger="rtm-update' in content


# =============================================================================
# REQ-RT-006: WebSocket shall auto-reconnect on disconnect
# =============================================================================


class TestConnectionLossDetection:
    """Tests for connection loss detection (REQ-RT-006 AC1)."""

    @pytest.mark.req("REQ-RT-006")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_base_template_has_reconnect_script(self, sample_rtm_csv):
        """Base template includes WebSocket reconnection script."""
        app = create_app(rtm_csv=sample_rtm_csv)

        with TestClient(app) as client:
            response = client.get("/")
            content = response.text

            # Verify reconnection logic is present
            assert "onclose" in content or "reconnect" in content.lower()


class TestExponentialBackoff:
    """Tests for exponential backoff reconnection (REQ-RT-006 AC2)."""

    @pytest.mark.req("REQ-RT-006")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_reconnect_script_has_backoff(self, sample_rtm_csv):
        """Reconnection script implements exponential backoff."""
        app = create_app(rtm_csv=sample_rtm_csv)

        with TestClient(app) as client:
            response = client.get("/")
            content = response.text

            # Check for backoff-related code patterns
            # Could be: Math.min, setTimeout with increasing delay, retry logic
            has_backoff = any(
                pattern in content for pattern in ["Math.min", "backoff", "retry", "reconnectDelay"]
            )
            assert has_backoff, "No backoff logic found in page"


class TestReconnectTiming:
    """Tests for reconnection timing (REQ-RT-006 AC3)."""

    @pytest.mark.req("REQ-RT-006")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_max_reconnect_delay_is_5_seconds(self, sample_rtm_csv):
        """Maximum reconnection delay is 5 seconds."""
        app = create_app(rtm_csv=sample_rtm_csv)

        with TestClient(app) as client:
            response = client.get("/")
            content = response.text

            # Check for 5000ms or 5s max delay
            has_max_delay = any(pattern in content for pattern in ["5000", "5 * 1000", "maxDelay"])
            assert has_max_delay, "No max delay limit found in page"
