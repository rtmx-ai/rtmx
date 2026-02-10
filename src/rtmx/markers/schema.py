"""JSON Schema for RTMX marker annotations.

Defines the canonical marker format that all language parsers normalize to.
"""

from __future__ import annotations

import re
from typing import Any


class MarkerValidationError(Exception):
    """Raised when marker data fails schema validation."""

    def __init__(self, message: str, field: str | None = None) -> None:
        self.field = field
        super().__init__(message)


# Canonical JSON Schema for marker annotations (v2020-12)
MARKER_SCHEMA: dict[str, Any] = {
    "$schema": "https://json-schema.org/draft/2020-12/schema",
    "$id": "https://rtmx.ai/schemas/marker/v1",
    "title": "RTMX Requirement Marker",
    "description": "Language-agnostic requirement marker annotation format",
    "type": "object",
    "required": ["req_id"],
    "properties": {
        "req_id": {
            "type": "string",
            "pattern": "^REQ-[A-Z]+-[0-9]+$",
            "description": "Requirement identifier in REQ-XXX-NNN format",
        },
        "scope": {
            "type": "string",
            "enum": ["unit", "integration", "system", "acceptance"],
            "description": "Test scope level",
        },
        "technique": {
            "type": "string",
            "enum": ["nominal", "parametric", "monte_carlo", "stress", "boundary"],
            "description": "Testing technique",
        },
        "env": {
            "type": "string",
            "enum": ["simulation", "hil", "anechoic", "field"],
            "description": "Test environment",
        },
        "file_path": {
            "type": "string",
            "description": "Source file containing the marker",
        },
        "line_number": {
            "type": "integer",
            "minimum": 1,
            "description": "Line number where marker appears",
        },
        "language": {
            "type": "string",
            "description": "Programming language of source file",
        },
        "function_name": {
            "type": "string",
            "description": "Name of test function or method",
        },
    },
    "additionalProperties": False,
}

# Compiled regex for req_id validation
REQ_ID_PATTERN = re.compile(r"^REQ-[A-Z]+-[0-9]+$")

# Valid enum values
VALID_SCOPES = {"unit", "integration", "system", "acceptance"}
VALID_TECHNIQUES = {"nominal", "parametric", "monte_carlo", "stress", "boundary"}
VALID_ENVS = {"simulation", "hil", "anechoic", "field"}


def validate_marker(data: dict[str, Any]) -> None:
    """Validate marker data against schema.

    Args:
        data: Marker data dictionary.

    Raises:
        MarkerValidationError: If validation fails.
    """
    # Required field: req_id
    if "req_id" not in data:
        raise MarkerValidationError("Missing required field: req_id", field="req_id")

    req_id = data["req_id"]
    if not isinstance(req_id, str):
        raise MarkerValidationError(
            f"req_id must be a string, got {type(req_id).__name__}", field="req_id"
        )

    if not REQ_ID_PATTERN.match(req_id):
        raise MarkerValidationError(
            f"Invalid req_id format: '{req_id}'. Expected REQ-XXX-NNN pattern.",
            field="req_id",
        )

    # Optional field: scope
    if "scope" in data and data["scope"] is not None:
        scope = data["scope"]
        if scope not in VALID_SCOPES:
            raise MarkerValidationError(
                f"Invalid scope: '{scope}'. Valid values: {sorted(VALID_SCOPES)}",
                field="scope",
            )

    # Optional field: technique
    if "technique" in data and data["technique"] is not None:
        technique = data["technique"]
        if technique not in VALID_TECHNIQUES:
            raise MarkerValidationError(
                f"Invalid technique: '{technique}'. Valid values: {sorted(VALID_TECHNIQUES)}",
                field="technique",
            )

    # Optional field: env
    if "env" in data and data["env"] is not None:
        env = data["env"]
        if env not in VALID_ENVS:
            raise MarkerValidationError(
                f"Invalid env: '{env}'. Valid values: {sorted(VALID_ENVS)}",
                field="env",
            )
