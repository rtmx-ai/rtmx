"""Marker validation utilities.

Provides validation and error suggestion functionality for markers.
"""

from __future__ import annotations

import re
from dataclasses import dataclass

from rtmx.markers.schema import (
    REQ_ID_PATTERN,
    VALID_ENVS,
    VALID_SCOPES,
    VALID_TECHNIQUES,
)


@dataclass
class ValidationResult:
    """Result of marker validation.

    Attributes:
        is_valid: Whether the marker is valid.
        field: Name of the invalid field (if any).
        message: Error message (if invalid).
        suggestion: Suggested fix (if invalid).
    """

    is_valid: bool
    field: str | None = None
    message: str | None = None
    suggestion: str | None = None


def validate_marker_format(
    req_id: str,
    scope: str | None = None,
    technique: str | None = None,
    env: str | None = None,
) -> ValidationResult:
    """Validate marker format and provide suggestions.

    Args:
        req_id: Requirement ID to validate.
        scope: Optional scope value.
        technique: Optional technique value.
        env: Optional env value.

    Returns:
        ValidationResult with validation status and suggestions.
    """
    # Validate req_id
    if not REQ_ID_PATTERN.match(req_id):
        suggestion = _suggest_req_id_fix(req_id)
        return ValidationResult(
            is_valid=False,
            field="req_id",
            message=f"Invalid req_id format: '{req_id}'",
            suggestion=suggestion,
        )

    # Validate scope
    if scope is not None and scope.lower() not in VALID_SCOPES:
        closest = _find_closest_match(scope, VALID_SCOPES)
        return ValidationResult(
            is_valid=False,
            field="scope",
            message=f"Invalid scope: '{scope}'",
            suggestion=f"Did you mean '{closest}'? Valid scopes: {sorted(VALID_SCOPES)}",
        )

    # Validate technique
    if technique is not None and technique.lower() not in VALID_TECHNIQUES:
        closest = _find_closest_match(technique, VALID_TECHNIQUES)
        return ValidationResult(
            is_valid=False,
            field="technique",
            message=f"Invalid technique: '{technique}'",
            suggestion=f"Did you mean '{closest}'? Valid techniques: {sorted(VALID_TECHNIQUES)}",
        )

    # Validate env
    if env is not None and env.lower() not in VALID_ENVS:
        closest = _find_closest_match(env, VALID_ENVS)
        return ValidationResult(
            is_valid=False,
            field="env",
            message=f"Invalid env: '{env}'",
            suggestion=f"Did you mean '{closest}'? Valid envs: {sorted(VALID_ENVS)}",
        )

    return ValidationResult(is_valid=True)


def _suggest_req_id_fix(req_id: str) -> str:
    """Suggest a fix for an invalid req_id.

    Args:
        req_id: Invalid requirement ID.

    Returns:
        Suggestion string.
    """
    # Check if it's just a case issue
    upper_id = req_id.upper()
    if REQ_ID_PATTERN.match(upper_id):
        return f"Use uppercase: '{upper_id}'"

    # Check if missing REQ- prefix
    if not req_id.upper().startswith("REQ-"):
        # Try to parse the pattern
        match = re.match(r"([A-Za-z]+)[_-]?([0-9]+)", req_id)
        if match:
            category = match.group(1).upper()
            number = match.group(2)
            suggested = f"REQ-{category}-{number.zfill(3)}"
            return f"Use format 'REQ-CATEGORY-NUMBER', e.g., '{suggested}'"

    # Check if missing number
    match = re.match(r"REQ-([A-Z]+)$", upper_id)
    if match:
        return f"Add a number: 'REQ-{match.group(1)}-001'"

    # Generic suggestion
    return "Use format 'REQ-CATEGORY-NUMBER', e.g., 'REQ-AUTH-001'"


def _find_closest_match(value: str, valid_values: set[str]) -> str:
    """Find the closest matching valid value.

    Uses simple string similarity based on common prefix.

    Args:
        value: Invalid value.
        valid_values: Set of valid values.

    Returns:
        Closest matching value.
    """
    value_lower = value.lower()

    # Check for prefix match
    for valid in valid_values:
        if valid.startswith(value_lower) or value_lower.startswith(valid):
            return valid

    # Check for substring match
    for valid in valid_values:
        if value_lower in valid or valid in value_lower:
            return valid

    # Return first valid value as fallback
    return sorted(valid_values)[0]
