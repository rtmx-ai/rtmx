"""RTMX CLI main entry point.

This module provides the main CLI command group and common options.
"""

from __future__ import annotations

import sys
from pathlib import Path

import click

from rtmx import __version__
from rtmx.cli.markers import markers as markers_group
from rtmx.config import RTMXConfig, load_config
from rtmx.deprecation import show_deprecation_warning
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
@click.option(
    "--no-migrate",
    is_flag=True,
    help="Suppress automatic migration from legacy layout to .rtmx/",
)
@click.pass_context
def main(
    ctx: click.Context,
    rtm_csv: Path | None,
    config_path: Path | None,
    no_color: bool,
    no_migrate: bool,
) -> None:
    """RTMX - Requirements Traceability Matrix toolkit.

    Manage requirements traceability for GenAI-driven development.
    """
    # Show deprecation warning about Go CLI transition (REQ-DIST-002)
    show_deprecation_warning()

    ctx.ensure_object(dict)

    # Store no_migrate flag for later use
    ctx.obj["no_migrate"] = no_migrate

    # Check for legacy layout and offer migration (unless suppressed)
    # Only check if we're in an interactive context and not running init/setup
    invoked_subcommand = ctx.invoked_subcommand
    skip_migration_commands = {"init", "setup", "version", None}

    if invoked_subcommand not in skip_migration_commands and sys.stdout.isatty() and not no_migrate:
        from rtmx.migration import run_migration_if_needed

        run_migration_if_needed(
            root_path=Path.cwd(),
            no_migrate=no_migrate,
            interactive=True,
        )

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
@click.option(
    "--rich/--no-rich",
    "use_rich",
    default=None,
    help="Force rich or plain output (auto-detects by default)",
)
@click.option(
    "--live",
    is_flag=True,
    default=False,
    help="Watch file and auto-refresh on changes",
)
@click.option(
    "--fail-under",
    type=float,
    default=None,
    help="Exit with code 1 if completion percentage is below this threshold (e.g., 80.0)",
)
@click.pass_context
def status(
    ctx: click.Context,
    verbose: int,
    json_output: Path | None,
    use_rich: bool | None,
    live: bool,
    fail_under: float | None,
) -> None:
    """Show RTM status.

    Displays completion status with pytest-style verbosity levels.
    Use --rich for enhanced terminal output with progress bars (requires rich library).
    Use --live to watch for file changes and auto-refresh.
    Use --fail-under to exit with code 1 if completion is below a threshold.
    """
    from rtmx.cli.status import run_status

    rtm_csv = ctx.obj.get("rtm_csv")
    run_status(rtm_csv, verbose, json_output, use_rich, live, fail_under)


@main.command()
@click.option(
    "--host",
    default="127.0.0.1",
    help="Bind address (default: 127.0.0.1)",
)
@click.option(
    "--port",
    type=int,
    default=8080,
    help="Port number (default: 8080)",
)
@click.option(
    "--reload",
    is_flag=True,
    help="Enable auto-reload on code changes",
)
@click.pass_context
def serve(
    ctx: click.Context,
    host: str,
    port: int,
    reload: bool,
) -> None:
    """Start the RTMX web dashboard server.

    Launches a FastAPI server with real-time WebSocket updates.
    The dashboard auto-refreshes when the RTM database changes.

    Requires the web dependencies: pip install rtmx[web]

    \b
    Examples:
        rtmx serve                      # Start on localhost:8080
        rtmx serve --port 3000          # Custom port
        rtmx serve --host 0.0.0.0       # Allow external connections
        rtmx serve --reload             # Auto-reload on code changes
    """
    from rtmx.cli.serve import run_serve

    rtm_csv = ctx.obj.get("rtm_csv")
    run_serve(rtm_csv, host, port, reload)


@main.command()
@click.pass_context
def tui(ctx: click.Context) -> None:
    """Launch interactive TUI dashboard.

    Provides a split-pane view with requirements list and detail panels.
    Navigate with vim-style keys (j/k/g/G), press 'q' to quit.

    Requires the textual library: pip install rtmx[tui]
    """
    from rtmx.cli.tui import run_tui

    rtm_csv = ctx.obj.get("rtm_csv")
    run_tui(rtm_csv)


@main.command()
@click.option(
    "--phase",
    type=int,
    help="Filter by phase number",
)
@click.option(
    "--view",
    type=click.Choice(["all", "critical", "quick-wins", "blockers", "list"]),
    default="all",
    help="View mode: all, critical (path), quick-wins, blockers, list",
)
@click.option(
    "--limit",
    "-n",
    type=int,
    default=10,
    help="Limit items in summary views (default: 10)",
)
@click.pass_context
def backlog(ctx: click.Context, phase: int | None, view: str, limit: int) -> None:
    """Show prioritized backlog.

    Displays incomplete requirements sorted by priority and blocking count.

    View modes:

      all        - Full backlog with all incomplete requirements

      critical   - Top items blocking the most other requirements

      quick-wins - HIGH/P0 priority, ≤1 week effort, unblocked

      blockers   - Requirements that block others

      list       - Complete list of all requirements for a phase
    """
    from rtmx.cli.backlog import BacklogView, run_backlog

    rtm_csv = ctx.obj.get("rtm_csv")
    run_backlog(rtm_csv, phase, BacklogView(view), limit)


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
    "--dry-run",
    is_flag=True,
    help="Preview changes without making them",
)
@click.option(
    "--minimal",
    is_flag=True,
    help="Only create config and RTM, skip agents and Makefile",
)
@click.option(
    "--force",
    is_flag=True,
    help="Overwrite existing files",
)
@click.option(
    "--skip-agents",
    is_flag=True,
    help="Skip agent config injection",
)
@click.option(
    "--skip-makefile",
    is_flag=True,
    help="Skip Makefile targets",
)
@click.option(
    "--branch",
    is_flag=True,
    help="Create git branch for isolation (safer for existing projects)",
)
@click.option(
    "--pr",
    "create_pr",
    is_flag=True,
    help="Create pull request after setup (implies --branch)",
)
@click.option(
    "--scaffold",
    is_flag=True,
    help="Auto-generate requirement spec files from database entries",
)
@click.pass_context
def setup(
    _ctx: click.Context,
    dry_run: bool,
    minimal: bool,
    force: bool,
    skip_agents: bool,
    skip_makefile: bool,
    branch: bool,
    create_pr: bool,
    scaffold: bool,
) -> None:
    """Complete rtmx setup in one command.

    Performs full integration: config, RTM, agent prompts, Makefile.
    Safe to run multiple times (idempotent). Creates backups before modifying files.

    \b
    Examples:
        rtmx setup              # Full setup with smart defaults
        rtmx setup --dry-run    # Preview what would happen
        rtmx setup --minimal    # Just config and RTM database
        rtmx setup --branch     # Create git branch for review workflow
        rtmx setup --pr         # Create branch and pull request
        rtmx setup --scaffold   # Generate spec files for all requirements
    """
    # Handle scaffold-only mode
    if scaffold:
        from pathlib import Path

        from rtmx.templates import run_scaffold

        scaffold_result = run_scaffold(
            project_path=Path.cwd(),
            force=force,
            dry_run=dry_run,
        )

        if not scaffold_result.success:
            import sys

            sys.exit(1)
        return

    from rtmx.cli.setup import run_setup

    setup_result = run_setup(
        dry_run=dry_run,
        minimal=minimal,
        force=force,
        skip_agents=skip_agents,
        skip_makefile=skip_makefile,
        branch="auto" if branch else None,
        create_pr=create_pr,
    )

    if not setup_result.success:
        import sys

        sys.exit(1)


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
@click.argument("test_path", required=False)
@click.option("--update", is_flag=True, help="Update RTM database with results")
@click.option("--dry-run", is_flag=True, help="Show changes without updating")
@click.option("-v", "--verbose", is_flag=True, help="Verbose output")
def verify(
    test_path: str | None,
    update: bool,
    dry_run: bool,
    verbose: bool,
) -> None:
    """Verify requirements by running tests and updating status.

    This is closed-loop verification: tests are run, and RTM status
    is automatically updated based on pass/fail results.

    \b
    Examples:
        rtmx verify                    # Run all tests, show results
        rtmx verify --update           # Run tests and update RTM
        rtmx verify tests/unit/ --update  # Verify specific tests
        rtmx verify --dry-run          # Show what would change
    """
    from rtmx.cli.verify import run_verify

    run_verify(
        test_path=test_path,
        update=update,
        dry_run=dry_run,
        verbose=verbose,
    )


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


# =============================================================================
# Remote Management Commands (Cross-Repo Dependencies)
# =============================================================================


@main.group()
def remote() -> None:
    """Manage cross-repository remote configurations.

    Remotes allow dependencies on requirements in other repositories.
    Use 'sync:REQ-001' format to reference requirements from the 'sync' remote.

    \b
    Examples:
        rtmx remote list                              # Show configured remotes
        rtmx remote add sync --repo rtmx-ai/rtmx-sync # Add remote
        rtmx remote add sync --repo rtmx-ai/rtmx-sync --path ../rtmx-sync
        rtmx remote remove sync                       # Remove remote
    """
    pass


@remote.command("list")
@click.pass_context
def remote_list(ctx: click.Context) -> None:
    """List configured remotes."""
    from rtmx.cli.remote import run_remote_list

    config: RTMXConfig = ctx.obj["config"]
    run_remote_list(config._config_path)


@remote.command("add")
@click.argument("alias")
@click.option(
    "--repo",
    required=True,
    help="Full repository path (e.g., 'rtmx-ai/rtmx-sync')",
)
@click.option(
    "--path",
    help="Local filesystem path for offline access",
)
@click.option(
    "--database",
    default=".rtmx/database.csv",
    help="Path to database within remote (default: .rtmx/database.csv)",
)
@click.pass_context
def remote_add(
    ctx: click.Context,
    alias: str,
    repo: str,
    path: str | None,
    database: str,
) -> None:
    """Add a new remote repository.

    ALIAS is the short name to use in references (e.g., 'sync' for 'sync:REQ-001').
    """
    from rtmx.cli.remote import run_remote_add

    config: RTMXConfig = ctx.obj["config"]
    run_remote_add(alias, repo, path, database, config._config_path)


@remote.command("remove")
@click.argument("alias")
@click.pass_context
def remote_remove(ctx: click.Context, alias: str) -> None:
    """Remove a remote repository.

    ALIAS is the name of the remote to remove.
    """
    from rtmx.cli.remote import run_remote_remove

    config: RTMXConfig = ctx.obj["config"]
    run_remote_remove(alias, config._config_path)


# =============================================================================
# Integration Commands (E2E Production Integration)
# =============================================================================


@main.command()
@click.option(
    "--format",
    "format_type",
    type=click.Choice(["terminal", "json", "ci"]),
    default="terminal",
    help="Output format",
)
@click.option(
    "--strict",
    is_flag=True,
    help="Treat warnings as errors",
)
@click.option(
    "--check",
    "checks",
    multiple=True,
    help="Specific checks to run (can be repeated)",
)
@click.pass_context
def health(
    ctx: click.Context,
    format_type: str,
    strict: bool,
    checks: tuple[str, ...],
) -> None:
    """Run integration health check.

    Single command to verify rtmx integration health for CI/CD pipelines.
    Exit codes: 0=healthy, 1=warnings (with --strict), 2=errors.
    """
    from rtmx.cli.health import run_health

    config: RTMXConfig = ctx.obj["config"]
    checks_list = list(checks) if checks else None
    run_health(format_type, strict, checks_list, config)


@main.command("diff")
@click.argument(
    "baseline",
    type=click.Path(exists=True, path_type=Path),
)
@click.argument(
    "current",
    required=False,
    type=click.Path(exists=True, path_type=Path),
)
@click.option(
    "--format",
    "format_type",
    type=click.Choice(["terminal", "markdown", "json"]),
    default="terminal",
    help="Output format",
)
@click.option(
    "--output",
    "-o",
    type=click.Path(path_type=Path),
    help="Output file",
)
@click.pass_context
def diff_cmd(
    _ctx: click.Context,
    baseline: Path,
    current: Path | None,
    format_type: str,
    output: Path | None,
) -> None:
    """Compare RTM databases before and after changes.

    Compares baseline with current (default: docs/rtm_database.csv).
    Exit codes: 0=stable/improved, 1=regressed/degraded, 2=breaking.
    """
    from rtmx.cli.diff import run_diff

    run_diff(baseline, current, format_type, output)


# =============================================================================
# Standalone Commands (formerly hidden)
# =============================================================================


@main.command()
@click.option(
    "--force",
    is_flag=True,
    help="Overwrite existing files",
)
@click.option(
    "--legacy",
    is_flag=True,
    help="Use legacy docs/ directory structure instead of .rtmx/",
)
def init(force: bool, legacy: bool) -> None:
    """Initialize RTM structure in current directory.

    Creates minimal RTM setup: config, database, and sample requirement.
    Use 'rtmx setup' for full integration including agents and Makefile.

    By default, uses the .rtmx/ directory structure. Use --legacy for the
    older docs/ structure.
    """
    from rtmx.cli.init import run_init

    run_init(force, use_rtmx_dir=not legacy)


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
    help="Merge with existing RTM (default: replace)",
)
@click.option(
    "--dry-run",
    is_flag=True,
    help="Preview without writing files",
)
@click.option(
    "--prefix",
    default="REQ",
    help="Requirement ID prefix (default: REQ)",
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

    Bootstrap requirements from existing tests, GitHub issues, or Jira tickets.

    \b
    Examples:
        rtmx bootstrap --from-tests        # Generate from test markers
        rtmx bootstrap --from-github       # Import from GitHub issues
        rtmx bootstrap --from-jira         # Import from Jira tickets
        rtmx bootstrap --from-tests --merge  # Merge with existing RTM
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
    "-y",
    "--yes",
    "non_interactive",
    is_flag=True,
    help="Skip confirmation prompts",
)
@click.option(
    "--force",
    is_flag=True,
    help="Overwrite existing RTMX sections",
)
@click.option(
    "--agents",
    multiple=True,
    type=click.Choice(["claude", "cursor", "copilot"]),
    help="Specific agents to install (can repeat)",
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
    help="Don't create backup files",
)
@click.option(
    "--hooks",
    is_flag=True,
    help="Install git hooks instead of agent configs",
)
@click.option(
    "--pre-push",
    is_flag=True,
    help="Also install pre-push hook (requires --hooks)",
)
@click.option(
    "--remove",
    is_flag=True,
    help="Remove installed hooks (requires --hooks)",
)
@click.option(
    "--validate",
    "validate_hook",
    is_flag=True,
    help="Install validation hook that checks staged RTM CSV files (requires --hooks)",
)
@click.option(
    "--claude",
    "claude_hooks",
    is_flag=True,
    help="Install Claude Code hooks for automatic context injection (requires --hooks)",
)
@click.pass_context
def install(
    ctx: click.Context,
    dry_run: bool,
    non_interactive: bool,
    force: bool,
    agents: tuple[str, ...],
    install_all: bool,
    skip_backup: bool,
    hooks: bool,
    pre_push: bool,
    remove: bool,
    validate_hook: bool,
    claude_hooks: bool,
) -> None:
    """Install RTM-aware prompts into AI agent configs or git hooks.

    Injects RTMX context and commands into Claude, Cursor, or Copilot configs.
    With --hooks, installs git hooks for automated validation.
    With --hooks --claude, installs Claude Code hooks for context injection.

    \b
    Examples:
        rtmx install                    # Interactive selection
        rtmx install --all              # Install to all detected agents
        rtmx install --agents claude    # Install only to Claude
        rtmx install --dry-run          # Preview changes
        rtmx install --hooks            # Install pre-commit hook (health check)
        rtmx install --hooks --validate # Install validation pre-commit hook
        rtmx install --hooks --pre-push # Install both hooks
        rtmx install --hooks --claude   # Install Claude Code hooks
        rtmx install --hooks --remove   # Remove rtmx hooks
    """
    if hooks:
        if claude_hooks:
            from rtmx.hooks import install_claude_hooks, uninstall_claude_hooks

            if remove:
                removed = uninstall_claude_hooks()
                if removed:
                    print(f"Removed {len(removed)} Claude Code hooks")
                    for path in removed:
                        print(f"  - {path}")
                else:
                    print("No RTMX Claude Code hooks found to remove")
            else:
                installed = install_claude_hooks(dry_run=dry_run, force=force)
                if installed:
                    action = "Would install" if dry_run else "Installed"
                    print(f"{action} {len(installed)} Claude Code hooks:")
                    for name, path in installed.items():
                        print(f"  - {name}: {path}")
                else:
                    print("No hooks installed (use --force to overwrite existing)")
            return

        from rtmx.cli.install import run_hooks

        run_hooks(dry_run=dry_run, pre_push=pre_push, remove=remove, validate=validate_hook)
    else:
        from rtmx.cli.install import run_install

        config: RTMXConfig = ctx.obj["config"]
        run_install(
            dry_run,
            non_interactive,
            force,
            list(agents) if agents else None,
            install_all,
            skip_backup,
            config,
        )


@main.command("validate-staged")
@click.argument("files", nargs=-1, type=click.Path(exists=True, path_type=Path))
@click.option("-v", "--verbose", is_flag=True, help="Show detailed output")
def validate_staged(files: tuple[Path, ...], verbose: bool) -> None:
    """Validate staged RTM CSV files (used by pre-commit hook).

    Validates only the specified CSV files. Designed to be called from
    a pre-commit hook to validate staged RTM database files.

    \b
    Examples:
        rtmx validate-staged docs/rtm_database.csv
        rtmx validate-staged *.csv
    """
    import sys

    from rtmx.cli.validate import run_validate_staged_cli

    exit_code = run_validate_staged_cli([str(f) for f in files], verbose)
    sys.exit(exit_code)


@main.command()
@click.option(
    "--output",
    "-o",
    type=click.Path(path_type=Path),
    help="Output file (default: stdout)",
)
def makefile(output: Path | None) -> None:
    """Generate Makefile targets for RTM commands.

    Outputs Makefile targets for common RTMX operations.

    \b
    Examples:
        rtmx makefile                   # Print to stdout
        rtmx makefile -o rtmx.mk        # Write to file
        rtmx makefile >> Makefile       # Append to Makefile
    """
    from rtmx.cli.makefile import run_makefile

    run_makefile(output)


@main.command("config")
@click.option(
    "--validate",
    "validate_config",
    is_flag=True,
    help="Validate configuration and check paths",
)
@click.option(
    "--format",
    "output_format",
    type=click.Choice(["terminal", "yaml", "json"]),
    default="terminal",
    help="Output format",
)
@click.pass_context
def config_cmd(
    ctx: click.Context,
    validate_config: bool,
    output_format: str,
) -> None:
    """Show or validate RTMX configuration.

    Displays the effective configuration after merging defaults with rtmx.yaml.

    \b
    Examples:
        rtmx config                     # Show current config
        rtmx config --validate          # Check config validity
        rtmx config --format yaml       # Output as YAML
    """
    from rtmx.cli.config_cmd import run_config

    config: RTMXConfig = ctx.obj["config"]
    run_config(config, validate_config, output_format)


# =============================================================================
# Documentation Generation Commands (REQ-DX-004)
# =============================================================================


@main.group()
def docs() -> None:
    """Generate documentation from RTMX internals.

    Auto-generate schema and configuration reference documentation.

    \b
    Examples:
        rtmx docs schema                # Generate schema.md
        rtmx docs config                # Generate config.md
        rtmx docs schema -o docs/       # Custom output location
    """
    pass


@docs.command("schema")
@click.option(
    "--output",
    "-o",
    type=click.Path(path_type=Path),
    help="Output file (default: .rtmx/cache/schema.md)",
)
def docs_schema(output: Path | None) -> None:
    """Generate schema documentation.

    Creates markdown documentation for all RTM schemas including
    core schema and Phoenix extension with column types, defaults,
    and descriptions.
    """
    from rtmx.cli.docs import run_docs_schema

    run_docs_schema(output)


@docs.command("config")
@click.option(
    "--output",
    "-o",
    type=click.Path(path_type=Path),
    help="Output file (default: .rtmx/cache/config.md)",
)
def docs_config(output: Path | None) -> None:
    """Generate configuration reference.

    Creates markdown documentation for RTMXConfig showing all
    configuration options with their defaults and descriptions.
    """
    from rtmx.cli.docs import run_docs_config

    run_docs_config(output)


# =============================================================================
# Authentication Commands (REQ-ZT-001)
# =============================================================================


@main.group()
def auth() -> None:
    """Manage RTMX authentication.

    Authenticate with RTMX sync services using Zitadel OIDC.

    \b
    Examples:
        rtmx auth login     # Login via browser
        rtmx auth logout    # Clear stored credentials
        rtmx auth status    # Show current auth status
    """
    pass


@auth.command("login")
@click.option(
    "--no-browser",
    is_flag=True,
    help="Print URL instead of opening browser",
)
def auth_login(no_browser: bool) -> None:
    """Login to RTMX sync services.

    Opens browser for Zitadel OIDC authentication.
    Uses PKCE flow for secure CLI authentication.
    """
    import asyncio

    from rtmx.auth import login
    from rtmx.formatting import Colors

    try:
        print(f"{Colors.CYAN}Starting authentication...{Colors.RESET}")
        tokens = asyncio.run(login(open_browser=not no_browser))
        print(f"{Colors.GREEN}✓ Authentication successful{Colors.RESET}")
        if tokens.id_token:
            print(f"{Colors.DIM}Token expires: {tokens.expires_at}{Colors.RESET}")
    except Exception as e:
        print(f"{Colors.RED}✗ Authentication failed: {e}{Colors.RESET}")
        raise SystemExit(1) from e


@auth.command("logout")
def auth_logout() -> None:
    """Logout and clear stored credentials.

    Removes all stored tokens from keychain/file storage.
    """
    from rtmx.auth import logout
    from rtmx.formatting import Colors

    logout()
    print(f"{Colors.GREEN}✓ Logged out successfully{Colors.RESET}")


@auth.command("status")
def auth_status() -> None:
    """Show current authentication status.

    Displays whether authenticated and token expiration.
    """
    from rtmx.auth import get_access_token, get_config, is_authenticated
    from rtmx.formatting import Colors

    config = get_config()

    print(f"{Colors.BOLD}Authentication Status{Colors.RESET}")
    print(f"Provider: {config.provider}")
    print(f"Issuer: {config.issuer}")
    print()

    if is_authenticated():
        token = get_access_token()
        if token:
            print(f"{Colors.GREEN}✓ Authenticated{Colors.RESET}")
            print(f"{Colors.DIM}Token: {token[:20]}...{Colors.RESET}")
        else:
            print(f"{Colors.YELLOW}⚠ Token expired, refresh required{Colors.RESET}")
    else:
        print(f"{Colors.RED}✗ Not authenticated{Colors.RESET}")
        print(f"{Colors.DIM}Run 'rtmx auth login' to authenticate{Colors.RESET}")


# =============================================================================
# Marker Discovery Commands (REQ-LANG-007)
# =============================================================================

# Register the markers command group (imported at top of file)
main.add_command(markers_group, name="markers")


# =============================================================================
# BDD Commands (REQ-BDD-001)
# =============================================================================


@main.command("parse-feature")
@click.argument(
    "path",
    type=click.Path(exists=True, path_type=Path),
)
@click.option(
    "--json",
    "output_json",
    is_flag=True,
    help="Output as JSON",
)
@click.option(
    "--format",
    "output_format",
    type=click.Choice(["text", "json"]),
    default="text",
    help="Output format",
)
@click.option(
    "--pattern",
    "-p",
    default="**/*.feature",
    help="Glob pattern for directory scanning (default: **/*.feature)",
)
@click.option(
    "--expand-outlines",
    is_flag=True,
    help="Expand Scenario Outlines into individual scenarios",
)
def parse_feature(
    path: Path,
    output_json: bool,
    output_format: str,
    pattern: str,
    expand_outlines: bool,
) -> None:
    """Parse Gherkin feature files and extract requirements.

    Parses .feature files and displays their structure including
    Features, Scenarios, Steps, and @REQ-XXX-NNN requirement tags.

    \b
    Examples:
        rtmx parse-feature test.feature              # Parse single file
        rtmx parse-feature features/                 # Scan directory recursively
        rtmx parse-feature features/ -p "*.feature"  # Non-recursive scan
        rtmx parse-feature test.feature --json       # JSON output
        rtmx parse-feature test.feature --expand-outlines  # Expand outlines
    """
    from rtmx.cli.parse_feature import run_parse_feature

    # Support both --json flag and --format json
    use_json = output_json or output_format == "json"

    exit_code = run_parse_feature(
        str(path),
        output_json=use_json,
        pattern=pattern,
        expand_outlines=expand_outlines,
    )
    if exit_code != 0:
        raise SystemExit(exit_code)


@main.command("discover-steps")
@click.argument(
    "path",
    type=click.Path(exists=True, path_type=Path),
)
@click.option(
    "--json",
    "output_json",
    is_flag=True,
    help="Output as JSON",
)
@click.option(
    "--pattern",
    "-p",
    default="**/*.py",
    help="Glob pattern for Python files (default: **/*.py)",
)
def discover_steps(
    path: Path,
    output_json: bool,
    pattern: str,
) -> None:
    """Discover step definitions in Python files.

    Scans Python files for pytest-bdd and behave step definitions
    (@given, @when, @then decorators) and reports them.

    \b
    Examples:
        rtmx discover-steps tests/steps/           # Scan directory
        rtmx discover-steps tests/ -p "conftest.py"  # Specific file pattern
        rtmx discover-steps tests/ --json          # JSON output
    """
    from rtmx.cli.discover_steps import run_discover_steps

    exit_code = run_discover_steps(
        str(path),
        output_json=output_json,
        pattern=pattern,
    )
    if exit_code != 0:
        raise SystemExit(exit_code)


# =============================================================================
# Claude Code Integration (REQ-CLAUDE-001)
# =============================================================================


@main.command()
@click.option(
    "--format",
    "format_type",
    type=click.Choice(["text", "json", "markdown"]),
    default="text",
    help="Output format",
)
@click.option(
    "--compact",
    is_flag=True,
    help="Minimal token-efficient output",
)
@click.option(
    "--phase",
    type=int,
    help="Filter to specific phase",
)
@click.option(
    "--verbose",
    "-v",
    is_flag=True,
    help="Include full requirement descriptions",
)
def context(
    format_type: str,
    compact: bool,
    phase: int | None,
    verbose: bool,
) -> None:
    """Generate RTM context for AI assistants.

    Produces token-efficient requirements context for use in AI coding sessions.
    Used by Claude Code hooks for automatic context injection.

    \b
    Examples:
        rtmx context                    # Text summary
        rtmx context --format json      # JSON for hooks
        rtmx context --compact          # Minimal output
        rtmx context --phase 10         # Filter to phase
    """
    from rtmx.cli.context import run_context

    exit_code = run_context(
        format_type=format_type,
        compact=compact,
        phase=phase,
        verbose=verbose,
    )
    if exit_code != 0:
        raise SystemExit(exit_code)


# =============================================================================
# Migration Command (REQ-DIST-002)
# =============================================================================


@main.command()
@click.option(
    "--verify-only",
    is_flag=True,
    help="Just verify Go CLI installation, don't install",
)
@click.option(
    "--install-dir",
    type=click.Path(path_type=Path),
    help="Custom installation directory",
)
@click.option(
    "--alias",
    is_flag=True,
    help="Show shell alias instructions",
)
def migrate(
    verify_only: bool,
    install_dir: Path | None,
    alias: bool,
) -> None:
    """Migrate from Python CLI to Go CLI.

    Downloads and installs the Go CLI binary for your platform.
    The Go CLI is faster, has no runtime dependencies, and will
    become the primary implementation.

    \b
    Examples:
        rtmx migrate                    # Install Go CLI
        rtmx migrate --verify-only      # Check installation
        rtmx migrate --install-dir ~/bin  # Custom location
        rtmx migrate --alias            # Show alias instructions
    """
    from rtmx.cli.migrate import run_migrate

    exit_code = run_migrate(
        verify_only=verify_only,
        install_dir=install_dir,
        alias=alias,
    )
    if exit_code != 0:
        raise SystemExit(exit_code)


if __name__ == "__main__":
    main()
