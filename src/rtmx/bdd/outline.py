"""Scenario Outline expansion functionality.

REQ-BDD-005: Scenario Outline Support

Provides expansion of Scenario Outlines into individual scenarios
using Examples table data.
"""

from __future__ import annotations

import re
from dataclasses import replace

from rtmx.bdd.parser import Scenario

# Placeholder pattern: <name>
PLACEHOLDER_PATTERN = re.compile(r"<([^>]+)>")


def expand_outline(outline: Scenario) -> list[Scenario]:
    """Expand a Scenario Outline into individual Scenarios.

    Args:
        outline: A Scenario Outline with examples

    Returns:
        List of expanded Scenarios with placeholders substituted

    Example:
        >>> outline = parse_feature("file.feature").scenarios[0]
        >>> expanded = expand_outline(outline)
        >>> len(expanded)  # One scenario per examples row
        3
    """
    if not outline.is_outline or not outline.examples_list:
        # Not an outline or no examples, return as-is in a list
        return [outline]

    expanded_scenarios: list[Scenario] = []
    scenario_index = 0

    for examples in outline.examples_list:
        if not examples.rows:
            continue

        # First row is header with column names
        header = examples.header
        if not header:
            continue

        # Create a scenario for each data row
        for row_idx, row in enumerate(examples.data_rows):
            scenario_index += 1

            # Create mapping from placeholder name to value
            substitutions = dict(zip(header, row, strict=False))

            # Substitute placeholders in step text
            expanded_steps = []
            for step in outline.steps:
                new_text = _substitute_placeholders(step.text, substitutions)
                expanded_step = replace(step, text=new_text)
                expanded_steps.append(expanded_step)

            # Combine outline tags with example tags
            combined_tags = list(outline.tags) + [
                tag for tag in examples.tags if tag not in outline.tags
            ]

            # Create expanded scenario
            expanded = Scenario(
                name=f"{outline.name} - {scenario_index}",
                steps=expanded_steps,
                tags=combined_tags,
                line=outline.line,
                description=outline.description,
                is_outline=False,
                examples_list=[],
                outline_line=outline.line,
                example_row_index=row_idx,
            )

            # Preserve inherited requirement tags
            expanded.inherited_requirement_tags = outline.inherited_requirement_tags

            expanded_scenarios.append(expanded)

    return expanded_scenarios


def _substitute_placeholders(text: str, substitutions: dict[str, str]) -> str:
    """Substitute placeholders in text with values.

    Args:
        text: Text containing <placeholder> patterns
        substitutions: Mapping of placeholder names to values

    Returns:
        Text with placeholders replaced by values
    """

    def replace_match(match: re.Match) -> str:
        placeholder = match.group(1)
        return substitutions.get(placeholder, match.group(0))

    return PLACEHOLDER_PATTERN.sub(replace_match, text)


def expand_all_outlines(scenarios: list[Scenario]) -> list[Scenario]:
    """Expand all Scenario Outlines in a list of scenarios.

    Args:
        scenarios: List of scenarios (may include outlines)

    Returns:
        List with all outlines expanded to individual scenarios
    """
    result: list[Scenario] = []
    for scenario in scenarios:
        if scenario.is_outline:
            result.extend(expand_outline(scenario))
        else:
            result.append(scenario)
    return result
