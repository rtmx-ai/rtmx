"""BDD support for RTMX - Gherkin parsing and feature file management.

REQ-BDD-001: Gherkin Parser for Feature Files
REQ-BDD-002: Step Definition Discovery
"""

from __future__ import annotations

from rtmx.bdd.parser import (
    Background,
    DataTable,
    DocString,
    Feature,
    GherkinParseError,
    Scenario,
    Step,
    discover_features,
    parse_feature,
)
from rtmx.bdd.steps import (
    StepDefinition,
    StepMatch,
    cucumber_to_regex,
    discover_step_definitions,
    extract_parameters,
    find_ambiguous_matches,
    find_unimplemented_steps,
    match_step_to_definition,
)

__all__ = [
    # Parser (REQ-BDD-001)
    "Background",
    "DataTable",
    "DocString",
    "Feature",
    "GherkinParseError",
    "Scenario",
    "Step",
    "discover_features",
    "parse_feature",
    # Steps (REQ-BDD-002)
    "StepDefinition",
    "StepMatch",
    "cucumber_to_regex",
    "discover_step_definitions",
    "extract_parameters",
    "find_ambiguous_matches",
    "find_unimplemented_steps",
    "match_step_to_definition",
]
