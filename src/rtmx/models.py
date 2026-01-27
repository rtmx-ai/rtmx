"""Core data models for RTMX.

This module provides the fundamental data structures for requirements traceability:
- Status: Requirement completion status
- Priority: Requirement priority level
- Requirement: Single requirement representation
- RTMDatabase: Collection of requirements with query/modification operations
"""

from __future__ import annotations

import contextlib
import sys
from collections.abc import Iterator
from dataclasses import dataclass, field
from enum import Enum
from pathlib import Path
from typing import TYPE_CHECKING, Any

if sys.version_info >= (3, 11):
    from typing import Self
else:
    from typing_extensions import Self

if TYPE_CHECKING:
    from rtmx.config import RTMXConfig
    from rtmx.graph import DependencyGraph


class RTMError(Exception):
    """Base exception for RTM-related errors."""

    pass


class RequirementNotFoundError(RTMError):
    """Raised when a requirement ID is not found in the RTM."""

    pass


class RTMValidationError(RTMError):
    """Raised when RTM data fails validation."""

    pass


class Status(str, Enum):
    """Requirement completion status."""

    COMPLETE = "COMPLETE"
    PARTIAL = "PARTIAL"
    MISSING = "MISSING"
    NOT_STARTED = "NOT_STARTED"

    @classmethod
    def from_string(cls, value: str) -> Status:
        """Parse status from string, handling variations."""
        normalized = value.strip().upper().replace("-", "_").replace(" ", "_")
        for member in cls:
            if member.value == normalized:
                return member
        # Default to MISSING for unknown values
        return cls.MISSING


class Priority(str, Enum):
    """Requirement priority level."""

    P0 = "P0"  # Critical
    HIGH = "HIGH"
    MEDIUM = "MEDIUM"
    LOW = "LOW"

    @classmethod
    def from_string(cls, value: str) -> Priority:
        """Parse priority from string, handling variations."""
        normalized = value.strip().upper()
        if normalized in ("P0", "CRITICAL"):
            return cls.P0
        for member in cls:
            if member.value == normalized:
                return member
        return cls.MEDIUM


class Visibility(str, Enum):
    """Visibility level for cross-repo requirements."""

    FULL = "full"  # Full access to requirement details
    SHADOW = "shadow"  # Limited access: status, hash, dependencies only
    HASH_ONLY = "hash_only"  # Minimal: only hash for verification


@dataclass
class ShadowRequirement:
    """Representation of a requirement from an external repository with partial visibility.

    When a user depends on a requirement in a repository they don't have full access to,
    RTMX provides a shadow view with limited information. This enables cross-repo dependency
    tracking across trust boundaries without exposing sensitive requirement details.

    Attributes:
        req_id: Requirement identifier in the external repository
        external_repo: Full repository path (e.g., "rtmx-ai/rtmx-sync")
        shadow_hash: SHA-256 hash of the requirement content for verification
        status: Current completion status (if visible)
        visibility: Level of access (full, shadow, hash_only)
        verified_at: ISO timestamp of last verification
        cached_dependencies: Visible dependency IDs (may be empty for hash_only)
    """

    req_id: str
    external_repo: str
    shadow_hash: str
    status: Status = Status.MISSING
    visibility: Visibility = Visibility.SHADOW
    verified_at: str = ""
    cached_dependencies: set[str] = field(default_factory=set)

    @property
    def is_accessible(self) -> bool:
        """Check if requirement details are accessible."""
        return self.visibility == Visibility.FULL

    @property
    def is_verifiable(self) -> bool:
        """Check if requirement can be verified via hash."""
        return bool(self.shadow_hash)

    @property
    def full_ref(self) -> str:
        """Get full cross-repo reference string."""
        return f"{self.external_repo}:{self.req_id}"

    def to_dict(self) -> dict[str, Any]:
        """Convert shadow requirement to dictionary for serialization."""
        return {
            "req_id": self.req_id,
            "external_repo": self.external_repo,
            "shadow_hash": self.shadow_hash,
            "status": self.status.value,
            "visibility": self.visibility.value,
            "verified_at": self.verified_at,
            "cached_dependencies": sorted(self.cached_dependencies),
        }

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> ShadowRequirement:
        """Create shadow requirement from dictionary."""
        return cls(
            req_id=str(data.get("req_id", "")),
            external_repo=str(data.get("external_repo", "")),
            shadow_hash=str(data.get("shadow_hash", "")),
            status=Status.from_string(str(data.get("status", "MISSING"))),
            visibility=Visibility(data.get("visibility", "shadow")),
            verified_at=str(data.get("verified_at", "")),
            cached_dependencies=set(data.get("cached_dependencies", [])),
        )

    @classmethod
    def from_requirement(
        cls,
        req: Requirement,
        external_repo: str,
        visibility: Visibility = Visibility.SHADOW,
    ) -> ShadowRequirement:
        """Create shadow requirement from a full requirement.

        Used when creating a shadow view for sharing across trust boundaries.

        Args:
            req: Full requirement to create shadow from
            external_repo: Repository path for this requirement
            visibility: Access level for the shadow

        Returns:
            ShadowRequirement with appropriate visibility restrictions
        """
        import hashlib
        from datetime import datetime

        # Create content hash from key fields
        content = f"{req.req_id}:{req.status.value}:{req.requirement_text}"
        shadow_hash = hashlib.sha256(content.encode()).hexdigest()[:16]

        return cls(
            req_id=req.req_id,
            external_repo=external_repo,
            shadow_hash=shadow_hash,
            status=req.status,
            visibility=visibility,
            verified_at=datetime.now().isoformat(),
            cached_dependencies=req.dependencies if visibility != Visibility.HASH_ONLY else set(),
        )


class DelegationRole(str, Enum):
    """Roles that can be delegated between repositories."""

    DEPENDENCY_VIEWER = "dependency_viewer"  # Can see deps and status
    REQUIREMENT_READER = "requirement_reader"  # Can read requirement details
    REQUIREMENT_EDITOR = "requirement_editor"  # Can modify requirements
    SHADOW_VIEWER = "shadow_viewer"  # Can only see shadow/hash


@dataclass
class GrantConstraint:
    """Constraints on a grant delegation.

    Limits what requirements or categories a delegation applies to.

    Attributes:
        categories: Limit to specific requirement categories
        requirement_ids: Limit to specific requirement IDs
        exclude_categories: Explicitly excluded categories
        expires_at: Optional expiration timestamp (ISO format)
    """

    categories: set[str] = field(default_factory=set)
    requirement_ids: set[str] = field(default_factory=set)
    exclude_categories: set[str] = field(default_factory=set)
    expires_at: str = ""

    @property
    def is_expired(self) -> bool:
        """Check if constraint has expired."""
        if not self.expires_at:
            return False
        from datetime import datetime

        try:
            expiry = datetime.fromisoformat(self.expires_at)
            return datetime.now() >= expiry
        except ValueError:
            return False

    def allows_requirement(self, req_id: str, category: str) -> bool:
        """Check if constraint allows access to a requirement.

        Args:
            req_id: Requirement identifier
            category: Requirement category

        Returns:
            True if access is allowed under this constraint
        """
        if self.is_expired:
            return False

        # Check exclusions first
        if category in self.exclude_categories:
            return False

        # If specific IDs listed, must match
        if self.requirement_ids and req_id not in self.requirement_ids:
            return False

        # If specific categories listed, must match
        return not (self.categories and category not in self.categories)

    def to_dict(self) -> dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "categories": sorted(self.categories),
            "requirement_ids": sorted(self.requirement_ids),
            "exclude_categories": sorted(self.exclude_categories),
            "expires_at": self.expires_at,
        }

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> GrantConstraint:
        """Create from dictionary."""
        return cls(
            categories=set(data.get("categories", [])),
            requirement_ids=set(data.get("requirement_ids", [])),
            exclude_categories=set(data.get("exclude_categories", [])),
            expires_at=data.get("expires_at", ""),
        )


@dataclass
class GrantDelegation:
    """Delegation of access from one repository to another.

    Enables controlled sharing of requirements across trust boundaries.
    Follows the Zitadel project grant pattern.

    Attributes:
        grantor: Repository granting access (e.g., "rtmx-ai/rtmx")
        grantee: Repository receiving access (e.g., "rtmx-ai/rtmx-sync")
        roles_delegated: Set of roles being delegated
        constraints: Optional constraints on the delegation
        created_at: When delegation was created (ISO format)
        created_by: User who created the delegation
        active: Whether delegation is currently active
    """

    grantor: str
    grantee: str
    roles_delegated: set[DelegationRole] = field(default_factory=set)
    constraints: GrantConstraint = field(default_factory=GrantConstraint)
    created_at: str = ""
    created_by: str = ""
    active: bool = True

    def __post_init__(self) -> None:
        """Set creation timestamp if not provided."""
        if not self.created_at:
            from datetime import datetime

            self.created_at = datetime.now().isoformat()

    @property
    def is_valid(self) -> bool:
        """Check if delegation is currently valid."""
        return self.active and not self.constraints.is_expired

    def has_role(self, role: DelegationRole) -> bool:
        """Check if delegation includes a specific role."""
        return role in self.roles_delegated

    def allows_access(self, req_id: str, category: str, role: DelegationRole) -> bool:
        """Check if delegation allows access to a requirement with given role.

        Args:
            req_id: Requirement identifier
            category: Requirement category
            role: Required role for access

        Returns:
            True if access is allowed
        """
        if not self.is_valid:
            return False
        if not self.has_role(role):
            return False
        return self.constraints.allows_requirement(req_id, category)

    def to_dict(self) -> dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "grantor": self.grantor,
            "grantee": self.grantee,
            "roles_delegated": [r.value for r in self.roles_delegated],
            "constraints": self.constraints.to_dict(),
            "created_at": self.created_at,
            "created_by": self.created_by,
            "active": self.active,
        }

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> GrantDelegation:
        """Create from dictionary."""
        roles = {DelegationRole(r) for r in data.get("roles_delegated", [])}
        constraints = GrantConstraint.from_dict(data.get("constraints", {}))
        return cls(
            grantor=data.get("grantor", ""),
            grantee=data.get("grantee", ""),
            roles_delegated=roles,
            constraints=constraints,
            created_at=data.get("created_at", ""),
            created_by=data.get("created_by", ""),
            active=data.get("active", True),
        )


@dataclass
class Requirement:
    """Single requirement in the RTM.

    Attributes:
        req_id: Unique identifier (e.g., REQ-SW-001)
        category: High-level grouping (e.g., SOFTWARE, MODE, PERFORMANCE)
        subcategory: Detailed classification within category
        requirement_text: Human-readable requirement description
        target_value: Quantitative acceptance criteria
        test_module: Python test file implementing validation
        test_function: Specific test function name
        validation_method: Testing approach description
        status: Current completion status
        priority: Criticality level
        phase: Development phase (1, 2, 3, etc.)
        notes: Additional context
        effort_weeks: Estimated effort in weeks
        dependencies: Set of requirement IDs this depends on
        blocks: Set of requirement IDs this blocks
        assignee: Person responsible
        sprint: Target sprint/version
        started_date: When work began (YYYY-MM-DD)
        completed_date: When completed (YYYY-MM-DD)
        requirement_file: Path to detailed specification markdown
        external_id: ID in external system (GitHub issue #, Jira key)
        extra: Additional fields not in core schema
    """

    req_id: str
    category: str = ""
    subcategory: str = ""
    requirement_text: str = ""
    target_value: str = ""
    test_module: str = ""
    test_function: str = ""
    validation_method: str = ""
    status: Status = Status.MISSING
    priority: Priority = Priority.MEDIUM
    phase: int | None = None
    notes: str = ""
    effort_weeks: float | None = None
    dependencies: set[str] = field(default_factory=set)
    blocks: set[str] = field(default_factory=set)
    assignee: str = ""
    sprint: str = ""
    started_date: str = ""
    completed_date: str = ""
    requirement_file: str = ""
    external_id: str = ""
    extra: dict[str, Any] = field(default_factory=dict)

    def has_test(self) -> bool:
        """Check if requirement has an associated test."""
        return self.test_module not in ("", "MISSING") and self.test_function not in ("", "MISSING")

    def is_complete(self) -> bool:
        """Check if requirement is fully complete."""
        return self.status == Status.COMPLETE

    def is_blocked(self, db: RTMDatabase, config: RTMXConfig | None = None) -> bool:
        """Check if requirement is blocked by incomplete dependencies.

        Args:
            db: Local RTM database
            config: Optional RTMX config for cross-repo dependency checking

        Returns:
            True if any dependency is incomplete, False otherwise
        """
        from rtmx.parser import parse_requirement_ref

        for dep_str in self.dependencies:
            ref = parse_requirement_ref(dep_str)

            if ref.is_local:
                # Local dependency - check local database
                try:
                    dep = db.get(ref.req_id)
                    if dep.status != Status.COMPLETE:
                        return True
                except RequirementNotFoundError:
                    pass
            elif config is not None:
                # Cross-repo dependency - check remote database
                alias = ref.remote_alias
                if alias is None and ref.full_repo:
                    # Try to find matching alias
                    for a, r in config.sync.remotes.items():
                        if r.repo == ref.full_repo:
                            alias = a
                            break

                if alias is None:
                    # Unknown remote - can't verify, assume not blocked
                    continue

                remote_config = config.sync.get_remote(alias)
                if remote_config is None or remote_config.path is None:
                    # Remote not accessible - can't verify, assume not blocked
                    continue

                # Try to load remote database
                remote_db_path = Path(remote_config.path) / remote_config.database
                if not remote_db_path.exists():
                    # Remote unavailable - can't verify, assume not blocked
                    continue

                try:
                    remote_db = RTMDatabase.load(remote_db_path)
                    if remote_db.exists(ref.req_id):
                        remote_req = remote_db.get(ref.req_id)
                        if remote_req.status != Status.COMPLETE:
                            return True
                except RTMError:
                    # Can't load remote - assume not blocked
                    pass

        return False

    # Convenience aliases for adapters
    @property
    def id(self) -> str:
        """Alias for req_id."""
        return self.req_id

    @property
    def text(self) -> str:
        """Alias for requirement_text."""
        return self.requirement_text

    @property
    def rationale(self) -> str:
        """Get rationale from notes or extra fields."""
        return self.extra.get("rationale", self.notes)

    @property
    def acceptance(self) -> str:
        """Get acceptance criteria from target_value or extra fields."""
        return self.extra.get("acceptance", self.target_value)

    def to_dict(self) -> dict[str, Any]:
        """Convert requirement to dictionary for serialization."""
        data = {
            "req_id": self.req_id,
            "category": self.category,
            "subcategory": self.subcategory,
            "requirement_text": self.requirement_text,
            "target_value": self.target_value,
            "test_module": self.test_module,
            "test_function": self.test_function,
            "validation_method": self.validation_method,
            "status": self.status.value,
            "priority": self.priority.value,
            "phase": self.phase if self.phase is not None else "",
            "notes": self.notes,
            "effort_weeks": self.effort_weeks if self.effort_weeks is not None else "",
            "dependencies": "|".join(sorted(self.dependencies)),
            "blocks": "|".join(sorted(self.blocks)),
            "assignee": self.assignee,
            "sprint": self.sprint,
            "started_date": self.started_date,
            "completed_date": self.completed_date,
            "requirement_file": self.requirement_file,
            "external_id": self.external_id,
        }
        # Add extra fields
        data.update(self.extra)
        return data

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> Self:
        """Create requirement from dictionary."""
        from rtmx.parser import parse_dependencies

        # Extract known fields
        req_id = str(data.get("req_id", data.get("Req_ID", "")))
        category = str(data.get("category", data.get("Category", "")))
        subcategory = str(data.get("subcategory", data.get("Subcategory", "")))

        # Parse status
        status_str = str(data.get("status", data.get("Status", "MISSING")))
        status = Status.from_string(status_str)

        # Parse priority
        priority_str = str(data.get("priority", data.get("Priority", "MEDIUM")))
        priority = Priority.from_string(priority_str)

        # Parse phase
        phase_val = data.get("phase", data.get("Phase"))
        phase: int | None = None
        if phase_val not in (None, "", "phase"):
            with contextlib.suppress(ValueError, TypeError):
                phase = int(str(phase_val))

        # Parse effort_weeks
        effort_val = data.get("effort_weeks", data.get("Effort_Weeks"))
        effort_weeks: float | None = None
        if effort_val not in (None, ""):
            with contextlib.suppress(ValueError, TypeError):
                effort_weeks = float(str(effort_val))

        # Parse dependencies and blocks
        deps_str = str(data.get("dependencies", data.get("Dependencies", "")))
        blocks_str = str(data.get("blocks", data.get("Blocks", "")))

        # Collect extra fields
        known_fields = {
            "req_id",
            "Req_ID",
            "category",
            "Category",
            "subcategory",
            "Subcategory",
            "requirement_text",
            "Requirement_Text",
            "target_value",
            "Target_Value",
            "test_module",
            "Test_Module",
            "test_function",
            "Test_Function",
            "validation_method",
            "Validation_Method",
            "status",
            "Status",
            "priority",
            "Priority",
            "phase",
            "Phase",
            "notes",
            "Notes",
            "effort_weeks",
            "Effort_Weeks",
            "dependencies",
            "Dependencies",
            "blocks",
            "Blocks",
            "assignee",
            "Assignee",
            "sprint",
            "Sprint",
            "started_date",
            "Started_Date",
            "completed_date",
            "Completed_Date",
            "requirement_file",
            "Requirement_File",
            "external_id",
            "External_ID",
        }
        extra = {k: v for k, v in data.items() if k not in known_fields}

        return cls(
            req_id=req_id,
            category=category,
            subcategory=subcategory,
            requirement_text=str(data.get("requirement_text", data.get("Requirement_Text", ""))),
            target_value=str(data.get("target_value", data.get("Target_Value", ""))),
            test_module=str(data.get("test_module", data.get("Test_Module", ""))),
            test_function=str(data.get("test_function", data.get("Test_Function", ""))),
            validation_method=str(data.get("validation_method", data.get("Validation_Method", ""))),
            status=status,
            priority=priority,
            phase=phase,
            notes=str(data.get("notes", data.get("Notes", ""))),
            effort_weeks=effort_weeks,
            dependencies=parse_dependencies(deps_str),
            blocks=parse_dependencies(blocks_str),
            assignee=str(data.get("assignee", data.get("Assignee", ""))),
            sprint=str(data.get("sprint", data.get("Sprint", ""))),
            started_date=str(data.get("started_date", data.get("Started_Date", ""))),
            completed_date=str(data.get("completed_date", data.get("Completed_Date", ""))),
            requirement_file=str(data.get("requirement_file", data.get("Requirement_File", ""))),
            external_id=str(data.get("external_id", data.get("External_ID", ""))),
            extra=extra,
        )


class RTMDatabase:
    """Collection of requirements with query and modification operations.

    The RTMDatabase is the primary interface for working with requirements data.
    It supports loading from CSV files, querying requirements, and saving changes.

    Example:
        >>> db = RTMDatabase.load("docs/rtm_database.csv")
        >>> req = db.get("REQ-SW-001")
        >>> incomplete = db.filter(status=Status.MISSING)
        >>> db.update("REQ-SW-001", status=Status.COMPLETE)
        >>> db.save()
    """

    def __init__(self, requirements: list[Requirement], path: Path | None = None) -> None:
        """Initialize RTM database.

        Args:
            requirements: List of requirements
            path: Optional path to the source CSV file
        """
        self._requirements: dict[str, Requirement] = {r.req_id: r for r in requirements}
        self._path = path
        self._graph: DependencyGraph | None = None

    @classmethod
    def load(cls, path: str | Path | None = None) -> Self:
        """Load RTM database from CSV file.

        Args:
            path: Path to CSV file. If None, searches for docs/rtm_database.csv

        Returns:
            RTMDatabase instance

        Raises:
            RTMError: If file not found or invalid
        """
        from rtmx.parser import find_rtm_database, load_csv

        resolved_path = find_rtm_database() if path is None else Path(path)

        if not resolved_path.exists():
            raise RTMError(f"RTM database not found: {resolved_path}")

        requirements = load_csv(resolved_path)
        return cls(requirements, resolved_path)

    def save(self, path: str | Path | None = None) -> None:
        """Save RTM database to CSV file.

        Args:
            path: Path to save to. If None, uses original load path.

        Raises:
            RTMError: If no path specified and database was not loaded from file
        """
        from rtmx.parser import save_csv

        save_path = Path(path) if path else self._path
        if save_path is None:
            raise RTMError("No save path specified and database was not loaded from file")

        save_csv(list(self._requirements.values()), save_path)
        self._path = save_path

    def get(self, req_id: str) -> Requirement:
        """Get requirement by ID.

        Args:
            req_id: Requirement identifier

        Returns:
            Requirement instance

        Raises:
            RequirementNotFoundError: If requirement not found
        """
        if req_id not in self._requirements:
            available = list(self._requirements.keys())[:5]
            raise RequirementNotFoundError(
                f"Requirement {req_id} not found. Available: {', '.join(available)}..."
            )
        return self._requirements[req_id]

    def exists(self, req_id: str) -> bool:
        """Check if requirement exists.

        Args:
            req_id: Requirement identifier

        Returns:
            True if requirement exists
        """
        return req_id in self._requirements

    def update(self, req_id: str, **fields: Any) -> Requirement:
        """Update requirement fields.

        Args:
            req_id: Requirement identifier
            **fields: Field name/value pairs to update

        Returns:
            Updated requirement

        Raises:
            RequirementNotFoundError: If requirement not found
        """
        req = self.get(req_id)

        for key, value in fields.items():
            if key == "status" and isinstance(value, str):
                value = Status.from_string(value)
            elif key == "priority" and isinstance(value, str):
                value = Priority.from_string(value)
            elif (
                key == "dependencies"
                and isinstance(value, str)
                or key == "blocks"
                and isinstance(value, str)
            ):
                from rtmx.parser import parse_dependencies

                value = parse_dependencies(value)

            if hasattr(req, key):
                setattr(req, key, value)
            else:
                req.extra[key] = value

        # Invalidate cached graph
        self._graph = None

        return req

    def add(self, requirement: Requirement) -> None:
        """Add a new requirement.

        Args:
            requirement: Requirement to add

        Raises:
            RTMError: If requirement ID already exists
        """
        if requirement.req_id in self._requirements:
            raise RTMError(f"Requirement {requirement.req_id} already exists")
        self._requirements[requirement.req_id] = requirement
        self._graph = None

    def remove(self, req_id: str) -> Requirement:
        """Remove a requirement.

        Args:
            req_id: Requirement identifier

        Returns:
            Removed requirement

        Raises:
            RequirementNotFoundError: If requirement not found
        """
        if req_id not in self._requirements:
            raise RequirementNotFoundError(f"Requirement {req_id} not found")
        req = self._requirements.pop(req_id)
        self._graph = None
        return req

    def filter(
        self,
        *,
        status: Status | None = None,
        priority: Priority | None = None,
        category: str | None = None,
        subcategory: str | None = None,
        phase: int | None = None,
        has_test: bool | None = None,
    ) -> list[Requirement]:
        """Filter requirements by criteria.

        Args:
            status: Filter by status
            priority: Filter by priority
            category: Filter by category
            subcategory: Filter by subcategory
            phase: Filter by phase
            has_test: Filter by test presence

        Returns:
            List of matching requirements
        """
        results = list(self._requirements.values())

        if status is not None:
            results = [r for r in results if r.status == status]
        if priority is not None:
            results = [r for r in results if r.priority == priority]
        if category is not None:
            results = [r for r in results if r.category == category]
        if subcategory is not None:
            results = [r for r in results if r.subcategory == subcategory]
        if phase is not None:
            results = [r for r in results if r.phase == phase]
        if has_test is not None:
            results = [r for r in results if r.has_test() == has_test]

        return results

    def all(self) -> list[Requirement]:
        """Get all requirements.

        Returns:
            List of all requirements
        """
        return list(self._requirements.values())

    def __len__(self) -> int:
        """Return number of requirements."""
        return len(self._requirements)

    def __iter__(self) -> Iterator[Requirement]:
        """Iterate over requirements."""
        return iter(self._requirements.values())

    def __contains__(self, req_id: str) -> bool:
        """Check if requirement exists."""
        return req_id in self._requirements

    # Graph operations (delegate to DependencyGraph)

    def _get_graph(self) -> DependencyGraph:
        """Get or create dependency graph."""
        if self._graph is None:
            from rtmx.graph import DependencyGraph

            self._graph = DependencyGraph.from_database(self)
        return self._graph

    def find_cycles(self) -> list[list[str]]:
        """Find circular dependency cycles.

        Returns:
            List of cycles, where each cycle is a list of requirement IDs
        """
        return self._get_graph().find_cycles()

    def transitive_blocks(self, req_id: str) -> set[str]:
        """Get all requirements transitively blocked by a requirement.

        Args:
            req_id: Requirement identifier

        Returns:
            Set of blocked requirement IDs
        """
        return self._get_graph().transitive_blocks(req_id)

    def critical_path(self) -> list[str]:
        """Get critical path through dependency graph.

        Returns:
            List of requirement IDs on critical path
        """
        return self._get_graph().critical_path()

    # Validation operations (delegate to validation module)

    def validate(self) -> list[str]:
        """Validate RTM structure and data.

        Returns:
            List of validation error messages (empty if valid)
        """
        from rtmx.validation import validate_schema

        return validate_schema(self)

    def check_reciprocity(self) -> list[tuple[str, str, str]]:
        """Check dependency/blocks reciprocity.

        Returns:
            List of (req_id, related_id, issue) tuples
        """
        from rtmx.validation import check_reciprocity

        return check_reciprocity(self)

    def fix_reciprocity(self) -> int:
        """Fix dependency/blocks reciprocity violations.

        Returns:
            Number of violations fixed
        """
        from rtmx.validation import fix_reciprocity

        return fix_reciprocity(self)

    # Statistics

    def status_counts(self) -> dict[Status, int]:
        """Get count of requirements by status.

        Returns:
            Dictionary mapping status to count
        """
        counts: dict[Status, int] = dict.fromkeys(Status, 0)
        for req in self._requirements.values():
            counts[req.status] += 1
        return counts

    def completion_percentage(self) -> float:
        """Calculate completion percentage.

        PARTIAL requirements count as 50%.

        Returns:
            Completion percentage (0-100)
        """
        if not self._requirements:
            return 0.0

        counts = self.status_counts()
        complete = counts[Status.COMPLETE]
        partial = counts[Status.PARTIAL]
        total = len(self._requirements)

        return ((complete + partial * 0.5) / total) * 100

    @property
    def path(self) -> Path | None:
        """Get path to source file."""
        return self._path
