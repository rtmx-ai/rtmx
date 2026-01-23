"""Validation operations for RTMX.

This module provides validation functions for RTM databases:
- Schema validation (required fields, valid values)
- Reciprocity validation (dependency/blocks consistency)
- Cycle detection warnings
- Cross-repo dependency validation
"""

from __future__ import annotations

from pathlib import Path
from typing import TYPE_CHECKING

from rtmx.models import Priority, Status

if TYPE_CHECKING:
    from rtmx.config import RTMXConfig
    from rtmx.models import RTMDatabase


def validate_schema(db: RTMDatabase) -> list[str]:
    """Validate RTM database against schema requirements.

    Checks:
    - Required fields are present and non-empty
    - Status values are valid
    - Priority values are valid
    - Phase values are valid integers
    - Duplicate requirement IDs
    - Dependencies reference existing requirements

    Args:
        db: RTM database to validate

    Returns:
        List of validation error messages (empty if valid)
    """
    errors: list[str] = []
    seen_ids: set[str] = set()

    for i, req in enumerate(db, start=1):
        row_prefix = f"Row {i} ({req.req_id})"

        # Check for duplicate IDs
        if req.req_id in seen_ids:
            errors.append(f"{row_prefix}: Duplicate requirement ID")
        seen_ids.add(req.req_id)

        # Check required fields
        if not req.req_id.strip():
            errors.append(f"Row {i}: Missing required field 'req_id'")

        if not req.category.strip():
            errors.append(f"{row_prefix}: Missing required field 'category'")

        if not req.requirement_text.strip():
            errors.append(f"{row_prefix}: Missing required field 'requirement_text'")

        # Validate status
        if req.status not in Status:
            errors.append(f"{row_prefix}: Invalid status '{req.status}'")

        # Validate priority
        if req.priority not in Priority:
            errors.append(f"{row_prefix}: Invalid priority '{req.priority}'")

        # Validate phase (if present)
        if req.phase is not None and req.phase < 1:
            errors.append(f"{row_prefix}: Invalid phase '{req.phase}' (must be >= 1)")

    # Second pass: validate dependency references
    for req in db:
        for dep_id in req.dependencies:
            if dep_id not in seen_ids:
                errors.append(
                    f"{req.req_id}: Dependency '{dep_id}' references non-existent requirement"
                )

        for block_id in req.blocks:
            if block_id not in seen_ids:
                errors.append(
                    f"{req.req_id}: Blocks '{block_id}' references non-existent requirement"
                )

    return errors


def check_reciprocity(db: RTMDatabase) -> list[tuple[str, str, str]]:
    """Check dependency/blocks reciprocity.

    For every A blocks B relationship, there should be a corresponding
    B depends on A relationship.

    Args:
        db: RTM database to check

    Returns:
        List of (req_id, related_id, issue) tuples describing violations
    """
    violations: list[tuple[str, str, str]] = []

    for req in db:
        # Check: if A blocks B, then B should depend on A
        for blocked_id in req.blocks:
            if not db.exists(blocked_id):
                violations.append((req.req_id, blocked_id, "blocks non-existent requirement"))
                continue

            blocked_req = db.get(blocked_id)
            if req.req_id not in blocked_req.dependencies:
                violations.append(
                    (
                        req.req_id,
                        blocked_id,
                        f"blocks {blocked_id} but {blocked_id} doesn't depend on {req.req_id}",
                    )
                )

        # Check: if A depends on B, then B should block A
        for dep_id in req.dependencies:
            if not db.exists(dep_id):
                violations.append((req.req_id, dep_id, "depends on non-existent requirement"))
                continue

            dep_req = db.get(dep_id)
            if req.req_id not in dep_req.blocks:
                violations.append(
                    (
                        dep_id,
                        req.req_id,
                        f"{req.req_id} depends on {dep_id} but {dep_id} doesn't block {req.req_id}",
                    )
                )

    return violations


def fix_reciprocity(db: RTMDatabase) -> int:
    """Fix dependency/blocks reciprocity violations.

    For every A blocks B relationship, ensures B depends on A.
    For every B depends on A relationship, ensures A blocks B.

    Args:
        db: RTM database to fix (modified in place)

    Returns:
        Number of violations fixed
    """
    fixed_count = 0

    for req in db:
        # If A blocks B, ensure B depends on A
        for blocked_id in list(req.blocks):
            if not db.exists(blocked_id):
                continue

            blocked_req = db.get(blocked_id)
            if req.req_id not in blocked_req.dependencies:
                blocked_req.dependencies.add(req.req_id)
                fixed_count += 1

        # If A depends on B, ensure B blocks A
        for dep_id in list(req.dependencies):
            if not db.exists(dep_id):
                continue

            dep_req = db.get(dep_id)
            if req.req_id not in dep_req.blocks:
                dep_req.blocks.add(req.req_id)
                fixed_count += 1

    return fixed_count


def validate_cycles(db: RTMDatabase) -> list[str]:
    """Check for circular dependencies.

    Args:
        db: RTM database to check

    Returns:
        List of warning messages about cycles
    """
    warnings: list[str] = []
    cycles = db.find_cycles()

    if cycles:
        warnings.append(f"Found {len(cycles)} circular dependency group(s)")

        for i, cycle in enumerate(cycles[:5], start=1):  # Show first 5
            if len(cycle) <= 5:
                path = " -> ".join(cycle)
            else:
                path = f"{' -> '.join(cycle[:3])} ... ({len(cycle)} total)"
            warnings.append(f"  Cycle {i}: {path}")

        if len(cycles) > 5:
            warnings.append(f"  ... and {len(cycles) - 5} more cycles")

    return warnings


def validate_all(db: RTMDatabase) -> dict[str, list[str]]:
    """Run all validations on RTM database.

    Args:
        db: RTM database to validate

    Returns:
        Dictionary with validation results:
        - "errors": Schema validation errors (blocking)
        - "warnings": Non-blocking issues (cycles, etc.)
        - "reciprocity": Reciprocity violations
    """
    return {
        "errors": validate_schema(db),
        "warnings": validate_cycles(db),
        "reciprocity": [f"{a} <-> {b}: {issue}" for a, b, issue in check_reciprocity(db)],
    }


def validate_cross_repo_deps(db: RTMDatabase, config: RTMXConfig) -> tuple[list[str], list[str]]:
    """Validate cross-repository dependencies.

    Checks:
    - Remote aliases are configured
    - Remote repositories are accessible (if path configured)
    - Referenced requirements exist in remote databases

    Graceful degradation: When a remote is unavailable, warnings are issued
    instead of errors to allow offline work.

    Args:
        db: Local RTM database to validate
        config: RTMX configuration with remote definitions

    Returns:
        Tuple of (errors, warnings) lists
    """
    from rtmx.parser import parse_requirement_ref

    errors: list[str] = []
    warnings: list[str] = []

    # Cache of loaded remote databases
    remote_dbs: dict[str, RTMDatabase | None] = {}

    def get_remote_db(alias: str) -> RTMDatabase | None:
        """Load and cache remote database."""
        if alias in remote_dbs:
            return remote_dbs[alias]

        remote_config = config.sync.get_remote(alias)
        if remote_config is None:
            return None

        if remote_config.path is None:
            # No local path configured - can't validate
            remote_dbs[alias] = None
            return None

        remote_path = Path(remote_config.path)
        db_path = remote_path / remote_config.database

        if not db_path.exists():
            remote_dbs[alias] = None
            return None

        try:
            from rtmx.models import RTMDatabase as DB

            remote_dbs[alias] = DB.load(db_path)
            return remote_dbs[alias]
        except Exception:
            remote_dbs[alias] = None
            return None

    # Scan all requirements for cross-repo dependencies
    for req in db:
        for dep_str in req.dependencies:
            ref = parse_requirement_ref(dep_str)

            if ref.is_local:
                continue  # Skip local dependencies

            alias = ref.remote_alias
            if alias is None and ref.full_repo:
                # Full repo format - try to find matching alias
                for a, r in config.sync.remotes.items():
                    if r.repo == ref.full_repo:
                        alias = a
                        break

            if alias is None:
                # Unknown alias or unmatched full repo
                if ref.remote_alias:
                    errors.append(
                        f"{req.req_id}: Unknown remote alias '{ref.remote_alias}' "
                        f"in dependency '{dep_str}'"
                    )
                else:
                    errors.append(
                        f"{req.req_id}: No remote configured for repository "
                        f"'{ref.full_repo}' in dependency '{dep_str}'"
                    )
                continue

            # Check if the alias actually exists in config
            remote_config = config.sync.get_remote(alias)
            if remote_config is None:
                errors.append(
                    f"{req.req_id}: Unknown remote alias '{alias}' " f"in dependency '{dep_str}'"
                )
                continue

            # Try to load remote database
            remote_db = get_remote_db(alias)

            if remote_db is None:
                # Remote unavailable - graceful degradation
                remote_config = config.sync.get_remote(alias)
                if remote_config and remote_config.path:
                    warnings.append(
                        f"{req.req_id}: Remote '{alias}' unavailable or not found "
                        f"at '{remote_config.path}' - cannot verify '{dep_str}'"
                    )
                else:
                    warnings.append(
                        f"{req.req_id}: Remote '{alias}' has no local path configured "
                        f"- cannot verify '{dep_str}'"
                    )
                continue

            # Verify requirement exists in remote
            if not remote_db.exists(ref.req_id):
                errors.append(
                    f"{req.req_id}: Dependency '{ref.req_id}' not found in " f"remote '{alias}'"
                )

    return errors, warnings
