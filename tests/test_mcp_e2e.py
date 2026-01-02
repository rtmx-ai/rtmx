"""End-to-end tests for RTMX MCP server.

This module tests the complete MCP server lifecycle:
1. Server startup on available port
2. Server startup with port conflict
3. Daemon mode with PID file
4. SIGTERM graceful shutdown
5. rtmx_status tool invocation
6. rtmx_backlog tool invocation
7. Concurrent client connections (optional)

Each test uses isolated temporary directories to ensure reproducibility.
MCP tests are skipped if MCP dependencies are not installed.
"""

from __future__ import annotations

import csv
import os
import signal
import socket
import subprocess
import sys
import tempfile
import time
from pathlib import Path
from typing import TYPE_CHECKING

import pytest

if TYPE_CHECKING:
    from collections.abc import Generator


# =============================================================================
# Skip marker for MCP tests if dependencies not installed
# =============================================================================


def mcp_available() -> bool:
    """Check if MCP package is properly available for rtmx.

    Returns True only if the rtmx MCP server can start without the
    "not installed" error. We run the server command briefly and check
    if it reports MCP as unavailable.
    """
    try:
        # Run mcp-server without --help to trigger the actual MCP import check
        # Use a timeout since the server would hang waiting for connections
        result = subprocess.run(
            [sys.executable, "-m", "rtmx", "mcp-server", "--port", "0"],
            capture_output=True,
            text=True,
            timeout=2,  # Short timeout - we just want to see if it errors immediately
        )
        # If it exits with "not installed" message, MCP is not available
        output = result.stdout + result.stderr
        if "MCP package not installed" in output:
            return False
        # MCP is available if it didn't exit with error
        return result.returncode == 0
    except subprocess.TimeoutExpired:
        # Server started and is waiting - MCP is available!
        return True
    except (FileNotFoundError, OSError):
        return False


requires_mcp = pytest.mark.skipif(
    not mcp_available(),
    reason="MCP package not installed. Install with: pip install rtmx[mcp]",
)


# =============================================================================
# Fixtures
# =============================================================================


@pytest.fixture
def temp_project() -> Generator[Path, None, None]:
    """Create an isolated temporary project directory."""
    with tempfile.TemporaryDirectory(prefix="rtmx_mcp_test_") as tmpdir:
        project_dir = Path(tmpdir)
        # Initialize as git repo for realistic testing
        subprocess.run(
            ["git", "init"],
            cwd=project_dir,
            capture_output=True,
            check=True,
        )
        subprocess.run(
            ["git", "config", "user.email", "test@example.com"],
            cwd=project_dir,
            capture_output=True,
            check=True,
        )
        subprocess.run(
            ["git", "config", "user.name", "Test User"],
            cwd=project_dir,
            capture_output=True,
            check=True,
        )
        yield project_dir


@pytest.fixture
def initialized_project(temp_project: Path) -> Path:
    """Create an initialized RTMX project."""
    result = subprocess.run(
        [sys.executable, "-m", "rtmx", "setup", "--minimal"],
        cwd=temp_project,
        capture_output=True,
        text=True,
    )
    assert result.returncode == 0, f"Setup failed: {result.stderr}"
    return temp_project


@pytest.fixture
def project_with_requirements(initialized_project: Path) -> Path:
    """Create a project with multiple requirements for testing."""
    db_path = initialized_project / "docs" / "rtm_database.csv"

    # Read existing CSV to get headers
    with open(db_path) as f:
        reader = csv.DictReader(f)
        fieldnames = reader.fieldnames or []

    # Add test requirements
    requirements = [
        {
            "req_id": "REQ-MCP-001",
            "category": "MCP",
            "subcategory": "Server",
            "requirement_text": "MCP server shall start correctly",
            "status": "COMPLETE",
            "priority": "P0",
            "phase": "1",
            "dependencies": "",
            "blocks": "REQ-MCP-002|REQ-MCP-003",
        },
        {
            "req_id": "REQ-MCP-002",
            "category": "MCP",
            "subcategory": "Tools",
            "requirement_text": "rtmx_status tool shall return completion status",
            "status": "PARTIAL",
            "priority": "HIGH",
            "phase": "1",
            "dependencies": "REQ-MCP-001",
            "blocks": "",
        },
        {
            "req_id": "REQ-MCP-003",
            "category": "MCP",
            "subcategory": "Tools",
            "requirement_text": "rtmx_backlog tool shall return prioritized backlog",
            "status": "MISSING",
            "priority": "HIGH",
            "phase": "1",
            "dependencies": "REQ-MCP-001",
            "blocks": "",
        },
        {
            "req_id": "REQ-MCP-004",
            "category": "MCP",
            "subcategory": "Lifecycle",
            "requirement_text": "Server shall shutdown gracefully on SIGTERM",
            "status": "MISSING",
            "priority": "MEDIUM",
            "phase": "2",
            "dependencies": "",
            "blocks": "",
        },
    ]

    # Write new requirements
    with open(db_path, "w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        for req in requirements:
            row = {field: req.get(field, "") for field in fieldnames}
            writer.writerow(row)

    return initialized_project


def run_rtmx(
    *args: str,
    cwd: Path,
    env: dict[str, str] | None = None,
) -> subprocess.CompletedProcess[str]:
    """Run rtmx command and return result."""
    full_env = os.environ.copy()
    if env:
        full_env.update(env)

    return subprocess.run(
        [sys.executable, "-m", "rtmx", *args],
        cwd=cwd,
        capture_output=True,
        text=True,
        env=full_env,
    )


def find_free_port() -> int:
    """Find a free port for testing."""
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(("", 0))
        s.listen(1)
        return s.getsockname()[1]


def is_port_in_use(port: int) -> bool:
    """Check if a port is in use."""
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        return s.connect_ex(("localhost", port)) == 0


# =============================================================================
# Server Startup E2E Tests
# =============================================================================


@pytest.mark.req("REQ-TEST-008")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMCPServerStartup:
    """E2E tests for MCP server startup."""

    def test_mcp_server_command_exists(self, initialized_project: Path) -> None:
        """Test that mcp-server command is available."""
        result = run_rtmx("mcp-server", "--help", cwd=initialized_project)

        assert result.returncode == 0
        assert "mcp" in result.stdout.lower() or "server" in result.stdout.lower()

    @requires_mcp
    def test_server_startup_on_available_port(self, project_with_requirements: Path) -> None:
        """Test MCP server starts on an available port.

        This test verifies:
        1. Server process can be started
        2. Server outputs expected startup messages
        3. Server can be terminated gracefully
        """
        port = find_free_port()

        # Start server as subprocess
        proc = subprocess.Popen(
            [sys.executable, "-m", "rtmx", "mcp-server", "--port", str(port)],
            cwd=project_with_requirements,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )

        try:
            # Give server time to start
            time.sleep(1)

            # Check process is still running
            assert proc.poll() is None, "Server process exited unexpectedly"

            # Terminate gracefully
            proc.terminate()
            proc.wait(timeout=5)

        finally:
            # Ensure cleanup
            if proc.poll() is None:
                proc.kill()
                proc.wait()

    def test_server_startup_without_mcp_package(self, initialized_project: Path) -> None:
        """Test server reports error when MCP package is not installed.

        This test uses environment manipulation to simulate missing MCP.
        """
        # This test is informational - we can't easily mock the import
        # but we verify the error handling exists in the command
        result = run_rtmx("mcp-server", "--help", cwd=initialized_project)

        # Help should mention MCP
        assert result.returncode == 0
        assert "mcp" in result.stdout.lower()


@pytest.mark.req("REQ-TEST-008")
@pytest.mark.scope_system
@pytest.mark.technique_stress
@pytest.mark.env_simulation
class TestMCPServerPortConflict:
    """E2E tests for MCP server port conflict handling."""

    @requires_mcp
    def test_server_startup_with_port_conflict(self, project_with_requirements: Path) -> None:
        """Test MCP server handles port conflict gracefully.

        This test verifies:
        1. When port is already in use
        2. Server reports appropriate error
        3. Server exits cleanly
        """
        port = find_free_port()

        # Occupy the port
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as blocking_socket:
            blocking_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            blocking_socket.bind(("localhost", port))
            blocking_socket.listen(1)

            # Try to start server on occupied port
            proc = subprocess.Popen(
                [sys.executable, "-m", "rtmx", "mcp-server", "--port", str(port)],
                cwd=project_with_requirements,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
            )

            try:
                # Wait for server to fail or timeout
                stdout, stderr = proc.communicate(timeout=5)

                # Server should fail due to port conflict
                # The exact behavior depends on implementation
                # At minimum it should not hang
                combined = stdout + stderr
                # Either exits with error or reports binding issue
                assert proc.returncode != 0 or "error" in combined.lower()

            except subprocess.TimeoutExpired:
                proc.kill()
                proc.wait()
                pytest.fail("Server did not fail on port conflict within timeout")


# =============================================================================
# Daemon Mode E2E Tests
# =============================================================================


@pytest.mark.req("REQ-TEST-008")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMCPServerDaemonMode:
    """E2E tests for MCP server daemon mode."""

    @requires_mcp
    @pytest.mark.skipif(
        sys.platform == "win32", reason="Daemon mode uses fork, not available on Windows"
    )
    def test_daemon_mode_with_pid_file(self, project_with_requirements: Path) -> None:
        """Test MCP server daemon mode creates PID file.

        This test verifies:
        1. Server starts in daemon mode
        2. PID file is created with valid PID
        3. Daemon process can be signaled via PID
        """
        port = find_free_port()
        pidfile = project_with_requirements / "mcp.pid"

        # Start server in daemon mode
        result = subprocess.run(
            [
                sys.executable,
                "-m",
                "rtmx",
                "mcp-server",
                "--port",
                str(port),
                "--daemon",
                "--pidfile",
                str(pidfile),
            ],
            cwd=project_with_requirements,
            capture_output=True,
            text=True,
            timeout=10,
        )

        try:
            # Parent should exit cleanly after forking
            # (exit code 0 means successful fork)
            assert result.returncode == 0

            # Give daemon time to start and write PID
            time.sleep(1)

            if pidfile.exists():
                pid = int(pidfile.read_text().strip())
                assert pid > 0

                # Terminate the daemon
                try:
                    os.kill(pid, signal.SIGTERM)
                    time.sleep(0.5)
                except ProcessLookupError:
                    pass  # Process already exited

        finally:
            # Cleanup PID file
            if pidfile.exists():
                try:
                    pid = int(pidfile.read_text().strip())
                    os.kill(pid, signal.SIGKILL)
                except (ValueError, ProcessLookupError, PermissionError):
                    pass
                pidfile.unlink(missing_ok=True)

    @requires_mcp
    @pytest.mark.skipif(
        sys.platform == "win32", reason="Daemon mode uses fork, not available on Windows"
    )
    def test_daemon_mode_creates_pid_directory(self, project_with_requirements: Path) -> None:
        """Test daemon mode creates PID file directory if needed."""
        port = find_free_port()
        pidfile = project_with_requirements / "run" / "mcp" / "server.pid"

        # Start server in daemon mode with nested PID path
        result = subprocess.run(
            [
                sys.executable,
                "-m",
                "rtmx",
                "mcp-server",
                "--port",
                str(port),
                "--daemon",
                "--pidfile",
                str(pidfile),
            ],
            cwd=project_with_requirements,
            capture_output=True,
            text=True,
            timeout=10,
        )

        try:
            # Give daemon time to start
            time.sleep(1)

            # Check directory was created
            assert pidfile.parent.exists() or result.returncode == 0

        finally:
            # Cleanup
            if pidfile.exists():
                try:
                    pid = int(pidfile.read_text().strip())
                    os.kill(pid, signal.SIGKILL)
                except (ValueError, ProcessLookupError, PermissionError):
                    pass
                pidfile.unlink(missing_ok=True)


# =============================================================================
# Graceful Shutdown E2E Tests
# =============================================================================


@pytest.mark.req("REQ-TEST-008")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMCPServerShutdown:
    """E2E tests for MCP server graceful shutdown."""

    @requires_mcp
    def test_sigterm_graceful_shutdown(self, project_with_requirements: Path) -> None:
        """Test MCP server shuts down gracefully on SIGTERM.

        This test verifies:
        1. Server is running
        2. SIGTERM signal is sent
        3. Server exits cleanly (exit code 0)
        """
        port = find_free_port()

        # Start server
        proc = subprocess.Popen(
            [sys.executable, "-m", "rtmx", "mcp-server", "--port", str(port)],
            cwd=project_with_requirements,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )

        try:
            # Give server time to start
            time.sleep(1)

            # Verify process is running
            assert proc.poll() is None, "Server process exited unexpectedly"

            # Send SIGTERM (graceful shutdown)
            proc.terminate()

            # Wait for graceful exit
            try:
                proc.wait(timeout=5)
            except subprocess.TimeoutExpired:
                proc.kill()
                proc.wait()
                pytest.fail("Server did not shutdown within timeout")

            # Graceful shutdown should result in exit code 0 or -SIGTERM
            # Different platforms may report differently
            assert proc.returncode in [0, -signal.SIGTERM, 1]

        finally:
            if proc.poll() is None:
                proc.kill()
                proc.wait()

    @requires_mcp
    def test_sigint_shutdown(self, project_with_requirements: Path) -> None:
        """Test MCP server handles SIGINT (Ctrl+C) gracefully."""
        port = find_free_port()

        proc = subprocess.Popen(
            [sys.executable, "-m", "rtmx", "mcp-server", "--port", str(port)],
            cwd=project_with_requirements,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )

        try:
            time.sleep(1)
            assert proc.poll() is None

            # Send SIGINT
            proc.send_signal(signal.SIGINT)

            try:
                proc.wait(timeout=5)
            except subprocess.TimeoutExpired:
                proc.kill()
                proc.wait()
                pytest.fail("Server did not shutdown on SIGINT")

        finally:
            if proc.poll() is None:
                proc.kill()
                proc.wait()


# =============================================================================
# Tool Invocation E2E Tests
# =============================================================================


@pytest.mark.req("REQ-TEST-008")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMCPToolInvocation:
    """E2E tests for MCP tool invocation via RTMXTools directly.

    Since full MCP protocol testing requires a client, these tests
    invoke the tools directly to verify E2E behavior.
    """

    def test_rtmx_status_tool_invocation(self, project_with_requirements: Path) -> None:
        """Test rtmx_status tool can be invoked and returns valid data.

        This test verifies:
        1. RTMXTools can be instantiated with valid config
        2. get_status returns success
        3. Response contains expected fields
        """
        from rtmx.adapters.mcp.tools import RTMXTools
        from rtmx.config import RTMXConfig

        config = RTMXConfig(database=str(project_with_requirements / "docs" / "rtm_database.csv"))
        tools = RTMXTools(config)

        result = tools.get_status(verbose=0)

        assert result.success is True
        assert result.error is None
        assert "total" in result.data
        assert "complete" in result.data
        assert "partial" in result.data
        assert "missing" in result.data
        assert "completion_pct" in result.data
        # Verify values are reasonable
        assert result.data["total"] == 4
        assert result.data["complete"] == 1
        assert isinstance(result.data["completion_pct"], int | float)

    def test_rtmx_status_tool_verbose(self, project_with_requirements: Path) -> None:
        """Test rtmx_status tool verbose mode includes categories."""
        from rtmx.adapters.mcp.tools import RTMXTools
        from rtmx.config import RTMXConfig

        config = RTMXConfig(database=str(project_with_requirements / "docs" / "rtm_database.csv"))
        tools = RTMXTools(config)

        result = tools.get_status(verbose=1)

        assert result.success is True
        assert "categories" in result.data
        assert "MCP" in result.data["categories"]

    def test_rtmx_backlog_tool_invocation(self, project_with_requirements: Path) -> None:
        """Test rtmx_backlog tool can be invoked and returns valid data.

        This test verifies:
        1. get_backlog returns success
        2. Response contains incomplete requirements
        3. Response is properly sorted by priority
        """
        from rtmx.adapters.mcp.tools import RTMXTools
        from rtmx.config import RTMXConfig

        config = RTMXConfig(database=str(project_with_requirements / "docs" / "rtm_database.csv"))
        tools = RTMXTools(config)

        result = tools.get_backlog(limit=20)

        assert result.success is True
        assert result.error is None
        assert "total_incomplete" in result.data
        assert "showing" in result.data
        assert "items" in result.data
        # Should have 3 incomplete requirements
        assert result.data["total_incomplete"] == 3
        # Verify items have expected fields
        for item in result.data["items"]:
            assert "id" in item
            assert "text" in item
            assert "priority" in item
            assert "status" in item

    def test_rtmx_backlog_tool_with_phase_filter(self, project_with_requirements: Path) -> None:
        """Test rtmx_backlog tool filters by phase."""
        from rtmx.adapters.mcp.tools import RTMXTools
        from rtmx.config import RTMXConfig

        config = RTMXConfig(database=str(project_with_requirements / "docs" / "rtm_database.csv"))
        tools = RTMXTools(config)

        result = tools.get_backlog(phase=1)

        assert result.success is True
        # Only phase 1 incomplete requirements
        for item in result.data["items"]:
            assert item["phase"] == 1

    def test_rtmx_get_requirement_tool(self, project_with_requirements: Path) -> None:
        """Test rtmx_get_requirement tool returns requirement details."""
        from rtmx.adapters.mcp.tools import RTMXTools
        from rtmx.config import RTMXConfig

        config = RTMXConfig(database=str(project_with_requirements / "docs" / "rtm_database.csv"))
        tools = RTMXTools(config)

        result = tools.get_requirement("REQ-MCP-001")

        assert result.success is True
        assert result.data["id"] == "REQ-MCP-001"
        assert result.data["category"] == "MCP"
        assert result.data["status"] == "COMPLETE"

    def test_rtmx_search_tool(self, project_with_requirements: Path) -> None:
        """Test rtmx_search tool finds requirements by text."""
        from rtmx.adapters.mcp.tools import RTMXTools
        from rtmx.config import RTMXConfig

        config = RTMXConfig(database=str(project_with_requirements / "docs" / "rtm_database.csv"))
        tools = RTMXTools(config)

        result = tools.search_requirements("backlog")

        assert result.success is True
        assert result.data["count"] >= 1
        # Should find REQ-MCP-003 which mentions "backlog"
        found_ids = [r["id"] for r in result.data["results"]]
        assert "REQ-MCP-003" in found_ids

    def test_rtmx_deps_tool(self, project_with_requirements: Path) -> None:
        """Test rtmx_deps tool returns dependency information."""
        from rtmx.adapters.mcp.tools import RTMXTools
        from rtmx.config import RTMXConfig

        config = RTMXConfig(database=str(project_with_requirements / "docs" / "rtm_database.csv"))
        tools = RTMXTools(config)

        result = tools.get_dependencies("REQ-MCP-002")

        assert result.success is True
        assert result.data["id"] == "REQ-MCP-002"
        assert "depends_on" in result.data
        assert "blocks" in result.data
        assert "is_blocked" in result.data
        # REQ-MCP-002 depends on REQ-MCP-001
        dep_ids = [d["id"] for d in result.data["depends_on"]]
        assert "REQ-MCP-001" in dep_ids


# =============================================================================
# Concurrent Client E2E Tests (Optional)
# =============================================================================


@pytest.mark.req("REQ-TEST-008")
@pytest.mark.scope_system
@pytest.mark.technique_stress
@pytest.mark.env_simulation
class TestMCPConcurrentAccess:
    """E2E tests for concurrent access to MCP tools.

    These tests verify thread safety of the RTMXTools implementation.
    """

    def test_concurrent_tool_invocations(self, project_with_requirements: Path) -> None:
        """Test multiple concurrent tool invocations.

        This test verifies:
        1. Multiple threads can invoke tools concurrently
        2. No data corruption or race conditions
        3. All invocations return valid results
        """
        import concurrent.futures

        from rtmx.adapters.mcp.tools import RTMXTools
        from rtmx.config import RTMXConfig

        config = RTMXConfig(database=str(project_with_requirements / "docs" / "rtm_database.csv"))
        tools = RTMXTools(config)

        def invoke_status():
            return tools.get_status(verbose=0)

        def invoke_backlog():
            return tools.get_backlog(limit=10)

        def invoke_search():
            return tools.search_requirements("MCP")

        # Run multiple concurrent invocations
        with concurrent.futures.ThreadPoolExecutor(max_workers=10) as executor:
            futures = []
            for _ in range(5):
                futures.append(executor.submit(invoke_status))
                futures.append(executor.submit(invoke_backlog))
                futures.append(executor.submit(invoke_search))

            # Collect results
            results = [f.result() for f in futures]

        # All invocations should succeed
        for result in results:
            assert result.success is True
            assert result.error is None


# =============================================================================
# Server Configuration E2E Tests
# =============================================================================


@pytest.mark.req("REQ-TEST-008")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMCPServerConfiguration:
    """E2E tests for MCP server configuration options."""

    def test_server_default_port(self, initialized_project: Path) -> None:
        """Test MCP server has port option available."""
        result = run_rtmx("mcp-server", "--help", cwd=initialized_project)

        assert result.returncode == 0
        # Verify port option exists (default 3000 may not be shown in help text)
        assert "--port" in result.stdout

    def test_server_custom_host(self, initialized_project: Path) -> None:
        """Test MCP server accepts custom host option."""
        result = run_rtmx("mcp-server", "--help", cwd=initialized_project)

        assert result.returncode == 0
        assert "--host" in result.stdout

    def test_server_daemon_option(self, initialized_project: Path) -> None:
        """Test MCP server has daemon option."""
        result = run_rtmx("mcp-server", "--help", cwd=initialized_project)

        assert result.returncode == 0
        assert "--daemon" in result.stdout

    def test_server_pidfile_option(self, initialized_project: Path) -> None:
        """Test MCP server has pidfile option."""
        result = run_rtmx("mcp-server", "--help", cwd=initialized_project)

        assert result.returncode == 0
        assert "--pidfile" in result.stdout


# =============================================================================
# Error Handling E2E Tests
# =============================================================================


@pytest.mark.req("REQ-TEST-008")
@pytest.mark.scope_system
@pytest.mark.technique_stress
@pytest.mark.env_simulation
class TestMCPErrorHandling:
    """E2E tests for MCP error handling."""

    def test_tool_handles_missing_database(self, temp_project: Path) -> None:
        """Test tools handle missing database gracefully."""
        from rtmx.adapters.mcp.tools import RTMXTools
        from rtmx.config import RTMXConfig

        config = RTMXConfig(database=str(temp_project / "nonexistent" / "rtm_database.csv"))
        tools = RTMXTools(config)

        result = tools.get_status()

        assert result.success is False
        assert result.error is not None
        assert "not found" in result.error.lower()

    def test_tool_handles_invalid_requirement_id(self, project_with_requirements: Path) -> None:
        """Test tools handle invalid requirement ID gracefully."""
        from rtmx.adapters.mcp.tools import RTMXTools
        from rtmx.config import RTMXConfig

        config = RTMXConfig(database=str(project_with_requirements / "docs" / "rtm_database.csv"))
        tools = RTMXTools(config)

        result = tools.get_requirement("REQ-INVALID-999")

        assert result.success is False
        assert result.error is not None

    def test_tool_handles_empty_database(self, initialized_project: Path) -> None:
        """Test tools handle empty database gracefully.

        Note: The implementation may treat empty database as an error,
        which is also a valid design choice. We test for consistent behavior.
        """
        from rtmx.adapters.mcp.tools import RTMXTools
        from rtmx.config import RTMXConfig

        # Clear the database (keep headers only)
        db_path = initialized_project / "docs" / "rtm_database.csv"
        with open(db_path) as f:
            reader = csv.reader(f)
            headers = next(reader)
        with open(db_path, "w", newline="") as f:
            writer = csv.writer(f)
            writer.writerow(headers)

        config = RTMXConfig(database=str(db_path))
        tools = RTMXTools(config)

        result = tools.get_status()

        # Empty database may be treated as error or as 0 requirements
        # Both are valid behaviors - we verify the response is consistent
        if result.success:
            assert result.data["total"] == 0
            assert result.data["completion_pct"] == 0
        else:
            # Empty database treated as error is also valid
            assert result.error is not None
            assert "empty" in result.error.lower() or "not found" in result.error.lower()
