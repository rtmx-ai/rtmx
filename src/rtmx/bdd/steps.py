"""Step definition discovery for BDD frameworks.

REQ-BDD-002: Step Definition Discovery Across Languages

Discovers step definitions from Python BDD frameworks (pytest-bdd, behave).
"""

from __future__ import annotations

import re
from dataclasses import dataclass, field
from pathlib import Path
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from rtmx.bdd.parser import Feature, Step


@dataclass
class StepDefinition:
    """A discovered step definition."""

    keyword: str  # "given", "when", "then"
    pattern: str  # Regex or cucumber expression
    function_name: str
    file_path: str
    line: int
    parameters: list[str] = field(default_factory=list)
    is_regex: bool = False  # True if pattern is raw regex


@dataclass
class StepMatch:
    """Result of matching a step to a definition."""

    definition: StepDefinition
    function_name: str
    parameters: list[str]


# Patterns for extracting step definitions

# pytest-bdd: @given("pattern"), @when("pattern"), @then("pattern")
PYTEST_BDD_PATTERN = re.compile(
    r"@(given|when|then)\s*\(\s*"
    r'(?:r?["\']([^"\']+)["\']|'  # Simple string
    r'r?"""([^"]+)"""|'  # Triple-quoted string
    r'r?"([^"]+)")'  # Multiline string
    r"\s*\)",
    re.IGNORECASE | re.MULTILINE,
)

# behave: @given(r"pattern"), @when(r"pattern"), @then(r"pattern")
BEHAVE_PATTERN = re.compile(
    r'@(given|when|then)\s*\(\s*r?["\']([^"\']+)["\']\s*\)',
    re.IGNORECASE,
)

# Alternative pattern for multiline string concatenation
MULTILINE_PATTERN = re.compile(
    r'@(given|when|then)\s*\(\s*\n\s*["\']([^"\']+)["\']\s*\n\s*["\']([^"\']+)["\']',
    re.IGNORECASE | re.MULTILINE,
)

# Cucumber expression type patterns
CUCUMBER_TYPES = {
    "{int}": r"(-?\d+)",
    "{float}": r"(-?\d+\.?\d*)",
    "{string}": r'"([^"]*)"',
    "{word}": r"(\w+)",
    "{any}": r"(.+)",
}


def cucumber_to_regex(pattern: str) -> str:
    """Convert cucumber expression to regex pattern.

    Args:
        pattern: Cucumber expression pattern

    Returns:
        Equivalent regex pattern
    """
    result = pattern
    for cucumber_type, regex in CUCUMBER_TYPES.items():
        result = result.replace(cucumber_type, regex)
    # Escape any remaining special regex characters
    return result


def discover_step_definitions(
    root: Path | str,
    pattern: str = "**/*.py",
) -> list[StepDefinition]:
    """Discover step definitions in Python files.

    Args:
        root: Root directory to search
        pattern: Glob pattern for Python files

    Returns:
        List of discovered step definitions
    """
    root = Path(root)
    if not root.exists():
        return []

    definitions: list[StepDefinition] = []

    for path in root.glob(pattern):
        if path.is_file():
            definitions.extend(_parse_python_file(path))

    return definitions


def _parse_python_file(path: Path) -> list[StepDefinition]:
    """Parse a Python file for step definitions."""
    try:
        content = path.read_text(encoding="utf-8")
    except (OSError, UnicodeDecodeError):
        return []

    definitions: list[StepDefinition] = []
    lines = content.split("\n")

    # Track line numbers for each match
    def find_line_number(match_start: int) -> int:
        pos = 0
        for line_num, line in enumerate(lines, 1):
            pos += len(line) + 1  # +1 for newline
            if pos > match_start:
                return line_num
        return len(lines)

    # Try pytest-bdd/behave patterns
    for match in PYTEST_BDD_PATTERN.finditer(content):
        keyword = match.group(1).lower()
        # Get the pattern from whichever group matched
        step_pattern = match.group(2) or match.group(3) or match.group(4)
        if step_pattern:
            line_num = find_line_number(match.start())
            # Find function name on next line(s)
            func_name = _find_function_name(content, match.end())
            definitions.append(
                StepDefinition(
                    keyword=keyword,
                    pattern=step_pattern,
                    function_name=func_name,
                    file_path=str(path),
                    line=line_num,
                    is_regex=step_pattern.startswith("^") or "\\d" in step_pattern,
                )
            )

    # Try behave pattern separately
    for match in BEHAVE_PATTERN.finditer(content):
        keyword = match.group(1).lower()
        step_pattern = match.group(2)
        # Check if we already have this definition
        if any(d.pattern == step_pattern and d.keyword == keyword for d in definitions):
            continue
        line_num = find_line_number(match.start())
        func_name = _find_function_name(content, match.end())
        definitions.append(
            StepDefinition(
                keyword=keyword,
                pattern=step_pattern,
                function_name=func_name,
                file_path=str(path),
                line=line_num,
                is_regex="(" in step_pattern or "\\d" in step_pattern,
            )
        )

    # Try multiline pattern
    for match in MULTILINE_PATTERN.finditer(content):
        keyword = match.group(1).lower()
        step_pattern = match.group(2) + match.group(3)
        # Check if we already have this definition
        if any(d.pattern == step_pattern and d.keyword == keyword for d in definitions):
            continue
        line_num = find_line_number(match.start())
        func_name = _find_function_name(content, match.end())
        definitions.append(
            StepDefinition(
                keyword=keyword,
                pattern=step_pattern,
                function_name=func_name,
                file_path=str(path),
                line=line_num,
            )
        )

    return definitions


def _find_function_name(content: str, after_pos: int) -> str:
    """Find the function name after a decorator."""
    func_pattern = re.compile(r"def\s+(\w+)\s*\(")
    remaining = content[after_pos:]
    match = func_pattern.search(remaining)
    if match:
        return match.group(1)
    return "unknown"


def match_step_to_definition(
    step_text: str,
    definitions: list[StepDefinition],
) -> StepMatch | None:
    """Match a step text to a step definition.

    Args:
        step_text: The step text from a feature file
        definitions: List of discovered step definitions

    Returns:
        StepMatch if found, None otherwise
    """
    for defn in definitions:
        try:
            pattern = defn.pattern
            # If it's a cucumber expression, convert to regex
            if "{" in pattern and "}" in pattern:
                pattern = cucumber_to_regex(pattern)
            # Try to match
            match = re.fullmatch(pattern, step_text)
            if match:
                return StepMatch(
                    definition=defn,
                    function_name=defn.function_name,
                    parameters=list(match.groups()),
                )
        except re.error:
            # Invalid regex, skip
            continue
    return None


def find_unimplemented_steps(
    feature: Feature,
    steps_root: Path | str,
) -> list[Step]:
    """Find steps in a feature that have no matching definition.

    Args:
        feature: Parsed feature file
        steps_root: Root directory with step definitions

    Returns:
        List of unimplemented steps
    """

    definitions = discover_step_definitions(steps_root)
    unimplemented: list[Step] = []

    for scenario in feature.scenarios:
        for step in scenario.steps:
            # Normalize keyword for matching
            keyword: str | None = step.keyword.strip().lower()
            if keyword == "and" or keyword == "but":
                # Need to track context from previous step
                # For simplicity, match against all keywords
                keyword = None

            # Try to match
            matched = False
            for defn in definitions:
                if keyword and defn.keyword != keyword:
                    continue
                try:
                    pattern = defn.pattern
                    if "{" in pattern:
                        pattern = cucumber_to_regex(pattern)
                    if re.fullmatch(pattern, step.text):
                        matched = True
                        break
                except re.error:
                    continue

            if not matched:
                unimplemented.append(step)

    return unimplemented


def find_ambiguous_matches(
    step_text: str,
    definitions: list[StepDefinition],
) -> list[StepDefinition]:
    """Find all definitions that match a step text.

    Args:
        step_text: The step text to match
        definitions: List of step definitions

    Returns:
        List of matching definitions (empty, one, or multiple)
    """
    matches: list[StepDefinition] = []

    for defn in definitions:
        try:
            pattern = defn.pattern
            if "{" in pattern:
                pattern = cucumber_to_regex(pattern)
            if re.fullmatch(pattern, step_text):
                matches.append(defn)
        except re.error:
            continue

    return matches


def extract_parameters(pattern: str, text: str) -> list[str]:
    """Extract captured parameters from a step match.

    Args:
        pattern: Regex pattern with capture groups
        text: Text to match against

    Returns:
        List of captured parameter values
    """
    try:
        match = re.fullmatch(pattern, text)
        if match:
            return list(match.groups())
    except re.error:
        pass
    return []
