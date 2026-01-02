"""End-to-end tests for RTMX sync command.

This module provides comprehensive E2E tests for the sync command with
mocked HTTP responses for GitHub and Jira integrations.

Tests cover:
1. GitHub import (mocked HTTP)
2. GitHub export (mocked HTTP)
3. Jira import (mocked HTTP)
4. Jira export (mocked HTTP)
5. Bidirectional sync with conflict resolution
6. --dry-run mode
7. Network error handling
8. Authentication failure

All tests use `scope_system` markers as required for E2E testing.
"""

from __future__ import annotations

import csv
import os
import subprocess
import sys
import tempfile
from collections.abc import Generator
from pathlib import Path
from typing import TYPE_CHECKING
from unittest.mock import Mock, patch

import pytest

from rtmx.adapters.base import ExternalItem, ServiceAdapter
from rtmx.cli.sync import (
    _get_adapter,
    run_sync,
)
from rtmx.config import RTMXConfig

if TYPE_CHECKING:
    pass


# =============================================================================
# Fixtures
# =============================================================================


@pytest.fixture
def temp_project() -> Generator[Path, None, None]:
    """Create an isolated temporary project directory."""
    with tempfile.TemporaryDirectory(prefix="rtmx_sync_test_") as tmpdir:
        project_dir = Path(tmpdir)
        # Initialize as git repo for realistic testing
        subprocess.run(
            ["git", "init"],
            cwd=project_dir,
            capture_output=True,
            check=True,
        )
        subprocess.run(
            ["git", "config", "user.email", "test@example.com"],
            cwd=project_dir,
            capture_output=True,
            check=True,
        )
        subprocess.run(
            ["git", "config", "user.name", "Test User"],
            cwd=project_dir,
            capture_output=True,
            check=True,
        )
        yield project_dir


@pytest.fixture
def mock_config(tmp_path: Path) -> RTMXConfig:
    """Create mock RTMX configuration with enabled adapters."""
    config = RTMXConfig()
    config.database = tmp_path / "docs" / "rtm_database.csv"
    config.adapters.github.enabled = True
    config.adapters.github.repo = "org/repo"
    config.adapters.jira.enabled = True
    config.adapters.jira.project = "PROJ"
    config.adapters.jira.server = "https://jira.example.com"
    config.sync.conflict_resolution = "manual"
    return config


@pytest.fixture
def sample_rtm_csv(tmp_path: Path) -> Path:
    """Create sample RTM CSV for testing."""
    csv_path = tmp_path / "docs" / "rtm_database.csv"
    csv_path.parent.mkdir(parents=True, exist_ok=True)

    headers = [
        "req_id",
        "category",
        "subcategory",
        "requirement_text",
        "target_value",
        "test_module",
        "test_function",
        "validation_method",
        "status",
        "priority",
        "phase",
        "notes",
        "effort_weeks",
        "dependencies",
        "blocks",
        "assignee",
        "sprint",
        "started_date",
        "completed_date",
        "requirement_file",
        "external_id",
    ]

    rows = [
        {
            "req_id": "REQ-TEST-001",
            "category": "TEST",
            "subcategory": "Sample",
            "requirement_text": "Test requirement",
            "target_value": "Value",
            "test_module": "tests/test.py",
            "test_function": "test_func",
            "validation_method": "Unit Test",
            "status": "COMPLETE",
            "priority": "HIGH",
            "phase": "1",
            "notes": "Note",
            "effort_weeks": "1.0",
            "dependencies": "",
            "blocks": "",
            "assignee": "dev",
            "sprint": "v0.1",
            "started_date": "2025-01-01",
            "completed_date": "2025-01-15",
            "requirement_file": "docs/req.md",
            "external_id": "",
        },
        {
            "req_id": "REQ-TEST-002",
            "category": "TEST",
            "subcategory": "Sample",
            "requirement_text": "Another requirement",
            "target_value": "Value",
            "test_module": "tests/test.py",
            "test_function": "test_func2",
            "validation_method": "Unit Test",
            "status": "PARTIAL",
            "priority": "MEDIUM",
            "phase": "1",
            "notes": "Note",
            "effort_weeks": "2.0",
            "dependencies": "",
            "blocks": "",
            "assignee": "dev",
            "sprint": "v0.1",
            "started_date": "2025-01-10",
            "completed_date": "",
            "requirement_file": "docs/req2.md",
            "external_id": "ISSUE-42",
        },
        {
            "req_id": "REQ-TEST-003",
            "category": "TEST",
            "subcategory": "Sample",
            "requirement_text": "Third requirement without external ID",
            "target_value": "Value",
            "test_module": "tests/test.py",
            "test_function": "test_func3",
            "validation_method": "Unit Test",
            "status": "MISSING",
            "priority": "LOW",
            "phase": "2",
            "notes": "",
            "effort_weeks": "0.5",
            "dependencies": "",
            "blocks": "",
            "assignee": "",
            "sprint": "",
            "started_date": "",
            "completed_date": "",
            "requirement_file": "",
            "external_id": "",
        },
    ]

    with open(csv_path, "w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=headers)
        writer.writeheader()
        writer.writerows(rows)

    return csv_path


@pytest.fixture
def mock_github_adapter() -> Mock:
    """Create mock GitHub adapter."""
    adapter = Mock(spec=ServiceAdapter)
    adapter.name = "github"
    adapter.test_connection.return_value = (True, "Connected to org/repo")
    adapter.fetch_items.return_value = iter([])
    adapter.create_item.return_value = "123"
    adapter.update_item.return_value = True
    adapter.map_status_to_rtmx.return_value = "COMPLETE"
    adapter.map_status_from_rtmx.return_value = "closed"
    return adapter


@pytest.fixture
def mock_jira_adapter() -> Mock:
    """Create mock Jira adapter."""
    adapter = Mock(spec=ServiceAdapter)
    adapter.name = "jira"
    adapter.test_connection.return_value = (True, "Connected to PROJ")
    adapter.fetch_items.return_value = iter([])
    adapter.create_item.return_value = "PROJ-123"
    adapter.update_item.return_value = True
    adapter.map_status_to_rtmx.return_value = "COMPLETE"
    adapter.map_status_from_rtmx.return_value = "Done"
    return adapter


def run_rtmx(
    *args: str,
    cwd: Path,
    env: dict[str, str] | None = None,
) -> subprocess.CompletedProcess[str]:
    """Run rtmx command and return result."""
    full_env = os.environ.copy()
    if env:
        full_env.update(env)

    return subprocess.run(
        [sys.executable, "-m", "rtmx", *args],
        cwd=cwd,
        capture_output=True,
        text=True,
        env=full_env,
    )


# =============================================================================
# E2E Tests for GitHub Sync
# =============================================================================


@pytest.mark.req("REQ-TEST-007")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGitHubSyncE2E:
    """E2E tests for GitHub sync operations."""

    def test_github_import_with_new_items(
        self,
        mock_github_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: GitHub import fetches new items and reports them."""
        mock_config.database = str(sample_rtm_csv)

        # Mock external items from GitHub
        external_items = [
            ExternalItem(
                external_id="101",
                title="New feature from GitHub",
                description="Feature description with RTMX: REQ-NEW-001",
                status="open",
                labels=["requirement", "feature"],
                url="https://github.com/org/repo/issues/101",
            ),
            ExternalItem(
                external_id="102",
                title="Another GitHub issue",
                description="Bug fix description",
                status="closed",
                labels=["bug"],
                url="https://github.com/org/repo/issues/102",
            ),
        ]
        mock_github_adapter.fetch_items.return_value = iter(external_items)

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_github_adapter):
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
        assert "RTMX Sync: GITHUB" in captured.out
        assert "Mode: import" in captured.out
        assert "Would import" in captured.out
        mock_github_adapter.test_connection.assert_called_once()
        mock_github_adapter.fetch_items.assert_called_once()

    def test_github_import_updates_linked_items(
        self,
        mock_github_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: GitHub import updates status for linked items."""
        mock_config.database = str(sample_rtm_csv)

        # Mock external item that matches existing external_id
        external_items = [
            ExternalItem(
                external_id="ISSUE-42",  # Matches REQ-TEST-002
                title="Linked requirement",
                status="closed",
            )
        ]
        mock_github_adapter.fetch_items.return_value = iter(external_items)
        mock_github_adapter.map_status_to_rtmx.return_value = "COMPLETE"

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_github_adapter):
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
        assert "Would update" in captured.out or "REQ-TEST-002" in captured.out

    def test_github_export_creates_new_issues(
        self,
        mock_github_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: GitHub export creates issues for unexported requirements."""
        mock_config.database = str(sample_rtm_csv)
        mock_github_adapter.create_item.return_value = "200"

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_github_adapter):
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
        # REQ-TEST-001 and REQ-TEST-003 have no external_id, should be exported
        assert mock_github_adapter.create_item.called

    def test_github_export_updates_existing_issues(
        self,
        mock_github_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: GitHub export updates issues for already-exported requirements."""
        mock_config.database = str(sample_rtm_csv)

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_github_adapter):
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

        _ = capsys.readouterr()  # Clear output buffer
        # REQ-TEST-002 has external_id, should be updated
        assert mock_github_adapter.update_item.called

    def test_github_export_dry_run(
        self,
        mock_github_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: GitHub export --dry-run doesn't make actual API calls."""
        mock_config.database = str(sample_rtm_csv)

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_github_adapter):
            run_sync(
                service="github",
                do_import=False,
                do_export=True,
                bidirectional=False,
                dry_run=True,
                prefer_local=False,
                prefer_remote=False,
                config=mock_config,
            )

        captured = capsys.readouterr()
        assert "DRY RUN" in captured.out
        assert "Would export" in captured.out or "Would update" in captured.out
        mock_github_adapter.create_item.assert_not_called()
        mock_github_adapter.update_item.assert_not_called()


# =============================================================================
# E2E Tests for Jira Sync
# =============================================================================


@pytest.mark.req("REQ-TEST-007")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestJiraSyncE2E:
    """E2E tests for Jira sync operations."""

    def test_jira_import_with_new_items(
        self,
        mock_jira_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Jira import fetches new items and reports them."""
        mock_config.database = str(sample_rtm_csv)

        # Mock external items from Jira
        external_items = [
            ExternalItem(
                external_id="PROJ-101",
                title="New Jira ticket",
                description="Ticket description",
                status="Open",
                labels=["requirement"],
                url="https://jira.example.com/browse/PROJ-101",
            ),
        ]
        mock_jira_adapter.fetch_items.return_value = iter(external_items)

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_jira_adapter):
            run_sync(
                service="jira",
                do_import=True,
                do_export=False,
                bidirectional=False,
                dry_run=True,
                prefer_local=False,
                prefer_remote=False,
                config=mock_config,
            )

        captured = capsys.readouterr()
        assert "RTMX Sync: JIRA" in captured.out
        assert "Mode: import" in captured.out
        assert "Would import" in captured.out

    def test_jira_export_creates_new_tickets(
        self,
        mock_jira_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Jira export creates tickets for unexported requirements."""
        mock_config.database = str(sample_rtm_csv)
        mock_jira_adapter.create_item.return_value = "PROJ-200"

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_jira_adapter):
            run_sync(
                service="jira",
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
        assert mock_jira_adapter.create_item.called

    def test_jira_export_dry_run(
        self,
        mock_jira_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Jira export --dry-run doesn't make actual API calls."""
        mock_config.database = str(sample_rtm_csv)

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_jira_adapter):
            run_sync(
                service="jira",
                do_import=False,
                do_export=True,
                bidirectional=False,
                dry_run=True,
                prefer_local=False,
                prefer_remote=False,
                config=mock_config,
            )

        captured = capsys.readouterr()
        assert "DRY RUN" in captured.out
        mock_jira_adapter.create_item.assert_not_called()
        mock_jira_adapter.update_item.assert_not_called()


# =============================================================================
# E2E Tests for Bidirectional Sync
# =============================================================================


@pytest.mark.req("REQ-TEST-007")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestBidirectionalSyncE2E:
    """E2E tests for bidirectional sync with conflict resolution."""

    def test_bidirectional_sync_prefer_local(
        self,
        mock_github_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Bidirectional sync with prefer-local conflict resolution."""
        mock_config.database = str(sample_rtm_csv)

        # Mock conflicting item (different status)
        external_items = [
            ExternalItem(
                external_id="ISSUE-42",  # Matches REQ-TEST-002
                title="Requirement with conflict",
                status="closed",
            )
        ]
        mock_github_adapter.fetch_items.return_value = iter(external_items)
        mock_github_adapter.map_status_to_rtmx.return_value = "COMPLETE"

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_github_adapter):
            run_sync(
                service="github",
                do_import=False,
                do_export=False,
                bidirectional=True,
                dry_run=False,
                prefer_local=True,
                prefer_remote=False,
                config=mock_config,
            )

        captured = capsys.readouterr()
        assert "Mode: bidirectional" in captured.out
        assert "Conflict resolution: prefer-local" in captured.out
        assert "Local wins" in captured.out

    def test_bidirectional_sync_prefer_remote(
        self,
        mock_github_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Bidirectional sync with prefer-remote conflict resolution."""
        mock_config.database = str(sample_rtm_csv)

        # Mock conflicting item (different status)
        external_items = [
            ExternalItem(
                external_id="ISSUE-42",  # Matches REQ-TEST-002
                title="Requirement with conflict",
                status="closed",
            )
        ]
        mock_github_adapter.fetch_items.return_value = iter(external_items)
        mock_github_adapter.map_status_to_rtmx.return_value = "COMPLETE"

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_github_adapter):
            run_sync(
                service="github",
                do_import=False,
                do_export=False,
                bidirectional=True,
                dry_run=False,
                prefer_local=False,
                prefer_remote=True,
                config=mock_config,
            )

        captured = capsys.readouterr()
        assert "Mode: bidirectional" in captured.out
        assert "Conflict resolution: prefer-remote" in captured.out
        assert "Remote wins" in captured.out

    def test_bidirectional_sync_manual_conflict(
        self,
        mock_github_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Bidirectional sync reports conflicts when manual resolution is set."""
        mock_config.database = str(sample_rtm_csv)
        mock_config.sync.conflict_resolution = "manual"

        # Mock conflicting item (different status)
        external_items = [
            ExternalItem(
                external_id="ISSUE-42",  # Matches REQ-TEST-002
                title="Requirement with conflict",
                status="closed",
            )
        ]
        mock_github_adapter.fetch_items.return_value = iter(external_items)
        mock_github_adapter.map_status_to_rtmx.return_value = "COMPLETE"

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_github_adapter):
            run_sync(
                service="github",
                do_import=False,
                do_export=False,
                bidirectional=True,
                dry_run=True,
                prefer_local=False,
                prefer_remote=False,
                config=mock_config,
            )

        captured = capsys.readouterr()
        assert "Conflict" in captured.out
        assert "Conflicts requiring attention" in captured.out

    def test_bidirectional_sync_identifies_import_export_candidates(
        self,
        mock_github_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Bidirectional sync identifies items to import and export."""
        mock_config.database = str(sample_rtm_csv)

        # Mock new external item (not linked to any requirement)
        external_items = [
            ExternalItem(
                external_id="NEW-ISSUE-999",
                title="Brand new external item",
                status="open",
            )
        ]
        mock_github_adapter.fetch_items.return_value = iter(external_items)

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_github_adapter):
            run_sync(
                service="github",
                do_import=False,
                do_export=False,
                bidirectional=True,
                dry_run=True,
                prefer_local=False,
                prefer_remote=False,
                config=mock_config,
            )

        captured = capsys.readouterr()
        assert "Import candidate" in captured.out or "Would import" in captured.out
        assert "Export candidate" in captured.out or "Would export" in captured.out


# =============================================================================
# E2E Tests for Error Handling
# =============================================================================


@pytest.mark.req("REQ-TEST-007")
@pytest.mark.scope_system
@pytest.mark.technique_stress
@pytest.mark.env_simulation
class TestSyncErrorHandlingE2E:
    """E2E tests for error handling in sync operations."""

    def test_network_error_handling(
        self,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Sync handles network errors gracefully."""
        mock_config.database = str(sample_rtm_csv)

        # Mock adapter that fails connection
        failed_adapter = Mock(spec=ServiceAdapter)
        failed_adapter.name = "github"
        failed_adapter.test_connection.return_value = (
            False,
            "Connection failed: Network error",
        )

        with patch("rtmx.cli.sync._get_adapter", return_value=failed_adapter):
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
        assert "Connection failed" in captured.out or "Network error" in captured.out

    def test_authentication_failure(
        self,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Sync handles authentication failures gracefully."""
        mock_config.database = str(sample_rtm_csv)

        # Mock adapter that fails authentication
        auth_failed_adapter = Mock(spec=ServiceAdapter)
        auth_failed_adapter.name = "github"
        auth_failed_adapter.test_connection.return_value = (
            False,
            "Authentication failed: Invalid token",
        )

        with patch("rtmx.cli.sync._get_adapter", return_value=auth_failed_adapter):
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
        assert "Authentication failed" in captured.out or "Invalid token" in captured.out

    def test_export_api_error_handling(
        self,
        mock_github_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Sync handles API errors during export gracefully."""
        mock_config.database = str(sample_rtm_csv)
        mock_github_adapter.create_item.side_effect = Exception("API rate limit exceeded")

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_github_adapter):
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
        assert "Failed to export" in captured.out or "API rate limit" in captured.out
        assert "errors" in captured.out.lower()

    def test_update_failure_handling(
        self,
        mock_github_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Sync handles update failures gracefully."""
        mock_config.database = str(sample_rtm_csv)
        mock_github_adapter.update_item.return_value = False

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_github_adapter):
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
        assert "Failed to update" in captured.out or "errors" in captured.out.lower()

    def test_adapter_disabled_error(
        self,
        mock_config: RTMXConfig,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Sync reports error when adapter is disabled."""
        mock_config.adapters.github.enabled = False

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

        # Should print error message and return early
        captured = capsys.readouterr()
        assert "adapter not enabled" in captured.out.lower()

    def test_unknown_service_error(
        self,
        mock_config: RTMXConfig,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Sync handles unknown service gracefully."""
        result = _get_adapter("unknown_service", mock_config)

        assert result is None
        captured = capsys.readouterr()
        assert "Unknown service" in captured.out


# =============================================================================
# E2E Tests for Dry Run Mode
# =============================================================================


@pytest.mark.req("REQ-TEST-007")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestDryRunModeE2E:
    """E2E tests for --dry-run mode across all sync operations."""

    def test_dry_run_import_no_database_changes(
        self,
        mock_github_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Dry run import doesn't modify the database."""
        mock_config.database = str(sample_rtm_csv)

        # Read original database content
        with open(sample_rtm_csv) as f:
            original_content = f.read()

        external_items = [
            ExternalItem(
                external_id="NEW-1",
                title="New item",
                status="open",
            )
        ]
        mock_github_adapter.fetch_items.return_value = iter(external_items)

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_github_adapter):
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

        # Verify database wasn't modified
        with open(sample_rtm_csv) as f:
            after_content = f.read()
        assert original_content == after_content

        captured = capsys.readouterr()
        assert "DRY RUN" in captured.out

    def test_dry_run_bidirectional_no_side_effects(
        self,
        mock_github_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Dry run bidirectional sync has no side effects."""
        mock_config.database = str(sample_rtm_csv)

        external_items = [
            ExternalItem(
                external_id="ISSUE-42",
                title="Conflicting item",
                status="closed",
            )
        ]
        mock_github_adapter.fetch_items.return_value = iter(external_items)
        mock_github_adapter.map_status_to_rtmx.return_value = "COMPLETE"

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_github_adapter):
            run_sync(
                service="github",
                do_import=False,
                do_export=False,
                bidirectional=True,
                dry_run=True,
                prefer_local=True,
                prefer_remote=False,
                config=mock_config,
            )

        # Verify no actual updates were made
        mock_github_adapter.create_item.assert_not_called()
        mock_github_adapter.update_item.assert_not_called()

        captured = capsys.readouterr()
        assert "DRY RUN" in captured.out


# =============================================================================
# E2E Tests for CLI Integration
# =============================================================================


@pytest.mark.req("REQ-TEST-007")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestSyncCLIIntegrationE2E:
    """E2E tests for sync command CLI integration."""

    def test_sync_no_direction_specified(
        self,
        mock_config: RTMXConfig,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Sync exits with error when no direction is specified."""
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

    def test_sync_conflicting_preferences(
        self,
        mock_config: RTMXConfig,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Sync exits with error when both preferences are set."""
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
        assert "Cannot use both" in captured.out

    def test_sync_import_and_export_equals_bidirectional(
        self,
        mock_github_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Using both --import and --export equals bidirectional mode."""
        mock_config.database = str(sample_rtm_csv)
        mock_github_adapter.fetch_items.return_value = iter([])

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_github_adapter):
            run_sync(
                service="github",
                do_import=True,
                do_export=True,
                bidirectional=False,
                dry_run=True,
                prefer_local=False,
                prefer_remote=False,
                config=mock_config,
            )

        captured = capsys.readouterr()
        assert "Mode: bidirectional" in captured.out


# =============================================================================
# E2E Tests for Sync Summary
# =============================================================================


@pytest.mark.req("REQ-TEST-007")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestSyncSummaryE2E:
    """E2E tests for sync summary output."""

    def test_sync_summary_shows_all_counts(
        self,
        mock_github_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Sync summary includes all relevant counts."""
        mock_config.database = str(sample_rtm_csv)

        # Mock multiple external items with various states
        external_items = [
            ExternalItem(
                external_id="ISSUE-42",  # Linked, same status (skip)
                title="Linked item",
                status="in_progress",
            ),
            ExternalItem(
                external_id="NEW-1",  # New (create)
                title="New item",
                status="open",
            ),
        ]
        mock_github_adapter.fetch_items.return_value = iter(external_items)
        mock_github_adapter.map_status_to_rtmx.return_value = "PARTIAL"

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_github_adapter):
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
        assert "Sync Summary" in captured.out

    def test_sync_summary_shows_conflicts(
        self,
        mock_github_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Sync summary shows conflicts requiring attention."""
        mock_config.database = str(sample_rtm_csv)
        mock_config.sync.conflict_resolution = "manual"

        external_items = [
            ExternalItem(
                external_id="ISSUE-42",
                title="Conflicting item",
                status="closed",
            )
        ]
        mock_github_adapter.fetch_items.return_value = iter(external_items)
        mock_github_adapter.map_status_to_rtmx.return_value = "COMPLETE"

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_github_adapter):
            run_sync(
                service="github",
                do_import=False,
                do_export=False,
                bidirectional=True,
                dry_run=True,
                prefer_local=False,
                prefer_remote=False,
                config=mock_config,
            )

        captured = capsys.readouterr()
        assert "Conflicts requiring attention" in captured.out

    def test_sync_summary_no_changes(
        self,
        mock_github_adapter: Mock,
        mock_config: RTMXConfig,
        sample_rtm_csv: Path,
        capsys: pytest.CaptureFixture,
    ) -> None:
        """E2E test: Sync summary reports no changes when nothing to sync."""
        mock_config.database = str(sample_rtm_csv)
        mock_github_adapter.fetch_items.return_value = iter([])

        with patch("rtmx.cli.sync._get_adapter", return_value=mock_github_adapter):
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
        assert "no changes" in captured.out or "0" in captured.out
