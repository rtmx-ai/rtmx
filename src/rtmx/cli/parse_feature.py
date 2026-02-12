"""CLI command for parsing Gherkin feature files.

REQ-BDD-001: Gherkin Parser for Feature Files
"""

from __future__ import annotations

import json
from pathlib import Path
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from rtmx.bdd.parser import Feature


def feature_to_dict(feature: Feature) -> dict:
    """Convert a Feature to a JSON-serializable dictionary."""
    return {
        "name": feature.name,
        "description": feature.description,
        "tags": feature.tags,
        "requirement_tags": feature.requirement_tags,
        "file_path": feature.file_path,
        "line": feature.line,
        "language": feature.language,
        "background": (
            {
                "name": feature.background.name,
                "steps": [
                    {
                        "keyword": s.keyword,
                        "text": s.text,
                        "line": s.line,
                    }
                    for s in feature.background.steps
                ],
                "line": feature.background.line,
            }
            if feature.background
            else None
        ),
        "scenarios": [
            {
                "name": scenario.name,
                "description": scenario.description,
                "tags": scenario.tags,
                "requirement_tags": scenario.requirement_tags,
                "inherited_requirement_tags": scenario.inherited_requirement_tags,
                "line": scenario.line,
                "is_outline": scenario.is_outline,
                "steps": [
                    {
                        "keyword": step.keyword,
                        "text": step.text,
                        "line": step.line,
                        "doc_string": (
                            {
                                "content": step.doc_string.content,
                                "media_type": step.doc_string.media_type,
                            }
                            if step.doc_string
                            else None
                        ),
                        "data_table": ({"rows": step.data_table.rows} if step.data_table else None),
                    }
                    for step in scenario.steps
                ],
                "examples": (
                    {
                        "name": scenario.examples.name,
                        "rows": scenario.examples.rows,
                        "tags": scenario.examples.tags,
                        "line": scenario.examples.line,
                    }
                    if scenario.examples
                    else None
                ),
            }
            for scenario in feature.scenarios
        ],
    }


def run_parse_feature(
    path: str,
    output_json: bool = False,
    pattern: str = "**/*.feature",
) -> int:
    """Parse Gherkin feature file(s) and display results.

    Args:
        path: Path to feature file or directory
        output_json: Output as JSON
        pattern: Glob pattern for directory scanning

    Returns:
        Exit code (0 for success)
    """
    from rtmx.bdd.parser import GherkinParseError, discover_features, parse_feature
    from rtmx.formatting import Colors

    path_obj = Path(path)

    try:
        if path_obj.is_dir():
            features = discover_features(path_obj, pattern=pattern)
        elif path_obj.is_file():
            features = [parse_feature(path_obj)]
        else:
            print(f"{Colors.RED}Error: {path} not found{Colors.RESET}")
            return 1
    except GherkinParseError as e:
        print(f"{Colors.RED}Error: {e}{Colors.RESET}")
        return 1

    if not features:
        print(f"{Colors.YELLOW}No feature files found{Colors.RESET}")
        return 0

    if output_json:
        if len(features) == 1:
            print(json.dumps(feature_to_dict(features[0]), indent=2))
        else:
            print(json.dumps([feature_to_dict(f) for f in features], indent=2))
    else:
        _print_features(features)

    return 0


def _print_features(features: list) -> None:
    """Print feature information in human-readable format."""
    from rtmx.formatting import Colors

    for feature in features:
        print(f"{Colors.CYAN}Feature:{Colors.RESET} {feature.name}")
        print(f"  File: {feature.file_path}")

        if feature.tags:
            print(f"  Tags: {', '.join(feature.tags)}")

        if feature.requirement_tags:
            print(
                f"  {Colors.GREEN}Requirements:{Colors.RESET} "
                f"{', '.join(feature.requirement_tags)}"
            )

        if feature.language != "en":
            print(f"  Language: {feature.language}")

        if feature.background:
            print(f"\n  {Colors.YELLOW}Background:{Colors.RESET}")
            for step in feature.background.steps:
                print(f"    {step.keyword}{step.text}")

        for scenario in feature.scenarios:
            outline_marker = " (Outline)" if scenario.is_outline else ""
            print(f"\n  {Colors.YELLOW}Scenario:{Colors.RESET} {scenario.name}{outline_marker}")

            if scenario.tags:
                print(f"    Tags: {', '.join(scenario.tags)}")

            if scenario.requirement_tags:
                print(
                    f"    {Colors.GREEN}Requirements:{Colors.RESET} "
                    f"{', '.join(scenario.requirement_tags)}"
                )

            for step in scenario.steps:
                print(f"    {step.keyword}{step.text}")
                if step.data_table:
                    for row in step.data_table.rows:
                        print(f"      | {' | '.join(row)} |")

            if scenario.examples:
                print(f"    {Colors.CYAN}Examples:{Colors.RESET}")
                for row in scenario.examples.rows:
                    print(f"      | {' | '.join(row)} |")

        print()
