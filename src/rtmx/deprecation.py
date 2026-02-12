"""Deprecation warnings for Python CLI transition to Go CLI.

This module provides deprecation warnings to inform users about the
transition from the Python CLI to the Go CLI (rtmx-go).

REQ-DIST-002: Go CLI shall replace Python CLI as primary implementation.
"""

from __future__ import annotations

import os
import sys
import warnings
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    pass

# Environment variable to suppress deprecation warnings
SUPPRESS_DEPRECATION_ENV = "RTMX_SUPPRESS_DEPRECATION"

# Version when Go CLI becomes primary
GO_CLI_PRIMARY_VERSION = "v1.0.0"

# Installation instructions by platform
INSTALL_INSTRUCTIONS = {
    "darwin": "brew install rtmx-ai/tap/rtmx",
    "linux": "curl -fsSL https://rtmx.ai/install.sh | bash",
    "win32": "scoop bucket add rtmx https://github.com/rtmx-ai/scoop-bucket && scoop install rtmx",
}

DEPRECATION_MESSAGE = """\
================================================================================
DEPRECATION NOTICE: Python CLI is being replaced by Go CLI
================================================================================

The RTMX Python CLI will be deprecated in {version}. A faster, standalone
Go CLI is now available with identical functionality and no runtime dependencies.

Install the Go CLI:
  {install_cmd}

Or download directly:
  https://github.com/rtmx-ai/rtmx-go/releases/latest

Migration guide:
  https://rtmx.ai/docs/go-migration

To suppress this warning:
  export {env_var}=1

The Python package will remain as a minimal pytest plugin for marker support.
================================================================================
"""


def _get_install_command() -> str:
    """Get platform-specific installation command."""
    platform = sys.platform
    if platform.startswith("linux"):
        return INSTALL_INSTRUCTIONS["linux"]
    elif platform == "darwin":
        return INSTALL_INSTRUCTIONS["darwin"]
    elif platform == "win32":
        return INSTALL_INSTRUCTIONS["win32"]
    else:
        return "go install github.com/rtmx-ai/rtmx-go/cmd/rtmx@latest"


def show_deprecation_warning(force: bool = False) -> None:
    """Show deprecation warning about Go CLI transition.

    Args:
        force: If True, show warning even if suppressed by environment variable.
    """
    # Check if suppressed
    if not force and os.environ.get(SUPPRESS_DEPRECATION_ENV):
        return

    # Only show in interactive terminals
    if not sys.stderr.isatty():
        return

    # Format and display warning
    message = DEPRECATION_MESSAGE.format(
        version=GO_CLI_PRIMARY_VERSION,
        install_cmd=_get_install_command(),
        env_var=SUPPRESS_DEPRECATION_ENV,
    )

    # Use ANSI colors if available
    try:
        yellow = "\033[33m"
        reset = "\033[0m"
        sys.stderr.write(f"{yellow}{message}{reset}\n")
    except Exception:
        sys.stderr.write(f"{message}\n")


def emit_command_deprecation(command_name: str) -> None:
    """Emit a Python DeprecationWarning for a specific command.

    This is for programmatic deprecation tracking.

    Args:
        command_name: Name of the deprecated CLI command.
    """
    warnings.warn(
        f"rtmx {command_name}: Python CLI will be replaced by Go CLI in {GO_CLI_PRIMARY_VERSION}. "
        f"Install Go CLI: {_get_install_command()}",
        DeprecationWarning,
        stacklevel=3,
    )


def check_go_cli_available() -> tuple[bool, str | None]:
    """Check if Go CLI is available on the system.

    Returns:
        Tuple of (is_available, version_string).
    """
    import shutil
    import subprocess

    rtmx_go = shutil.which("rtmx-go") or shutil.which("rtmx")
    if not rtmx_go:
        return False, None

    try:
        result = subprocess.run(
            [rtmx_go, "version", "--short"],
            capture_output=True,
            text=True,
            timeout=5,
        )
        if result.returncode == 0 and "go" in result.stdout.lower():
            return True, result.stdout.strip()
    except Exception:
        pass

    return False, None
