"""RTMX CLI remote management commands.

Commands for managing cross-repository remote configurations.
"""

from __future__ import annotations

from pathlib import Path

import click

from rtmx.config import RemoteConfig, load_config, save_config
from rtmx.formatting import Colors


def run_remote_list(config_path: Path | None = None) -> None:
    """List configured remotes.

    Args:
        config_path: Optional path to config file
    """
    config = load_config(config_path)

    if not config.sync.remotes:
        click.echo("No remotes configured.")
        click.echo("")
        click.echo("Add a remote with:")
        click.echo("  rtmx remote add <alias> --repo <org/repo> [--path <local-path>]")
        return

    click.echo(f"Configured remotes ({len(config.sync.remotes)}):")
    click.echo("")

    for alias, remote in sorted(config.sync.remotes.items()):
        status = ""
        if remote.path:
            remote_path = Path(remote.path)
            db_path = remote_path / remote.database
            if db_path.exists():
                status = f" {Colors.GREEN}✓ available{Colors.RESET}"
            else:
                status = f" {Colors.YELLOW}✗ not found{Colors.RESET}"
        else:
            status = f" {Colors.DIM}(no local path){Colors.RESET}"

        click.echo(f"  {Colors.BOLD}{alias}{Colors.RESET}: {remote.repo}{status}")
        if remote.path:
            click.echo(f"    path: {remote.path}")
        click.echo(f"    database: {remote.database}")


def run_remote_add(
    alias: str,
    repo: str,
    path: str | None = None,
    database: str = ".rtmx/database.csv",
    config_path: Path | None = None,
) -> None:
    """Add a new remote configuration.

    Args:
        alias: Short alias for the remote (e.g., 'sync')
        repo: Full repository path (e.g., 'sync-server')
        path: Optional local filesystem path
        database: Path to database within remote
        config_path: Optional path to config file
    """
    config = load_config(config_path)

    if alias in config.sync.remotes:
        click.echo(f"Remote '{alias}' already exists. Use 'rtmx remote remove' first.")
        raise SystemExit(1)

    remote = RemoteConfig(
        alias=alias,
        repo=repo,
        path=path,
        database=database,
    )

    config.sync.remotes[alias] = remote

    # Determine save path
    save_path = config._config_path
    if save_path is None:
        # No existing config - create in .rtmx/config.yaml
        save_path = Path.cwd() / ".rtmx" / "config.yaml"

    save_config(config, save_path)

    click.echo(f"Added remote '{alias}' -> {repo}")
    if path:
        click.echo(f"  local path: {path}")


def run_remote_remove(alias: str, config_path: Path | None = None) -> None:
    """Remove a remote configuration.

    Args:
        alias: Alias of the remote to remove
        config_path: Optional path to config file
    """
    config = load_config(config_path)

    if alias not in config.sync.remotes:
        click.echo(f"Remote '{alias}' not found.")
        raise SystemExit(1)

    del config.sync.remotes[alias]

    # Determine save path
    save_path = config._config_path
    if save_path is None:
        save_path = Path.cwd() / ".rtmx" / "config.yaml"

    save_config(config, save_path)

    click.echo(f"Removed remote '{alias}'")
