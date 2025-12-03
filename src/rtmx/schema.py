"""RTM schema definitions and validation.

This module defines the core RTM schema and provides an extension mechanism
for project-specific columns like Phoenix's validation taxonomy.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Callable


class ColumnType(str, Enum):
    """Column data types."""

    STRING = "string"
    INTEGER = "integer"
    FLOAT = "float"
    BOOLEAN = "boolean"
    DATE = "date"  # YYYY-MM-DD format
    LIST = "list"  # Pipe-separated values


@dataclass
class Column:
    """Schema column definition.

    Attributes:
        name: Column name (snake_case)
        type: Data type
        required: Whether column must have a non-empty value
        default: Default value if not provided
        validator: Optional validation function
        description: Human-readable description
    """

    name: str
    type: ColumnType = ColumnType.STRING
    required: bool = False
    default: Any = ""
    validator: Callable[[Any], bool] | None = None
    description: str = ""


@dataclass
class Schema:
    """RTM schema definition.

    A schema defines the columns expected in an RTM database,
    including core required columns and optional extension columns.
    """

    name: str
    columns: dict[str, Column] = field(default_factory=dict)
    description: str = ""

    def add_column(self, column: Column) -> None:
        """Add a column to the schema."""
        self.columns[column.name] = column

    def remove_column(self, name: str) -> None:
        """Remove a column from the schema."""
        if name in self.columns:
            del self.columns[name]

    def has_column(self, name: str) -> bool:
        """Check if schema has a column."""
        return name in self.columns

    def required_columns(self) -> list[str]:
        """Get list of required column names."""
        return [col.name for col in self.columns.values() if col.required]

    def validate_row(self, row: dict[str, Any]) -> list[str]:
        """Validate a row against the schema.

        Args:
            row: Row data as dictionary

        Returns:
            List of validation error messages
        """
        errors = []

        # Check required columns
        for col in self.columns.values():
            if col.required:
                value = row.get(col.name)
                if value is None or (isinstance(value, str) and not value.strip()):
                    errors.append(f"Missing required column: {col.name}")

        # Check validators
        for col in self.columns.values():
            if col.validator and col.name in row:
                value = row[col.name]
                try:
                    if not col.validator(value):
                        errors.append(f"Invalid value for {col.name}: {value}")
                except Exception as e:
                    errors.append(f"Validation error for {col.name}: {e}")

        return errors

    def extend(self, other: Schema) -> Schema:
        """Create a new schema by extending this one with another.

        Args:
            other: Schema to extend with

        Returns:
            New combined schema
        """
        combined = Schema(
            name=f"{self.name}+{other.name}",
            columns=dict(self.columns),
            description=f"{self.description} Extended with {other.description}",
        )
        combined.columns.update(other.columns)
        return combined


# Core schema - 20 columns that every RTM must have
CORE_SCHEMA = Schema(
    name="core",
    description="Core RTM schema with essential columns for requirements traceability",
    columns={
        "req_id": Column(
            name="req_id",
            type=ColumnType.STRING,
            required=True,
            description="Unique requirement identifier (e.g., REQ-SW-001)",
        ),
        "category": Column(
            name="category",
            type=ColumnType.STRING,
            required=True,
            description="High-level grouping (e.g., SOFTWARE, MODE, PERFORMANCE)",
        ),
        "subcategory": Column(
            name="subcategory",
            type=ColumnType.STRING,
            required=False,
            description="Detailed classification within category",
        ),
        "requirement_text": Column(
            name="requirement_text",
            type=ColumnType.STRING,
            required=True,
            description="Human-readable requirement description",
        ),
        "target_value": Column(
            name="target_value",
            type=ColumnType.STRING,
            required=False,
            description="Quantitative acceptance criteria",
        ),
        "test_module": Column(
            name="test_module",
            type=ColumnType.STRING,
            required=False,
            description="Python test file implementing validation",
        ),
        "test_function": Column(
            name="test_function",
            type=ColumnType.STRING,
            required=False,
            description="Specific test function name",
        ),
        "validation_method": Column(
            name="validation_method",
            type=ColumnType.STRING,
            required=False,
            description="Testing approach (Analysis, Test, Design, Inspection)",
        ),
        "status": Column(
            name="status",
            type=ColumnType.STRING,
            required=True,
            default="MISSING",
            description="Completion status (COMPLETE, PARTIAL, MISSING)",
            validator=lambda v: v in ("COMPLETE", "PARTIAL", "MISSING", "NOT_STARTED", ""),
        ),
        "priority": Column(
            name="priority",
            type=ColumnType.STRING,
            required=False,
            default="MEDIUM",
            description="Priority level (P0, HIGH, MEDIUM, LOW)",
            validator=lambda v: v in ("P0", "HIGH", "MEDIUM", "LOW", ""),
        ),
        "phase": Column(
            name="phase",
            type=ColumnType.INTEGER,
            required=False,
            description="Development phase (1, 2, 3, etc.)",
        ),
        "notes": Column(
            name="notes",
            type=ColumnType.STRING,
            required=False,
            description="Additional context and notes",
        ),
        "effort_weeks": Column(
            name="effort_weeks",
            type=ColumnType.FLOAT,
            required=False,
            description="Estimated effort in weeks",
        ),
        "dependencies": Column(
            name="dependencies",
            type=ColumnType.LIST,
            required=False,
            description="Pipe-separated list of requirement IDs this depends on",
        ),
        "blocks": Column(
            name="blocks",
            type=ColumnType.LIST,
            required=False,
            description="Pipe-separated list of requirement IDs this blocks",
        ),
        "assignee": Column(
            name="assignee",
            type=ColumnType.STRING,
            required=False,
            description="Person responsible for the requirement",
        ),
        "sprint": Column(
            name="sprint",
            type=ColumnType.STRING,
            required=False,
            description="Target sprint or version",
        ),
        "started_date": Column(
            name="started_date",
            type=ColumnType.DATE,
            required=False,
            description="Date work began (YYYY-MM-DD)",
        ),
        "completed_date": Column(
            name="completed_date",
            type=ColumnType.DATE,
            required=False,
            description="Date completed (YYYY-MM-DD)",
        ),
        "requirement_file": Column(
            name="requirement_file",
            type=ColumnType.STRING,
            required=False,
            description="Path to detailed specification markdown file",
        ),
    },
)


# Phoenix validation taxonomy extension
PHOENIX_EXTENSION = Schema(
    name="phoenix",
    description="Phoenix project validation taxonomy with scope, technique, and environment markers",
    columns={
        # Validation type columns (legacy)
        "unit_test": Column(
            name="unit_test",
            type=ColumnType.BOOLEAN,
            default=False,
            description="Has unit test coverage",
        ),
        "integration_test": Column(
            name="integration_test",
            type=ColumnType.BOOLEAN,
            default=False,
            description="Has integration test coverage",
        ),
        "parametric_test": Column(
            name="parametric_test",
            type=ColumnType.BOOLEAN,
            default=False,
            description="Has parametric sweep test",
        ),
        "monte_carlo_test": Column(
            name="monte_carlo_test",
            type=ColumnType.BOOLEAN,
            default=False,
            description="Has Monte Carlo test",
        ),
        "stress_test": Column(
            name="stress_test",
            type=ColumnType.BOOLEAN,
            default=False,
            description="Has stress/boundary test",
        ),
        # Scope columns (REQ-VAL-001)
        "scope_unit": Column(
            name="scope_unit",
            type=ColumnType.BOOLEAN,
            default=False,
            description="Single component isolation test",
        ),
        "scope_integration": Column(
            name="scope_integration",
            type=ColumnType.BOOLEAN,
            default=False,
            description="Multi-component interaction test",
        ),
        "scope_system": Column(
            name="scope_system",
            type=ColumnType.BOOLEAN,
            default=False,
            description="End-to-end system test",
        ),
        # Technique columns (REQ-VAL-001)
        "technique_nominal": Column(
            name="technique_nominal",
            type=ColumnType.BOOLEAN,
            default=False,
            description="Typical operating parameters (happy path)",
        ),
        "technique_parametric": Column(
            name="technique_parametric",
            type=ColumnType.BOOLEAN,
            default=False,
            description="Systematic parameter space exploration",
        ),
        "technique_monte_carlo": Column(
            name="technique_monte_carlo",
            type=ColumnType.BOOLEAN,
            default=False,
            description="Random scenario testing",
        ),
        "technique_stress": Column(
            name="technique_stress",
            type=ColumnType.BOOLEAN,
            default=False,
            description="Boundary/edge case testing",
        ),
        # Environment columns (REQ-VAL-001)
        "env_simulation": Column(
            name="env_simulation",
            type=ColumnType.BOOLEAN,
            default=False,
            description="Pure software synthetic signals",
        ),
        "env_hil": Column(
            name="env_hil",
            type=ColumnType.BOOLEAN,
            default=False,
            description="Hardware-in-loop with controlled signals",
        ),
        "env_anechoic": Column(
            name="env_anechoic",
            type=ColumnType.BOOLEAN,
            default=False,
            description="RF anechoic chamber characterization",
        ),
        "env_static_field": Column(
            name="env_static_field",
            type=ColumnType.BOOLEAN,
            default=False,
            description="Outdoor stationary targets",
        ),
        "env_dynamic_field": Column(
            name="env_dynamic_field",
            type=ColumnType.BOOLEAN,
            default=False,
            description="Outdoor moving targets",
        ),
        # Metrics columns
        "baseline_metric": Column(
            name="baseline_metric",
            type=ColumnType.FLOAT,
            required=False,
            description="Previous measured value",
        ),
        "current_metric": Column(
            name="current_metric",
            type=ColumnType.FLOAT,
            required=False,
            description="Latest measured value",
        ),
        "target_metric": Column(
            name="target_metric",
            type=ColumnType.FLOAT,
            required=False,
            description="Acceptance threshold",
        ),
        "metric_unit": Column(
            name="metric_unit",
            type=ColumnType.STRING,
            required=False,
            description="Units for metrics (Hz, m, m/s, etc.)",
        ),
        "lead_time_weeks": Column(
            name="lead_time_weeks",
            type=ColumnType.FLOAT,
            required=False,
            description="Procurement lead time",
        ),
        "supplier_part": Column(
            name="supplier_part",
            type=ColumnType.STRING,
            required=False,
            description="Hardware part number if applicable",
        ),
    },
)


# Full Phoenix schema
PHOENIX_SCHEMA = CORE_SCHEMA.extend(PHOENIX_EXTENSION)


# Schema registry
_schemas: dict[str, Schema] = {
    "core": CORE_SCHEMA,
    "phoenix": PHOENIX_SCHEMA,
}


def get_schema(name: str) -> Schema:
    """Get a schema by name.

    Args:
        name: Schema name ("core", "phoenix", etc.)

    Returns:
        Schema instance

    Raises:
        KeyError: If schema not found
    """
    if name not in _schemas:
        available = ", ".join(_schemas.keys())
        raise KeyError(f"Schema '{name}' not found. Available: {available}")
    return _schemas[name]


def register_schema(schema: Schema) -> None:
    """Register a custom schema.

    Args:
        schema: Schema to register
    """
    _schemas[schema.name] = schema


def list_schemas() -> list[str]:
    """List available schema names.

    Returns:
        List of schema names
    """
    return list(_schemas.keys())
