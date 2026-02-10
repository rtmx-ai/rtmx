"""Parser registry for language-specific marker parsers.

Manages registration and lookup of parsers for different programming languages.
"""

from __future__ import annotations

from abc import ABC, abstractmethod
from pathlib import Path
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from rtmx.markers.models import MarkerInfo


class BaseParser(ABC):
    """Base class for language-specific marker parsers.

    Subclasses must implement the parse() method to extract markers
    from source code in their target language.
    """

    @abstractmethod
    def parse(self, content: str, file_path: Path) -> list[MarkerInfo]:
        """Parse markers from source code.

        Args:
            content: Source code content as string.
            file_path: Path to the source file (for error reporting).

        Returns:
            List of MarkerInfo objects found in the source.
        """
        pass

    def get_language(self) -> str:
        """Get the language this parser handles.

        Returns:
            Language name (lowercase).
        """
        return "unknown"


class ParserRegistry:
    """Registry for language-specific marker parsers.

    Manages mapping from file extensions and language names to parser instances.
    Built-in parsers are registered automatically; custom parsers can be added
    via configuration or programmatically.
    """

    def __init__(self) -> None:
        """Initialize registry with built-in parsers."""
        self._extension_map: dict[str, BaseParser] = {}
        self._language_map: dict[str, BaseParser] = {}
        self._custom_parsers: dict[str, BaseParser] = {}

        # Register built-in parsers
        self._register_builtins()

    def _register_builtins(self) -> None:
        """Register built-in parsers for supported languages."""
        # Import here to avoid circular imports
        from rtmx.markers.parsers.comment import CommentParser
        from rtmx.markers.parsers.python import PythonParser

        # Python parser
        python_parser = PythonParser()
        self._register_builtin("python", python_parser, [".py", ".pyw", ".pyi"])

        # Comment-based parsers for other languages
        # These use a generic comment parser that looks for @req patterns
        comment_parser = CommentParser()

        # JavaScript/TypeScript
        self._register_builtin("javascript", comment_parser, [".js", ".mjs", ".cjs", ".jsx"])
        self._register_builtin("typescript", comment_parser, [".ts", ".tsx", ".mts", ".cts"])

        # Go
        self._register_builtin("go", comment_parser, [".go"])

        # Rust
        self._register_builtin("rust", comment_parser, [".rs"])

        # Java/Kotlin
        self._register_builtin("java", comment_parser, [".java"])
        self._register_builtin("kotlin", comment_parser, [".kt", ".kts"])

        # C#
        self._register_builtin("csharp", comment_parser, [".cs"])

        # Ruby
        self._register_builtin("ruby", comment_parser, [".rb"])

    def _register_builtin(self, language: str, parser: BaseParser, extensions: list[str]) -> None:
        """Register a built-in parser.

        Args:
            language: Language name.
            parser: Parser instance.
            extensions: File extensions this parser handles.
        """
        self._language_map[language] = parser
        for ext in extensions:
            self._extension_map[ext] = parser

    @property
    def custom_parsers(self) -> dict[str, BaseParser]:
        """Get custom (non-builtin) parsers.

        Returns:
            Dictionary of custom parsers by extension.
        """
        return self._custom_parsers.copy()

    def register(self, extension: str, parser: BaseParser) -> None:
        """Register a parser for a file extension.

        Args:
            extension: File extension (with dot, e.g., '.py').
            parser: Parser instance.
        """
        self._extension_map[extension] = parser
        self._custom_parsers[extension] = parser

    def register_language(self, language: str, parser: BaseParser, extensions: list[str]) -> None:
        """Register a parser for a language with its extensions.

        Args:
            language: Language name.
            parser: Parser instance.
            extensions: File extensions for this language.
        """
        self._language_map[language] = parser
        for ext in extensions:
            self._extension_map[ext] = parser
            self._custom_parsers[ext] = parser

    def get_parser_for_extension(self, extension: str) -> BaseParser | None:
        """Get parser for a file extension.

        Args:
            extension: File extension (with dot).

        Returns:
            Parser instance or None if not found.
        """
        return self._extension_map.get(extension.lower())

    def get_parser_for_language(self, language: str) -> BaseParser | None:
        """Get parser for a language.

        Args:
            language: Language name.

        Returns:
            Parser instance or None if not found.
        """
        return self._language_map.get(language.lower())

    def get_supported_extensions(self) -> list[str]:
        """Get all supported file extensions.

        Returns:
            List of file extensions.
        """
        return sorted(self._extension_map.keys())

    def get_supported_languages(self) -> list[str]:
        """Get all supported language names.

        Returns:
            List of language names.
        """
        return sorted(self._language_map.keys())


# Global registry instance
_default_registry: ParserRegistry | None = None


def get_default_registry() -> ParserRegistry:
    """Get the default parser registry.

    Returns:
        Global ParserRegistry instance.
    """
    global _default_registry
    if _default_registry is None:
        _default_registry = ParserRegistry()
    return _default_registry
