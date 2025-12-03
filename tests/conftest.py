"""Pytest configuration for RTMX tests."""

from pathlib import Path

import pytest


@pytest.fixture
def fixtures_dir() -> Path:
    """Return path to test fixtures directory."""
    return Path(__file__).parent / "fixtures"


@pytest.fixture
def core_rtm_path(fixtures_dir: Path) -> Path:
    """Return path to core RTM test fixture."""
    return fixtures_dir / "core_rtm.csv"
