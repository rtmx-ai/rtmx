"""Integration tests for rtmx.adapters module with realistic mocks.

These tests use mock HTTP servers/responses to test adapter behavior
in realistic scenarios, covering GitHub, Jira, and MCP adapters.

REQ-TEST-004: Adapter tests shall use mock HTTP servers
"""

from __future__ import annotations

import sys
from datetime import datetime, timezone
from typing import TYPE_CHECKING
from unittest.mock import MagicMock, Mock, patch

import pytest

# Mock external dependencies before importing adapters
sys.modules["github"] = MagicMock()
sys.modules["jira"] = MagicMock()

from rtmx.adapters.base import (
    ExternalItem,
    SyncResult,
)
from rtmx.adapters.github import GitHubAdapter
from rtmx.adapters.jira import JiraAdapter
from rtmx.config import GitHubAdapterConfig, JiraAdapterConfig, RTMXConfig
from rtmx.models import Requirement

if TYPE_CHECKING:
    from pathlib import Path


# =============================================================================
# Helper Functions
# =============================================================================


def create_test_requirement(
    req_id: str = "REQ-SW-001",
    status: str = "MISSING",
    **kwargs,
) -> Requirement:
    """Create a test requirement with sensible defaults."""
    defaults = {
        "req_id": req_id,
        "category": "SOFTWARE",
        "subcategory": "API",
        "requirement_text": "Test requirement for integration testing",
        "target_value": "v1",
        "test_module": "tests/test.py",
        "test_function": "test_func",
        "validation_method": "Unit Test",
        "status": status,
        "priority": "HIGH",
        "phase": 1,
        "notes": "Integration test requirement",
        "assignee": "developer@example.com",
        "sprint": "v1.0",
        "requirement_file": "docs/requirements/SW/REQ-SW-001.md",
        "external_id": "",
    }
    defaults.update(kwargs)
    return Requirement(**defaults)


def create_mock_github_issue(
    number: int = 42,
    title: str = "Test Issue",
    body: str = "Issue body with details",
    state: str = "open",
    labels: list | None = None,
    assignee: str | None = None,
    pull_request: bool = False,
) -> Mock:
    """Create a mock GitHub issue with realistic structure."""
    mock_issue = Mock()
    mock_issue.number = number
    mock_issue.title = title
    mock_issue.body = body
    mock_issue.state = state
    mock_issue.html_url = f"https://github.com/test/repo/issues/{number}"
    mock_issue.created_at = datetime(2024, 1, 1, 12, 0, 0)
    mock_issue.updated_at = datetime(2024, 1, 15, 14, 30, 0)
    mock_issue.pull_request = Mock() if pull_request else None

    # Labels
    if labels is None:
        labels = ["bug", "priority:high"]
    mock_labels = []
    for label_name in labels:
        mock_label = Mock()
        mock_label.name = label_name
        mock_labels.append(mock_label)
    mock_issue.labels = mock_labels

    # Assignee
    if assignee:
        mock_assignee = Mock()
        mock_assignee.login = assignee
        mock_issue.assignee = mock_assignee
    else:
        mock_issue.assignee = None

    return mock_issue


def create_mock_jira_issue(
    key: str = "PROJ-123",
    summary: str = "Test Jira Issue",
    description: str = "Detailed description",
    status: str = "To Do",
    labels: list | None = None,
    assignee: str | None = None,
    priority: str | None = None,
) -> Mock:
    """Create a mock Jira issue with realistic structure."""
    mock_issue = Mock()
    mock_issue.key = key

    # Fields
    mock_issue.fields = Mock()
    mock_issue.fields.summary = summary
    mock_issue.fields.description = description
    mock_issue.fields.created = "2024-01-01T12:00:00+00:00"
    mock_issue.fields.updated = "2024-01-15T14:30:00+00:00"

    # Status
    mock_status = Mock()
    mock_status.name = status
    mock_issue.fields.status = mock_status

    # Labels
    mock_issue.fields.labels = labels or []

    # Assignee
    if assignee:
        mock_assignee = Mock()
        mock_assignee.displayName = assignee
        mock_issue.fields.assignee = mock_assignee
    else:
        mock_issue.fields.assignee = None

    # Priority
    if priority:
        mock_priority = Mock()
        mock_priority.name = priority
        mock_issue.fields.priority = mock_priority
    else:
        mock_issue.fields.priority = None

    return mock_issue


# =============================================================================
# GitHub Adapter Integration Tests
# =============================================================================


@pytest.mark.req("REQ-TEST-004")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGitHubAdapterIntegration:
    """Integration tests for GitHubAdapter with realistic mock responses."""

    @patch.dict("os.environ", {"GITHUB_TOKEN": "ghp_test_token_12345"})
    def test_full_fetch_workflow(self):
        """Test complete workflow of fetching and processing GitHub issues."""
        config = GitHubAdapterConfig(
            enabled=True,
            repo="testorg/testrepo",
            labels=["requirement"],
            status_mapping={"open": "MISSING", "closed": "COMPLETE"},
        )
        adapter = GitHubAdapter(config)

        # Create realistic set of issues
        mock_issues = [
            create_mock_github_issue(
                number=1,
                title="[REQ-SW-001] Implement API endpoint",
                body="API endpoint for user management\n\nRTMX: REQ-SW-001",
                state="open",
                labels=["requirement", "priority:high"],
                assignee="developer1",
            ),
            create_mock_github_issue(
                number=2,
                title="[REQ-SW-002] Add authentication",
                body="OAuth2 authentication flow",
                state="closed",
                labels=["requirement", "priority:medium"],
            ),
            create_mock_github_issue(
                number=3,
                title="Fix bug in login",
                body="Login bug fix",
                state="open",
                labels=["bug"],
            ),
        ]

        with patch("github.Github") as mock_github_class:
            mock_repo = Mock()
            mock_repo.get_issues.return_value = mock_issues
            mock_client = Mock()
            mock_client.get_repo.return_value = mock_repo
            mock_github_class.return_value = mock_client

            # Fetch all items
            items = list(adapter.fetch_items())

            # Verify correct number (excluding PRs)
            assert len(items) == 3
            assert items[0].external_id == "1"
            assert items[0].requirement_id == "REQ-SW-001"
            assert items[1].external_id == "2"

    @patch.dict("os.environ", {"GITHUB_TOKEN": "ghp_test_token_12345"})
    def test_create_and_link_workflow(self):
        """Test creating issue from requirement and linking them."""
        config = GitHubAdapterConfig(
            enabled=True,
            repo="testorg/testrepo",
            labels=["rtmx-requirement"],
        )
        adapter = GitHubAdapter(config)

        requirement = create_test_requirement(
            req_id="REQ-API-001",
            requirement_text="API shall return JSON responses",
            notes="Standardization on JSON format; Acceptance: All endpoints return application/json",
        )

        with patch("github.Github") as mock_github_class:
            mock_issue = Mock()
            mock_issue.number = 42
            mock_repo = Mock()
            mock_repo.create_issue.return_value = mock_issue
            mock_client = Mock()
            mock_client.get_repo.return_value = mock_repo
            mock_github_class.return_value = mock_client

            # Create issue
            external_id = adapter.create_item(requirement)

            assert external_id == "42"
            mock_repo.create_issue.assert_called_once()

            # Verify issue structure
            call_kwargs = mock_repo.create_issue.call_args
            assert "REQ-API-001" in call_kwargs.kwargs.get("title", "")
            assert "RTMX: REQ-API-001" in call_kwargs.kwargs.get("body", "")

    @patch.dict("os.environ", {"GITHUB_TOKEN": "ghp_test_token_12345"})
    def test_update_and_close_workflow(self):
        """Test updating issue and transitioning state."""
        config = GitHubAdapterConfig(
            enabled=True,
            repo="testorg/testrepo",
            status_mapping={"open": "MISSING", "closed": "COMPLETE"},
        )
        adapter = GitHubAdapter(config)

        # Requirement marked as complete
        requirement = create_test_requirement(
            req_id="REQ-SW-001",
            status="COMPLETE",
            requirement_text="Completed requirement",
        )

        with patch("github.Github") as mock_github_class:
            mock_issue = Mock()
            mock_issue.state = "open"
            mock_repo = Mock()
            mock_repo.get_issue.return_value = mock_issue
            mock_client = Mock()
            mock_client.get_repo.return_value = mock_repo
            mock_github_class.return_value = mock_client

            # Update issue
            result = adapter.update_item("42", requirement)

            assert result is True
            # Verify edit was called to update content and close
            mock_issue.edit.assert_called()

    @patch.dict("os.environ", {"GITHUB_TOKEN": "ghp_test_token_12345"})
    def test_filter_by_labels(self):
        """Test fetching issues filtered by labels."""
        config = GitHubAdapterConfig(
            enabled=True,
            repo="testorg/testrepo",
            labels=["requirement"],
        )
        adapter = GitHubAdapter(config)

        with patch("github.Github") as mock_github_class:
            mock_repo = Mock()
            mock_repo.get_issues.return_value = []
            mock_client = Mock()
            mock_client.get_repo.return_value = mock_repo
            mock_github_class.return_value = mock_client

            # Fetch with label filter
            list(adapter.fetch_items({"labels": ["requirement", "P0"]}))

            # Verify label filter was passed
            mock_repo.get_issues.assert_called_once()
            call_kwargs = mock_repo.get_issues.call_args
            assert call_kwargs.kwargs.get("labels") == ["requirement", "P0"]

    @patch.dict("os.environ", {"GITHUB_TOKEN": "ghp_test_token_12345"})
    def test_error_handling_get_item(self):
        """Test graceful error handling when fetching non-existent issue."""
        config = GitHubAdapterConfig(enabled=True, repo="testorg/testrepo")
        adapter = GitHubAdapter(config)

        with patch("github.Github") as mock_github_class:
            mock_repo = Mock()
            mock_repo.get_issue.side_effect = Exception("Issue not found")
            mock_client = Mock()
            mock_client.get_repo.return_value = mock_repo
            mock_github_class.return_value = mock_client

            # Should return None on error
            result = adapter.get_item("99999")
            assert result is None


# =============================================================================
# Jira Adapter Integration Tests
# =============================================================================


@pytest.mark.req("REQ-TEST-004")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestJiraAdapterIntegration:
    """Integration tests for JiraAdapter with realistic mock responses."""

    @patch.dict(
        "os.environ",
        {"JIRA_API_TOKEN": "jira_test_token", "JIRA_EMAIL": "test@example.com"},
    )
    def test_full_fetch_workflow(self):
        """Test complete workflow of fetching and processing Jira issues."""
        config = JiraAdapterConfig(
            enabled=True,
            server="https://testorg.atlassian.net",
            project="RTMX",
            issue_type="Story",
            status_mapping={
                "To Do": "MISSING",
                "In Progress": "PARTIAL",
                "Done": "COMPLETE",
            },
        )
        adapter = JiraAdapter(config)

        # Create realistic set of issues
        mock_issues = [
            create_mock_jira_issue(
                key="RTMX-101",
                summary="Implement core API",
                description="Core API implementation\n\nRTMX: REQ-AP-001",
                status="In Progress",
                labels=["requirement"],
                assignee="John Doe",
                priority="High",
            ),
            create_mock_jira_issue(
                key="RTMX-102",
                summary="Add documentation",
                description="API documentation",
                status="To Do",
                labels=["documentation"],
            ),
        ]

        with patch("jira.JIRA") as mock_jira_class:
            mock_client = Mock()
            mock_client.search_issues.return_value = mock_issues
            mock_jira_class.return_value = mock_client

            # Fetch all items
            items = list(adapter.fetch_items())

            assert len(items) == 2
            assert items[0].external_id == "RTMX-101"
            assert items[0].requirement_id == "REQ-AP-001"
            assert items[0].assignee == "John Doe"
            assert items[0].priority == "High"

    @patch.dict(
        "os.environ",
        {"JIRA_API_TOKEN": "jira_test_token", "JIRA_EMAIL": "test@example.com"},
    )
    def test_create_issue_workflow(self):
        """Test creating Jira issue from requirement."""
        config = JiraAdapterConfig(
            enabled=True,
            server="https://testorg.atlassian.net",
            project="RTMX",
            issue_type="Story",
            labels=["rtmx-tracked"],
        )
        adapter = JiraAdapter(config)

        requirement = create_test_requirement(
            req_id="REQ-FEAT-001",
            requirement_text="Feature requirement",
            notes="Business need; Acceptance: Feature works correctly",
        )

        with patch("jira.JIRA") as mock_jira_class:
            mock_issue = Mock()
            mock_issue.key = "RTMX-500"
            mock_client = Mock()
            mock_client.create_issue.return_value = mock_issue
            mock_jira_class.return_value = mock_client

            external_id = adapter.create_item(requirement)

            assert external_id == "RTMX-500"
            mock_client.create_issue.assert_called_once()

            # Verify issue fields
            call_kwargs = mock_client.create_issue.call_args
            fields = call_kwargs.kwargs.get("fields", {})
            assert fields.get("project", {}).get("key") == "RTMX"
            assert "REQ-FEAT-001" in fields.get("summary", "")

    @patch.dict(
        "os.environ",
        {"JIRA_API_TOKEN": "jira_test_token", "JIRA_EMAIL": "test@example.com"},
    )
    def test_update_with_transition(self):
        """Test updating Jira issue with status transition."""
        config = JiraAdapterConfig(
            enabled=True,
            server="https://testorg.atlassian.net",
            project="RTMX",
            status_mapping={"Done": "COMPLETE", "In Progress": "PARTIAL"},
        )
        adapter = JiraAdapter(config)

        requirement = create_test_requirement(status="COMPLETE")

        with patch("jira.JIRA") as mock_jira_class:
            mock_status = Mock()
            mock_status.name = "In Progress"
            mock_issue = Mock()
            mock_issue.fields = Mock()
            mock_issue.fields.status = mock_status

            mock_transition = {"id": "31", "to": {"name": "Done"}}

            mock_client = Mock()
            mock_client.issue.return_value = mock_issue
            mock_client.transitions.return_value = [mock_transition]
            mock_jira_class.return_value = mock_client

            result = adapter.update_item("RTMX-100", requirement)

            assert result is True
            mock_issue.update.assert_called_once()
            mock_client.transition_issue.assert_called_once()

    @patch.dict(
        "os.environ",
        {"JIRA_API_TOKEN": "jira_test_token", "JIRA_EMAIL": "test@example.com"},
    )
    def test_custom_jql_query(self):
        """Test fetching with custom JQL query."""
        config = JiraAdapterConfig(
            enabled=True,
            server="https://testorg.atlassian.net",
            project="RTMX",
        )
        adapter = JiraAdapter(config)

        with patch("jira.JIRA") as mock_jira_class:
            mock_client = Mock()
            mock_client.search_issues.return_value = []
            mock_jira_class.return_value = mock_client

            custom_jql = "project = RTMX AND priority = High ORDER BY created DESC"
            list(adapter.fetch_items({"jql": custom_jql}))

            mock_client.search_issues.assert_called()
            call_args = mock_client.search_issues.call_args
            assert custom_jql in call_args[0]

    @patch.dict(
        "os.environ",
        {"JIRA_API_TOKEN": "jira_test_token", "JIRA_EMAIL": "test@example.com"},
    )
    def test_connection_test_success(self):
        """Test connection validation."""
        config = JiraAdapterConfig(
            enabled=True,
            server="https://testorg.atlassian.net",
            project="RTMX",
        )
        adapter = JiraAdapter(config)

        with patch("jira.JIRA") as mock_jira_class:
            mock_project = Mock()
            mock_project.name = "RTMX Project"
            mock_project.key = "RTMX"
            mock_client = Mock()
            mock_client.project.return_value = mock_project
            mock_jira_class.return_value = mock_client

            success, message = adapter.test_connection()

            assert success is True
            assert "RTMX Project" in message

    @patch.dict(
        "os.environ",
        {"JIRA_API_TOKEN": "jira_test_token", "JIRA_EMAIL": "test@example.com"},
    )
    def test_error_handling_update(self):
        """Test graceful error handling on update failure."""
        config = JiraAdapterConfig(
            enabled=True,
            server="https://testorg.atlassian.net",
            project="RTMX",
        )
        adapter = JiraAdapter(config)

        requirement = create_test_requirement()

        with patch("jira.JIRA") as mock_jira_class:
            mock_client = Mock()
            mock_client.issue.side_effect = Exception("Issue not found")
            mock_jira_class.return_value = mock_client

            result = adapter.update_item("RTMX-999", requirement)

            assert result is False


# =============================================================================
# MCP Adapter Integration Tests
# =============================================================================


@pytest.mark.req("REQ-TEST-004")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMCPToolsIntegration:
    """Integration tests for MCP tools with realistic database."""

    @pytest.fixture
    def sample_db_path(self, tmp_path: Path) -> Path:
        """Create a sample RTM database for testing."""
        import csv

        db_path = tmp_path / "rtm_database.csv"

        requirements = [
            {
                "req_id": "REQ-MCP-001",
                "category": "MCP",
                "subcategory": "Tools",
                "requirement_text": "MCP shall provide status tool",
                "target_value": "Implemented",
                "test_module": "tests/test_mcp.py",
                "test_function": "test_status",
                "validation_method": "Unit Test",
                "status": "COMPLETE",
                "priority": "HIGH",
                "phase": "1",
                "notes": "Status tool working",
                "effort_weeks": "0.5",
                "dependencies": "",
                "blocks": "",
                "assignee": "dev1",
                "sprint": "v0.1",
                "started_date": "2024-01-01",
                "completed_date": "2024-01-15",
                "requirement_file": "",
            },
            {
                "req_id": "REQ-MCP-002",
                "category": "MCP",
                "subcategory": "Tools",
                "requirement_text": "MCP shall provide backlog tool",
                "target_value": "Implemented",
                "test_module": "",
                "test_function": "",
                "validation_method": "Unit Test",
                "status": "MISSING",
                "priority": "MEDIUM",
                "phase": "2",
                "notes": "Not started",
                "effort_weeks": "1.0",
                "dependencies": "REQ-MCP-001",
                "blocks": "",
                "assignee": "",
                "sprint": "",
                "started_date": "",
                "completed_date": "",
                "requirement_file": "",
            },
        ]

        with open(db_path, "w", newline="") as f:
            writer = csv.DictWriter(f, fieldnames=list(requirements[0].keys()))
            writer.writeheader()
            writer.writerows(requirements)

        return db_path

    def test_mcp_get_status_workflow(self, sample_db_path: Path):
        """Test MCP get_status tool with real database."""
        from rtmx.adapters.mcp.tools import RTMXTools

        config = RTMXConfig(database=str(sample_db_path))
        tools = RTMXTools(config)

        result = tools.get_status(verbose=0)

        assert result.success is True
        assert result.data["total"] == 2
        assert result.data["complete"] == 1
        assert result.data["missing"] == 1
        assert result.data["completion_pct"] == 50.0

    def test_mcp_get_status_verbose(self, sample_db_path: Path):
        """Test MCP get_status with category breakdown."""
        from rtmx.adapters.mcp.tools import RTMXTools

        config = RTMXConfig(database=str(sample_db_path))
        tools = RTMXTools(config)

        result = tools.get_status(verbose=1)

        assert result.success is True
        assert "categories" in result.data
        assert "MCP" in result.data["categories"]
        assert result.data["categories"]["MCP"]["total"] == 2

    def test_mcp_get_backlog_workflow(self, sample_db_path: Path):
        """Test MCP get_backlog tool."""
        from rtmx.adapters.mcp.tools import RTMXTools

        config = RTMXConfig(database=str(sample_db_path))
        tools = RTMXTools(config)

        result = tools.get_backlog()

        assert result.success is True
        # Should only include incomplete requirements
        assert result.data["total_incomplete"] == 1
        backlog = result.data.get("items", [])
        assert len(backlog) == 1
        assert backlog[0]["id"] == "REQ-MCP-002"

    def test_mcp_get_requirement(self, sample_db_path: Path):
        """Test MCP get_requirement tool."""
        from rtmx.adapters.mcp.tools import RTMXTools

        config = RTMXConfig(database=str(sample_db_path))
        tools = RTMXTools(config)

        result = tools.get_requirement("REQ-MCP-001")

        assert result.success is True
        assert result.data["id"] == "REQ-MCP-001"
        assert result.data["status"] == "COMPLETE"

    def test_mcp_get_requirement_not_found(self, sample_db_path: Path):
        """Test MCP get_requirement with non-existent ID."""
        from rtmx.adapters.mcp.tools import RTMXTools

        config = RTMXConfig(database=str(sample_db_path))
        tools = RTMXTools(config)

        result = tools.get_requirement("REQ-NONEXISTENT-999")

        assert result.success is False
        assert result.error is not None

    def test_mcp_database_not_found(self, tmp_path: Path):
        """Test MCP tools handle missing database gracefully."""
        from rtmx.adapters.mcp.tools import RTMXTools

        config = RTMXConfig(database=str(tmp_path / "nonexistent.csv"))
        tools = RTMXTools(config)

        result = tools.get_status()

        assert result.success is False
        assert "not found" in result.error.lower()


# =============================================================================
# Cross-Adapter Integration Tests
# =============================================================================


@pytest.mark.req("REQ-TEST-004")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCrossAdapterIntegration:
    """Integration tests for adapter interoperability."""

    def test_external_item_consistency(self):
        """Test that ExternalItem works consistently across adapters."""
        # Create items representing the same requirement from different sources
        github_item = ExternalItem(
            external_id="42",
            title="[REQ-SW-001] API Implementation",
            description="Implement the API\n\nRTMX: REQ-SW-001",
            status="open",
            labels=["requirement"],
            url="https://github.com/org/repo/issues/42",
            created_at=datetime(2024, 1, 1, tzinfo=timezone.utc),
            requirement_id="REQ-SW-001",
        )

        jira_item = ExternalItem(
            external_id="PROJ-100",
            title="[REQ-SW-001] API Implementation",
            description="Implement the API\n\nRTMX: REQ-SW-001",
            status="To Do",
            labels=["requirement"],
            url="https://jira.example.com/browse/PROJ-100",
            created_at=datetime(2024, 1, 1, tzinfo=timezone.utc),
            requirement_id="REQ-SW-001",
        )

        # Both should have same requirement ID
        assert github_item.requirement_id == jira_item.requirement_id == "REQ-SW-001"

        # Both should serialize consistently
        github_dict = github_item.to_dict()
        jira_dict = jira_item.to_dict()

        assert github_dict["requirement_id"] == jira_dict["requirement_id"]

    def test_sync_result_aggregation(self):
        """Test aggregating sync results from multiple adapters."""
        github_result = SyncResult(
            created=["REQ-001", "REQ-002"],
            updated=["REQ-003"],
        )

        jira_result = SyncResult(
            created=["REQ-004"],
            errors=[("REQ-005", "Permission denied")],
        )

        # Aggregate results
        total_created = len(github_result.created) + len(jira_result.created)
        total_errors = len(github_result.errors) + len(jira_result.errors)

        assert total_created == 3
        assert total_errors == 1

        # Check success status
        assert github_result.success is True
        assert jira_result.success is False  # Has errors

    @patch.dict(
        "os.environ", {"GITHUB_TOKEN": "token", "JIRA_API_TOKEN": "token", "JIRA_EMAIL": "e@e.com"}
    )
    def test_adapter_configuration_independence(self):
        """Test that adapters can be configured independently."""
        github_config = GitHubAdapterConfig(
            enabled=True,
            repo="org/repo",
            status_mapping={"open": "MISSING"},
        )
        jira_config = JiraAdapterConfig(
            enabled=True,
            server="https://jira.example.com",
            project="PROJ",
            status_mapping={"To Do": "MISSING"},
        )

        github_adapter = GitHubAdapter(github_config)
        jira_adapter = JiraAdapter(jira_config)

        # Each adapter should have its own status mapping
        assert github_adapter.map_status_to_rtmx("open") == "MISSING"
        assert jira_adapter.map_status_to_rtmx("To Do") == "MISSING"

        # Different status names map correctly
        assert github_adapter.map_status_from_rtmx("MISSING") == "open"
        assert jira_adapter.map_status_from_rtmx("MISSING") == "To Do"
