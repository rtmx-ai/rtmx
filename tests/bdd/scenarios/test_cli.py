"""BDD scenarios for all RTMX CLI commands.

This module links all Gherkin feature files in features/cli/ to step definitions
and runs as part of the pytest test suite.
"""

from __future__ import annotations

import sys
from pathlib import Path

import pytest
from pytest_bdd import scenarios

# Add tests directory to path for imports
tests_dir = Path(__file__).parents[2]
if str(tests_dir) not in sys.path:
    sys.path.insert(0, str(tests_dir))

# Import step definitions - these register with pytest-bdd
from bdd.steps.cli_steps import *  # noqa: F403, E402
from bdd.steps.common_steps import *  # noqa: F403, E402

# Load all scenarios from all CLI feature files
FEATURES_DIR = Path(__file__).parents[3] / "features" / "cli"
for feature_file in sorted(FEATURES_DIR.glob("*.feature")):
    scenarios(str(feature_file))


# Mark all tests in this module with scope markers
pytestmark = [
    pytest.mark.scope_system,
    pytest.mark.env_simulation,
]
