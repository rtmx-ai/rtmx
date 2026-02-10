"""RTMX Markers Module.

Language-agnostic marker annotation specification and discovery.

This module provides:
- JSON Schema for canonical marker format
- Parser registry for language-specific parsers
- Language auto-detection
- Marker discovery across codebases
"""

from rtmx.markers.detection import detect_language
from rtmx.markers.discover import discover_markers
from rtmx.markers.models import MarkerInfo
from rtmx.markers.registry import ParserRegistry
from rtmx.markers.schema import MARKER_SCHEMA, MarkerValidationError, validate_marker

__all__ = [
    "MarkerInfo",
    "ParserRegistry",
    "MARKER_SCHEMA",
    "MarkerValidationError",
    "validate_marker",
    "discover_markers",
    "detect_language",
]
