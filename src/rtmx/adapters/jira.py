"""Jira adapter.

Sync requirements with Jira tickets.
"""

from __future__ import annotations

import os
from collections.abc import Iterator
from datetime import datetime
from typing import TYPE_CHECKING

from rtmx.adapters.base import ExternalItem, ServiceAdapter

if TYPE_CHECKING:
    from jira import JIRA
    from jira.resources import Issue

    from rtmx.config import JiraAdapterConfig
    from rtmx.models import Requirement


class JiraAdapter(ServiceAdapter):
    """Adapter for Jira.

    Provides bidirectional sync between RTMX requirements and Jira tickets.
    """

    def __init__(self, config: JiraAdapterConfig) -> None:
        """Initialize Jira adapter.

        Args:
            config: Jira adapter configuration
        """
        self._config = config
        self._client: JIRA | None = None

    @property
    def name(self) -> str:
        """Return adapter name."""
        return "jira"

    @property
    def is_configured(self) -> bool:
        """Check if adapter is properly configured."""
        if not self._config.enabled:
            return False
        if not self._config.server:
            return False
        if not self._config.project:
            return False
        # Check for token
        token = os.environ.get(self._config.token_env, "")
        email = os.environ.get(self._config.email_env, "")
        return bool(token and email)

    def _get_client(self) -> JIRA:
        """Get or create Jira client."""
        if self._client is None:
            try:
                from jira import JIRA
            except ImportError as e:
                raise ImportError(
                    "jira package is required for Jira integration. "
                    "Install with: pip install rtmx[jira]"
                ) from e

            token = os.environ.get(self._config.token_env, "")
            email = os.environ.get(self._config.email_env, "")

            if not token:
                raise ValueError(
                    f"Jira API token not found. Set {self._config.token_env} environment variable."
                )
            if not email:
                raise ValueError(
                    f"Jira email not found. Set {self._config.email_env} environment variable."
                )

            self._client = JIRA(
                server=self._config.server,
                basic_auth=(email, token),
            )

        return self._client

    def test_connection(self) -> tuple[bool, str]:
        """Test connection to Jira."""
        try:
            client = self._get_client()
            # Try to get project info
            project = client.project(self._config.project)
            return True, f"Connected to {project.name} ({project.key})"
        except ImportError as e:
            return False, str(e)
        except Exception as e:
            return False, f"Connection failed: {e}"

    def _issue_to_item(self, issue: Issue) -> ExternalItem:
        """Convert Jira Issue to ExternalItem."""
        # Check if issue has RTMX requirement ID in description
        requirement_id = None
        description = issue.fields.description or ""
        if description:
            import re

            match = re.search(r"(?:RTMX:|REQ-)\s*(REQ-[A-Z]{2}-\d{3})", description)
            if match:
                requirement_id = match.group(1)

        # Parse dates
        created_at = None
        if issue.fields.created:
            created_at = datetime.fromisoformat(issue.fields.created.replace("Z", "+00:00"))

        updated_at = None
        if issue.fields.updated:
            updated_at = datetime.fromisoformat(issue.fields.updated.replace("Z", "+00:00"))

        # Get labels
        labels = []
        if hasattr(issue.fields, "labels") and issue.fields.labels:
            labels = list(issue.fields.labels)

        # Get priority
        priority = None
        if hasattr(issue.fields, "priority") and issue.fields.priority:
            priority = issue.fields.priority.name

        # Get assignee
        assignee = None
        if hasattr(issue.fields, "assignee") and issue.fields.assignee:
            assignee = issue.fields.assignee.displayName

        return ExternalItem(
            external_id=issue.key,
            title=issue.fields.summary,
            description=description,
            status=issue.fields.status.name if issue.fields.status else "Open",
            labels=labels,
            url=f"{self._config.server}/browse/{issue.key}",
            created_at=created_at,
            updated_at=updated_at,
            assignee=assignee,
            priority=priority,
            requirement_id=requirement_id,
        )

    def fetch_items(self, query: dict | None = None) -> Iterator[ExternalItem]:
        """Fetch issues from Jira.

        Args:
            query: Optional filter parameters:
                - jql: Custom JQL query
                - status: Filter by status
                - labels: Filter by labels

        Yields:
            ExternalItem for each matching issue
        """
        client = self._get_client()

        # Build JQL query
        if query and "jql" in query:
            jql = query["jql"]
        else:
            jql_parts = [f"project = {self._config.project}"]

            if self._config.issue_type:
                jql_parts.append(f"issuetype = '{self._config.issue_type}'")

            if query:
                if "status" in query:
                    jql_parts.append(f"status = '{query['status']}'")
                if "labels" in query:
                    for label in query["labels"]:
                        jql_parts.append(f"labels = '{label}'")

            jql = " AND ".join(jql_parts)

        # Fetch issues
        start_at = 0
        max_results = 50

        while True:
            issues = client.search_issues(
                jql,
                startAt=start_at,
                maxResults=max_results,
            )

            if not issues:
                break

            for issue in issues:
                yield self._issue_to_item(issue)

            if len(issues) < max_results:
                break

            start_at += max_results

    def get_item(self, external_id: str) -> ExternalItem | None:
        """Get a single issue by key."""
        try:
            client = self._get_client()
            issue = client.issue(external_id)
            return self._issue_to_item(issue)
        except Exception:
            return None

    def create_item(self, requirement: Requirement) -> str:
        """Create a Jira issue from a requirement.

        Args:
            requirement: The requirement to export

        Returns:
            Issue key (e.g., "PROJ-123")
        """
        client = self._get_client()

        # Build description
        desc_parts = [requirement.text]

        if requirement.rationale:
            desc_parts.append(f"\nh2. Rationale\n{requirement.rationale}")

        if requirement.acceptance:
            desc_parts.append(f"\nh2. Acceptance Criteria\n{requirement.acceptance}")

        # Add RTMX tracking
        desc_parts.append(f"\n----\nRTMX: {requirement.id}")

        description = "\n".join(desc_parts)

        # Build issue fields
        issue_dict = {
            "project": {"key": self._config.project},
            "summary": f"[{requirement.id}] {requirement.text[:80]}",
            "description": description,
            "issuetype": {"name": self._config.issue_type or "Task"},
        }

        # Add labels if configured
        if self._config.labels:
            issue_dict["labels"] = list(self._config.labels)

        # Create issue
        issue = client.create_issue(fields=issue_dict)

        return issue.key

    def update_item(self, external_id: str, requirement: Requirement) -> bool:
        """Update a Jira issue from a requirement.

        Args:
            external_id: Issue key
            requirement: Updated requirement data

        Returns:
            True if update succeeded
        """
        try:
            client = self._get_client()
            issue = client.issue(external_id)

            # Build updated description
            desc_parts = [requirement.text]

            if requirement.rationale:
                desc_parts.append(f"\nh2. Rationale\n{requirement.rationale}")

            if requirement.acceptance:
                desc_parts.append(f"\nh2. Acceptance Criteria\n{requirement.acceptance}")

            desc_parts.append(f"\n----\nRTMX: {requirement.id}")

            # Update issue
            issue.update(
                summary=f"[{requirement.id}] {requirement.text[:80]}",
                description="\n".join(desc_parts),
            )

            # Transition status if needed
            target_status = self.map_status_from_rtmx(requirement.status)
            current_status = issue.fields.status.name

            if target_status != current_status:
                self._transition_issue(client, issue, target_status)

            return True

        except Exception:
            return False

    def _transition_issue(self, client: JIRA, issue: Issue, target_status: str) -> bool:
        """Transition an issue to a new status."""
        try:
            # Get available transitions
            transitions = client.transitions(issue)

            for transition in transitions:
                if transition["to"]["name"].lower() == target_status.lower():
                    client.transition_issue(issue, transition["id"])
                    return True

            return False
        except Exception:
            return False

    def map_status_to_rtmx(self, external_status: str) -> str:
        """Map Jira status to RTMX status."""
        mapping = self._config.status_mapping
        return mapping.get(external_status, "MISSING")

    def map_status_from_rtmx(self, rtmx_status: str) -> str:
        """Map RTMX status to Jira status."""
        # Reverse the status mapping
        reverse_mapping = {v: k for k, v in self._config.status_mapping.items()}
        return reverse_mapping.get(rtmx_status, "Open")
