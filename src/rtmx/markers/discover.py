"""Marker discovery across codebases.

Scans source files and extracts requirement markers.
"""

from __future__ import annotations

import fnmatch
from collections.abc import Iterator
from pathlib import Path

from rtmx.markers.config import MarkerConfig, load_marker_config
from rtmx.markers.detection import EXTENSION_MAP, detect_language
from rtmx.markers.models import MarkerInfo
from rtmx.markers.registry import ParserRegistry, get_default_registry


def discover_markers(
    path: Path,
    config_path: Path | None = None,
    include_errors: bool = False,
    registry: ParserRegistry | None = None,
) -> list[MarkerInfo]:
    """Discover requirement markers in a directory or file.

    Args:
        path: Path to scan (file or directory).
        config_path: Optional path to rtmx.yaml config.
        include_errors: Include markers with validation errors.
        registry: Parser registry to use (default: global registry).

    Returns:
        List of MarkerInfo objects.
    """
    if registry is None:
        registry = get_default_registry()

    config = load_marker_config(config_path)

    # Update registry with custom extensions
    for ext, lang in config.custom_extensions.items():
        parser = registry.get_parser_for_language(lang)
        if parser:
            registry.register(ext, parser)

    markers: list[MarkerInfo] = []

    if path.is_file():
        file_markers = _parse_file(path, registry, config)
        markers.extend(file_markers)
    else:
        for file_path in _iter_source_files(path, config):
            file_markers = _parse_file(file_path, registry, config)
            markers.extend(file_markers)

    # Filter out errors if not requested
    if not include_errors:
        markers = [m for m in markers if not m.error]

    return markers


def _iter_source_files(root: Path, config: MarkerConfig) -> Iterator[Path]:
    """Iterate over source files in a directory.

    Respects .gitignore and config exclude patterns.

    Args:
        root: Root directory to scan.
        config: Marker configuration.

    Yields:
        Paths to source files.
    """
    # Load .gitignore patterns
    gitignore_patterns = _load_gitignore(root)
    all_exclude = config.exclude + gitignore_patterns

    # Get supported extensions
    supported_extensions = set(EXTENSION_MAP.keys())
    supported_extensions.update(config.custom_extensions.keys())

    for file_path in root.rglob("*"):
        if not file_path.is_file():
            continue

        # Check extension
        if file_path.suffix.lower() not in supported_extensions:
            continue

        # Check exclude patterns
        relative_path = file_path.relative_to(root)
        str_path = str(relative_path)

        if _matches_any_pattern(str_path, all_exclude):
            continue

        # Check include patterns (if specified)
        if config.include and not _matches_any_pattern(str_path, config.include):
            continue

        yield file_path


def _matches_any_pattern(path: str, patterns: list[str]) -> bool:
    """Check if path matches any glob pattern.

    Args:
        path: Path string to check.
        patterns: List of glob patterns.

    Returns:
        True if path matches any pattern.
    """
    # Normalize path separators
    path_normalized = path.replace("\\", "/")
    path_parts = path_normalized.split("/")

    for pattern in patterns:
        pattern_normalized = pattern.replace("\\", "/")

        # Direct match
        if fnmatch.fnmatch(path_normalized, pattern_normalized):
            return True

        # Handle ** patterns (match any number of directories)
        if "**" in pattern_normalized:
            # Convert ** pattern to regex-style matching
            # **/ at start means any directory prefix
            # /**/ in middle means any directory in between
            # /** at end means any file/dir suffix

            # Simple handling: check if any part of the path matches
            # the non-** portion
            pattern_parts = [p for p in pattern_normalized.split("/") if p and p != "**"]
            if pattern_parts:
                # Check if the key part of pattern appears in path
                key_part = pattern_parts[-1] if pattern_parts else ""
                if key_part:
                    for path_part in path_parts:
                        if fnmatch.fnmatch(path_part, key_part):
                            return True
                    # Also check full path match with fnmatch
                    if fnmatch.fnmatch(path_normalized, pattern_normalized.replace("**", "*")):
                        return True

        # Check if path starts with a directory that matches pattern
        # e.g., pattern "node_modules/" should match "node_modules/test.js"
        if pattern_normalized.endswith("/"):
            dir_pattern = pattern_normalized.rstrip("/")
            for part in path_parts[:-1]:  # Exclude filename
                if fnmatch.fnmatch(part, dir_pattern):
                    return True

        # Check if any directory component matches (for patterns like "vendor")
        if (
            "/" not in pattern_normalized
            and "*" not in pattern_normalized
            and pattern_normalized in path_parts
        ):
            return True

    return False


def _load_gitignore(root: Path) -> list[str]:
    """Load patterns from .gitignore file.

    Args:
        root: Root directory.

    Returns:
        List of gitignore patterns.
    """
    gitignore_path = root / ".gitignore"
    if not gitignore_path.exists():
        return []

    patterns: list[str] = []
    try:
        with open(gitignore_path, encoding="utf-8") as f:
            for line in f:
                line = line.strip()
                # Skip comments and empty lines
                if not line or line.startswith("#"):
                    continue
                # Strip trailing slash for directory patterns
                line = line.rstrip("/")
                # Add the pattern itself and with ** prefix/suffix
                patterns.append(line)
                patterns.append(f"**/{line}")
                patterns.append(f"**/{line}/**")
                patterns.append(f"{line}/**")
    except OSError:
        pass

    return patterns


def _parse_file(
    file_path: Path, registry: ParserRegistry, config: MarkerConfig
) -> list[MarkerInfo]:
    """Parse markers from a single file.

    Args:
        file_path: Path to the source file.
        registry: Parser registry.
        config: Marker configuration.

    Returns:
        List of MarkerInfo objects.
    """
    # Determine language
    ext = file_path.suffix.lower()
    override_lang = config.custom_extensions.get(ext)
    language = detect_language(file_path, override_language=override_lang)

    if not language:
        return []

    # Get parser
    parser = registry.get_parser_for_extension(ext)
    if parser is None:
        parser = registry.get_parser_for_language(language)

    if parser is None:
        return []

    # Read and parse file
    try:
        content = file_path.read_text(encoding="utf-8", errors="ignore")
    except OSError:
        return []

    return parser.parse(content, file_path)
