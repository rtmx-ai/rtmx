"""Tests for RTMX sync CLI command.

This module tests the run_sync function from src/rtmx/cli/sync.py
which handles bi-directional synchronization with external services.
"""

from __future__ import annotations

from pathlib import Path
from unittest.mock import Mock, patch

import pytest

from rtmx.adapters.base import ExternalItem, ServiceAdapter, SyncResult
from rtmx.cli.sync import (
    _get_adapter,
    _print_summary,
    _run_bidirectional,
    _run_export,
    _run_import,
    run_sync,
)
from rtmx.config import RTMXConfig

# =============================================================================
# Fixtures
# =============================================================================


@pytest.fixture
def mock_config(tmp_path: Path) -> RTMXConfig:
    """Create mock RTMX configuration."""
    config = RTMXConfig()
    config.database = tmp_path / "docs" / "rtm_database.csv"
    config.adapters.github.enabled = True
    config.adapters.github.repo = "org/repo"
    config.adapters.jira.enabled = True
    config.adapters.jira.project = "PROJ"
    config.sync.conflict_resolution = "manual"
    return config


@pytest.fixture
def mock_adapter() -> Mock:
    """Create mock service adapter."""
    adapter = Mock(spec=ServiceAdapter)
    adapter.name = "MockAdapter"
    adapter.test_connection.return_value = (True, "Connection successful")
    adapter.fetch_items.return_value = []
    adapter.create_item.return_value = "EXT-001"
    adapter.update_item.return_value = True
    adapter.map_status_to_rtmx.return_value = "COMPLETE"
    adapter.map_status_from_rtmx.return_value = "closed"
    return adapter


@pytest.fixture
def sample_rtm_csv(tmp_path: Path) -> Path:
    """Create sample RTM CSV for testing."""
    csv_path = tmp_path / "docs" / "rtm_database.csv"
    csv_path.parent.mkdir(parents=True, exist_ok=True)

    content = """req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id
REQ-TEST-001,TEST,Sample,Test requirement,Value,tests/test.py,test_func,Unit Test,COMPLETE,HIGH,1,Note,1.0,,,dev,v0.1,2025-01-01,2025-01-15,docs/req.md,
REQ-TEST-002,TEST,Sample,Another requirement,Value,tests/test.py,test_func2,Unit Test,PARTIAL,MEDIUM,1,Note,2.0,,,dev,v0.1,2025-01-10,,docs/req2.md,ISSUE-42
"""
    csv_path.write_text(content)
    return csv_path


# =============================================================================
# Tests for run_sync
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_sync_no_direction_exits(
    mock_config: RTMXConfig, capsys: pytest.CaptureFixture
) -> None:
    """Test that run_sync exits when no sync direction is specified."""
    with pytest.raises(SystemExit) as exc_info:
        run_sync(
            service="github",
            do_import=False,
            do_export=False,
            bidirectional=False,
            dry_run=False,
            prefer_local=False,
            prefer_remote=False,
            config=mock_config,
        )

    assert exc_info.value.code == 1
    captured = capsys.readouterr()
    assert "No sync direction specified" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_sync_both_preferences_exits(
    mock_config: RTMXConfig, capsys: pytest.CaptureFixture
) -> None:
    """Test that run_sync exits when both prefer_local and prefer_remote are set."""
    with pytest.raises(SystemExit) as exc_info:
        run_sync(
            service="github",
            do_import=True,
            do_export=False,
            bidirectional=False,
            dry_run=False,
            prefer_local=True,
            prefer_remote=True,
            config=mock_config,
        )

    assert exc_info.value.code == 1
    captured = capsys.readouterr()
    assert "Cannot use both --prefer-local and --prefer-remote" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
@patch("rtmx.cli.sync._get_adapter")
@patch("rtmx.cli.sync._run_import")
def test_run_sync_import_mode(
    mock_run_import: Mock,
    mock_get_adapter: Mock,
    mock_adapter: Mock,
    mock_config: RTMXConfig,
    capsys: pytest.CaptureFixture,
) -> None:
    """Test run_sync in import mode."""
    mock_get_adapter.return_value = mock_adapter
    mock_run_import.return_value = SyncResult(created=["EXT-1"], updated=["EXT-2"])

    run_sync(
        service="github",
        do_import=True,
        do_export=False,
        bidirectional=False,
        dry_run=False,
        prefer_local=False,
        prefer_remote=False,
        config=mock_config,
    )

    captured = capsys.readouterr()
    assert "RTMX Sync: GITHUB" in captured.out
    assert "Mode: import" in captured.out
    mock_adapter.test_connection.assert_called_once()
    mock_run_import.assert_called_once()


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
@patch("rtmx.cli.sync._get_adapter")
@patch("rtmx.cli.sync._run_export")
def test_run_sync_export_mode(
    mock_run_export: Mock,
    mock_get_adapter: Mock,
    mock_adapter: Mock,
    mock_config: RTMXConfig,
    capsys: pytest.CaptureFixture,
) -> None:
    """Test run_sync in export mode."""
    mock_get_adapter.return_value = mock_adapter
    mock_run_export.return_value = SyncResult(created=["REQ-1"], updated=["REQ-2"])

    run_sync(
        service="github",
        do_import=False,
        do_export=True,
        bidirectional=False,
        dry_run=False,
        prefer_local=False,
        prefer_remote=False,
        config=mock_config,
    )

    captured = capsys.readouterr()
    assert "Mode: export" in captured.out
    mock_run_export.assert_called_once()


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
@patch("rtmx.cli.sync._get_adapter")
@patch("rtmx.cli.sync._run_bidirectional")
def test_run_sync_bidirectional_mode(
    mock_run_bidirectional: Mock,
    mock_get_adapter: Mock,
    mock_adapter: Mock,
    mock_config: RTMXConfig,
    capsys: pytest.CaptureFixture,
) -> None:
    """Test run_sync in bidirectional mode."""
    mock_get_adapter.return_value = mock_adapter
    mock_run_bidirectional.return_value = SyncResult()

    run_sync(
        service="github",
        do_import=False,
        do_export=False,
        bidirectional=True,
        dry_run=False,
        prefer_local=False,
        prefer_remote=False,
        config=mock_config,
    )

    captured = capsys.readouterr()
    assert "Mode: bidirectional" in captured.out
    mock_run_bidirectional.assert_called_once()


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
@patch("rtmx.cli.sync._get_adapter")
def test_run_sync_dry_run_flag(
    mock_get_adapter: Mock,
    mock_adapter: Mock,
    mock_config: RTMXConfig,
    capsys: pytest.CaptureFixture,
) -> None:
    """Test run_sync with dry_run flag."""
    mock_get_adapter.return_value = mock_adapter

    run_sync(
        service="github",
        do_import=True,
        do_export=False,
        bidirectional=False,
        dry_run=True,
        prefer_local=False,
        prefer_remote=False,
        config=mock_config,
    )

    captured = capsys.readouterr()
    assert "DRY RUN" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
@patch("rtmx.cli.sync._get_adapter")
def test_run_sync_prefer_local(
    mock_get_adapter: Mock,
    mock_adapter: Mock,
    mock_config: RTMXConfig,
    capsys: pytest.CaptureFixture,
) -> None:
    """Test run_sync with prefer_local conflict resolution."""
    mock_get_adapter.return_value = mock_adapter

    run_sync(
        service="github",
        do_import=True,
        do_export=True,
        bidirectional=False,
        dry_run=False,
        prefer_local=True,
        prefer_remote=False,
        config=mock_config,
    )

    captured = capsys.readouterr()
    assert "Conflict resolution: prefer-local" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
@patch("rtmx.cli.sync._get_adapter")
def test_run_sync_connection_failure(
    mock_get_adapter: Mock,
    mock_config: RTMXConfig,
    capsys: pytest.CaptureFixture,
) -> None:
    """Test run_sync when connection test fails."""
    adapter = Mock(spec=ServiceAdapter)
    adapter.name = "MockAdapter"
    adapter.test_connection.return_value = (False, "Connection failed")
    mock_get_adapter.return_value = adapter

    run_sync(
        service="github",
        do_import=True,
        do_export=False,
        bidirectional=False,
        dry_run=False,
        prefer_local=False,
        prefer_remote=False,
        config=mock_config,
    )

    captured = capsys.readouterr()
    assert "Connection failed" in captured.out


# =============================================================================
# Tests for _get_adapter
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_get_adapter_github_disabled(capsys: pytest.CaptureFixture) -> None:
    """Test _get_adapter when GitHub adapter is disabled."""
    config = RTMXConfig()
    config.adapters.github.enabled = False

    result = _get_adapter("github", config)

    assert result is None
    captured = capsys.readouterr()
    assert "GitHub adapter not enabled" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_get_adapter_github_no_repo(capsys: pytest.CaptureFixture) -> None:
    """Test _get_adapter when GitHub repo not configured."""
    config = RTMXConfig()
    config.adapters.github.enabled = True
    config.adapters.github.repo = ""

    result = _get_adapter("github", config)

    assert result is None
    captured = capsys.readouterr()
    assert "GitHub repo not configured" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_get_adapter_jira_disabled(capsys: pytest.CaptureFixture) -> None:
    """Test _get_adapter when Jira adapter is disabled."""
    config = RTMXConfig()
    config.adapters.jira.enabled = False

    result = _get_adapter("jira", config)

    assert result is None
    captured = capsys.readouterr()
    assert "Jira adapter not enabled" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_get_adapter_jira_no_project(capsys: pytest.CaptureFixture) -> None:
    """Test _get_adapter when Jira project not configured."""
    config = RTMXConfig()
    config.adapters.jira.enabled = True
    config.adapters.jira.project = ""

    result = _get_adapter("jira", config)

    assert result is None
    captured = capsys.readouterr()
    assert "Jira project not configured" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_get_adapter_unknown_service(capsys: pytest.CaptureFixture) -> None:
    """Test _get_adapter with unknown service."""
    config = RTMXConfig()

    result = _get_adapter("unknown", config)

    assert result is None
    captured = capsys.readouterr()
    assert "Unknown service: unknown" in captured.out


# =============================================================================
# Tests for _run_import
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_import_new_items(
    mock_adapter: Mock,
    mock_config: RTMXConfig,
    sample_rtm_csv: Path,
    capsys: pytest.CaptureFixture,
) -> None:
    """Test _run_import with new items from external service."""
    mock_config.database = str(sample_rtm_csv)

    # Mock external items
    external_items = [
        ExternalItem(
            external_id="ISSUE-100",
            title="New feature request",
            status="open",
        )
    ]
    mock_adapter.fetch_items.return_value = iter(external_items)

    result = _run_import(mock_adapter, mock_config, dry_run=True)

    assert len(result.created) == 1
    assert "ISSUE-100" in result.created
    captured = capsys.readouterr()
    assert "Would import" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_import_linked_items(
    mock_adapter: Mock,
    mock_config: RTMXConfig,
    sample_rtm_csv: Path,
) -> None:
    """Test _run_import with items already linked to requirements."""
    mock_config.database = str(sample_rtm_csv)

    # Mock external item with different status
    external_items = [
        ExternalItem(
            external_id="ISSUE-42",
            title="Linked requirement",
            status="closed",
        )
    ]
    mock_adapter.fetch_items.return_value = iter(external_items)
    mock_adapter.map_status_to_rtmx.return_value = "COMPLETE"

    result = _run_import(mock_adapter, mock_config, dry_run=True)

    # Should update status
    assert len(result.updated) == 1


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_import_no_rtm_file(
    mock_adapter: Mock,
    mock_config: RTMXConfig,
    tmp_path: Path,
) -> None:
    """Test _run_import when RTM database doesn't exist yet."""
    mock_config.database = str(tmp_path / "nonexistent.csv")

    external_items = [ExternalItem(external_id="ISSUE-1", title="First item", status="open")]
    mock_adapter.fetch_items.return_value = iter(external_items)

    result = _run_import(mock_adapter, mock_config, dry_run=True)

    assert len(result.created) == 1


# =============================================================================
# Tests for _run_export
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_export_no_database(
    mock_adapter: Mock,
    mock_config: RTMXConfig,
    tmp_path: Path,
    capsys: pytest.CaptureFixture,
) -> None:
    """Test _run_export when RTM database doesn't exist."""
    mock_config.database = str(tmp_path / "nonexistent.csv")

    result = _run_export(mock_adapter, mock_config, dry_run=False)

    assert len(result.created) == 0
    assert len(result.updated) == 0
    captured = capsys.readouterr()
    assert "RTM database not found" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_export_new_requirements(
    mock_adapter: Mock,
    mock_config: RTMXConfig,
    sample_rtm_csv: Path,
) -> None:
    """Test _run_export with requirements not yet exported."""
    mock_config.database = str(sample_rtm_csv)
    mock_adapter.create_item.return_value = "ISSUE-NEW"

    result = _run_export(mock_adapter, mock_config, dry_run=False)

    # REQ-TEST-001 has no external_id, should be created
    assert len(result.created) >= 1
    mock_adapter.create_item.assert_called()


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_export_existing_requirements(
    mock_adapter: Mock,
    mock_config: RTMXConfig,
    sample_rtm_csv: Path,
) -> None:
    """Test _run_export with requirements already exported."""
    mock_config.database = str(sample_rtm_csv)
    mock_adapter.update_item.return_value = True

    result = _run_export(mock_adapter, mock_config, dry_run=False)

    # REQ-TEST-002 has external_id, should be updated
    assert len(result.updated) >= 1


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_export_dry_run(
    mock_adapter: Mock,
    mock_config: RTMXConfig,
    sample_rtm_csv: Path,
    capsys: pytest.CaptureFixture,
) -> None:
    """Test _run_export in dry run mode."""
    mock_config.database = str(sample_rtm_csv)

    _run_export(mock_adapter, mock_config, dry_run=True)

    # Should not actually call adapter methods
    mock_adapter.create_item.assert_not_called()
    mock_adapter.update_item.assert_not_called()
    captured = capsys.readouterr()
    assert "Would export" in captured.out or "Would update" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_export_create_failure(
    mock_adapter: Mock,
    mock_config: RTMXConfig,
    sample_rtm_csv: Path,
) -> None:
    """Test _run_export when create_item raises exception."""
    mock_config.database = str(sample_rtm_csv)
    mock_adapter.create_item.side_effect = Exception("API Error")

    result = _run_export(mock_adapter, mock_config, dry_run=False)

    assert len(result.errors) >= 1


# =============================================================================
# Tests for _run_bidirectional
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_bidirectional_no_conflicts(
    mock_adapter: Mock,
    mock_config: RTMXConfig,
    sample_rtm_csv: Path,
) -> None:
    """Test _run_bidirectional when statuses match."""
    mock_config.database = str(sample_rtm_csv)

    # External item matches local status
    external_items = [
        ExternalItem(
            external_id="ISSUE-42",
            title="Synced requirement",
            status="in_progress",
        )
    ]
    mock_adapter.fetch_items.return_value = iter(external_items)
    mock_adapter.map_status_to_rtmx.return_value = "PARTIAL"

    result = _run_bidirectional(mock_adapter, mock_config, "manual", dry_run=True)

    assert len(result.conflicts) == 0
    assert len(result.skipped) >= 1


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_bidirectional_prefer_local(
    mock_adapter: Mock,
    mock_config: RTMXConfig,
    sample_rtm_csv: Path,
    capsys: pytest.CaptureFixture,
) -> None:
    """Test _run_bidirectional with prefer-local conflict resolution."""
    mock_config.database = str(sample_rtm_csv)

    # External item has different status
    external_items = [
        ExternalItem(
            external_id="ISSUE-42",
            title="Conflicted requirement",
            status="closed",
        )
    ]
    mock_adapter.fetch_items.return_value = iter(external_items)
    mock_adapter.map_status_to_rtmx.return_value = "COMPLETE"

    _run_bidirectional(mock_adapter, mock_config, "prefer-local", dry_run=True)

    captured = capsys.readouterr()
    assert "Local wins" in captured.out or "Would update" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_bidirectional_prefer_remote(
    mock_adapter: Mock,
    mock_config: RTMXConfig,
    sample_rtm_csv: Path,
    capsys: pytest.CaptureFixture,
) -> None:
    """Test _run_bidirectional with prefer-remote conflict resolution."""
    mock_config.database = str(sample_rtm_csv)

    # External item has different status
    external_items = [
        ExternalItem(
            external_id="ISSUE-42",
            title="Conflicted requirement",
            status="closed",
        )
    ]
    mock_adapter.fetch_items.return_value = iter(external_items)
    mock_adapter.map_status_to_rtmx.return_value = "COMPLETE"

    _run_bidirectional(mock_adapter, mock_config, "prefer-remote", dry_run=True)

    captured = capsys.readouterr()
    assert "Remote wins" in captured.out or "Would update" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_bidirectional_manual_conflict(
    mock_adapter: Mock,
    mock_config: RTMXConfig,
    sample_rtm_csv: Path,
    capsys: pytest.CaptureFixture,
) -> None:
    """Test _run_bidirectional with manual conflict resolution."""
    mock_config.database = str(sample_rtm_csv)

    # External item has different status
    external_items = [
        ExternalItem(
            external_id="ISSUE-42",
            title="Conflicted requirement",
            status="closed",
        )
    ]
    mock_adapter.fetch_items.return_value = iter(external_items)
    mock_adapter.map_status_to_rtmx.return_value = "COMPLETE"

    result = _run_bidirectional(mock_adapter, mock_config, "manual", dry_run=True)

    assert len(result.conflicts) >= 1
    captured = capsys.readouterr()
    assert "Conflict" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_bidirectional_import_candidates(
    mock_adapter: Mock,
    mock_config: RTMXConfig,
    sample_rtm_csv: Path,
    capsys: pytest.CaptureFixture,
) -> None:
    """Test _run_bidirectional identifies import candidates."""
    mock_config.database = str(sample_rtm_csv)

    # External item not linked to any requirement
    external_items = [
        ExternalItem(
            external_id="ISSUE-999",
            title="New external item",
            status="open",
        )
    ]
    mock_adapter.fetch_items.return_value = iter(external_items)

    result = _run_bidirectional(mock_adapter, mock_config, "manual", dry_run=True)

    assert len(result.created) >= 1
    captured = capsys.readouterr()
    assert "Import candidate" in captured.out or "Would import" in captured.out


# =============================================================================
# Tests for _print_summary
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_print_summary_basic(capsys: pytest.CaptureFixture) -> None:
    """Test _print_summary with basic sync result."""
    result = SyncResult(created=["ID1", "ID2"], updated=["ID3"])

    _print_summary(result)

    captured = capsys.readouterr()
    assert "Sync Summary" in captured.out
    assert "2 created" in captured.out
    assert "1 updated" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_print_summary_with_conflicts(capsys: pytest.CaptureFixture) -> None:
    """Test _print_summary with conflicts."""
    result = SyncResult(conflicts=[("REQ-1", "Status mismatch"), ("REQ-2", "Priority conflict")])

    _print_summary(result)

    captured = capsys.readouterr()
    assert "Conflicts requiring attention" in captured.out
    assert "REQ-1" in captured.out
    assert "Status mismatch" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_print_summary_with_errors(capsys: pytest.CaptureFixture) -> None:
    """Test _print_summary with errors."""
    result = SyncResult(errors=[("REQ-1", "API timeout"), ("REQ-2", "Invalid data")])

    _print_summary(result)

    captured = capsys.readouterr()
    assert "Errors" in captured.out
    assert "REQ-1" in captured.out
    assert "API timeout" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_print_summary_no_changes(capsys: pytest.CaptureFixture) -> None:
    """Test _print_summary when no changes occurred."""
    result = SyncResult()

    _print_summary(result)

    captured = capsys.readouterr()
    assert "no changes" in captured.out
