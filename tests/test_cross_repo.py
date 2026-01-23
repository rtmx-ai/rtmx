"""Tests for cross-repository dependencies MVP (Phase 1).

This module tests the local cross-repo dependency functionality:
- RemoteConfig and SyncConfig with remotes
- parse_requirement_ref() for parsing cross-repo references
- validate_cross_repo_deps() for cross-repo dependency validation
- is_blocked() with cross-repo dependencies
- CLI remote commands
"""

from pathlib import Path

import pytest

from rtmx.config import (
    RemoteConfig,
    RTMXConfig,
    SyncConfig,
    load_config,
    save_config,
)
from rtmx.models import RTMDatabase
from rtmx.parser import (
    RequirementRef,
    parse_requirement_ref,
)
from rtmx.validation import validate_cross_repo_deps

# =============================================================================
# RemoteConfig Tests
# =============================================================================


@pytest.mark.req("REQ-COLLAB-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRemoteConfig:
    """Tests for RemoteConfig dataclass."""

    def test_remote_config_defaults(self):
        """Test RemoteConfig has sensible defaults."""
        config = RemoteConfig(alias="sync", repo="sync-server")

        assert config.alias == "sync"
        assert config.repo == "sync-server"
        assert config.path is None
        assert config.database == ".rtmx/database.csv"

    def test_remote_config_with_path(self):
        """Test RemoteConfig with local path."""
        config = RemoteConfig(
            alias="sync",
            repo="sync-server",
            path="../rtmx-sync",
            database="docs/rtm_database.csv",
        )

        assert config.alias == "sync"
        assert config.repo == "sync-server"
        assert config.path == "../rtmx-sync"
        assert config.database == "docs/rtm_database.csv"

    def test_remote_config_from_dict(self):
        """Test RemoteConfig.from_dict creates config from dictionary."""
        data = {
            "alias": "upstream",
            "repo": "owner/repo",
            "path": "/path/to/repo",
            "database": "custom.csv",
        }
        config = RemoteConfig.from_dict(data)

        assert config.alias == "upstream"
        assert config.repo == "owner/repo"
        assert config.path == "/path/to/repo"
        assert config.database == "custom.csv"

    def test_remote_config_from_dict_minimal(self):
        """Test RemoteConfig.from_dict with minimal data uses defaults."""
        data = {"alias": "sync", "repo": "sync-server"}
        config = RemoteConfig.from_dict(data)

        assert config.alias == "sync"
        assert config.repo == "sync-server"
        assert config.path is None
        assert config.database == ".rtmx/database.csv"


# =============================================================================
# SyncConfig with Remotes Tests
# =============================================================================


@pytest.mark.req("REQ-COLLAB-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestSyncConfigRemotes:
    """Tests for SyncConfig with remotes support."""

    def test_sync_config_remotes_default_empty(self):
        """Test SyncConfig.remotes defaults to empty dict."""
        config = SyncConfig()

        assert config.remotes == {}

    def test_sync_config_with_remotes(self):
        """Test SyncConfig with configured remotes."""
        remote = RemoteConfig(alias="sync", repo="sync-server")
        config = SyncConfig(remotes={"sync": remote})

        assert "sync" in config.remotes
        assert config.remotes["sync"].repo == "sync-server"

    def test_sync_config_from_dict_with_remotes(self):
        """Test SyncConfig.from_dict loads remotes configuration."""
        data = {
            "conflict_resolution": "prefer-local",
            "remotes": {
                "sync": {
                    "repo": "sync-server",
                    "path": "../rtmx-sync",
                },
                "upstream": {
                    "repo": "company/requirements",
                },
            },
        }
        config = SyncConfig.from_dict(data)

        assert config.conflict_resolution == "prefer-local"
        assert len(config.remotes) == 2
        assert config.remotes["sync"].repo == "sync-server"
        assert config.remotes["sync"].path == "../rtmx-sync"
        assert config.remotes["upstream"].repo == "company/requirements"

    def test_sync_config_get_remote(self):
        """Test SyncConfig.get_remote returns configured remote."""
        remote = RemoteConfig(alias="sync", repo="sync-server")
        config = SyncConfig(remotes={"sync": remote})

        result = config.get_remote("sync")
        assert result is not None
        assert result.repo == "sync-server"

    def test_sync_config_get_remote_not_found(self):
        """Test SyncConfig.get_remote returns None for unknown alias."""
        config = SyncConfig()

        result = config.get_remote("unknown")
        assert result is None


# =============================================================================
# RTMXConfig Integration Tests
# =============================================================================


@pytest.mark.req("REQ-COLLAB-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRTMXConfigRemotes:
    """Tests for RTMXConfig integration with remotes."""

    def test_rtmx_config_from_dict_with_remotes(self):
        """Test RTMXConfig.from_dict loads sync.remotes configuration."""
        data = {
            "rtmx": {
                "sync": {
                    "conflict_resolution": "manual",
                    "remotes": {
                        "sync": {
                            "repo": "sync-server",
                            "path": "../rtmx-sync",
                            "database": ".rtmx/database.csv",
                        }
                    },
                }
            }
        }
        config = RTMXConfig.from_dict(data)

        assert "sync" in config.sync.remotes
        assert config.sync.remotes["sync"].repo == "sync-server"
        assert config.sync.remotes["sync"].path == "../rtmx-sync"

    def test_rtmx_config_to_dict_includes_remotes(self):
        """Test RTMXConfig.to_dict serializes remotes."""
        remote = RemoteConfig(
            alias="sync",
            repo="sync-server",
            path="../rtmx-sync",
        )
        config = RTMXConfig()
        config.sync.remotes = {"sync": remote}

        result = config.to_dict()

        assert "remotes" in result["rtmx"]["sync"]
        assert "sync" in result["rtmx"]["sync"]["remotes"]

    def test_remotes_config_roundtrip(self, tmp_path):
        """Test remotes can be saved and loaded without loss."""
        original = RTMXConfig()
        original.sync.remotes = {
            "sync": RemoteConfig(
                alias="sync",
                repo="sync-server",
                path="../rtmx-sync",
            ),
            "upstream": RemoteConfig(
                alias="upstream",
                repo="company/reqs",
            ),
        }

        save_path = tmp_path / "config.yaml"
        save_config(original, save_path)
        loaded = load_config(save_path)

        assert len(loaded.sync.remotes) == 2
        assert loaded.sync.remotes["sync"].repo == "sync-server"
        assert loaded.sync.remotes["sync"].path == "../rtmx-sync"
        assert loaded.sync.remotes["upstream"].repo == "company/reqs"


# =============================================================================
# RequirementRef and parse_requirement_ref Tests
# =============================================================================


@pytest.mark.req("REQ-COLLAB-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestRequirementRef:
    """Tests for RequirementRef dataclass."""

    def test_requirement_ref_local(self):
        """Test RequirementRef for local requirement."""
        ref = RequirementRef(req_id="REQ-SW-001")

        assert ref.req_id == "REQ-SW-001"
        assert ref.remote_alias is None
        assert ref.full_repo is None
        assert ref.is_local is True
        assert ref.is_cross_repo is False

    def test_requirement_ref_with_alias(self):
        """Test RequirementRef with remote alias."""
        ref = RequirementRef(req_id="REQ-SYNC-001", remote_alias="sync")

        assert ref.req_id == "REQ-SYNC-001"
        assert ref.remote_alias == "sync"
        assert ref.full_repo is None
        assert ref.is_local is False
        assert ref.is_cross_repo is True

    def test_requirement_ref_with_full_repo(self):
        """Test RequirementRef with full repository path."""
        ref = RequirementRef(req_id="REQ-SYNC-001", full_repo="sync-server")

        assert ref.req_id == "REQ-SYNC-001"
        assert ref.remote_alias is None
        assert ref.full_repo == "sync-server"
        assert ref.is_local is False
        assert ref.is_cross_repo is True

    def test_requirement_ref_str_local(self):
        """Test RequirementRef string representation for local."""
        ref = RequirementRef(req_id="REQ-SW-001")
        assert str(ref) == "REQ-SW-001"

    def test_requirement_ref_str_with_alias(self):
        """Test RequirementRef string representation with alias."""
        ref = RequirementRef(req_id="REQ-SYNC-001", remote_alias="sync")
        assert str(ref) == "sync:REQ-SYNC-001"

    def test_requirement_ref_str_with_full_repo(self):
        """Test RequirementRef string representation with full repo."""
        ref = RequirementRef(req_id="REQ-SYNC-001", full_repo="sync-server")
        assert str(ref) == "sync-server:REQ-SYNC-001"


@pytest.mark.req("REQ-COLLAB-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestParseRequirementRef:
    """Tests for parse_requirement_ref function."""

    def test_parse_local_requirement(self):
        """Test parsing local requirement reference."""
        ref = parse_requirement_ref("REQ-SW-001")

        assert ref.req_id == "REQ-SW-001"
        assert ref.is_local is True

    def test_parse_aliased_requirement(self):
        """Test parsing aliased remote requirement reference."""
        ref = parse_requirement_ref("sync:REQ-SYNC-001")

        assert ref.req_id == "REQ-SYNC-001"
        assert ref.remote_alias == "sync"
        assert ref.is_cross_repo is True

    def test_parse_full_repo_requirement(self):
        """Test parsing full repository requirement reference."""
        ref = parse_requirement_ref("sync-server:REQ-SYNC-001")

        assert ref.req_id == "REQ-SYNC-001"
        assert ref.full_repo == "sync-server"
        assert ref.is_cross_repo is True

    def test_parse_complex_req_id(self):
        """Test parsing requirement with complex ID."""
        ref = parse_requirement_ref("upstream:REQ-COLLAB-007-A")

        assert ref.req_id == "REQ-COLLAB-007-A"
        assert ref.remote_alias == "upstream"

    def test_parse_preserves_case(self):
        """Test parsing preserves case in requirement ID."""
        ref = parse_requirement_ref("sync:req-lower-001")

        assert ref.req_id == "req-lower-001"

    def test_parse_whitespace_trimmed(self):
        """Test parsing trims whitespace."""
        ref = parse_requirement_ref("  sync:REQ-001  ")

        assert ref.req_id == "REQ-001"
        assert ref.remote_alias == "sync"

    def test_parse_empty_string_raises(self):
        """Test parsing empty string raises ValueError."""
        with pytest.raises(ValueError, match="empty"):
            parse_requirement_ref("")

    def test_parse_invalid_format_raises(self):
        """Test parsing invalid format raises ValueError."""
        with pytest.raises(ValueError, match="Invalid"):
            parse_requirement_ref("too:many:colons:here")


# =============================================================================
# Cross-Repo Validation Tests
# =============================================================================


@pytest.mark.req("REQ-COLLAB-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestValidateCrossRepoDeps:
    """Tests for validate_cross_repo_deps function."""

    def test_validate_no_cross_repo_deps(self, core_rtm_path: Path):
        """Test validation passes with no cross-repo dependencies."""
        db = RTMDatabase.load(core_rtm_path)
        config = RTMXConfig()

        errors, warnings = validate_cross_repo_deps(db, config)

        assert errors == []
        # May have warnings about missing remotes config

    def test_validate_cross_repo_dep_remote_available(self, tmp_path):
        """Test validation with available remote repository."""
        # Create local database
        local_db_path = tmp_path / "local" / "database.csv"
        local_db_path.parent.mkdir(parents=True)

        # Create remote database
        remote_path = tmp_path / "remote"
        remote_db_path = remote_path / ".rtmx" / "database.csv"
        remote_db_path.parent.mkdir(parents=True)

        # Write remote database with REQ-SYNC-001
        remote_db_path.write_text(
            "req_id,category,subcategory,requirement_text,status,priority,phase\n"
            "REQ-SYNC-001,SYNC,Core,Sync requirement,COMPLETE,HIGH,1\n"
        )

        # Write local database with cross-repo dependency
        local_db_path.write_text(
            "req_id,category,subcategory,requirement_text,status,priority,phase,dependencies\n"
            "REQ-LOCAL-001,LOCAL,Core,Local requirement,MISSING,HIGH,1,sync:REQ-SYNC-001\n"
        )

        db = RTMDatabase.load(local_db_path)

        # Configure remote
        config = RTMXConfig()
        config.sync.remotes = {
            "sync": RemoteConfig(
                alias="sync",
                repo="sync-server",
                path=str(remote_path),
            )
        }

        errors, warnings = validate_cross_repo_deps(db, config)

        assert errors == []
        assert warnings == []

    def test_validate_cross_repo_dep_remote_unavailable(self, tmp_path):
        """Test validation with unavailable remote repository (graceful warning)."""
        # Create local database with cross-repo dependency
        local_db_path = tmp_path / "database.csv"
        local_db_path.write_text(
            "req_id,category,subcategory,requirement_text,status,priority,phase,dependencies\n"
            "REQ-LOCAL-001,LOCAL,Core,Local requirement,MISSING,HIGH,1,sync:REQ-SYNC-001\n"
        )

        db = RTMDatabase.load(local_db_path)

        # Configure remote pointing to non-existent path
        config = RTMXConfig()
        config.sync.remotes = {
            "sync": RemoteConfig(
                alias="sync",
                repo="sync-server",
                path="/nonexistent/path",
            )
        }

        errors, warnings = validate_cross_repo_deps(db, config)

        # Should warn, not error (graceful degradation)
        assert errors == []
        assert len(warnings) > 0
        assert any("unavailable" in w.lower() or "not found" in w.lower() for w in warnings)

    def test_validate_cross_repo_dep_unknown_alias(self, tmp_path):
        """Test validation with unknown remote alias."""
        # Create local database with unknown alias
        local_db_path = tmp_path / "database.csv"
        local_db_path.write_text(
            "req_id,category,subcategory,requirement_text,status,priority,phase,dependencies\n"
            "REQ-LOCAL-001,LOCAL,Core,Local requirement,MISSING,HIGH,1,unknown:REQ-001\n"
        )

        db = RTMDatabase.load(local_db_path)
        config = RTMXConfig()  # No remotes configured

        errors, warnings = validate_cross_repo_deps(db, config)

        # Unknown alias should be an error
        assert len(errors) > 0
        assert any("unknown" in e.lower() for e in errors)

    def test_validate_cross_repo_dep_missing_in_remote(self, tmp_path):
        """Test validation when referenced requirement doesn't exist in remote."""
        # Create remote database (empty of requested requirement)
        remote_path = tmp_path / "remote"
        remote_db_path = remote_path / ".rtmx" / "database.csv"
        remote_db_path.parent.mkdir(parents=True)
        remote_db_path.write_text(
            "req_id,category,subcategory,requirement_text,status,priority,phase\n"
            "REQ-OTHER-001,OTHER,Core,Other requirement,COMPLETE,HIGH,1\n"
        )

        # Create local database
        local_db_path = tmp_path / "local" / "database.csv"
        local_db_path.parent.mkdir(parents=True)
        local_db_path.write_text(
            "req_id,category,subcategory,requirement_text,status,priority,phase,dependencies\n"
            "REQ-LOCAL-001,LOCAL,Core,Local requirement,MISSING,HIGH,1,sync:REQ-NONEXISTENT\n"
        )

        db = RTMDatabase.load(local_db_path)

        config = RTMXConfig()
        config.sync.remotes = {
            "sync": RemoteConfig(
                alias="sync",
                repo="sync-server",
                path=str(remote_path),
            )
        }

        errors, warnings = validate_cross_repo_deps(db, config)

        assert len(errors) > 0
        assert any("REQ-NONEXISTENT" in e for e in errors)


# =============================================================================
# is_blocked() Cross-Repo Tests
# =============================================================================


@pytest.mark.req("REQ-COLLAB-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestIsBlockedCrossRepo:
    """Tests for Requirement.is_blocked with cross-repo dependencies."""

    def test_is_blocked_local_only(self, core_rtm_path: Path):
        """Test is_blocked with local dependencies only."""
        db = RTMDatabase.load(core_rtm_path)
        req = db.get("REQ-SW-002")  # Has local dependency

        # Should work without cross-repo context
        result = req.is_blocked(db)
        assert isinstance(result, bool)

    def test_is_blocked_cross_repo_complete(self, tmp_path):
        """Test is_blocked when cross-repo dependency is complete."""
        # Create remote database with complete requirement
        remote_path = tmp_path / "remote"
        remote_db_path = remote_path / ".rtmx" / "database.csv"
        remote_db_path.parent.mkdir(parents=True)
        remote_db_path.write_text(
            "req_id,category,subcategory,requirement_text,status,priority,phase\n"
            "REQ-SYNC-001,SYNC,Core,Sync requirement,COMPLETE,HIGH,1\n"
        )

        # Create local database
        local_db_path = tmp_path / "local" / "database.csv"
        local_db_path.parent.mkdir(parents=True)
        local_db_path.write_text(
            "req_id,category,subcategory,requirement_text,status,priority,phase,dependencies\n"
            "REQ-LOCAL-001,LOCAL,Core,Local requirement,MISSING,HIGH,1,sync:REQ-SYNC-001\n"
        )

        db = RTMDatabase.load(local_db_path)
        req = db.get("REQ-LOCAL-001")

        # Configure cross-repo context
        config = RTMXConfig()
        config.sync.remotes = {
            "sync": RemoteConfig(
                alias="sync",
                repo="sync-server",
                path=str(remote_path),
            )
        }

        # is_blocked with cross-repo context
        result = req.is_blocked(db, config)
        assert result is False  # Not blocked because SYNC-001 is COMPLETE

    def test_is_blocked_cross_repo_incomplete(self, tmp_path):
        """Test is_blocked when cross-repo dependency is incomplete."""
        # Create remote database with incomplete requirement
        remote_path = tmp_path / "remote"
        remote_db_path = remote_path / ".rtmx" / "database.csv"
        remote_db_path.parent.mkdir(parents=True)
        remote_db_path.write_text(
            "req_id,category,subcategory,requirement_text,status,priority,phase\n"
            "REQ-SYNC-001,SYNC,Core,Sync requirement,MISSING,HIGH,1\n"
        )

        # Create local database
        local_db_path = tmp_path / "local" / "database.csv"
        local_db_path.parent.mkdir(parents=True)
        local_db_path.write_text(
            "req_id,category,subcategory,requirement_text,status,priority,phase,dependencies\n"
            "REQ-LOCAL-001,LOCAL,Core,Local requirement,MISSING,HIGH,1,sync:REQ-SYNC-001\n"
        )

        db = RTMDatabase.load(local_db_path)
        req = db.get("REQ-LOCAL-001")

        config = RTMXConfig()
        config.sync.remotes = {
            "sync": RemoteConfig(
                alias="sync",
                repo="sync-server",
                path=str(remote_path),
            )
        }

        result = req.is_blocked(db, config)
        assert result is True  # Blocked because SYNC-001 is MISSING

    def test_is_blocked_cross_repo_unavailable_graceful(self, tmp_path):
        """Test is_blocked gracefully handles unavailable remote."""
        # Create local database with cross-repo dependency
        local_db_path = tmp_path / "database.csv"
        local_db_path.write_text(
            "req_id,category,subcategory,requirement_text,status,priority,phase,dependencies\n"
            "REQ-LOCAL-001,LOCAL,Core,Local requirement,MISSING,HIGH,1,sync:REQ-SYNC-001\n"
        )

        db = RTMDatabase.load(local_db_path)
        req = db.get("REQ-LOCAL-001")

        config = RTMXConfig()
        config.sync.remotes = {
            "sync": RemoteConfig(
                alias="sync",
                repo="sync-server",
                path="/nonexistent",  # Unavailable
            )
        }

        # Should not raise, assume not blocked when can't verify
        result = req.is_blocked(db, config)
        # When remote unavailable, we can't verify - default to not blocked
        # (user gets warning from validation)
        assert isinstance(result, bool)


# =============================================================================
# parse_dependencies Cross-Repo Tests
# =============================================================================


@pytest.mark.req("REQ-COLLAB-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestParseDependenciesCrossRepo:
    """Tests for parse_dependencies with cross-repo references."""

    def test_parse_mixed_dependencies(self):
        """Test parsing mixed local and cross-repo dependencies."""
        from rtmx.parser import parse_dependencies

        deps = parse_dependencies("REQ-LOCAL-001|sync:REQ-SYNC-001|upstream:REQ-UP-002")

        assert "REQ-LOCAL-001" in deps
        assert "sync:REQ-SYNC-001" in deps
        assert "upstream:REQ-UP-002" in deps

    def test_parse_full_repo_dependencies(self):
        """Test parsing full repository path dependencies."""
        from rtmx.parser import parse_dependencies

        deps = parse_dependencies("sync-server:REQ-SYNC-001")

        assert "sync-server:REQ-SYNC-001" in deps


# =============================================================================
# CLI Remote Command Tests
# =============================================================================


@pytest.mark.req("REQ-COLLAB-001")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCLIRemoteCommands:
    """Tests for rtmx remote CLI commands."""

    def test_remote_list_empty(self, tmp_path, monkeypatch):
        """Test 'rtmx remote list' with no remotes configured."""
        from click.testing import CliRunner

        from rtmx.cli.main import main

        # Create minimal config
        config_path = tmp_path / ".rtmx" / "config.yaml"
        config_path.parent.mkdir(parents=True)
        config_path.write_text("rtmx:\n  database: database.csv\n")

        # Create empty database
        db_path = tmp_path / "database.csv"
        db_path.write_text(
            "req_id,category,subcategory,requirement_text,status,priority,phase\n"
            "REQ-001,TEST,Core,Test req,MISSING,HIGH,1\n"
        )

        monkeypatch.chdir(tmp_path)
        runner = CliRunner()
        result = runner.invoke(main, ["remote", "list"])

        assert result.exit_code == 0
        assert "no remotes" in result.output.lower() or "empty" in result.output.lower()

    def test_remote_add(self, tmp_path, monkeypatch):
        """Test 'rtmx remote add' adds remote configuration."""
        from click.testing import CliRunner

        from rtmx.cli.main import main

        # Create minimal config
        config_path = tmp_path / ".rtmx" / "config.yaml"
        config_path.parent.mkdir(parents=True)
        config_path.write_text("rtmx:\n  database: database.csv\n")

        # Create empty database
        db_path = tmp_path / "database.csv"
        db_path.write_text(
            "req_id,category,subcategory,requirement_text,status,priority,phase\n"
            "REQ-001,TEST,Core,Test req,MISSING,HIGH,1\n"
        )

        monkeypatch.chdir(tmp_path)
        runner = CliRunner()
        result = runner.invoke(
            main,
            ["remote", "add", "sync", "--repo", "sync-server", "--path", "../rtmx-sync"],
        )

        assert result.exit_code == 0

        # Verify config was updated
        config = load_config(config_path)
        assert "sync" in config.sync.remotes
        assert config.sync.remotes["sync"].repo == "sync-server"
        assert config.sync.remotes["sync"].path == "../rtmx-sync"

    def test_remote_remove(self, tmp_path, monkeypatch):
        """Test 'rtmx remote remove' removes remote configuration."""
        from click.testing import CliRunner

        from rtmx.cli.main import main

        # Create config with a remote
        config_path = tmp_path / ".rtmx" / "config.yaml"
        config_path.parent.mkdir(parents=True)
        config_path.write_text(
            """rtmx:
  database: database.csv
  sync:
    remotes:
      sync:
        repo: sync-server
        path: ../rtmx-sync
"""
        )

        # Create empty database
        db_path = tmp_path / "database.csv"
        db_path.write_text(
            "req_id,category,subcategory,requirement_text,status,priority,phase\n"
            "REQ-001,TEST,Core,Test req,MISSING,HIGH,1\n"
        )

        monkeypatch.chdir(tmp_path)
        runner = CliRunner()
        result = runner.invoke(main, ["remote", "remove", "sync"])

        assert result.exit_code == 0

        # Verify remote was removed
        config = load_config(config_path)
        assert "sync" not in config.sync.remotes

    def test_remote_list_with_remotes(self, tmp_path, monkeypatch):
        """Test 'rtmx remote list' shows configured remotes."""
        from click.testing import CliRunner

        from rtmx.cli.main import main

        # Create config with remotes
        config_path = tmp_path / ".rtmx" / "config.yaml"
        config_path.parent.mkdir(parents=True)
        config_path.write_text(
            """rtmx:
  database: database.csv
  sync:
    remotes:
      sync:
        repo: sync-server
        path: ../rtmx-sync
      upstream:
        repo: company/requirements
"""
        )

        # Create empty database
        db_path = tmp_path / "database.csv"
        db_path.write_text(
            "req_id,category,subcategory,requirement_text,status,priority,phase\n"
            "REQ-001,TEST,Core,Test req,MISSING,HIGH,1\n"
        )

        monkeypatch.chdir(tmp_path)
        runner = CliRunner()
        result = runner.invoke(main, ["remote", "list"])

        assert result.exit_code == 0
        assert "sync" in result.output
        assert "sync-server" in result.output
        assert "upstream" in result.output
        assert "company/requirements" in result.output
