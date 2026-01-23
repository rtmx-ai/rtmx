"""CSV parsing and serialization for RTMX.

This module handles loading and saving RTM data from CSV files,
with support for multiple column naming conventions and dependency formats.
Also provides cross-repo requirement reference parsing.
"""

from __future__ import annotations

import csv
import re
from dataclasses import dataclass
from pathlib import Path
from typing import TYPE_CHECKING

from rtmx.models import Requirement, RTMError

if TYPE_CHECKING:
    pass


# =============================================================================
# Cross-Repo Requirement Reference Parsing
# =============================================================================


@dataclass
class RequirementRef:
    """Reference to a requirement, either local or cross-repo.

    Formats:
        - Local: "REQ-SW-001"
        - Aliased: "sync:REQ-SYNC-001"
        - Full repo: "sync-server:REQ-SYNC-001"

    Attributes:
        req_id: The requirement ID (e.g., "REQ-SYNC-001")
        remote_alias: Short alias if using aliased format (e.g., "sync")
        full_repo: Full repository path if using full format (e.g., "sync-server")
    """

    req_id: str
    remote_alias: str | None = None
    full_repo: str | None = None

    @property
    def is_local(self) -> bool:
        """Check if this is a local requirement reference."""
        return self.remote_alias is None and self.full_repo is None

    @property
    def is_cross_repo(self) -> bool:
        """Check if this is a cross-repo requirement reference."""
        return not self.is_local

    def __str__(self) -> str:
        """Return string representation of the reference."""
        if self.full_repo:
            return f"{self.full_repo}:{self.req_id}"
        elif self.remote_alias:
            return f"{self.remote_alias}:{self.req_id}"
        return self.req_id


# Pattern to match full repo format: org/repo:REQ-ID
_FULL_REPO_PATTERN = re.compile(r"^([a-zA-Z0-9_-]+/[a-zA-Z0-9._-]+):(.+)$")

# Pattern to match aliased format: alias:REQ-ID (alias is simple identifier)
_ALIAS_PATTERN = re.compile(r"^([a-zA-Z][a-zA-Z0-9_-]*):(.+)$")


def parse_requirement_ref(ref_str: str) -> RequirementRef:
    """Parse a requirement reference string.

    Args:
        ref_str: Reference string in one of these formats:
            - "REQ-SW-001" (local)
            - "sync:REQ-SYNC-001" (aliased remote)
            - "sync-server:REQ-SYNC-001" (full repo)

    Returns:
        RequirementRef with parsed components

    Raises:
        ValueError: If the format is invalid or empty
    """
    ref_str = ref_str.strip()

    if not ref_str:
        raise ValueError("Requirement reference cannot be empty")

    # Count colons to detect format
    colon_count = ref_str.count(":")

    if colon_count == 0:
        # Local reference
        return RequirementRef(req_id=ref_str)

    if colon_count == 1:
        # Could be aliased or full repo
        # Check for full repo pattern first (contains /)
        full_match = _FULL_REPO_PATTERN.match(ref_str)
        if full_match:
            return RequirementRef(
                req_id=full_match.group(2),
                full_repo=full_match.group(1),
            )

        # Try aliased pattern
        alias_match = _ALIAS_PATTERN.match(ref_str)
        if alias_match:
            return RequirementRef(
                req_id=alias_match.group(2),
                remote_alias=alias_match.group(1),
            )

        # Fallback: treat as local with colon in name (unusual but possible)
        return RequirementRef(req_id=ref_str)

    # Multiple colons - invalid format
    raise ValueError(f"Invalid requirement reference format: {ref_str}")


# Directory and file name constants
RTMX_DIR_NAME = ".rtmx"
DATABASE_FILE_NAME = "database.csv"

# Legacy RTM database location
DEFAULT_RTM_PATH = Path("docs/rtm_database.csv")


def find_rtm_database(start_path: Path | None = None) -> Path:
    """Find the RTM database by searching upward from start_path.

    Checks for database files in this order:
    1. .rtmx/database.csv (new standard)
    2. docs/rtm_database.csv (legacy)

    Args:
        start_path: Starting directory for search. Defaults to cwd.

    Returns:
        Path to RTM database

    Raises:
        RTMError: If database not found
    """
    start_path = Path.cwd() if start_path is None else Path(start_path).resolve()

    current = start_path
    for _ in range(20):  # Limit search depth
        # Check .rtmx/database.csv first (new standard)
        rtmx_db = current / RTMX_DIR_NAME / DATABASE_FILE_NAME
        if rtmx_db.exists():
            return rtmx_db

        # Fall back to legacy location
        legacy_db = current / DEFAULT_RTM_PATH
        if legacy_db.exists():
            return legacy_db

        parent = current.parent
        if parent == current:  # Reached filesystem root
            break
        current = parent

    raise RTMError(
        f"Could not find RTM database (looking for {RTMX_DIR_NAME}/{DATABASE_FILE_NAME} or {DEFAULT_RTM_PATH}). Started from {start_path}"
    )


def parse_dependencies(dep_str: str) -> set[str]:
    """Parse dependency string into set of requirement IDs.

    Handles both pipe-separated (REQ-A|REQ-B) and space-separated (REQ-A REQ-B) formats.

    Args:
        dep_str: Dependency string from CSV

    Returns:
        Set of requirement IDs
    """
    if not dep_str or not dep_str.strip():
        return set()

    dep_str = dep_str.strip()

    # Detect separator
    parts = dep_str.split("|") if "|" in dep_str else dep_str.split()

    # Clean and filter
    return {p.strip() for p in parts if p.strip()}


def format_dependencies(deps: set[str]) -> str:
    """Format dependency set as pipe-separated string.

    Args:
        deps: Set of requirement IDs

    Returns:
        Pipe-separated string
    """
    return "|".join(sorted(deps))


def detect_column_format(fieldnames: list[str]) -> str:
    """Detect column naming convention.

    Args:
        fieldnames: List of column names from CSV header

    Returns:
        "snake_case" or "PascalCase"
    """
    if "req_id" in fieldnames:
        return "snake_case"
    elif "Req_ID" in fieldnames:
        return "PascalCase"
    else:
        # Check for any snake_case columns
        for name in fieldnames:
            if "_" in name and name.islower():
                return "snake_case"
        return "PascalCase"


def normalize_column_name(name: str, target_format: str = "snake_case") -> str:
    """Normalize column name to target format.

    Args:
        name: Original column name
        target_format: "snake_case" or "PascalCase"

    Returns:
        Normalized column name
    """
    # Common mappings
    mappings = {
        # PascalCase -> snake_case
        "Req_ID": "req_id",
        "Category": "category",
        "Subcategory": "subcategory",
        "Requirement_Text": "requirement_text",
        "Target_Value": "target_value",
        "Test_Module": "test_module",
        "Test_Function": "test_function",
        "Validation_Method": "validation_method",
        "Status": "status",
        "Priority": "priority",
        "Phase": "phase",
        "Notes": "notes",
        "Effort_Weeks": "effort_weeks",
        "Dependencies": "dependencies",
        "Blocks": "blocks",
        "Assignee": "assignee",
        "Sprint": "sprint",
        "Started_Date": "started_date",
        "Completed_Date": "completed_date",
        "Requirement_File": "requirement_file",
    }

    if target_format == "snake_case":
        return mappings.get(name, name.lower())
    else:
        # Reverse mapping
        reverse = {v: k for k, v in mappings.items()}
        return reverse.get(name, name)


def load_csv(path: Path) -> list[Requirement]:
    """Load requirements from CSV file.

    Args:
        path: Path to CSV file

    Returns:
        List of Requirement objects

    Raises:
        RTMError: If file cannot be loaded
    """
    requirements: list[Requirement] = []

    try:
        with path.open(encoding="utf-8") as f:
            reader = csv.DictReader(f)

            if reader.fieldnames is None:
                raise RTMError(f"CSV file has no header: {path}")

            # Detect and normalize column format
            col_format = detect_column_format(list(reader.fieldnames))

            for row in reader:
                # Normalize row keys if needed
                if col_format == "PascalCase":
                    row = {normalize_column_name(k): v for k, v in row.items()}

                # Convert boolean fields
                row = _convert_booleans(row)

                req = Requirement.from_dict(row)
                requirements.append(req)

    except FileNotFoundError:
        raise RTMError(f"RTM database not found: {path}") from None
    except csv.Error as e:
        raise RTMError(f"Failed to parse CSV: {e}") from e
    except Exception as e:
        raise RTMError(f"Failed to load RTM database: {e}") from e

    if not requirements:
        raise RTMError(f"RTM database is empty: {path}")

    return requirements


def save_csv(requirements: list[Requirement], path: Path) -> None:
    """Save requirements to CSV file.

    Args:
        requirements: List of requirements to save
        path: Path to CSV file

    Raises:
        RTMError: If file cannot be saved
    """
    if not requirements:
        raise RTMError("Cannot save empty RTM database")

    # Get all field names from first requirement + any extra fields
    sample = requirements[0].to_dict()
    fieldnames = list(sample.keys())

    # Ensure consistent field order across all requirements
    for req in requirements:
        data = req.to_dict()
        for key in data:
            if key not in fieldnames:
                fieldnames.append(key)

    try:
        path.parent.mkdir(parents=True, exist_ok=True)
        with path.open("w", encoding="utf-8", newline="") as f:
            writer = csv.DictWriter(f, fieldnames=fieldnames)
            writer.writeheader()

            for req in requirements:
                row = req.to_dict()
                # Convert booleans back to strings
                row = _format_booleans(row)
                writer.writerow(row)

    except Exception as e:
        raise RTMError(f"Failed to save RTM database: {e}") from e


def _convert_booleans(row: dict[str, str]) -> dict[str, str | bool]:
    """Convert boolean string fields to actual booleans.

    Args:
        row: CSV row dictionary

    Returns:
        Row with booleans converted
    """
    boolean_fields = {
        # Phoenix validation taxonomy
        "unit_test",
        "integration_test",
        "parametric_test",
        "monte_carlo_test",
        "stress_test",
        # Environment columns
        "env_simulation",
        "env_hil",
        "env_anechoic",
        "env_static_field",
        "env_dynamic_field",
        # Scope columns
        "scope_unit",
        "scope_integration",
        "scope_system",
        # Technique columns
        "technique_nominal",
        "technique_parametric",
        "technique_monte_carlo",
        "technique_stress",
    }

    result: dict[str, str | bool] = {}
    for key, value in row.items():
        if key in boolean_fields:
            result[key] = value.strip().lower() == "true" if value else False
        else:
            result[key] = value

    return result


def _format_booleans(row: dict[str, object]) -> dict[str, str]:
    """Convert boolean values to strings for CSV.

    Args:
        row: Row dictionary

    Returns:
        Row with booleans as strings
    """
    result: dict[str, str] = {}
    for key, value in row.items():
        if isinstance(value, bool):
            result[key] = "True" if value else "False"
        elif value is None:
            result[key] = ""
        else:
            result[key] = str(value)
    return result
