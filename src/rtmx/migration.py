"""Auto-migration from legacy layout to .rtmx/ directory.

This module handles automatic migration from the legacy layout (rtmx.yaml in root,
docs/rtm_database.csv) to the new .rtmx/ directory structure on first run.
"""

from __future__ import annotations

import os
import shutil
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path

import yaml

from rtmx.config import CONFIG_FILE_NAME, LEGACY_CONFIG_NAMES, RTMX_DIR_NAME

# Database file locations
LEGACY_DATABASE_PATH = Path("docs/rtm_database.csv")
NEW_DATABASE_NAME = "database.csv"
LEGACY_REQUIREMENTS_PATH = Path("docs/requirements")
NEW_REQUIREMENTS_DIR = "requirements"


@dataclass
class LegacyLayoutResult:
    """Result of detecting legacy layout."""

    has_legacy_config: bool = False
    legacy_config_path: Path | None = None
    has_legacy_database: bool = False
    legacy_database_path: Path | None = None
    has_legacy_requirements: bool = False
    legacy_requirements_path: Path | None = None

    @property
    def needs_migration(self) -> bool:
        """Check if any legacy elements need migration."""
        return self.has_legacy_config or self.has_legacy_database or self.has_legacy_requirements


@dataclass
class MigrationAction:
    """Describes a single migration action."""

    action_type: str  # "move", "create", "update"
    source: Path | None
    destination: Path
    description: str


@dataclass
class MigrationResult:
    """Result of migration execution."""

    success: bool = False
    already_migrated: bool = False
    planned_actions: list[MigrationAction] = field(default_factory=list)
    completed_actions: list[MigrationAction] = field(default_factory=list)
    backup_paths: list[Path] = field(default_factory=list)
    summary: str | None = None
    error: str | None = None


def detect_legacy_layout(root_path: Path) -> LegacyLayoutResult:
    """Detect if legacy RTMX layout exists.

    Checks for:
    1. rtmx.yaml or rtmx.yml in root (legacy config)
    2. docs/rtm_database.csv (legacy database)
    3. docs/requirements/ (legacy requirements directory)

    Does NOT consider it legacy if .rtmx/config.yaml exists.

    Args:
        root_path: Project root directory

    Returns:
        LegacyLayoutResult with detection findings
    """
    result = LegacyLayoutResult()

    # Check if already migrated
    new_config = root_path / RTMX_DIR_NAME / CONFIG_FILE_NAME
    if new_config.exists():
        return result  # Already using new structure

    # Check for legacy config files
    for config_name in LEGACY_CONFIG_NAMES:
        legacy_config = root_path / config_name
        if legacy_config.exists():
            result.has_legacy_config = True
            result.legacy_config_path = legacy_config
            break

    # Check for legacy database
    legacy_db = root_path / LEGACY_DATABASE_PATH
    if legacy_db.exists():
        result.has_legacy_database = True
        result.legacy_database_path = legacy_db

    # Check for legacy requirements directory
    legacy_reqs = root_path / LEGACY_REQUIREMENTS_PATH
    if legacy_reqs.exists() and legacy_reqs.is_dir():
        result.has_legacy_requirements = True
        result.legacy_requirements_path = legacy_reqs

    return result


def should_migrate(root_path: Path, no_migrate: bool = False) -> bool:
    """Check if migration should proceed.

    Args:
        root_path: Project root directory
        no_migrate: If True, suppress migration (--no-migrate flag)

    Returns:
        True if migration should proceed
    """
    # Check flag first
    if no_migrate:
        return False

    # Check environment variable
    env_no_migrate = os.environ.get("RTMX_NO_MIGRATE", "").lower()
    if env_no_migrate in ("1", "true", "yes"):
        return False

    # Check if migration is needed
    layout = detect_legacy_layout(root_path)
    return layout.needs_migration


def create_backup(source_path: Path) -> Path | None:
    """Create a backup of a file or directory.

    Creates backup with timestamp suffix: .rtmx-backup-YYYYMMDD_HHMMSS

    Args:
        source_path: Path to file or directory to backup

    Returns:
        Path to backup, or None if source doesn't exist
    """
    if not source_path.exists():
        return None

    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    backup_name = f"{source_path.name}.rtmx-backup-{timestamp}"
    backup_path = source_path.parent / backup_name

    if source_path.is_dir():
        shutil.copytree(source_path, backup_path)
    else:
        shutil.copy2(source_path, backup_path)

    return backup_path


def _update_config_paths(config_content: str) -> str:
    """Update path references in config content.

    Converts:
    - docs/rtm_database.csv -> .rtmx/database.csv
    - docs/requirements -> .rtmx/requirements

    Args:
        config_content: Original config YAML content

    Returns:
        Updated config content
    """
    try:
        config_data = yaml.safe_load(config_content) or {}
    except yaml.YAMLError:
        return config_content

    rtmx_data = config_data.get("rtmx", config_data)

    # Update database path
    if "database" in rtmx_data:
        db_path = rtmx_data["database"]
        if db_path == "docs/rtm_database.csv":
            rtmx_data["database"] = ".rtmx/database.csv"

    # Update requirements_dir path
    if "requirements_dir" in rtmx_data:
        req_dir = rtmx_data["requirements_dir"]
        if req_dir == "docs/requirements":
            rtmx_data["requirements_dir"] = ".rtmx/requirements"

    # Preserve structure
    if "rtmx" not in config_data:
        config_data = {"rtmx": rtmx_data}
    else:
        config_data["rtmx"] = rtmx_data

    return yaml.dump(config_data, default_flow_style=False, sort_keys=False)


def _plan_migration(root_path: Path, layout: LegacyLayoutResult) -> list[MigrationAction]:
    """Plan migration actions without executing them.

    Args:
        root_path: Project root directory
        layout: Detected legacy layout

    Returns:
        List of planned migration actions
    """
    actions: list[MigrationAction] = []
    rtmx_dir = root_path / RTMX_DIR_NAME

    # Create .rtmx directory
    if not rtmx_dir.exists():
        actions.append(
            MigrationAction(
                action_type="create",
                source=None,
                destination=rtmx_dir,
                description=f"Create {RTMX_DIR_NAME}/ directory",
            )
        )

    # Create cache directory
    cache_dir = rtmx_dir / "cache"
    if not cache_dir.exists():
        actions.append(
            MigrationAction(
                action_type="create",
                source=None,
                destination=cache_dir,
                description=f"Create {RTMX_DIR_NAME}/cache/ directory",
            )
        )

    # Create .gitignore
    gitignore = rtmx_dir / ".gitignore"
    if not gitignore.exists():
        actions.append(
            MigrationAction(
                action_type="create",
                source=None,
                destination=gitignore,
                description=f"Create {RTMX_DIR_NAME}/.gitignore",
            )
        )

    # Move config
    if layout.has_legacy_config and layout.legacy_config_path:
        new_config = rtmx_dir / CONFIG_FILE_NAME
        if not new_config.exists():
            actions.append(
                MigrationAction(
                    action_type="move",
                    source=layout.legacy_config_path,
                    destination=new_config,
                    description=f"Move {layout.legacy_config_path.name} -> {RTMX_DIR_NAME}/{CONFIG_FILE_NAME}",
                )
            )

    # Move database
    if layout.has_legacy_database and layout.legacy_database_path:
        new_db = rtmx_dir / NEW_DATABASE_NAME
        if not new_db.exists():
            actions.append(
                MigrationAction(
                    action_type="move",
                    source=layout.legacy_database_path,
                    destination=new_db,
                    description=f"Move docs/rtm_database.csv -> {RTMX_DIR_NAME}/{NEW_DATABASE_NAME}",
                )
            )

    # Move requirements directory
    if layout.has_legacy_requirements and layout.legacy_requirements_path:
        new_reqs = rtmx_dir / NEW_REQUIREMENTS_DIR
        if not new_reqs.exists():
            actions.append(
                MigrationAction(
                    action_type="move",
                    source=layout.legacy_requirements_path,
                    destination=new_reqs,
                    description=f"Move docs/requirements/ -> {RTMX_DIR_NAME}/{NEW_REQUIREMENTS_DIR}/",
                )
            )

    return actions


def _execute_migration(
    actions: list[MigrationAction],
) -> tuple[list[MigrationAction], list[Path]]:
    """Execute planned migration actions.

    Args:
        actions: List of planned actions

    Returns:
        Tuple of (completed_actions, backup_paths)
    """
    completed: list[MigrationAction] = []
    backups: list[Path] = []

    for action in actions:
        if action.action_type == "create":
            if action.destination.name == ".gitignore":
                # .gitignore is a file, not a directory
                action.destination.parent.mkdir(parents=True, exist_ok=True)
                action.destination.write_text("# RTMX cache and generated files\ncache/\n")
            else:
                # It's a directory
                action.destination.mkdir(parents=True, exist_ok=True)
            completed.append(action)

        elif action.action_type == "move" and action.source:
            # Create backup first
            backup = create_backup(action.source)
            if backup:
                backups.append(backup)

            # Move the file/directory
            action.destination.parent.mkdir(parents=True, exist_ok=True)

            if action.source.is_dir():
                shutil.copytree(action.source, action.destination)
                shutil.rmtree(action.source)
            else:
                # For config file, update paths before moving
                if action.source.name in LEGACY_CONFIG_NAMES:
                    content = action.source.read_text()
                    updated_content = _update_config_paths(content)
                    action.destination.write_text(updated_content)
                    action.source.unlink()
                else:
                    shutil.move(str(action.source), str(action.destination))

            completed.append(action)

    return completed, backups


def _generate_summary(completed_actions: list[MigrationAction], backup_paths: list[Path]) -> str:
    """Generate human-readable migration summary.

    Args:
        completed_actions: List of completed actions
        backup_paths: List of backup file paths

    Returns:
        Summary string
    """
    lines = ["Migration Summary:", ""]

    if completed_actions:
        lines.append("Completed actions:")
        for action in completed_actions:
            lines.append(f"  - {action.description}")

    if backup_paths:
        lines.append("")
        lines.append("Backup files created:")
        for path in backup_paths:
            lines.append(f"  - {path}")

    return "\n".join(lines)


def migrate_layout(root_path: Path, confirm: bool = False) -> MigrationResult:
    """Migrate from legacy layout to .rtmx/ directory.

    Args:
        root_path: Project root directory
        confirm: If True, execute migration. If False, return preview only.

    Returns:
        MigrationResult with details
    """
    result = MigrationResult()

    # Check if already migrated
    new_config = root_path / RTMX_DIR_NAME / CONFIG_FILE_NAME
    if new_config.exists():
        result.success = True
        result.already_migrated = True
        result.planned_actions = []
        return result

    # Detect legacy layout
    layout = detect_legacy_layout(root_path)

    if not layout.needs_migration:
        result.success = True
        result.already_migrated = True
        result.planned_actions = []
        return result

    # Plan migration
    actions = _plan_migration(root_path, layout)
    result.planned_actions = actions

    if not confirm:
        return result

    # Execute migration
    try:
        completed, backups = _execute_migration(actions)
        result.completed_actions = completed
        result.backup_paths = backups
        result.summary = _generate_summary(completed, backups)
        result.success = True
    except Exception as e:
        result.error = str(e)
        result.success = False

    return result


def prompt_for_migration(root_path: Path) -> bool:
    """Display migration prompt and get user confirmation.

    This function is meant to be called from the CLI layer.

    Args:
        root_path: Project root directory

    Returns:
        True if user confirms migration
    """
    import click

    layout = detect_legacy_layout(root_path)

    if not layout.needs_migration:
        return False

    click.echo()
    click.echo(click.style("Legacy RTMX layout detected", fg="yellow", bold=True))
    click.echo()
    click.echo("Found legacy files that can be migrated to .rtmx/ directory:")

    if layout.has_legacy_config:
        click.echo(f"  - {layout.legacy_config_path}")
    if layout.has_legacy_database:
        click.echo(f"  - {layout.legacy_database_path}")
    if layout.has_legacy_requirements:
        click.echo(f"  - {layout.legacy_requirements_path}/")

    click.echo()
    click.echo("Migration will:")
    click.echo("  1. Create backups of all files")
    click.echo("  2. Move files to .rtmx/ directory")
    click.echo("  3. Update path references in config")
    click.echo()

    return click.confirm("Proceed with migration?", default=True)


def run_migration_if_needed(
    root_path: Path | None = None,
    no_migrate: bool = False,
    interactive: bool = True,
) -> MigrationResult | None:
    """Check for legacy layout and migrate if confirmed.

    This is the main entry point for migration, called from CLI.

    Args:
        root_path: Project root directory (defaults to cwd)
        no_migrate: If True, skip migration entirely
        interactive: If True, prompt for confirmation

    Returns:
        MigrationResult if migration was attempted, None otherwise
    """
    if root_path is None:
        root_path = Path.cwd()

    if not should_migrate(root_path, no_migrate):
        return None

    confirmed = prompt_for_migration(root_path) if interactive else True

    if not confirmed:
        return None

    result = migrate_layout(root_path, confirm=True)

    if result.success and result.summary and not result.already_migrated:
        import click

        click.echo()
        click.echo(click.style("Migration completed successfully!", fg="green"))
        click.echo()
        click.echo(result.summary)
        click.echo()

    return result
