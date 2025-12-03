"""Base adapter interface for external services.

Defines the abstract interface that all service adapters must implement.
"""

from __future__ import annotations

from abc import ABC, abstractmethod
from collections.abc import Iterator
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from rtmx.models import Requirement


class SyncDirection(Enum):
    """Direction of sync operation."""

    IMPORT = "import"
    EXPORT = "export"
    BIDIRECTIONAL = "bidirectional"


class ConflictResolution(Enum):
    """How to resolve conflicts during sync."""

    MANUAL = "manual"
    PREFER_LOCAL = "prefer-local"
    PREFER_REMOTE = "prefer-remote"


@dataclass
class ExternalItem:
    """Represents an item from an external service.

    This is the common format used to exchange data between
    RTMX and external services like GitHub Issues or Jira.
    """

    external_id: str
    title: str
    description: str = ""
    status: str = "open"
    labels: list[str] = field(default_factory=list)
    url: str = ""
    created_at: datetime | None = None
    updated_at: datetime | None = None
    assignee: str | None = None
    priority: str | None = None

    # Mapping to RTMX fields
    requirement_id: str | None = None  # If already linked to a requirement

    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "external_id": self.external_id,
            "title": self.title,
            "description": self.description,
            "status": self.status,
            "labels": self.labels,
            "url": self.url,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
            "assignee": self.assignee,
            "priority": self.priority,
            "requirement_id": self.requirement_id,
        }


@dataclass
class SyncResult:
    """Result of a sync operation."""

    created: list[str] = field(default_factory=list)
    updated: list[str] = field(default_factory=list)
    skipped: list[str] = field(default_factory=list)
    conflicts: list[tuple[str, str]] = field(default_factory=list)  # (id, reason)
    errors: list[tuple[str, str]] = field(default_factory=list)  # (id, error)

    @property
    def success(self) -> bool:
        """Check if sync completed without errors."""
        return len(self.errors) == 0

    @property
    def summary(self) -> str:
        """Get human-readable summary."""
        parts = []
        if self.created:
            parts.append(f"{len(self.created)} created")
        if self.updated:
            parts.append(f"{len(self.updated)} updated")
        if self.skipped:
            parts.append(f"{len(self.skipped)} skipped")
        if self.conflicts:
            parts.append(f"{len(self.conflicts)} conflicts")
        if self.errors:
            parts.append(f"{len(self.errors)} errors")
        return ", ".join(parts) if parts else "no changes"


class ServiceAdapter(ABC):
    """Abstract base class for external service adapters.

    All service adapters (GitHub, Jira, etc.) must implement this interface.
    """

    @property
    @abstractmethod
    def name(self) -> str:
        """Return the adapter name (e.g., 'github', 'jira')."""
        ...

    @property
    @abstractmethod
    def is_configured(self) -> bool:
        """Check if the adapter is properly configured."""
        ...

    @abstractmethod
    def test_connection(self) -> tuple[bool, str]:
        """Test the connection to the external service.

        Returns:
            Tuple of (success, message)
        """
        ...

    @abstractmethod
    def fetch_items(self, query: dict | None = None) -> Iterator[ExternalItem]:
        """Fetch items from the external service.

        Args:
            query: Optional filter/query parameters

        Yields:
            ExternalItem instances
        """
        ...

    @abstractmethod
    def get_item(self, external_id: str) -> ExternalItem | None:
        """Get a single item by its external ID.

        Args:
            external_id: The ID in the external system

        Returns:
            ExternalItem or None if not found
        """
        ...

    @abstractmethod
    def create_item(self, requirement: Requirement) -> str:
        """Create an item in the external service from a requirement.

        Args:
            requirement: The requirement to export

        Returns:
            The external ID of the created item
        """
        ...

    @abstractmethod
    def update_item(self, external_id: str, requirement: Requirement) -> bool:
        """Update an existing item in the external service.

        Args:
            external_id: The ID in the external system
            requirement: The requirement with updated data

        Returns:
            True if update succeeded
        """
        ...

    @abstractmethod
    def map_status_to_rtmx(self, external_status: str) -> str:
        """Map external status to RTMX status.

        Args:
            external_status: Status from external service

        Returns:
            RTMX status string
        """
        ...

    @abstractmethod
    def map_status_from_rtmx(self, rtmx_status: str) -> str:
        """Map RTMX status to external status.

        Args:
            rtmx_status: RTMX status string

        Returns:
            External service status
        """
        ...
