"""GitHub Issues adapter.

Sync requirements with GitHub Issues.
"""

from __future__ import annotations

import os
from collections.abc import Iterator
from datetime import timezone
from typing import TYPE_CHECKING

from rtmx.adapters.base import ExternalItem, ServiceAdapter

if TYPE_CHECKING:
    from github import Github
    from github.Issue import Issue

    from rtmx.config import GitHubAdapterConfig
    from rtmx.models import Requirement


class GitHubAdapter(ServiceAdapter):
    """Adapter for GitHub Issues.

    Provides bidirectional sync between RTMX requirements and GitHub Issues.
    """

    def __init__(self, config: GitHubAdapterConfig) -> None:
        """Initialize GitHub adapter.

        Args:
            config: GitHub adapter configuration
        """
        self._config = config
        self._client: Github | None = None
        self._repo = None

    @property
    def name(self) -> str:
        """Return adapter name."""
        return "github"

    @property
    def is_configured(self) -> bool:
        """Check if adapter is properly configured."""
        if not self._config.enabled:
            return False
        if not self._config.repo:
            return False
        # Check for token
        token = os.environ.get(self._config.token_env, "")
        return bool(token)

    def _get_client(self) -> Github:
        """Get or create GitHub client."""
        if self._client is None:
            try:
                from github import Github
            except ImportError as e:
                raise ImportError(
                    "PyGithub is required for GitHub integration. "
                    "Install with: pip install rtmx[github]"
                ) from e

            token = os.environ.get(self._config.token_env, "")
            if not token:
                raise ValueError(
                    f"GitHub token not found. Set {self._config.token_env} environment variable."
                )

            self._client = Github(token)

        return self._client

    def _get_repo(self):
        """Get repository object."""
        if self._repo is None:
            client = self._get_client()
            self._repo = client.get_repo(self._config.repo)
        return self._repo

    def test_connection(self) -> tuple[bool, str]:
        """Test connection to GitHub."""
        try:
            repo = self._get_repo()
            return True, f"Connected to {repo.full_name}"
        except ImportError as e:
            return False, str(e)
        except Exception as e:
            return False, f"Connection failed: {e}"

    def _issue_to_item(self, issue: Issue) -> ExternalItem:
        """Convert GitHub Issue to ExternalItem."""
        # Check if issue has RTMX requirement ID in body
        requirement_id = None
        if issue.body:
            # Look for pattern like "RTMX: REQ-XX-NNN" or "[REQ-XX-NNN]"
            import re

            match = re.search(r"(?:RTMX:|REQ-)\s*(REQ-[A-Z]{2}-\d{3})", issue.body)
            if match:
                requirement_id = match.group(1)

        return ExternalItem(
            external_id=str(issue.number),
            title=issue.title,
            description=issue.body or "",
            status=issue.state,
            labels=[label.name for label in issue.labels],
            url=issue.html_url,
            created_at=issue.created_at.replace(tzinfo=timezone.utc) if issue.created_at else None,
            updated_at=issue.updated_at.replace(tzinfo=timezone.utc) if issue.updated_at else None,
            assignee=issue.assignee.login if issue.assignee else None,
            priority=self._extract_priority(issue),
            requirement_id=requirement_id,
        )

    def _extract_priority(self, issue: Issue) -> str | None:
        """Extract priority from issue labels."""
        priority_labels = {
            "priority:critical": "CRITICAL",
            "priority:high": "HIGH",
            "priority:medium": "MEDIUM",
            "priority:low": "LOW",
            "P0": "CRITICAL",
            "P1": "HIGH",
            "P2": "MEDIUM",
            "P3": "LOW",
        }

        for label in issue.labels:
            if label.name in priority_labels:
                return priority_labels[label.name]

        return None

    def fetch_items(self, query: dict | None = None) -> Iterator[ExternalItem]:
        """Fetch issues from GitHub.

        Args:
            query: Optional filter parameters:
                - state: 'open', 'closed', 'all'
                - labels: list of label names
                - since: datetime to filter by update time

        Yields:
            ExternalItem for each matching issue
        """
        repo = self._get_repo()

        # Build query parameters
        state = "all"
        labels = self._config.labels or []

        if query:
            state = query.get("state", state)
            if "labels" in query:
                labels = query["labels"]

        # Fetch issues
        if labels:
            issues = repo.get_issues(state=state, labels=labels)
        else:
            issues = repo.get_issues(state=state)

        for issue in issues:
            # Skip pull requests (they show up in issues API)
            if issue.pull_request is not None:
                continue
            yield self._issue_to_item(issue)

    def get_item(self, external_id: str) -> ExternalItem | None:
        """Get a single issue by number."""
        try:
            repo = self._get_repo()
            issue = repo.get_issue(int(external_id))
            return self._issue_to_item(issue)
        except Exception:
            return None

    def create_item(self, requirement: Requirement) -> str:
        """Create a GitHub issue from a requirement.

        Args:
            requirement: The requirement to export

        Returns:
            Issue number as string
        """
        repo = self._get_repo()

        # Build issue body
        body_parts = [requirement.text]

        if requirement.rationale:
            body_parts.append(f"\n## Rationale\n{requirement.rationale}")

        if requirement.acceptance:
            body_parts.append(f"\n## Acceptance Criteria\n{requirement.acceptance}")

        # Add RTMX tracking
        body_parts.append(f"\n---\nRTMX: {requirement.id}")

        body = "\n".join(body_parts)

        # Determine labels
        labels = list(self._config.labels) if self._config.labels else []

        # Add status label if configured
        status_label = self._get_status_label(requirement.status)
        if status_label:
            labels.append(status_label)

        # Create issue
        issue = repo.create_issue(
            title=f"[{requirement.id}] {requirement.text[:80]}",
            body=body,
            labels=labels,
        )

        return str(issue.number)

    def update_item(self, external_id: str, requirement: Requirement) -> bool:
        """Update a GitHub issue from a requirement.

        Args:
            external_id: Issue number
            requirement: Updated requirement data

        Returns:
            True if update succeeded
        """
        try:
            repo = self._get_repo()
            issue = repo.get_issue(int(external_id))

            # Build updated body
            body_parts = [requirement.text]

            if requirement.rationale:
                body_parts.append(f"\n## Rationale\n{requirement.rationale}")

            if requirement.acceptance:
                body_parts.append(f"\n## Acceptance Criteria\n{requirement.acceptance}")

            body_parts.append(f"\n---\nRTMX: {requirement.id}")

            # Update issue
            issue.edit(
                title=f"[{requirement.id}] {requirement.text[:80]}",
                body="\n".join(body_parts),
            )

            # Update state based on status
            target_state = self.map_status_from_rtmx(requirement.status)
            if target_state == "closed" and issue.state == "open":
                issue.edit(state="closed")
            elif target_state == "open" and issue.state == "closed":
                issue.edit(state="open")

            return True

        except Exception:
            return False

    def _get_status_label(self, status: str) -> str | None:
        """Get label for RTMX status."""
        status_labels = {
            "MISSING": "status:todo",
            "PARTIAL": "status:in-progress",
            "COMPLETE": "status:done",
        }
        return status_labels.get(status)

    def map_status_to_rtmx(self, external_status: str) -> str:
        """Map GitHub issue state to RTMX status."""
        mapping = self._config.status_mapping
        return mapping.get(external_status, "MISSING")

    def map_status_from_rtmx(self, rtmx_status: str) -> str:
        """Map RTMX status to GitHub issue state."""
        # Reverse the status mapping
        reverse_mapping = {v: k for k, v in self._config.status_mapping.items()}
        return reverse_mapping.get(rtmx_status, "open")
