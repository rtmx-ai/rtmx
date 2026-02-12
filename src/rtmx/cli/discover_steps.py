"""CLI command for discovering step definitions.

REQ-BDD-002: Step Definition Discovery Across Languages
"""

from __future__ import annotations

import json
from pathlib import Path
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from rtmx.bdd.steps import StepDefinition


def step_definition_to_dict(definition: StepDefinition) -> dict:
    """Convert a StepDefinition to a JSON-serializable dictionary."""
    return {
        "keyword": definition.keyword,
        "pattern": definition.pattern,
        "function_name": definition.function_name,
        "file_path": definition.file_path,
        "line": definition.line,
        "is_regex": definition.is_regex,
    }


def run_discover_steps(
    path: str,
    output_json: bool = False,
    pattern: str = "**/*.py",
) -> int:
    """Discover step definitions and display results.

    Args:
        path: Path to directory containing step definitions
        output_json: Output as JSON
        pattern: Glob pattern for Python files

    Returns:
        Exit code (0 for success)
    """
    from rtmx.bdd.steps import discover_step_definitions
    from rtmx.formatting import Colors

    path_obj = Path(path)

    if not path_obj.exists():
        print(f"{Colors.RED}Error: {path} not found{Colors.RESET}")
        return 1

    definitions = discover_step_definitions(path_obj, pattern=pattern)

    if not definitions:
        print(f"{Colors.YELLOW}No step definitions found{Colors.RESET}")
        return 0

    if output_json:
        print(json.dumps([step_definition_to_dict(d) for d in definitions], indent=2))
    else:
        _print_definitions(definitions)

    return 0


def _print_definitions(definitions: list) -> None:
    """Print step definitions in human-readable format."""
    from rtmx.formatting import Colors

    # Group by file
    by_file: dict[str, list] = {}
    for d in definitions:
        by_file.setdefault(d.file_path, []).append(d)

    for file_path, file_defs in sorted(by_file.items()):
        print(f"\n{Colors.CYAN}File:{Colors.RESET} {file_path}")

        for d in sorted(file_defs, key=lambda x: x.line):
            keyword_color = {
                "given": Colors.GREEN,
                "when": Colors.YELLOW,
                "then": Colors.BLUE,
            }.get(d.keyword, Colors.RESET)

            print(
                f"  {Colors.DIM}L{d.line:4d}{Colors.RESET} "
                f"{keyword_color}{d.keyword.upper():5s}{Colors.RESET} "
                f'"{d.pattern}"'
            )
            print(f"        → {d.function_name}()")

    print(f"\n{Colors.GREEN}Total: {len(definitions)} step definitions{Colors.RESET}")
