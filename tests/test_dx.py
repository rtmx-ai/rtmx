"""Tests for developer experience requirements (REQ-DX-*).

This module tests the .rtmx/ directory structure and related DX features.
"""

from __future__ import annotations

import os
from pathlib import Path
from unittest.mock import patch

import pytest

from rtmx.config import find_config_file, load_config
from rtmx.parser import find_rtm_database


@pytest.mark.req("REQ-DX-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRtmxDirectoryStructure:
    """Test .rtmx/ directory structure discovery."""

    def test_config_discovery_prefers_rtmx_dir(self, tmp_path: Path) -> None:
        """Config discovery should prefer .rtmx/config.yaml over rtmx.yaml."""
        # Create both config locations
        old_config = tmp_path / "rtmx.yaml"
        old_config.write_text("rtmx:\n  database: old/path.csv\n")

        rtmx_dir = tmp_path / ".rtmx"
        rtmx_dir.mkdir()
        new_config = rtmx_dir / "config.yaml"
        new_config.write_text("rtmx:\n  database: .rtmx/database.csv\n")

        # Should find .rtmx/config.yaml first
        result = find_config_file(tmp_path)
        assert result == new_config

    def test_config_discovery_falls_back_to_rtmx_yaml(self, tmp_path: Path) -> None:
        """Config discovery should fall back to rtmx.yaml if .rtmx/ doesn't exist."""
        old_config = tmp_path / "rtmx.yaml"
        old_config.write_text("rtmx:\n  database: docs/rtm_database.csv\n")

        result = find_config_file(tmp_path)
        assert result == old_config

    def test_database_discovery_prefers_rtmx_dir(self, tmp_path: Path) -> None:
        """Database discovery should prefer .rtmx/database.csv over docs/rtm_database.csv."""
        # Create both database locations with valid content
        old_db = tmp_path / "docs" / "rtm_database.csv"
        old_db.parent.mkdir(parents=True)
        old_db.write_text(
            "req_id,category,subcategory,requirement_text,target_value,test_module,test_function,"
            "validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,"
            "assignee,sprint,started_date,completed_date,requirement_file\n"
            "REQ-OLD-001,TEST,OLD,Old requirement,Target,tests/test.py,test_old,Unit Test,"
            "MISSING,HIGH,1,Notes,1.0,,,,,,,\n"
        )

        rtmx_dir = tmp_path / ".rtmx"
        rtmx_dir.mkdir()
        new_db = rtmx_dir / "database.csv"
        new_db.write_text(
            "req_id,category,subcategory,requirement_text,target_value,test_module,test_function,"
            "validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,"
            "assignee,sprint,started_date,completed_date,requirement_file\n"
            "REQ-NEW-001,TEST,NEW,New requirement,Target,tests/test.py,test_new,Unit Test,"
            "MISSING,HIGH,1,Notes,1.0,,,,,,,\n"
        )

        result = find_rtm_database(tmp_path)
        assert result == new_db

    def test_database_discovery_falls_back_to_docs(self, tmp_path: Path) -> None:
        """Database discovery should fall back to docs/rtm_database.csv if .rtmx/ doesn't exist."""
        old_db = tmp_path / "docs" / "rtm_database.csv"
        old_db.parent.mkdir(parents=True)
        old_db.write_text(
            "req_id,category,subcategory,requirement_text,target_value,test_module,test_function,"
            "validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,"
            "assignee,sprint,started_date,completed_date,requirement_file\n"
            "REQ-OLD-001,TEST,OLD,Old requirement,Target,tests/test.py,test_old,Unit Test,"
            "MISSING,HIGH,1,Notes,1.0,,,,,,,\n"
        )

        result = find_rtm_database(tmp_path)
        assert result == old_db

    def test_config_loads_from_rtmx_dir(self, tmp_path: Path) -> None:
        """load_config should work with .rtmx/config.yaml."""
        rtmx_dir = tmp_path / ".rtmx"
        rtmx_dir.mkdir()
        config_path = rtmx_dir / "config.yaml"
        config_path.write_text(
            "rtmx:\n"
            "  database: .rtmx/database.csv\n"
            "  requirements_dir: .rtmx/requirements\n"
            "  schema: core\n"
        )

        with patch("rtmx.config.find_config_file", return_value=config_path):
            config = load_config()
            assert config.database == Path(".rtmx/database.csv")
            assert config.requirements_dir == Path(".rtmx/requirements")


@pytest.mark.req("REQ-DX-001")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRtmxInitDirectory:
    """Test init command creates .rtmx/ directory structure."""

    def test_init_creates_rtmx_directory(self, tmp_path: Path) -> None:
        """rtmx init should create .rtmx/ directory structure."""
        from rtmx.cli.init import run_init

        os.chdir(tmp_path)
        run_init(force=True, use_rtmx_dir=True)

        # Check directory structure
        assert (tmp_path / ".rtmx").is_dir()
        assert (tmp_path / ".rtmx" / "config.yaml").is_file()
        assert (tmp_path / ".rtmx" / "database.csv").is_file()
        assert (tmp_path / ".rtmx" / "requirements").is_dir()
        assert (tmp_path / ".rtmx" / ".gitignore").is_file()

    def test_init_creates_cache_directory(self, tmp_path: Path) -> None:
        """rtmx init should create cache directory."""
        from rtmx.cli.init import run_init

        os.chdir(tmp_path)
        run_init(force=True, use_rtmx_dir=True)

        assert (tmp_path / ".rtmx" / "cache").is_dir()

    def test_init_gitignore_ignores_cache(self, tmp_path: Path) -> None:
        """.rtmx/.gitignore should ignore cache/ directory."""
        from rtmx.cli.init import run_init

        os.chdir(tmp_path)
        run_init(force=True, use_rtmx_dir=True)

        gitignore = (tmp_path / ".rtmx" / ".gitignore").read_text()
        assert "cache/" in gitignore

    def test_init_legacy_mode_creates_docs_structure(self, tmp_path: Path) -> None:
        """rtmx init without --use-rtmx-dir should create legacy docs/ structure."""
        from rtmx.cli.init import run_init

        os.chdir(tmp_path)
        run_init(force=True, use_rtmx_dir=False)

        # Check legacy structure
        assert (tmp_path / "docs" / "rtm_database.csv").is_file()
        assert (tmp_path / "rtmx.yaml").is_file()
        assert not (tmp_path / ".rtmx").exists()


@pytest.mark.req("REQ-DX-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRtmxDirectoryConstants:
    """Test directory constants and helpers."""

    def test_rtmx_dir_name_constant(self) -> None:
        """RTMX_DIR_NAME should be .rtmx."""
        from rtmx.config import RTMX_DIR_NAME

        assert RTMX_DIR_NAME == ".rtmx"

    def test_config_file_name_constant(self) -> None:
        """CONFIG_FILE_NAME should be config.yaml."""
        from rtmx.config import CONFIG_FILE_NAME

        assert CONFIG_FILE_NAME == "config.yaml"

    def test_database_file_name_constant(self) -> None:
        """DATABASE_FILE_NAME should be database.csv."""
        from rtmx.parser import DATABASE_FILE_NAME

        assert DATABASE_FILE_NAME == "database.csv"
