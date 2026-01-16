"""BDD test configuration and fixtures.

This module provides shared fixtures for pytest-bdd scenarios.
Fixtures follow patterns portable to other BDD runners.
"""

from __future__ import annotations

from pathlib import Path
from typing import TYPE_CHECKING, Any

import pytest

if TYPE_CHECKING:
    pass


@pytest.fixture
def bdd_context() -> dict[str, Any]:
    """Shared context for BDD scenarios.

    This fixture provides a dictionary for storing state between
    Given/When/Then steps. The pattern is portable across BDD frameworks.
    """
    return {}


@pytest.fixture
def project_dir(tmp_path: Path, bdd_context: dict[str, Any]) -> Path:
    """Create a temporary project directory for CLI tests.

    Returns:
        Path to temporary directory with rtmx.yaml and RTM database.
    """
    # Store in context for step access
    bdd_context["project_dir"] = tmp_path
    return tmp_path


@pytest.fixture
def rtmx_yaml_content() -> str:
    """Default rtmx.yaml content for test projects."""
    return """\
database_path: docs/rtm_database.csv
phases:
  1: Foundation
  2: Core Features
  10: Collaboration
"""
