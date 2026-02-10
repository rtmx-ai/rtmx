"""Data models for RTMX markers.

Defines the MarkerInfo class that represents a normalized requirement marker.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

from rtmx.markers.schema import (
    REQ_ID_PATTERN,
    VALID_ENVS,
    VALID_SCOPES,
    VALID_TECHNIQUES,
    MarkerValidationError,
)


@dataclass
class MarkerInfo:
    """Normalized requirement marker information.

    This is the canonical representation of a requirement marker,
    regardless of the source language or marker syntax.

    Attributes:
        req_id: Requirement identifier (REQ-XXX-NNN format).
        file_path: Path to the source file containing the marker.
        line_number: Line number where the marker appears.
        language: Programming language of the source file.
        scope: Optional test scope (unit, integration, system, acceptance).
        technique: Optional test technique (nominal, parametric, etc.).
        env: Optional test environment (simulation, hil, anechoic, field).
        function_name: Optional name of the test function/method.
        error: Optional error message if marker has issues.
    """

    req_id: str
    file_path: Path
    line_number: int
    language: str
    scope: str | None = None
    technique: str | None = None
    env: str | None = None
    function_name: str | None = None
    error: str | None = field(default=None, repr=False)

    def __post_init__(self) -> None:
        """Validate marker fields after initialization."""
        # Skip validation if this is an error marker
        if self.error:
            return

        # Validate req_id format
        if not REQ_ID_PATTERN.match(self.req_id):
            raise MarkerValidationError(
                f"Invalid req_id format: '{self.req_id}'. Expected REQ-XXX-NNN pattern.",
                field="req_id",
            )

        # Validate scope if provided
        if self.scope is not None and self.scope not in VALID_SCOPES:
            raise MarkerValidationError(
                f"Invalid scope: '{self.scope}'. Valid values: {sorted(VALID_SCOPES)}",
                field="scope",
            )

        # Validate technique if provided
        if self.technique is not None and self.technique not in VALID_TECHNIQUES:
            raise MarkerValidationError(
                f"Invalid technique: '{self.technique}'. "
                f"Valid values: {sorted(VALID_TECHNIQUES)}",
                field="technique",
            )

        # Validate env if provided
        if self.env is not None and self.env not in VALID_ENVS:
            raise MarkerValidationError(
                f"Invalid env: '{self.env}'. Valid values: {sorted(VALID_ENVS)}",
                field="env",
            )

    def to_dict(self) -> dict[str, Any]:
        """Convert to dictionary representation.

        Returns:
            Dictionary with all marker fields.
        """
        result: dict[str, Any] = {
            "req_id": self.req_id,
            "file_path": str(self.file_path),
            "line_number": self.line_number,
            "language": self.language,
        }

        if self.scope:
            result["scope"] = self.scope
        if self.technique:
            result["technique"] = self.technique
        if self.env:
            result["env"] = self.env
        if self.function_name:
            result["function_name"] = self.function_name
        if self.error:
            result["error"] = self.error

        return result

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> MarkerInfo:
        """Create MarkerInfo from dictionary.

        Args:
            data: Dictionary with marker fields.

        Returns:
            MarkerInfo instance.
        """
        return cls(
            req_id=data["req_id"],
            file_path=Path(data["file_path"]),
            line_number=data["line_number"],
            language=data["language"],
            scope=data.get("scope"),
            technique=data.get("technique"),
            env=data.get("env"),
            function_name=data.get("function_name"),
            error=data.get("error"),
        )

    @classmethod
    def error_marker(
        cls,
        raw_req_id: str,
        file_path: Path,
        line_number: int,
        language: str,
        error_message: str,
    ) -> MarkerInfo:
        """Create a marker that represents a parsing/validation error.

        Args:
            raw_req_id: The raw (possibly invalid) requirement ID.
            file_path: Path to the source file.
            line_number: Line number of the invalid marker.
            language: Programming language of the source file.
            error_message: Description of the error.

        Returns:
            MarkerInfo instance with error field set.
        """
        # Use object.__new__ to bypass __post_init__ validation
        instance = object.__new__(cls)
        instance.req_id = raw_req_id
        instance.file_path = file_path
        instance.line_number = line_number
        instance.language = language
        instance.scope = None
        instance.technique = None
        instance.env = None
        instance.function_name = None
        instance.error = error_message
        return instance


# Re-export MarkerValidationError for convenience
__all__ = ["MarkerInfo", "MarkerValidationError"]
