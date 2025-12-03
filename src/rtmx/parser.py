"""CSV parsing and serialization for RTMX.

This module handles loading and saving RTM data from CSV files,
with support for multiple column naming conventions and dependency formats.
"""

from __future__ import annotations

import csv
from pathlib import Path
from typing import TYPE_CHECKING

from rtmx.models import RTMError, Requirement

if TYPE_CHECKING:
    pass


# Default RTM database location
DEFAULT_RTM_PATH = Path("docs/rtm_database.csv")


def find_rtm_database(start_path: Path | None = None) -> Path:
    """Find the RTM database by searching upward from start_path.

    Searches for docs/rtm_database.csv starting from start_path
    (or current directory) and moving up the directory tree.

    Args:
        start_path: Starting directory for search. Defaults to cwd.

    Returns:
        Path to RTM database

    Raises:
        RTMError: If database not found
    """
    if start_path is None:
        start_path = Path.cwd()
    else:
        start_path = Path(start_path).resolve()

    current = start_path
    for _ in range(20):  # Limit search depth
        rtm_path = current / DEFAULT_RTM_PATH
        if rtm_path.exists():
            return rtm_path

        parent = current.parent
        if parent == current:  # Reached filesystem root
            break
        current = parent

    raise RTMError(
        f"Could not find RTM database (looking for {DEFAULT_RTM_PATH}). "
        f"Started from {start_path}"
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
    if "|" in dep_str:
        parts = dep_str.split("|")
    else:
        parts = dep_str.split()

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
                    row = {
                        normalize_column_name(k): v
                        for k, v in row.items()
                    }

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
        "unit_test", "integration_test", "parametric_test",
        "monte_carlo_test", "stress_test",
        # Environment columns
        "env_simulation", "env_hil", "env_anechoic",
        "env_static_field", "env_dynamic_field",
        # Scope columns
        "scope_unit", "scope_integration", "scope_system",
        # Technique columns
        "technique_nominal", "technique_parametric",
        "technique_monte_carlo", "technique_stress",
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
