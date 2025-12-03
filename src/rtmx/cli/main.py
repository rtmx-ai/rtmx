"""RTMX CLI main entry point.

This module provides the main CLI command group and common options.
"""

from __future__ import annotations

import sys
from pathlib import Path

import click

from rtmx import __version__
from rtmx.config import RTMXConfig, load_config
from rtmx.formatting import Colors


@click.group()
@click.version_option(version=__version__, prog_name="rtmx")
@click.option(
    "--rtm-csv",
    type=click.Path(exists=False, path_type=Path),
    help="Path to RTM database CSV",
)
@click.option(
    "--config",
    "-c",
    "config_path",
    type=click.Path(exists=True, path_type=Path),
    help="Path to rtmx.yaml config file",
)
@click.option(
    "--no-color",
    is_flag=True,
    help="Disable colored output",
)
@click.pass_context
def main(
    ctx: click.Context,
    rtm_csv: Path | None,
    config_path: Path | None,
    no_color: bool,
) -> None:
    """RTMX - Requirements Traceability Matrix toolkit.

    Manage requirements traceability for GenAI-driven development.
    """
    ctx.ensure_object(dict)

    # Load configuration
    config = load_config(config_path)
    ctx.obj["config"] = config

    # RTM CSV path precedence: CLI arg > config > default discovery
    ctx.obj["rtm_csv"] = rtm_csv or config.database
    ctx.obj["no_color"] = no_color

    if no_color or not sys.stdout.isatty():
        Colors.disable()


@main.command()
@click.option(
    "-v",
    "--verbose",
    count=True,
    help="Increase verbosity (-v categories, -vv subcategories, -vvv all)",
)
@click.option(
    "--json",
    "json_output",
    type=click.Path(path_type=Path),
    help="Export status as JSON",
)
@click.pass_context
def status(ctx: click.Context, verbose: int, json_output: Path | None) -> None:
    """Show RTM status.

    Displays completion status with pytest-style verbosity levels.
    """
    from rtmx.cli.status import run_status

    rtm_csv = ctx.obj.get("rtm_csv")
    run_status(rtm_csv, verbose, json_output)


@main.command()
@click.option(
    "--phase",
    type=int,
    help="Filter by phase number",
)
@click.option(
    "--critical",
    is_flag=True,
    help="Show critical path only",
)
@click.pass_context
def backlog(ctx: click.Context, phase: int | None, critical: bool) -> None:
    """Show prioritized backlog.

    Displays incomplete requirements sorted by priority and blocking count.
    """
    from rtmx.cli.backlog import run_backlog

    rtm_csv = ctx.obj.get("rtm_csv")
    run_backlog(rtm_csv, phase, critical)


@main.command()
@click.option(
    "--execute",
    is_flag=True,
    help="Execute fixes (default is dry-run)",
)
@click.pass_context
def reconcile(ctx: click.Context, execute: bool) -> None:
    """Check and fix dependency reciprocity.

    Ensures dependency/blocks relationships are consistent.
    """
    from rtmx.cli.reconcile import run_reconcile

    rtm_csv = ctx.obj.get("rtm_csv")
    run_reconcile(rtm_csv, execute)


@main.command()
@click.option(
    "--category",
    help="Filter by category",
)
@click.option(
    "--phase",
    type=int,
    help="Filter by phase",
)
@click.option(
    "--req",
    "req_id",
    help="Show dependencies for specific requirement",
)
@click.pass_context
def deps(ctx: click.Context, category: str | None, phase: int | None, req_id: str | None) -> None:
    """Show dependency graph.

    Visualize requirement dependencies.
    """
    from rtmx.cli.deps import run_deps

    rtm_csv = ctx.obj.get("rtm_csv")
    run_deps(rtm_csv, category, phase, req_id)


@main.command()
@click.pass_context
def cycles(ctx: click.Context) -> None:
    """Detect circular dependencies.

    Find and report dependency cycles using Tarjan's algorithm.
    """
    from rtmx.cli.cycles import run_cycles

    rtm_csv = ctx.obj.get("rtm_csv")
    run_cycles(rtm_csv)


@main.command()
@click.option(
    "--force",
    is_flag=True,
    help="Overwrite existing files",
)
@click.pass_context
def init(_ctx: click.Context, force: bool) -> None:
    """Initialize RTM in current project.

    Creates docs/rtm_database.csv and docs/requirements/ structure.
    """
    from rtmx.cli.init import run_init

    run_init(force)


@main.command("from-tests")
@click.argument(
    "test_path",
    required=False,
    type=click.Path(exists=True, path_type=Path),
)
@click.option(
    "--all",
    "show_all",
    is_flag=True,
    help="Show all markers found",
)
@click.option(
    "--missing",
    "show_missing",
    is_flag=True,
    help="Show requirements not in database",
)
@click.option(
    "--update",
    is_flag=True,
    help="Update RTM database with test information",
)
@click.pass_context
def from_tests(
    ctx: click.Context,
    test_path: Path | None,
    show_all: bool,
    show_missing: bool,
    update: bool,
) -> None:
    """Scan test files for requirement markers.

    Extracts @pytest.mark.req() markers from test files and reports coverage.
    """
    from rtmx.cli.from_tests import run_from_tests

    rtm_csv = ctx.obj.get("rtm_csv")
    run_from_tests(
        str(test_path) if test_path else None,
        str(rtm_csv) if rtm_csv else None,
        show_all,
        show_missing,
        update,
    )


@main.command()
@click.option(
    "--output",
    "-o",
    type=click.Path(path_type=Path),
    help="Output file (default: stdout)",
)
@click.pass_context
def makefile(_ctx: click.Context, output: Path | None) -> None:
    """Generate Makefile targets for RTM commands.

    Outputs Makefile targets that can be appended to your project's Makefile.
    """
    from rtmx.cli.makefile import run_makefile

    run_makefile(output)


# =============================================================================
# Agent Integration Commands (Phase 6+)
# =============================================================================


@main.command()
@click.argument(
    "path",
    required=False,
    type=click.Path(exists=True, path_type=Path),
)
@click.option(
    "--output",
    "-o",
    type=click.Path(path_type=Path),
    help="Output analysis report",
)
@click.option(
    "--format",
    "output_format",
    type=click.Choice(["terminal", "json", "markdown"]),
    default="terminal",
    help="Output format",
)
@click.option(
    "--deep",
    is_flag=True,
    help="Include source code analysis",
)
@click.pass_context
def analyze(
    ctx: click.Context,
    path: Path | None,
    output: Path | None,
    output_format: str,
    deep: bool,
) -> None:
    """Analyze project for requirements artifacts.

    Discovers tests, issues, documentation that could become requirements.
    """
    from rtmx.cli.analyze import run_analyze

    config: RTMXConfig = ctx.obj["config"]
    run_analyze(path, output, output_format, deep, config)


@main.command()
@click.option(
    "--from-tests",
    "from_tests",
    is_flag=True,
    help="Generate requirements from test functions",
)
@click.option(
    "--from-github",
    "from_github",
    is_flag=True,
    help="Import from GitHub issues",
)
@click.option(
    "--from-jira",
    "from_jira",
    is_flag=True,
    help="Import from Jira tickets",
)
@click.option(
    "--merge",
    is_flag=True,
    help="Merge with existing RTM (don't overwrite)",
)
@click.option(
    "--dry-run",
    is_flag=True,
    help="Preview without writing files",
)
@click.option(
    "--prefix",
    default="REQ",
    help="Requirement ID prefix",
)
@click.pass_context
def bootstrap(
    ctx: click.Context,
    from_tests: bool,
    from_github: bool,
    from_jira: bool,
    merge: bool,
    dry_run: bool,
    prefix: str,
) -> None:
    """Generate initial RTM from project artifacts.

    Bootstrap requirements traceability from existing tests, issues, or tickets.
    """
    from rtmx.cli.bootstrap import run_bootstrap

    config: RTMXConfig = ctx.obj["config"]
    run_bootstrap(from_tests, from_github, from_jira, merge, dry_run, prefix, config)


@main.command()
@click.option(
    "--dry-run",
    is_flag=True,
    help="Preview changes without writing",
)
@click.option(
    "--non-interactive",
    is_flag=True,
    help="Skip confirmation prompts (for CI)",
)
@click.option(
    "--force",
    is_flag=True,
    help="Overwrite existing rtmx sections",
)
@click.option(
    "--agents",
    help="Comma-separated agent list (claude,cursor,copilot)",
)
@click.option(
    "--all",
    "install_all",
    is_flag=True,
    help="Install to all detected agents",
)
@click.option(
    "--skip-backup",
    is_flag=True,
    help="Don't backup modified files",
)
@click.pass_context
def install(
    ctx: click.Context,
    dry_run: bool,
    non_interactive: bool,
    force: bool,
    agents: str | None,
    install_all: bool,
    skip_backup: bool,
) -> None:
    """Inject RTM-aware prompts into AI agent configurations.

    Detects and enhances CLAUDE.md, .cursorrules, copilot-instructions.md.
    """
    from rtmx.cli.install import run_install

    config: RTMXConfig = ctx.obj["config"]
    agent_list = agents.split(",") if agents else None
    run_install(dry_run, non_interactive, force, agent_list, install_all, skip_backup, config)


@main.command()
@click.argument(
    "service",
    type=click.Choice(["github", "jira"]),
)
@click.option(
    "--import",
    "do_import",
    is_flag=True,
    help="Pull from service into RTM",
)
@click.option(
    "--export",
    "do_export",
    is_flag=True,
    help="Push RTM to service",
)
@click.option(
    "--bidirectional",
    is_flag=True,
    help="Two-way sync",
)
@click.option(
    "--dry-run",
    is_flag=True,
    help="Preview changes",
)
@click.option(
    "--prefer-local",
    is_flag=True,
    help="RTM wins on conflicts",
)
@click.option(
    "--prefer-remote",
    is_flag=True,
    help="Service wins on conflicts",
)
@click.pass_context
def sync(
    ctx: click.Context,
    service: str,
    do_import: bool,
    do_export: bool,
    bidirectional: bool,
    dry_run: bool,
    prefer_local: bool,
    prefer_remote: bool,
) -> None:
    """Synchronize RTM with external services.

    Bi-directional sync with GitHub Issues or Jira.
    """
    from rtmx.cli.sync import run_sync

    config: RTMXConfig = ctx.obj["config"]
    run_sync(
        service, do_import, do_export, bidirectional, dry_run, prefer_local, prefer_remote, config
    )


@main.command("mcp-server")
@click.option(
    "--port",
    type=int,
    default=3000,
    help="Server port",
)
@click.option(
    "--host",
    default="localhost",
    help="Bind address",
)
@click.option(
    "--daemon",
    is_flag=True,
    help="Run as background daemon",
)
@click.option(
    "--pidfile",
    type=click.Path(path_type=Path),
    help="Write PID file for daemon management",
)
@click.pass_context
def mcp_server(
    ctx: click.Context,
    port: int,
    host: str,
    daemon: bool,
    pidfile: Path | None,
) -> None:
    """Start MCP protocol server for AI agent integration.

    Exposes rtmx operations as MCP tools.
    """
    from rtmx.cli.mcp_server import run_mcp_server

    config: RTMXConfig = ctx.obj["config"]
    run_mcp_server(port, host, daemon, pidfile, config)


if __name__ == "__main__":
    main()
