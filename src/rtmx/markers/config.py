"""Configuration for marker discovery.

Loads marker-related configuration from rtmx.yaml.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path

import yaml


@dataclass
class MarkerConfig:
    """Configuration for marker discovery.

    Attributes:
        custom_extensions: Mapping of custom extensions to languages.
        exclude: Glob patterns to exclude from discovery.
        include: Glob patterns to include (if set, only these are scanned).
    """

    custom_extensions: dict[str, str] = field(default_factory=dict)
    exclude: list[str] = field(default_factory=list)
    include: list[str] = field(default_factory=list)

    @classmethod
    def default(cls) -> MarkerConfig:
        """Create default configuration.

        Returns:
            MarkerConfig with sensible defaults.
        """
        return cls(
            custom_extensions={},
            exclude=[
                "**/node_modules/**",
                "**/.git/**",
                "**/venv/**",
                "**/.venv/**",
                "**/env/**",
                "**/__pycache__/**",
                "**/dist/**",
                "**/build/**",
                "**/target/**",
                "**/.tox/**",
                "**/.pytest_cache/**",
                "**/.mypy_cache/**",
                "**/.ruff_cache/**",
            ],
            include=[],
        )


def load_marker_config(config_path: Path | None = None) -> MarkerConfig:
    """Load marker configuration from rtmx.yaml.

    Args:
        config_path: Path to config file. If None, searches for rtmx.yaml.

    Returns:
        MarkerConfig instance.
    """
    # Start with defaults
    config = MarkerConfig.default()

    # Find config file
    if config_path is None:
        config_path = _find_config_file()

    if config_path is None or not config_path.exists():
        return config

    try:
        with open(config_path, encoding="utf-8") as f:
            data = yaml.safe_load(f) or {}
    except (OSError, yaml.YAMLError):
        return config

    # Extract markers section
    rtmx_config = data.get("rtmx", {})
    markers_config = rtmx_config.get("markers", {})

    # Update config with file values
    if "custom_extensions" in markers_config:
        config.custom_extensions = markers_config["custom_extensions"]

    if "exclude" in markers_config:
        config.exclude.extend(markers_config["exclude"])

    if "include" in markers_config:
        config.include = markers_config["include"]

    return config


def _find_config_file() -> Path | None:
    """Find rtmx.yaml in current or parent directories.

    Returns:
        Path to config file or None.
    """
    current = Path.cwd()

    for parent in [current, *current.parents]:
        config_path = parent / "rtmx.yaml"
        if config_path.exists():
            return config_path

        # Also check .rtmx/config.yaml
        alt_path = parent / ".rtmx" / "config.yaml"
        if alt_path.exists():
            return alt_path

    return None
