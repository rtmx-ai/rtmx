"""Gherkin parser for feature files.

REQ-BDD-001: Gherkin Parser for Feature Files

Uses the official gherkin-official library for parsing.
"""

from __future__ import annotations

import re
from dataclasses import dataclass, field
from pathlib import Path
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from collections.abc import Iterator

# Requirement tag pattern
REQ_TAG_PATTERN = re.compile(r"^REQ-[A-Z]+-\d+$")


class GherkinParseError(Exception):
    """Error parsing a Gherkin feature file."""

    def __init__(self, message: str, file_path: str | None = None) -> None:
        self.file_path = file_path
        if file_path:
            message = f"{file_path}: {message}"
        super().__init__(message)


@dataclass
class DocString:
    """A doc string argument to a step."""

    content: str
    media_type: str = ""


@dataclass
class DataTable:
    """A data table argument to a step."""

    rows: list[list[str]] = field(default_factory=list)


@dataclass
class Step:
    """A step in a scenario."""

    keyword: str
    text: str
    line: int = 0
    doc_string: DocString | None = None
    data_table: DataTable | None = None


@dataclass
class Background:
    """Background steps that run before each scenario."""

    name: str = ""
    steps: list[Step] = field(default_factory=list)
    line: int = 0


@dataclass
class Examples:
    """Examples table for Scenario Outline."""

    name: str = ""
    rows: list[list[str]] = field(default_factory=list)
    tags: list[str] = field(default_factory=list)
    line: int = 0
    description: str = ""

    @property
    def header(self) -> list[str]:
        """Get the header row (first row)."""
        return self.rows[0] if self.rows else []

    @property
    def data_rows(self) -> list[list[str]]:
        """Get data rows (all rows except header)."""
        return self.rows[1:] if len(self.rows) > 1 else []


@dataclass
class Scenario:
    """A scenario or scenario outline in a feature."""

    name: str
    steps: list[Step] = field(default_factory=list)
    tags: list[str] = field(default_factory=list)
    line: int = 0
    description: str = ""
    is_outline: bool = False
    examples_list: list[Examples] = field(default_factory=list)
    # For expanded scenarios, track origin
    outline_line: int = 0
    example_row_index: int = -1

    @property
    def examples(self) -> Examples | None:
        """Get the first Examples table (for backward compatibility)."""
        return self.examples_list[0] if self.examples_list else None

    @property
    def requirement_tags(self) -> list[str]:
        """Get requirement tags directly on this scenario."""
        return [tag for tag in self.tags if REQ_TAG_PATTERN.match(tag)]

    @property
    def inherited_requirement_tags(self) -> list[str]:
        """Get requirement tags inherited from feature level.

        Note: This is populated during parsing based on feature-level tags.
        """
        return getattr(self, "_inherited_req_tags", [])

    @inherited_requirement_tags.setter
    def inherited_requirement_tags(self, tags: list[str]) -> None:
        self._inherited_req_tags = tags


@dataclass
class Feature:
    """A parsed Gherkin feature file."""

    name: str
    description: str = ""
    tags: list[str] = field(default_factory=list)
    scenarios: list[Scenario] = field(default_factory=list)
    background: Background | None = None
    file_path: str = ""
    line: int = 0
    language: str = "en"

    @property
    def requirement_tags(self) -> list[str]:
        """Get requirement tags from this feature."""
        return [tag for tag in self.tags if REQ_TAG_PATTERN.match(tag)]


def _extract_tags(tags: list) -> list[str]:
    """Extract tag names from gherkin tag objects."""
    result = []
    for tag in tags:
        name = tag.get("name", "")
        # Remove @ prefix if present
        if name.startswith("@"):
            name = name[1:]
        result.append(name)
    return result


def _parse_data_table(table: dict) -> DataTable:
    """Parse a data table from gherkin AST."""
    rows = []
    for row in table.get("rows", []):
        cells = [cell.get("value", "") for cell in row.get("cells", [])]
        rows.append(cells)
    return DataTable(rows=rows)


def _parse_doc_string(doc_string: dict) -> DocString:
    """Parse a doc string from gherkin AST."""
    return DocString(
        content=doc_string.get("content", ""),
        media_type=doc_string.get("mediaType", ""),
    )


def _parse_step(step: dict) -> Step:
    """Parse a step from gherkin AST."""
    result = Step(
        keyword=step.get("keyword", ""),
        text=step.get("text", ""),
        line=step.get("location", {}).get("line", 0),
    )

    if "dataTable" in step:
        result.data_table = _parse_data_table(step["dataTable"])

    if "docString" in step:
        result.doc_string = _parse_doc_string(step["docString"])

    return result


def _parse_background(background: dict) -> Background:
    """Parse a background from gherkin AST."""
    steps = [_parse_step(s) for s in background.get("steps", [])]
    return Background(
        name=background.get("name", ""),
        steps=steps,
        line=background.get("location", {}).get("line", 0),
    )


def _parse_examples(examples: dict) -> Examples:
    """Parse examples table from gherkin AST."""
    rows = []
    table_header = examples.get("tableHeader", {})
    if table_header:
        header_cells = [cell.get("value", "") for cell in table_header.get("cells", [])]
        rows.append(header_cells)

    for row in examples.get("tableBody", []):
        cells = [cell.get("value", "") for cell in row.get("cells", [])]
        rows.append(cells)

    return Examples(
        name=examples.get("name", ""),
        rows=rows,
        tags=_extract_tags(examples.get("tags", [])),
        line=examples.get("location", {}).get("line", 0),
        description=examples.get("description", "").strip(),
    )


def _parse_scenario(scenario: dict, feature_req_tags: list[str]) -> Scenario:
    """Parse a scenario from gherkin AST."""
    keyword = scenario.get("keyword", "")
    # Check for outline keywords in multiple languages
    is_outline = keyword.strip() in (
        "Scenario Outline",
        "Scenario Template",
        "Szenariovorlage",  # German
        "Plan du Scénario",  # French
        "Esquema del escenario",  # Spanish
    )

    steps = [_parse_step(s) for s in scenario.get("steps", [])]
    tags = _extract_tags(scenario.get("tags", []))

    # Parse all examples tables
    examples_data = scenario.get("examples", [])
    examples_list = [_parse_examples(ex) for ex in examples_data]

    result = Scenario(
        name=scenario.get("name", ""),
        steps=steps,
        tags=tags,
        line=scenario.get("location", {}).get("line", 0),
        description=scenario.get("description", "").strip(),
        is_outline=is_outline or len(examples_list) > 0,  # Also outline if has examples
        examples_list=examples_list,
    )

    # Set inherited requirement tags from feature
    result.inherited_requirement_tags = feature_req_tags

    return result


def parse_feature(path: Path | str) -> Feature:
    """Parse a Gherkin feature file.

    Args:
        path: Path to the .feature file

    Returns:
        Parsed Feature object

    Raises:
        GherkinParseError: If the file cannot be parsed
    """
    try:
        from gherkin.parser import Parser
        from gherkin.token_scanner import TokenScanner
    except ImportError as e:
        raise GherkinParseError(
            "gherkin-official package not installed. " "Install with: pip install rtmx[bdd]"
        ) from e

    path = Path(path)
    if not path.exists():
        raise GherkinParseError(f"File not found: {path}", str(path))

    try:
        content = path.read_text(encoding="utf-8")
    except OSError as e:
        raise GherkinParseError(f"Cannot read file: {e}", str(path)) from e

    parser = Parser()
    try:
        gherkin_document = parser.parse(TokenScanner(content))
    except Exception as e:
        raise GherkinParseError(f"Parse error: {e}", str(path)) from e

    feature_dict = gherkin_document.get("feature")
    if not feature_dict:
        raise GherkinParseError("No Feature found in file", str(path))

    # Extract language from document
    language = feature_dict.get("language", "en")

    # Extract tags
    tags = _extract_tags(feature_dict.get("tags", []))
    feature_req_tags = [t for t in tags if REQ_TAG_PATTERN.match(t)]

    # Parse children (Background, Scenarios, Rules)
    scenarios: list[Scenario] = []
    background: Background | None = None

    for child in feature_dict.get("children", []):
        if "background" in child:
            background = _parse_background(child["background"])
        elif "scenario" in child:
            scenarios.append(_parse_scenario(child["scenario"], feature_req_tags))
        elif "rule" in child:
            # Rules can contain scenarios
            rule = child["rule"]
            for rule_child in rule.get("children", []):
                if "background" in rule_child and background is None:
                    background = _parse_background(rule_child["background"])
                elif "scenario" in rule_child:
                    scenarios.append(_parse_scenario(rule_child["scenario"], feature_req_tags))

    return Feature(
        name=feature_dict.get("name", ""),
        description=feature_dict.get("description", "").strip(),
        tags=tags,
        scenarios=scenarios,
        background=background,
        file_path=str(path),
        line=feature_dict.get("location", {}).get("line", 0),
        language=language,
    )


def discover_features(
    root: Path | str,
    pattern: str = "**/*.feature",
) -> list[Feature]:
    """Discover and parse all feature files in a directory.

    Args:
        root: Root directory to search
        pattern: Glob pattern for feature files

    Returns:
        List of parsed Feature objects

    Raises:
        GherkinParseError: If any file cannot be parsed
    """
    root = Path(root)
    if not root.exists():
        raise GherkinParseError(f"Directory not found: {root}")

    features: list[Feature] = []
    for path in root.glob(pattern):
        if path.is_file():
            features.append(parse_feature(path))

    return features


def iter_features(
    root: Path | str,
    pattern: str = "**/*.feature",
) -> Iterator[Feature]:
    """Iterate over feature files in a directory.

    Args:
        root: Root directory to search
        pattern: Glob pattern for feature files

    Yields:
        Parsed Feature objects
    """
    root = Path(root)
    if not root.exists():
        return

    for path in root.glob(pattern):
        if path.is_file():
            try:
                yield parse_feature(path)
            except GherkinParseError:
                # Skip unparseable files in iteration mode
                continue
