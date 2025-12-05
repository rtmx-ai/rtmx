"""Comprehensive tests for rtmx.adapters module - achieving high coverage.

This test suite covers all adapter classes:
- BaseAdapter: ExternalItem, SyncResult, ServiceAdapter interface
- GitHubAdapter: GitHub Issues integration with mocked HTTP calls
- JiraAdapter: Jira integration with mocked HTTP calls

All external dependencies (PyGithub, jira-python) are mocked to avoid actual API calls.
"""

import sys
from datetime import datetime, timezone
from unittest.mock import MagicMock, Mock, patch

import pytest

# Mock github and jira modules before importing adapters
sys.modules["github"] = MagicMock()
sys.modules["jira"] = MagicMock()

from rtmx.adapters.base import (
    ConflictResolution,
    ExternalItem,
    ServiceAdapter,
    SyncDirection,
    SyncResult,
)
from rtmx.adapters.github import GitHubAdapter
from rtmx.adapters.jira import JiraAdapter
from rtmx.config import GitHubAdapterConfig, JiraAdapterConfig
from rtmx.models import Requirement

# ===============================================================================
# Helper Functions
# ===============================================================================


def create_test_requirement(req_id="REQ-SW-001", **kwargs):
    """Create a test requirement with sensible defaults."""
    defaults = {
        "req_id": req_id,
        "category": "SOFTWARE",
        "subcategory": "API",
        "requirement_text": "Test requirement",
        "target_value": "v1",
        "test_module": "tests/test.py",
        "test_function": "test_func",
        "validation_method": "Unit Test",
        "status": "MISSING",
        "priority": "HIGH",
        "phase": 1,
        "notes": "",
        "assignee": "dev",
        "sprint": "v1.0",
        "requirement_file": "",
        "external_id": "",
    }
    defaults.update(kwargs)
    return Requirement(**defaults)


# ===============================================================================
# BaseAdapter Tests
# ===============================================================================


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestSyncDirection:
    """Tests for SyncDirection enum."""

    def test_sync_direction_values(self):
        """Test SyncDirection enum has correct values."""
        assert SyncDirection.IMPORT.value == "import"
        assert SyncDirection.EXPORT.value == "export"
        assert SyncDirection.BIDIRECTIONAL.value == "bidirectional"

    def test_sync_direction_members(self):
        """Test all SyncDirection members exist."""
        directions = [d.value for d in SyncDirection]
        assert "import" in directions
        assert "export" in directions
        assert "bidirectional" in directions


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestConflictResolution:
    """Tests for ConflictResolution enum."""

    def test_conflict_resolution_values(self):
        """Test ConflictResolution enum has correct values."""
        assert ConflictResolution.MANUAL.value == "manual"
        assert ConflictResolution.PREFER_LOCAL.value == "prefer-local"
        assert ConflictResolution.PREFER_REMOTE.value == "prefer-remote"


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestExternalItem:
    """Tests for ExternalItem dataclass."""

    def test_external_item_minimal(self):
        """Test creating ExternalItem with minimal required fields."""
        item = ExternalItem(external_id="123", title="Test Item")
        assert item.external_id == "123"
        assert item.title == "Test Item"
        assert item.description == ""
        assert item.status == "open"

    def test_external_item_full(self):
        """Test creating ExternalItem with all fields."""
        created = datetime(2024, 1, 1, 12, 0, 0, tzinfo=timezone.utc)
        updated = datetime(2024, 1, 2, 12, 0, 0, tzinfo=timezone.utc)

        item = ExternalItem(
            external_id="456",
            title="Full Item",
            description="Detailed description",
            status="closed",
            labels=["bug"],
            created_at=created,
            updated_at=updated,
            assignee="user@example.com",
            priority="HIGH",
            requirement_id="REQ-SW-001",
        )

        assert item.external_id == "456"
        assert item.title == "Full Item"
        assert item.assignee == "user@example.com"

    def test_external_item_to_dict(self):
        """Test converting ExternalItem to dictionary."""
        created = datetime(2024, 1, 1, 12, 0, 0, tzinfo=timezone.utc)
        item = ExternalItem(external_id="789", title="Dict Test", created_at=created)
        result = item.to_dict()

        assert result["external_id"] == "789"
        assert result["title"] == "Dict Test"
        assert result["created_at"] == "2024-01-01T12:00:00+00:00"


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestSyncResult:
    """Tests for SyncResult dataclass."""

    def test_sync_result_empty(self):
        """Test empty SyncResult."""
        result = SyncResult()
        assert result.success is True
        assert result.summary == "no changes"

    def test_sync_result_with_operations(self):
        """Test SyncResult with multiple operation types."""
        result = SyncResult(
            created=["REQ-001", "REQ-002"],
            updated=["REQ-003"],
            errors=[("REQ-007", "error")],
        )
        assert result.success is False  # Has errors
        assert "2 created" in result.summary
        assert "1 updated" in result.summary
        assert "1 errors" in result.summary


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestServiceAdapter:
    """Tests for ServiceAdapter abstract base class."""

    def test_service_adapter_is_abstract(self):
        """Test that ServiceAdapter cannot be instantiated directly."""
        with pytest.raises(TypeError):
            ServiceAdapter()  # type: ignore


# ===============================================================================
# GitHubAdapter Tests
# ===============================================================================


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGitHubAdapterInit:
    """Tests for GitHubAdapter initialization."""

    def test_github_adapter_initialization(self):
        """Test GitHubAdapter initializes correctly."""
        config = GitHubAdapterConfig(enabled=True, repo="owner/repo")
        adapter = GitHubAdapter(config)

        assert adapter.name == "github"
        assert adapter._config == config

    @patch.dict("os.environ", {"GITHUB_TOKEN": "test-token"})
    def test_github_adapter_is_configured_with_token(self):
        """Test is_configured returns True with valid config and token."""
        config = GitHubAdapterConfig(enabled=True, repo="owner/repo")
        adapter = GitHubAdapter(config)
        assert adapter.is_configured is True

    @patch.dict("os.environ", {}, clear=True)
    def test_github_adapter_is_configured_no_token(self):
        """Test is_configured returns False without token."""
        config = GitHubAdapterConfig(enabled=True, repo="owner/repo")
        adapter = GitHubAdapter(config)
        assert adapter.is_configured is False


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGitHubAdapterClient:
    """Tests for GitHubAdapter client management."""

    @patch.dict("os.environ", {"GITHUB_TOKEN": "test-token"})
    def test_get_client_creates_client(self):
        """Test _get_client creates a GitHub client."""
        config = GitHubAdapterConfig(enabled=True, repo="owner/repo")
        adapter = GitHubAdapter(config)

        with patch("github.Github") as mock_github_class:
            client = adapter._get_client()
            mock_github_class.assert_called_once_with("test-token")
            assert client == mock_github_class.return_value

    @patch.dict("os.environ", {}, clear=True)
    def test_get_client_no_token_raises(self):
        """Test _get_client raises ValueError without token."""
        config = GitHubAdapterConfig(enabled=True, repo="owner/repo")
        adapter = GitHubAdapter(config)

        with patch("github.Github"):
            with pytest.raises(ValueError, match="GitHub token not found"):
                adapter._get_client()

    @patch.dict("os.environ", {"GITHUB_TOKEN": "test-token"})
    def test_get_repo(self):
        """Test _get_repo returns repository object."""
        config = GitHubAdapterConfig(enabled=True, repo="owner/repo")
        adapter = GitHubAdapter(config)

        with patch("github.Github") as mock_github_class:
            mock_client = Mock()
            mock_repo = Mock()
            mock_client.get_repo.return_value = mock_repo
            mock_github_class.return_value = mock_client

            repo = adapter._get_repo()
            mock_client.get_repo.assert_called_once_with("owner/repo")
            assert repo == mock_repo


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGitHubAdapterConnection:
    """Tests for GitHubAdapter connection testing."""

    @patch.dict("os.environ", {"GITHUB_TOKEN": "test-token"})
    def test_test_connection_success(self):
        """Test test_connection succeeds with valid credentials."""
        config = GitHubAdapterConfig(enabled=True, repo="owner/repo")
        adapter = GitHubAdapter(config)

        with patch("github.Github") as mock_github_class:
            mock_repo = Mock()
            mock_repo.full_name = "owner/repo"
            mock_client = Mock()
            mock_client.get_repo.return_value = mock_repo
            mock_github_class.return_value = mock_client

            success, message = adapter.test_connection()
            assert success is True
            assert "Connected to owner/repo" in message


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGitHubAdapterMapping:
    """Tests for GitHubAdapter status mapping."""

    def test_map_status_to_rtmx(self):
        """Test mapping GitHub status to RTMX status."""
        config = GitHubAdapterConfig(
            enabled=True, repo="owner/repo", status_mapping={"open": "MISSING"}
        )
        adapter = GitHubAdapter(config)
        assert adapter.map_status_to_rtmx("open") == "MISSING"

    def test_map_status_from_rtmx(self):
        """Test mapping RTMX status to GitHub status."""
        config = GitHubAdapterConfig(
            enabled=True, repo="owner/repo", status_mapping={"open": "MISSING"}
        )
        adapter = GitHubAdapter(config)
        assert adapter.map_status_from_rtmx("MISSING") == "open"


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGitHubAdapterFetch:
    """Tests for GitHubAdapter fetch operations."""

    @patch.dict("os.environ", {"GITHUB_TOKEN": "test-token"})
    def test_fetch_items_basic(self):
        """Test fetching items from GitHub."""
        config = GitHubAdapterConfig(enabled=True, repo="owner/repo")
        adapter = GitHubAdapter(config)

        with patch("github.Github") as mock_github_class:
            mock_issue = Mock()
            mock_issue.number = 42
            mock_issue.title = "Test Issue"
            mock_issue.body = "Issue body"
            mock_issue.state = "open"
            mock_issue.labels = []
            mock_issue.html_url = "https://github.com/owner/repo/issues/42"
            mock_issue.created_at = datetime(2024, 1, 1, 12, 0, 0)
            mock_issue.updated_at = datetime(2024, 1, 2, 12, 0, 0)
            mock_issue.assignee = None
            mock_issue.pull_request = None

            mock_repo = Mock()
            mock_repo.get_issues.return_value = [mock_issue]
            mock_client = Mock()
            mock_client.get_repo.return_value = mock_repo
            mock_github_class.return_value = mock_client

            items = list(adapter.fetch_items())
            assert len(items) == 1
            assert items[0].external_id == "42"

    @patch.dict("os.environ", {"GITHUB_TOKEN": "test-token"})
    def test_get_item_success(self):
        """Test getting a single item by ID."""
        config = GitHubAdapterConfig(enabled=True, repo="owner/repo")
        adapter = GitHubAdapter(config)

        with patch("github.Github") as mock_github_class:
            mock_issue = Mock()
            mock_issue.number = 99
            mock_issue.title = "Single Issue"
            mock_issue.body = ""
            mock_issue.state = "open"
            mock_issue.labels = []
            mock_issue.html_url = "https://github.com/owner/repo/issues/99"
            mock_issue.created_at = None
            mock_issue.updated_at = None
            mock_issue.assignee = None

            mock_repo = Mock()
            mock_repo.get_issue.return_value = mock_issue
            mock_client = Mock()
            mock_client.get_repo.return_value = mock_repo
            mock_github_class.return_value = mock_client

            item = adapter.get_item("99")
            assert item is not None
            assert item.external_id == "99"


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGitHubAdapterCreate:
    """Tests for GitHubAdapter create operations."""

    @patch.dict("os.environ", {"GITHUB_TOKEN": "test-token"})
    def test_create_item_basic(self):
        """Test creating a GitHub issue from requirement."""
        config = GitHubAdapterConfig(enabled=True, repo="owner/repo")
        adapter = GitHubAdapter(config)
        requirement = create_test_requirement()

        with patch("github.Github") as mock_github_class:
            mock_issue = Mock()
            mock_issue.number = 123
            mock_repo = Mock()
            mock_repo.create_issue.return_value = mock_issue
            mock_client = Mock()
            mock_client.get_repo.return_value = mock_repo
            mock_github_class.return_value = mock_client

            external_id = adapter.create_item(requirement)
            assert external_id == "123"
            mock_repo.create_issue.assert_called_once()


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGitHubAdapterUpdate:
    """Tests for GitHubAdapter update operations."""

    @patch.dict("os.environ", {"GITHUB_TOKEN": "test-token"})
    def test_update_item_success(self):
        """Test updating a GitHub issue."""
        config = GitHubAdapterConfig(enabled=True, repo="owner/repo")
        adapter = GitHubAdapter(config)
        requirement = create_test_requirement(status="COMPLETE")

        with patch("github.Github") as mock_github_class:
            mock_issue = Mock()
            mock_issue.state = "open"
            mock_repo = Mock()
            mock_repo.get_issue.return_value = mock_issue
            mock_client = Mock()
            mock_client.get_repo.return_value = mock_repo
            mock_github_class.return_value = mock_client

            result = adapter.update_item("100", requirement)
            assert result is True
            mock_issue.edit.assert_called()


# ===============================================================================
# JiraAdapter Tests
# ===============================================================================


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestJiraAdapterInit:
    """Tests for JiraAdapter initialization."""

    def test_jira_adapter_initialization(self):
        """Test JiraAdapter initializes correctly."""
        config = JiraAdapterConfig(enabled=True, server="https://jira.example.com", project="PROJ")
        adapter = JiraAdapter(config)

        assert adapter.name == "jira"
        assert adapter._config == config

    @patch.dict("os.environ", {"JIRA_API_TOKEN": "test-token", "JIRA_EMAIL": "user@example.com"})
    def test_jira_adapter_is_configured_with_credentials(self):
        """Test is_configured returns True with valid config and credentials."""
        config = JiraAdapterConfig(enabled=True, server="https://jira.example.com", project="PROJ")
        adapter = JiraAdapter(config)
        assert adapter.is_configured is True


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestJiraAdapterClient:
    """Tests for JiraAdapter client management."""

    @patch.dict("os.environ", {"JIRA_API_TOKEN": "test-token", "JIRA_EMAIL": "user@example.com"})
    def test_get_client_creates_client(self):
        """Test _get_client creates a Jira client."""
        config = JiraAdapterConfig(enabled=True, server="https://jira.example.com", project="PROJ")
        adapter = JiraAdapter(config)

        with patch("jira.JIRA") as mock_jira_class:
            client = adapter._get_client()
            mock_jira_class.assert_called_once_with(
                server="https://jira.example.com", basic_auth=("user@example.com", "test-token")
            )
            assert client == mock_jira_class.return_value

    @patch.dict("os.environ", {}, clear=True)
    def test_get_client_no_token_raises(self):
        """Test _get_client raises ValueError without token."""
        config = JiraAdapterConfig(enabled=True, server="https://jira.example.com", project="PROJ")
        adapter = JiraAdapter(config)

        with patch("jira.JIRA"):
            with pytest.raises(ValueError, match="Jira API token not found"):
                adapter._get_client()


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestJiraAdapterConnection:
    """Tests for JiraAdapter connection testing."""

    @patch.dict("os.environ", {"JIRA_API_TOKEN": "test-token", "JIRA_EMAIL": "user@example.com"})
    def test_test_connection_success(self):
        """Test test_connection succeeds with valid credentials."""
        config = JiraAdapterConfig(enabled=True, server="https://jira.example.com", project="PROJ")
        adapter = JiraAdapter(config)

        with patch("jira.JIRA") as mock_jira_class:
            mock_project = Mock()
            mock_project.name = "Test Project"
            mock_project.key = "PROJ"
            mock_client = Mock()
            mock_client.project.return_value = mock_project
            mock_jira_class.return_value = mock_client

            success, message = adapter.test_connection()
            assert success is True
            assert "Connected to Test Project (PROJ)" in message


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestJiraAdapterMapping:
    """Tests for JiraAdapter status mapping."""

    def test_map_status_to_rtmx(self):
        """Test mapping Jira status to RTMX status."""
        config = JiraAdapterConfig(
            enabled=True,
            server="https://jira.example.com",
            project="PROJ",
            status_mapping={"To Do": "MISSING"},
        )
        adapter = JiraAdapter(config)
        assert adapter.map_status_to_rtmx("To Do") == "MISSING"

    def test_map_status_from_rtmx(self):
        """Test mapping RTMX status to Jira status."""
        config = JiraAdapterConfig(
            enabled=True,
            server="https://jira.example.com",
            project="PROJ",
            status_mapping={"To Do": "MISSING"},
        )
        adapter = JiraAdapter(config)
        assert adapter.map_status_from_rtmx("MISSING") == "To Do"


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestJiraAdapterFetch:
    """Tests for JiraAdapter fetch operations."""

    @patch.dict("os.environ", {"JIRA_API_TOKEN": "test-token", "JIRA_EMAIL": "user@example.com"})
    def test_fetch_items_basic(self):
        """Test fetching items from Jira."""
        config = JiraAdapterConfig(enabled=True, server="https://jira.example.com", project="PROJ")
        adapter = JiraAdapter(config)

        with patch("jira.JIRA") as mock_jira_class:
            mock_status = Mock()
            mock_status.name = "To Do"
            mock_issue = Mock()
            mock_issue.key = "PROJ-123"
            mock_issue.fields = Mock()
            mock_issue.fields.summary = "Test Issue"
            mock_issue.fields.description = "Issue description"
            mock_issue.fields.status = mock_status
            mock_issue.fields.labels = []
            mock_issue.fields.created = "2024-01-01T12:00:00Z"
            mock_issue.fields.updated = "2024-01-02T12:00:00Z"
            mock_issue.fields.assignee = None
            mock_issue.fields.priority = None

            mock_client = Mock()
            mock_client.search_issues.return_value = [mock_issue]
            mock_jira_class.return_value = mock_client

            items = list(adapter.fetch_items())
            assert len(items) == 1
            assert items[0].external_id == "PROJ-123"

    @patch.dict("os.environ", {"JIRA_API_TOKEN": "test-token", "JIRA_EMAIL": "user@example.com"})
    def test_get_item_success(self):
        """Test getting a single item by key."""
        config = JiraAdapterConfig(enabled=True, server="https://jira.example.com", project="PROJ")
        adapter = JiraAdapter(config)

        with patch("jira.JIRA") as mock_jira_class:
            mock_status = Mock()
            mock_status.name = "Open"
            mock_issue = Mock()
            mock_issue.key = "PROJ-999"
            mock_issue.fields = Mock()
            mock_issue.fields.summary = "Single Issue"
            mock_issue.fields.description = ""
            mock_issue.fields.status = mock_status
            mock_issue.fields.labels = []
            mock_issue.fields.created = None
            mock_issue.fields.updated = None
            mock_issue.fields.assignee = None
            mock_issue.fields.priority = None

            mock_client = Mock()
            mock_client.issue.return_value = mock_issue
            mock_jira_class.return_value = mock_client

            item = adapter.get_item("PROJ-999")
            assert item is not None
            assert item.external_id == "PROJ-999"


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestJiraAdapterCreate:
    """Tests for JiraAdapter create operations."""

    @patch.dict("os.environ", {"JIRA_API_TOKEN": "test-token", "JIRA_EMAIL": "user@example.com"})
    def test_create_item_basic(self):
        """Test creating a Jira issue from requirement."""
        config = JiraAdapterConfig(enabled=True, server="https://jira.example.com", project="PROJ")
        adapter = JiraAdapter(config)
        requirement = create_test_requirement()

        with patch("jira.JIRA") as mock_jira_class:
            mock_issue = Mock()
            mock_issue.key = "PROJ-456"
            mock_client = Mock()
            mock_client.create_issue.return_value = mock_issue
            mock_jira_class.return_value = mock_client

            external_id = adapter.create_item(requirement)
            assert external_id == "PROJ-456"
            mock_client.create_issue.assert_called_once()


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestJiraAdapterUpdate:
    """Tests for JiraAdapter update operations."""

    @patch.dict("os.environ", {"JIRA_API_TOKEN": "test-token", "JIRA_EMAIL": "user@example.com"})
    def test_update_item_success(self):
        """Test updating a Jira issue."""
        config = JiraAdapterConfig(enabled=True, server="https://jira.example.com", project="PROJ")
        adapter = JiraAdapter(config)
        requirement = create_test_requirement(status="PARTIAL")

        with patch("jira.JIRA") as mock_jira_class:
            mock_status = Mock()
            mock_status.name = "In Progress"
            mock_issue = Mock()
            mock_issue.fields = Mock()
            mock_issue.fields.status = mock_status
            mock_client = Mock()
            mock_client.issue.return_value = mock_issue
            mock_jira_class.return_value = mock_client

            result = adapter.update_item("PROJ-100", requirement)
            assert result is True
            mock_issue.update.assert_called_once()


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGitHubAdapterHelpers:
    """Tests for GitHubAdapter helper methods."""

    def test_extract_priority_from_labels(self):
        """Test _extract_priority extracts priority from labels."""
        config = GitHubAdapterConfig(enabled=True, repo="owner/repo")
        adapter = GitHubAdapter(config)

        mock_label = Mock()
        mock_label.name = "priority:high"
        mock_issue = Mock()
        mock_issue.labels = [mock_label]

        priority = adapter._extract_priority(mock_issue)
        assert priority == "HIGH"

    def test_get_status_label(self):
        """Test _get_status_label returns correct label for status."""
        config = GitHubAdapterConfig(enabled=True, repo="owner/repo")
        adapter = GitHubAdapter(config)

        assert adapter._get_status_label("MISSING") == "status:todo"
        assert adapter._get_status_label("COMPLETE") == "status:done"


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestJiraAdapterHelpers:
    """Tests for JiraAdapter helper methods."""

    @patch.dict("os.environ", {"JIRA_API_TOKEN": "test-token", "JIRA_EMAIL": "user@example.com"})
    def test_transition_issue_success(self):
        """Test _transition_issue successfully transitions issue."""
        config = JiraAdapterConfig(enabled=True, server="https://jira.example.com", project="PROJ")
        adapter = JiraAdapter(config)

        with patch("jira.JIRA") as mock_jira_class:
            mock_issue = Mock()
            mock_transition = {"id": "21", "to": {"name": "Done"}}
            mock_client = Mock()
            mock_client.transitions.return_value = [mock_transition]
            mock_jira_class.return_value = mock_client

            result = adapter._transition_issue(mock_client, mock_issue, "Done")
            assert result is True
            mock_client.transition_issue.assert_called_once_with(mock_issue, "21")


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGitHubAdapterIssueConversion:
    """Tests for GitHubAdapter issue to ExternalItem conversion."""

    def test_issue_to_item_with_requirement_id(self):
        """Test _issue_to_item extracts requirement ID from body."""
        config = GitHubAdapterConfig(enabled=True, repo="owner/repo")
        adapter = GitHubAdapter(config)

        mock_label = Mock()
        mock_label.name = "bug"
        mock_issue = Mock()
        mock_issue.number = 42
        mock_issue.title = "Test Issue"
        mock_issue.body = "Issue body\n\nRTMX: REQ-SW-001"
        mock_issue.state = "open"
        mock_issue.labels = [mock_label]
        mock_issue.html_url = "https://github.com/owner/repo/issues/42"
        mock_issue.created_at = None
        mock_issue.updated_at = None
        mock_issue.assignee = None

        item = adapter._issue_to_item(mock_issue)
        assert item.requirement_id == "REQ-SW-001"


@pytest.mark.req("REQ-ADAPTER-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestJiraAdapterIssueConversion:
    """Tests for JiraAdapter issue to ExternalItem conversion."""

    def test_issue_to_item_with_assignee_and_priority(self):
        """Test _issue_to_item includes assignee and priority."""
        config = JiraAdapterConfig(enabled=True, server="https://jira.example.com", project="PROJ")
        adapter = JiraAdapter(config)

        mock_status = Mock()
        mock_status.name = "Open"
        mock_assignee = Mock()
        mock_assignee.displayName = "John Doe"
        mock_priority = Mock()
        mock_priority.name = "High"
        mock_issue = Mock()
        mock_issue.key = "PROJ-456"
        mock_issue.fields = Mock()
        mock_issue.fields.summary = "Test Issue"
        mock_issue.fields.description = ""
        mock_issue.fields.status = mock_status
        mock_issue.fields.labels = ["bug", "urgent"]
        mock_issue.fields.created = None
        mock_issue.fields.updated = None
        mock_issue.fields.assignee = mock_assignee
        mock_issue.fields.priority = mock_priority

        item = adapter._issue_to_item(mock_issue)
        assert item.assignee == "John Doe"
        assert item.priority == "High"
        assert item.labels == ["bug", "urgent"]
