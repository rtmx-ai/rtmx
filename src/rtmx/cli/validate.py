"""RTMX validate commands.

Provides validation for staged RTM CSV files (pre-commit hook support).
"""

from __future__ import annotations

import csv
from pathlib import Path

from rtmx.formatting import Colors
from rtmx.models import Priority, RTMDatabase, Status
from rtmx.validation import validate_all, validate_cycles

# Valid status values (strict validation)
VALID_STATUS_VALUES = {s.value for s in Status}

# Valid priority values (strict validation)
VALID_PRIORITY_VALUES = {p.value for p in Priority}


def validate_csv_raw(file_path: Path) -> list[str]:
    """Validate CSV content at the raw level before loading into model.

    This catches issues that the model would silently handle (e.g., invalid status).

    Args:
        file_path: Path to CSV file

    Returns:
        List of error messages
    """
    errors: list[str] = []
    required_columns = {"req_id", "category", "requirement_text", "status"}

    try:
        with open(file_path, newline="", encoding="utf-8") as f:
            reader = csv.DictReader(f)
            fieldnames = set(reader.fieldnames or [])

            # Check required columns
            missing_cols = required_columns - fieldnames
            if missing_cols:
                errors.append(f"Missing required column(s): {', '.join(sorted(missing_cols))}")
                return errors  # Can't continue without required columns

            seen_ids: set[str] = set()
            for row_num, row in enumerate(reader, start=2):  # Start at 2 (header is 1)
                req_id = row.get("req_id", "").strip()

                # Check for duplicate IDs
                if req_id:
                    if req_id in seen_ids:
                        errors.append(f"Row {row_num}: Duplicate requirement ID '{req_id}'")
                    seen_ids.add(req_id)

                # Validate status value (strict)
                status_val = row.get("status", "").strip().upper()
                if status_val and status_val not in VALID_STATUS_VALUES:
                    errors.append(
                        f"Row {row_num} ({req_id}): Invalid status '{status_val}' "
                        f"(valid: {', '.join(sorted(VALID_STATUS_VALUES))})"
                    )

                # Validate priority value (strict)
                priority_val = row.get("priority", "").strip().upper()
                if priority_val and priority_val not in VALID_PRIORITY_VALUES:
                    errors.append(
                        f"Row {row_num} ({req_id}): Invalid priority '{priority_val}' "
                        f"(valid: {', '.join(sorted(VALID_PRIORITY_VALUES))})"
                    )

    except csv.Error as e:
        errors.append(f"CSV parsing error: {e}")

    return errors


def run_validate_staged(files: list[Path]) -> list[str]:
    """Validate staged RTM CSV files.

    This function is designed to be called from a pre-commit hook.
    It validates only the specified files (which should be staged files).

    Args:
        files: List of paths to RTM CSV files to validate

    Returns:
        List of error messages (empty if validation passes)
    """
    if not files:
        return []

    errors: list[str] = []

    for file_path in files:
        if not file_path.exists():
            errors.append(f"{file_path}: File not found")
            continue

        if file_path.suffix != ".csv":
            # Skip non-CSV files
            continue

        # First, validate raw CSV content (catches issues model would ignore)
        raw_errors = validate_csv_raw(file_path)
        for error in raw_errors:
            errors.append(f"{file_path}: {error}")

        # If raw validation found errors, skip model-level validation
        if raw_errors:
            continue

        try:
            db = RTMDatabase.load(file_path)
        except KeyError as e:
            # Missing required column
            errors.append(f"{file_path}: Missing required column: {e}")
            continue
        except Exception as e:
            errors.append(f"{file_path}: Failed to parse CSV: {e}")
            continue

        # Run all model-level validations
        results = validate_all(db)

        # Collect errors (schema validation errors are blocking)
        for error in results["errors"]:
            errors.append(f"{file_path}: {error}")

        # Cycle detection is a blocking error
        cycle_warnings = validate_cycles(db)
        for warning in cycle_warnings:
            if "Found" in warning and "circular" in warning.lower():
                errors.append(f"{file_path}: {warning}")

    return errors


def run_validate_staged_cli(files: list[str], verbose: bool = False) -> int:
    """CLI wrapper for validate-staged command.

    Args:
        files: List of file paths as strings
        verbose: Show detailed output

    Returns:
        Exit code (0 = success, 1 = validation failed)
    """
    file_paths = [Path(f) for f in files]

    if verbose:
        print(f"Validating {len(file_paths)} file(s)...")

    errors = run_validate_staged(file_paths)

    if errors:
        print(f"{Colors.RED}Validation failed:{Colors.RESET}")
        for error in errors:
            print(f"  {Colors.RED}✗{Colors.RESET} {error}")
        return 1

    if verbose:
        print(f"{Colors.GREEN}✓ All files passed validation{Colors.RESET}")

    return 0
