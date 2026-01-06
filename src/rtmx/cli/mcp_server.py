"""RTMX MCP server command.

Start MCP protocol server for AI agent integration.
"""

from __future__ import annotations

import asyncio
import os
import sys
from pathlib import Path

from rtmx.config import RTMXConfig
from rtmx.formatting import Colors


def run_mcp_server(
    port: int,
    host: str,
    daemon: bool,
    pidfile: Path | None,
    config: RTMXConfig,
) -> None:
    """Run MCP server command.

    Start MCP protocol server exposing rtmx operations as tools.

    Args:
        port: Server port
        host: Bind address
        daemon: Run as background daemon
        pidfile: PID file path for daemon management
        config: RTMX configuration
    """
    # Check if mcp package is available
    try:
        import mcp  # noqa: F401
    except ImportError:
        print(f"{Colors.RED}MCP package not installed.{Colors.RESET}")
        print("Install with: pip install rtmx[mcp]")
        sys.exit(1)

    # Check if stdin is a TTY (interactive terminal)
    # MCP servers communicate over stdin/stdout with JSON-RPC, not for interactive use
    if sys.stdin.isatty() and not daemon:
        print("=== RTMX MCP Server ===")
        print()
        print(f"{Colors.BOLD}Server Configuration:{Colors.RESET}")
        print(f"  Host: {host}")
        print(f"  Port: {port}")
        print()
        print(f"{Colors.BOLD}Available Tools:{Colors.RESET}")
        print("  rtmx_status           - Get completion status")
        print("  rtmx_backlog          - Get prioritized backlog")
        print("  rtmx_get_requirement  - Get requirement details")
        print("  rtmx_update_status    - Update requirement status")
        print("  rtmx_deps             - Get dependencies")
        print("  rtmx_search           - Search requirements")
        print()
        print(
            f"{Colors.YELLOW}Note: MCP server is designed to be run by an MCP client,{Colors.RESET}"
        )
        print(f"{Colors.YELLOW}not directly from a terminal.{Colors.RESET}")
        print()
        print("To use the MCP server:")
        print("  1. Configure your MCP client (e.g., Claude Desktop) to run:")
        print(f"     {sys.executable} -m rtmx mcp-server")
        print("  2. Or pipe commands for testing:")
        print("     echo '{}' | rtmx mcp-server")
        print()
        print("See https://iotactical.github.io/rtmx/adapters/mcp for setup instructions.")
        sys.exit(0)

    # Handle daemon mode
    if daemon:
        _daemonize(pidfile)

    # Run the server
    print(f"{Colors.GREEN}Starting MCP server...{Colors.RESET}", file=sys.stderr)

    try:
        from rtmx.adapters.mcp.server import run_server

        asyncio.run(run_server(config))
    except KeyboardInterrupt:
        print(f"{Colors.YELLOW}Server stopped{Colors.RESET}", file=sys.stderr)
    except (EOFError, ConnectionError, BrokenPipeError):
        # Client disconnected - normal shutdown
        print(f"{Colors.YELLOW}Client disconnected{Colors.RESET}", file=sys.stderr)
        sys.exit(0)
    except Exception as e:
        # Handle asyncio TaskGroup errors gracefully (Python 3.11+ ExceptionGroup)
        # This typically happens when the MCP client disconnects
        if type(e).__name__ == "ExceptionGroup" and hasattr(e, "exceptions"):
            for exc in e.exceptions:
                if isinstance(exc, EOFError | ConnectionError | BrokenPipeError):
                    print(f"{Colors.YELLOW}Client disconnected{Colors.RESET}", file=sys.stderr)
                    sys.exit(0)
        print(f"{Colors.RED}Server error: {e}{Colors.RESET}", file=sys.stderr)
        sys.exit(1)


def _daemonize(pidfile: Path | None) -> None:
    """Daemonize the process.

    Uses double-fork technique to properly detach from terminal.

    Args:
        pidfile: Path to write PID file
    """
    # First fork
    try:
        pid = os.fork()
        if pid > 0:
            # Parent exits
            sys.exit(0)
    except OSError as e:
        print(f"{Colors.RED}Fork #1 failed: {e}{Colors.RESET}")
        sys.exit(1)

    # Decouple from parent environment
    os.chdir("/")
    os.setsid()
    os.umask(0)

    # Second fork
    try:
        pid = os.fork()
        if pid > 0:
            # Parent exits
            sys.exit(0)
    except OSError as e:
        print(f"{Colors.RED}Fork #2 failed: {e}{Colors.RESET}")
        sys.exit(1)

    # Redirect standard file descriptors
    sys.stdout.flush()
    sys.stderr.flush()

    with open("/dev/null") as devnull:
        os.dup2(devnull.fileno(), sys.stdin.fileno())

    # Write PID file
    if pidfile:
        pidfile.parent.mkdir(parents=True, exist_ok=True)
        pidfile.write_text(str(os.getpid()))
        print(f"Daemon started with PID {os.getpid()}")
