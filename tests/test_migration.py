"""Tests for auto-migration to .rtmx/ directory (REQ-DX-002).

This module tests the automatic migration from legacy layout to .rtmx/ on first run.
"""

from __future__ import annotations

import os
from datetime import datetime
from pathlib import Path
from unittest.mock import patch

import pytest


@pytest.mark.req("REQ-DX-002")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMigrationDetection:
    """Test detection of legacy layout."""

    def test_detect_legacy_layout_with_rtmx_yaml(self, tmp_path: Path) -> None:
        """Should detect legacy layout when rtmx.yaml exists in root."""
        from rtmx.migration import detect_legacy_layout

        # Create legacy config
        legacy_config = tmp_path / "rtmx.yaml"
        legacy_config.write_text("rtmx:\n  database: docs/rtm_database.csv\n")

        result = detect_legacy_layout(tmp_path)
        assert result.has_legacy_config is True
        assert result.legacy_config_path == legacy_config

    def test_detect_legacy_layout_with_docs_database(self, tmp_path: Path) -> None:
        """Should detect legacy layout when docs/rtm_database.csv exists."""
        from rtmx.migration import detect_legacy_layout

        # Create legacy database
        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()
        legacy_db = docs_dir / "rtm_database.csv"
        legacy_db.write_text("req_id,category\nREQ-001,TEST\n")

        result = detect_legacy_layout(tmp_path)
        assert result.has_legacy_database is True
        assert result.legacy_database_path == legacy_db

    def test_detect_legacy_layout_with_docs_requirements(self, tmp_path: Path) -> None:
        """Should detect legacy layout when docs/requirements/ exists."""
        from rtmx.migration import detect_legacy_layout

        # Create legacy requirements directory
        req_dir = tmp_path / "docs" / "requirements"
        req_dir.mkdir(parents=True)
        (req_dir / "sample.md").write_text("# Sample")

        result = detect_legacy_layout(tmp_path)
        assert result.has_legacy_requirements is True
        assert result.legacy_requirements_path == tmp_path / "docs" / "requirements"

    def test_detect_legacy_layout_returns_none_for_new_layout(self, tmp_path: Path) -> None:
        """Should return no legacy elements when .rtmx/ structure exists."""
        from rtmx.migration import detect_legacy_layout

        # Create new structure
        rtmx_dir = tmp_path / ".rtmx"
        rtmx_dir.mkdir()
        (rtmx_dir / "config.yaml").write_text("rtmx:\n  schema: core\n")
        (rtmx_dir / "database.csv").write_text("req_id,category\nREQ-001,TEST\n")

        result = detect_legacy_layout(tmp_path)
        assert result.has_legacy_config is False
        assert result.has_legacy_database is False
        assert result.has_legacy_requirements is False

    def test_detect_no_rtmx_layout(self, tmp_path: Path) -> None:
        """Should detect no RTMX layout when neither exists."""
        from rtmx.migration import detect_legacy_layout

        result = detect_legacy_layout(tmp_path)
        assert result.has_legacy_config is False
        assert result.has_legacy_database is False
        assert result.has_legacy_requirements is False
        assert not result.needs_migration

    def test_needs_migration_property(self, tmp_path: Path) -> None:
        """Should indicate migration needed when legacy elements exist."""
        from rtmx.migration import detect_legacy_layout

        # Create legacy config
        legacy_config = tmp_path / "rtmx.yaml"
        legacy_config.write_text("rtmx:\n  database: docs/rtm_database.csv\n")

        result = detect_legacy_layout(tmp_path)
        assert result.needs_migration is True


@pytest.mark.req("REQ-DX-002")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMigrationBackup:
    """Test backup creation before migration."""

    def test_creates_backup_of_config(self, tmp_path: Path) -> None:
        """Should create backup of rtmx.yaml before migration."""
        from rtmx.migration import create_backup

        # Create legacy config
        legacy_config = tmp_path / "rtmx.yaml"
        legacy_config.write_text("rtmx:\n  database: docs/rtm_database.csv\n")

        backup_path = create_backup(legacy_config)

        assert backup_path is not None
        assert backup_path.exists()
        assert ".rtmx-backup-" in backup_path.name
        assert backup_path.read_text() == legacy_config.read_text()

    def test_creates_backup_of_directory(self, tmp_path: Path) -> None:
        """Should create backup of docs/requirements/ directory."""
        from rtmx.migration import create_backup

        # Create legacy requirements directory
        req_dir = tmp_path / "docs" / "requirements"
        req_dir.mkdir(parents=True)
        (req_dir / "TEST" / "REQ-001.md").parent.mkdir()
        (req_dir / "TEST" / "REQ-001.md").write_text("# REQ-001")

        backup_path = create_backup(req_dir)

        assert backup_path is not None
        assert backup_path.is_dir()
        assert ".rtmx-backup-" in backup_path.name
        assert (backup_path / "TEST" / "REQ-001.md").exists()

    def test_backup_uses_timestamp(self, tmp_path: Path) -> None:
        """Backup should include timestamp in name."""
        from rtmx.migration import create_backup

        legacy_config = tmp_path / "rtmx.yaml"
        legacy_config.write_text("rtmx:\n  database: docs/rtm_database.csv\n")

        backup_path = create_backup(legacy_config)

        # Check timestamp format YYYYMMDD_HHMMSS
        assert backup_path is not None
        backup_suffix = backup_path.name.replace("rtmx.yaml.rtmx-backup-", "")
        # Should be parseable as datetime
        datetime.strptime(backup_suffix, "%Y%m%d_%H%M%S")


@pytest.mark.req("REQ-DX-002")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMigrationExecution:
    """Test migration execution."""

    def test_moves_config_to_rtmx_dir(self, tmp_path: Path) -> None:
        """Should move rtmx.yaml to .rtmx/config.yaml."""
        from rtmx.migration import migrate_layout

        # Create legacy config
        legacy_config = tmp_path / "rtmx.yaml"
        legacy_config.write_text("rtmx:\n  database: docs/rtm_database.csv\n")

        result = migrate_layout(tmp_path, confirm=True)

        assert result.success is True
        assert (tmp_path / ".rtmx" / "config.yaml").exists()
        # Original should be gone (backed up)
        assert not legacy_config.exists()

    def test_moves_database_to_rtmx_dir(self, tmp_path: Path) -> None:
        """Should move docs/rtm_database.csv to .rtmx/database.csv."""
        from rtmx.migration import migrate_layout

        # Create legacy database
        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()
        legacy_db = docs_dir / "rtm_database.csv"
        legacy_db.write_text("req_id,category\nREQ-001,TEST\n")

        result = migrate_layout(tmp_path, confirm=True)

        assert result.success is True
        assert (tmp_path / ".rtmx" / "database.csv").exists()
        assert not legacy_db.exists()

    def test_moves_requirements_to_rtmx_dir(self, tmp_path: Path) -> None:
        """Should move docs/requirements/ to .rtmx/requirements/."""
        from rtmx.migration import migrate_layout

        # Create legacy requirements directory
        req_dir = tmp_path / "docs" / "requirements" / "TEST"
        req_dir.mkdir(parents=True)
        (req_dir / "REQ-001.md").write_text("# REQ-001")

        result = migrate_layout(tmp_path, confirm=True)

        assert result.success is True
        assert (tmp_path / ".rtmx" / "requirements" / "TEST" / "REQ-001.md").exists()
        assert not (tmp_path / "docs" / "requirements").exists()

    def test_updates_config_paths(self, tmp_path: Path) -> None:
        """Should update path references in config file."""
        from rtmx.migration import migrate_layout

        # Create legacy config with old paths
        legacy_config = tmp_path / "rtmx.yaml"
        legacy_config.write_text(
            "rtmx:\n  database: docs/rtm_database.csv\n  requirements_dir: docs/requirements\n"
        )

        result = migrate_layout(tmp_path, confirm=True)

        assert result.success is True
        new_config = (tmp_path / ".rtmx" / "config.yaml").read_text()
        assert ".rtmx/database.csv" in new_config
        assert ".rtmx/requirements" in new_config

    def test_creates_cache_directory(self, tmp_path: Path) -> None:
        """Should create .rtmx/cache/ directory during migration."""
        from rtmx.migration import migrate_layout

        # Create legacy config
        legacy_config = tmp_path / "rtmx.yaml"
        legacy_config.write_text("rtmx:\n  database: docs/rtm_database.csv\n")

        result = migrate_layout(tmp_path, confirm=True)

        assert result.success is True
        assert (tmp_path / ".rtmx" / "cache").is_dir()

    def test_creates_gitignore(self, tmp_path: Path) -> None:
        """Should create .rtmx/.gitignore during migration."""
        from rtmx.migration import migrate_layout

        # Create legacy config
        legacy_config = tmp_path / "rtmx.yaml"
        legacy_config.write_text("rtmx:\n  database: docs/rtm_database.csv\n")

        result = migrate_layout(tmp_path, confirm=True)

        assert result.success is True
        gitignore = tmp_path / ".rtmx" / ".gitignore"
        assert gitignore.exists()
        assert "cache/" in gitignore.read_text()

    def test_no_confirm_returns_preview(self, tmp_path: Path) -> None:
        """Should return preview without making changes when confirm=False."""
        from rtmx.migration import migrate_layout

        # Create legacy config
        legacy_config = tmp_path / "rtmx.yaml"
        legacy_config.write_text("rtmx:\n  database: docs/rtm_database.csv\n")

        result = migrate_layout(tmp_path, confirm=False)

        # Should not have migrated
        assert legacy_config.exists()
        assert not (tmp_path / ".rtmx" / "config.yaml").exists()
        # But should have migration plan
        assert result.planned_actions is not None
        assert len(result.planned_actions) > 0


@pytest.mark.req("REQ-DX-002")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMigrationSummary:
    """Test migration summary display."""

    def test_summary_includes_moved_files(self, tmp_path: Path) -> None:
        """Migration summary should list all moved files."""
        from rtmx.migration import migrate_layout

        # Create full legacy layout
        (tmp_path / "rtmx.yaml").write_text("rtmx:\n  database: docs/rtm_database.csv\n")
        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()
        (docs_dir / "rtm_database.csv").write_text("req_id,category\nREQ-001,TEST\n")
        req_dir = docs_dir / "requirements"
        req_dir.mkdir()

        result = migrate_layout(tmp_path, confirm=True)

        assert result.summary is not None
        assert "rtmx.yaml" in result.summary
        assert "database.csv" in result.summary

    def test_summary_includes_backup_locations(self, tmp_path: Path) -> None:
        """Migration summary should include backup file paths."""
        from rtmx.migration import migrate_layout

        # Create legacy config
        (tmp_path / "rtmx.yaml").write_text("rtmx:\n  database: docs/rtm_database.csv\n")

        result = migrate_layout(tmp_path, confirm=True)

        assert result.backup_paths is not None
        assert len(result.backup_paths) > 0


@pytest.mark.req("REQ-DX-002")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMigrationSuppression:
    """Test --no-migrate flag and RTMX_NO_MIGRATE env var."""

    def test_no_migrate_flag_suppresses_migration(self, tmp_path: Path) -> None:
        """--no-migrate flag should suppress migration."""
        from rtmx.migration import should_migrate

        # Create legacy config
        (tmp_path / "rtmx.yaml").write_text("rtmx:\n  database: docs/rtm_database.csv\n")

        result = should_migrate(tmp_path, no_migrate=True)
        assert result is False

    def test_env_var_suppresses_migration(self, tmp_path: Path) -> None:
        """RTMX_NO_MIGRATE env var should suppress migration."""
        from rtmx.migration import should_migrate

        # Create legacy config
        (tmp_path / "rtmx.yaml").write_text("rtmx:\n  database: docs/rtm_database.csv\n")

        with patch.dict(os.environ, {"RTMX_NO_MIGRATE": "1"}):
            result = should_migrate(tmp_path, no_migrate=False)
            assert result is False

    def test_env_var_accepts_various_truthy_values(self, tmp_path: Path) -> None:
        """RTMX_NO_MIGRATE should accept 1, true, yes."""
        from rtmx.migration import should_migrate

        # Create legacy config
        (tmp_path / "rtmx.yaml").write_text("rtmx:\n  database: docs/rtm_database.csv\n")

        for value in ["1", "true", "True", "TRUE", "yes", "Yes", "YES"]:
            with patch.dict(os.environ, {"RTMX_NO_MIGRATE": value}):
                result = should_migrate(tmp_path, no_migrate=False)
                assert result is False, f"Failed for value: {value}"


@pytest.mark.req("REQ-DX-002")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMigrationIdempotency:
    """Test migration is idempotent - running again doesn't break anything."""

    def test_migration_is_idempotent(self, tmp_path: Path) -> None:
        """Running migration twice should not break anything."""
        from rtmx.migration import migrate_layout

        # Create legacy config
        (tmp_path / "rtmx.yaml").write_text("rtmx:\n  database: docs/rtm_database.csv\n")

        # First migration
        result1 = migrate_layout(tmp_path, confirm=True)
        assert result1.success is True

        # Second migration should be a no-op
        result2 = migrate_layout(tmp_path, confirm=True)
        assert result2.success is True
        assert result2.already_migrated is True

    def test_already_migrated_returns_quickly(self, tmp_path: Path) -> None:
        """If already migrated, should return immediately with no actions."""
        from rtmx.migration import migrate_layout

        # Create new structure directly
        rtmx_dir = tmp_path / ".rtmx"
        rtmx_dir.mkdir()
        (rtmx_dir / "config.yaml").write_text("rtmx:\n  schema: core\n")

        result = migrate_layout(tmp_path, confirm=True)

        assert result.already_migrated is True
        assert result.planned_actions is None or len(result.planned_actions) == 0


@pytest.mark.req("REQ-DX-002")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMigrationCLIIntegration:
    """Test CLI integration of migration."""

    def test_migration_prompt_function_works(self, tmp_path: Path) -> None:
        """Migration prompt function should detect legacy layout and return True on confirm."""
        from rtmx.migration import prompt_for_migration

        # Create legacy config
        (tmp_path / "rtmx.yaml").write_text("rtmx:\n  database: docs/rtm_database.csv\n")

        # Mock click.confirm to return True
        with patch("click.confirm", return_value=True):
            result = prompt_for_migration(tmp_path)
            assert result is True

    def test_migration_prompt_returns_false_on_decline(self, tmp_path: Path) -> None:
        """Migration prompt should return False when user declines."""
        from rtmx.migration import prompt_for_migration

        # Create legacy config
        (tmp_path / "rtmx.yaml").write_text("rtmx:\n  database: docs/rtm_database.csv\n")

        # Mock click.confirm to return False
        with patch("click.confirm", return_value=False):
            result = prompt_for_migration(tmp_path)
            assert result is False

    def test_cli_no_migrate_flag_skips_prompt(self, tmp_path: Path) -> None:
        """--no-migrate should skip migration prompt entirely."""
        from click.testing import CliRunner

        from rtmx.cli.main import main

        # Create legacy config
        (tmp_path / "rtmx.yaml").write_text("rtmx:\n  database: docs/rtm_database.csv\n")
        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()
        (docs_dir / "rtm_database.csv").write_text(
            "req_id,category,subcategory,requirement_text,target_value,test_module,"
            "test_function,validation_method,status,priority,phase,notes,effort_weeks,"
            "dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file\n"
            "REQ-001,TEST,SUB,Test req,Target,test.py,test_func,Unit Test,MISSING,HIGH,1,"
            "Notes,1.0,,,dev,v1,,,\n"
        )

        runner = CliRunner()
        os.chdir(tmp_path)

        runner.invoke(main, ["--no-migrate", "status"])

        # Should not have prompted about migration
        # and legacy files should still exist
        assert (tmp_path / "rtmx.yaml").exists()

    def test_run_migration_if_needed_executes(self, tmp_path: Path) -> None:
        """run_migration_if_needed should perform migration when not suppressed."""
        from rtmx.migration import run_migration_if_needed

        # Create legacy config
        (tmp_path / "rtmx.yaml").write_text("rtmx:\n  database: docs/rtm_database.csv\n")
        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()
        (docs_dir / "rtm_database.csv").write_text(
            "req_id,category,subcategory,requirement_text,target_value,test_module,"
            "test_function,validation_method,status,priority,phase,notes,effort_weeks,"
            "dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file\n"
            "REQ-001,TEST,SUB,Test req,Target,test.py,test_func,Unit Test,MISSING,HIGH,1,"
            "Notes,1.0,,,dev,v1,,,\n"
        )

        # Run with interactive=False to skip prompting
        result = run_migration_if_needed(
            root_path=tmp_path,
            no_migrate=False,
            interactive=False,
        )

        # Migration should have occurred
        assert result is not None
        assert result.success is True
        assert (tmp_path / ".rtmx" / "config.yaml").exists()
        assert not (tmp_path / "rtmx.yaml").exists()


@pytest.mark.req("REQ-DX-002")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMigrationEdgeCases:
    """Test edge cases in migration."""

    def test_handles_partial_legacy_layout(self, tmp_path: Path) -> None:
        """Should handle case where only some legacy files exist."""
        from rtmx.migration import migrate_layout

        # Only create config, no database or requirements
        (tmp_path / "rtmx.yaml").write_text("rtmx:\n  database: docs/rtm_database.csv\n")

        result = migrate_layout(tmp_path, confirm=True)

        assert result.success is True
        assert (tmp_path / ".rtmx" / "config.yaml").exists()

    def test_handles_empty_requirements_directory(self, tmp_path: Path) -> None:
        """Should handle empty requirements directory."""
        from rtmx.migration import migrate_layout

        # Create legacy structure with empty requirements
        (tmp_path / "rtmx.yaml").write_text("rtmx:\n  database: docs/rtm_database.csv\n")
        req_dir = tmp_path / "docs" / "requirements"
        req_dir.mkdir(parents=True)

        result = migrate_layout(tmp_path, confirm=True)

        assert result.success is True
        assert (tmp_path / ".rtmx" / "requirements").is_dir()

    def test_handles_nested_requirements(self, tmp_path: Path) -> None:
        """Should preserve nested directory structure in requirements."""
        from rtmx.migration import migrate_layout

        # Create nested requirements structure
        (tmp_path / "rtmx.yaml").write_text("rtmx:\n  database: docs/rtm_database.csv\n")
        req_path = tmp_path / "docs" / "requirements" / "FEATURE" / "AUTH"
        req_path.mkdir(parents=True)
        (req_path / "REQ-AUTH-001.md").write_text("# Authentication")

        result = migrate_layout(tmp_path, confirm=True)

        assert result.success is True
        assert (
            tmp_path / ".rtmx" / "requirements" / "FEATURE" / "AUTH" / "REQ-AUTH-001.md"
        ).exists()

    def test_handles_rtmx_yml_extension(self, tmp_path: Path) -> None:
        """Should detect rtmx.yml as legacy config."""
        from rtmx.migration import detect_legacy_layout

        # Create legacy config with .yml extension
        (tmp_path / "rtmx.yml").write_text("rtmx:\n  database: docs/rtm_database.csv\n")

        result = detect_legacy_layout(tmp_path)

        assert result.has_legacy_config is True
        assert result.legacy_config_path == tmp_path / "rtmx.yml"

    def test_preserves_extra_config_fields(self, tmp_path: Path) -> None:
        """Should preserve extra configuration fields during migration."""
        from rtmx.migration import migrate_layout

        # Create legacy config with extra fields
        (tmp_path / "rtmx.yaml").write_text(
            "rtmx:\n"
            "  database: docs/rtm_database.csv\n"
            "  schema: phoenix\n"
            "  adapters:\n"
            "    github:\n"
            "      enabled: true\n"
        )

        result = migrate_layout(tmp_path, confirm=True)

        assert result.success is True
        new_config = (tmp_path / ".rtmx" / "config.yaml").read_text()
        assert "phoenix" in new_config
        assert "github" in new_config

    def test_handles_existing_rtmx_dir_with_different_content(self, tmp_path: Path) -> None:
        """Should handle case where .rtmx/ exists but has different content."""
        from rtmx.migration import migrate_layout

        # Create both old and new structures
        (tmp_path / "rtmx.yaml").write_text("rtmx:\n  database: docs/rtm_database.csv\n")
        rtmx_dir = tmp_path / ".rtmx"
        rtmx_dir.mkdir()
        # But no config.yaml inside

        result = migrate_layout(tmp_path, confirm=True)

        assert result.success is True
        assert (tmp_path / ".rtmx" / "config.yaml").exists()
