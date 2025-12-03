"""Core data models for RTMX.

This module provides the fundamental data structures for requirements traceability:
- Status: Requirement completion status
- Priority: Requirement priority level
- Requirement: Single requirement representation
- RTMDatabase: Collection of requirements with query/modification operations
"""

from __future__ import annotations

from dataclasses import dataclass, field
from enum import Enum
from pathlib import Path
from typing import TYPE_CHECKING, Any, Iterator, Self

if TYPE_CHECKING:
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
    extra: dict[str, Any] = field(default_factory=dict)

    def has_test(self) -> bool:
        """Check if requirement has an associated test."""
        return (
            self.test_module not in ("", "MISSING")
            and self.test_function not in ("", "MISSING")
        )

    def is_complete(self) -> bool:
        """Check if requirement is fully complete."""
        return self.status == Status.COMPLETE

    def is_blocked(self, db: RTMDatabase) -> bool:
        """Check if requirement is blocked by incomplete dependencies."""
        for dep_id in self.dependencies:
            try:
                dep = db.get(dep_id)
                if dep.status != Status.COMPLETE:
                    return True
            except RequirementNotFoundError:
                pass
        return False

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
            try:
                phase = int(phase_val)
            except (ValueError, TypeError):
                pass

        # Parse effort_weeks
        effort_val = data.get("effort_weeks", data.get("Effort_Weeks"))
        effort_weeks: float | None = None
        if effort_val not in (None, ""):
            try:
                effort_weeks = float(effort_val)
            except (ValueError, TypeError):
                pass

        # Parse dependencies and blocks
        deps_str = str(data.get("dependencies", data.get("Dependencies", "")))
        blocks_str = str(data.get("blocks", data.get("Blocks", "")))

        # Collect extra fields
        known_fields = {
            "req_id", "Req_ID", "category", "Category", "subcategory", "Subcategory",
            "requirement_text", "Requirement_Text", "target_value", "Target_Value",
            "test_module", "Test_Module", "test_function", "Test_Function",
            "validation_method", "Validation_Method", "status", "Status",
            "priority", "Priority", "phase", "Phase", "notes", "Notes",
            "effort_weeks", "Effort_Weeks", "dependencies", "Dependencies",
            "blocks", "Blocks", "assignee", "Assignee", "sprint", "Sprint",
            "started_date", "Started_Date", "completed_date", "Completed_Date",
            "requirement_file", "Requirement_File",
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
        from rtmx.parser import load_csv, find_rtm_database

        if path is None:
            resolved_path = find_rtm_database()
        else:
            resolved_path = Path(path)

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
            elif key == "dependencies" and isinstance(value, str):
                from rtmx.parser import parse_dependencies
                value = parse_dependencies(value)
            elif key == "blocks" and isinstance(value, str):
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
        counts: dict[Status, int] = {s: 0 for s in Status}
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
