"""RTMX CLI main entry point.

This module provides the main CLI command group and common options.
"""

from __future__ import annotations

import sys
from pathlib import Path

import click

from rtmx import __version__
from rtmx.formatting import Colors


@click.group()
@click.version_option(version=__version__, prog_name="rtmx")
@click.option(
    "--rtm-csv",
    type=click.Path(exists=False, path_type=Path),
    help="Path to RTM database CSV",
)
@click.option(
    "--no-color",
    is_flag=True,
    help="Disable colored output",
)
@click.pass_context
def main(ctx: click.Context, rtm_csv: Path | None, no_color: bool) -> None:
    """RTMX - Requirements Traceability Matrix toolkit.

    Manage requirements traceability for GenAI-driven development.
    """
    ctx.ensure_object(dict)
    ctx.obj["rtm_csv"] = rtm_csv
    ctx.obj["no_color"] = no_color

    if no_color or not sys.stdout.isatty():
        Colors.disable()


@main.command()
@click.option(
    "-v", "--verbose",
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
def init(ctx: click.Context, force: bool) -> None:
    """Initialize RTM in current project.

    Creates docs/rtm_database.csv and docs/requirements/ structure.
    """
    from rtmx.cli.init import run_init

    run_init(force)


if __name__ == "__main__":
    main()
