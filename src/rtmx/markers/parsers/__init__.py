"""Language-specific marker parsers.

This package contains parser implementations for different programming languages.
Each parser extracts requirement markers from source code and returns normalized
MarkerInfo objects.
"""

from rtmx.markers.parsers.comment import CommentParser
from rtmx.markers.parsers.python import PythonParser

__all__ = ["PythonParser", "CommentParser"]
