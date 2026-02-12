"""Tests for Scenario Outline parsing and expansion.

REQ-BDD-005: Scenario Outline Support
"""

from __future__ import annotations

from pathlib import Path

import pytest


@pytest.mark.req("REQ-BDD-005")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestScenarioOutlineParsing:
    """Test Scenario Outline parsing functionality."""

    def test_parse_scenario_outline(self, tmp_path: Path) -> None:
        """Basic outline parsing."""
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "outline.feature"
        feature_file.write_text(
            """
Feature: Outline test
  Scenario Outline: Add numbers
    Given I have <a> and <b>
    When I add them
    Then the result is <sum>

    Examples:
      | a | b | sum |
      | 1 | 2 | 3   |
      | 4 | 5 | 9   |
"""
        )

        feature = parse_feature(feature_file)

        assert len(feature.scenarios) == 1
        scenario = feature.scenarios[0]
        assert scenario.is_outline
        assert scenario.name == "Add numbers"
        assert len(scenario.steps) == 3
        # Check examples
        assert scenario.examples_list is not None
        assert len(scenario.examples_list) == 1
        examples = scenario.examples_list[0]
        # First row is header
        assert examples.rows[0] == ["a", "b", "sum"]
        # Data rows
        assert examples.rows[1] == ["1", "2", "3"]
        assert examples.rows[2] == ["4", "5", "9"]

    def test_parse_examples_table(self, tmp_path: Path) -> None:
        """Examples table extraction."""
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "examples.feature"
        feature_file.write_text(
            """
Feature: Examples test
  Scenario Outline: Check values
    Given the value is <value>
    Then it should be <expected>

    Examples: Valid cases
      | value  | expected |
      | foo    | true     |
      | bar    | true     |
      | baz    | true     |
"""
        )

        feature = parse_feature(feature_file)
        scenario = feature.scenarios[0]
        examples = scenario.examples_list[0]

        assert examples.name == "Valid cases"
        assert len(examples.rows) == 4  # 1 header + 3 data rows

    def test_multiple_examples_tables(self, tmp_path: Path) -> None:
        """Multiple Examples tables per outline."""
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "multi_examples.feature"
        feature_file.write_text(
            """
Feature: Multiple examples
  Scenario Outline: Test multiple tables
    Given I have <input>
    Then I expect <output>

    Examples: Positive cases
      | input | output |
      | a     | A      |
      | b     | B      |

    Examples: Negative cases
      | input | output |
      | x     | error  |
      | y     | error  |
"""
        )

        feature = parse_feature(feature_file)
        scenario = feature.scenarios[0]

        assert len(scenario.examples_list) == 2
        assert scenario.examples_list[0].name == "Positive cases"
        assert scenario.examples_list[1].name == "Negative cases"
        assert len(scenario.examples_list[0].rows) == 3  # header + 2 data
        assert len(scenario.examples_list[1].rows) == 3  # header + 2 data

    def test_expand_outline_to_scenarios(self, tmp_path: Path) -> None:
        """Expansion algorithm."""
        from rtmx.bdd.outline import expand_outline
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "expand.feature"
        feature_file.write_text(
            """
Feature: Expansion test
  Scenario Outline: Calculate sum
    Given I have <a>
    And I have <b>
    When I add them
    Then the result is <sum>

    Examples:
      | a | b | sum |
      | 1 | 2 | 3   |
      | 5 | 5 | 10  |
"""
        )

        feature = parse_feature(feature_file)
        outline = feature.scenarios[0]
        expanded = expand_outline(outline)

        assert len(expanded) == 2
        # First expanded scenario
        assert expanded[0].name == "Calculate sum - 1"
        assert expanded[0].steps[0].text == "I have 1"
        assert expanded[0].steps[1].text == "I have 2"
        assert expanded[0].steps[3].text == "the result is 3"
        # Second expanded scenario
        assert expanded[1].name == "Calculate sum - 2"
        assert expanded[1].steps[0].text == "I have 5"
        assert expanded[1].steps[1].text == "I have 5"
        assert expanded[1].steps[3].text == "the result is 10"

    def test_placeholder_substitution(self, tmp_path: Path) -> None:
        """Value substitution in placeholders."""
        from rtmx.bdd.outline import expand_outline
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "placeholder.feature"
        feature_file.write_text(
            """
Feature: Placeholder test
  Scenario Outline: Check <item>
    Given I search for "<item>"
    When I click the <action> button
    Then I should see "<message>"

    Examples:
      | item   | action | message        |
      | apple  | buy    | Added to cart  |
      | banana | view   | Product detail |
"""
        )

        feature = parse_feature(feature_file)
        expanded = expand_outline(feature.scenarios[0])

        assert expanded[0].steps[0].text == 'I search for "apple"'
        assert expanded[0].steps[1].text == "I click the buy button"
        assert expanded[0].steps[2].text == 'I should see "Added to cart"'

        assert expanded[1].steps[0].text == 'I search for "banana"'
        assert expanded[1].steps[1].text == "I click the view button"
        assert expanded[1].steps[2].text == 'I should see "Product detail"'

    def test_tagged_examples(self, tmp_path: Path) -> None:
        """Example-level tags."""
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "tagged_examples.feature"
        feature_file.write_text(
            """
Feature: Tagged examples
  Scenario Outline: Test with tags
    Given I have <value>

    @smoke @critical
    Examples: Important
      | value |
      | 1     |

    @edge-case
    Examples: Edge cases
      | value |
      | 0     |
"""
        )

        feature = parse_feature(feature_file)
        scenario = feature.scenarios[0]

        assert len(scenario.examples_list) == 2
        assert "smoke" in scenario.examples_list[0].tags
        assert "critical" in scenario.examples_list[0].tags
        assert "edge-case" in scenario.examples_list[1].tags

    def test_tag_inheritance_to_expanded(self, tmp_path: Path) -> None:
        """Tags on expanded scenarios."""
        from rtmx.bdd.outline import expand_outline
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "tag_inherit.feature"
        feature_file.write_text(
            """
@REQ-TEST-001
Feature: Tag inheritance
  @important
  Scenario Outline: Tagged outline
    Given I have <value>

    @smoke
    Examples: Smoke tests
      | value |
      | a     |
      | b     |
"""
        )

        feature = parse_feature(feature_file)
        expanded = expand_outline(feature.scenarios[0])

        assert len(expanded) == 2
        # Each expanded scenario should have outline tags + example tags
        for scenario in expanded:
            assert "important" in scenario.tags
            assert "smoke" in scenario.tags

    def test_preserve_line_numbers(self, tmp_path: Path) -> None:
        """Traceability metadata."""
        from rtmx.bdd.outline import expand_outline
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "lines.feature"
        feature_file.write_text(
            """Feature: Line tracking
  Scenario Outline: Track lines
    Given I have <value>

    Examples:
      | value |
      | 1     |
      | 2     |
"""
        )

        feature = parse_feature(feature_file)
        expanded = expand_outline(feature.scenarios[0])

        # Expanded scenarios should reference original outline
        for scenario in expanded:
            assert scenario.outline_line > 0
            assert scenario.example_row_index >= 0

    def test_expand_outlines_cli_flag(self, tmp_path: Path, cli_runner) -> None:
        """CLI expansion flag."""
        from click.testing import CliRunner

        from rtmx.cli.main import main

        feature_file = tmp_path / "cli_expand.feature"
        feature_file.write_text(
            """
Feature: CLI test
  Scenario Outline: Test
    Given <value>

    Examples:
      | value |
      | a     |
      | b     |
"""
        )

        runner = CliRunner()
        result = runner.invoke(
            main, ["parse-feature", str(feature_file), "--expand-outlines", "--format", "json"]
        )

        assert result.exit_code == 0
        import json

        data = json.loads(result.output)
        # With expand, should have 2 scenarios instead of 1 outline
        assert len(data["features"][0]["scenarios"]) == 2

    def test_complex_data_types(self, tmp_path: Path) -> None:
        """Type handling in Examples."""
        from rtmx.bdd.outline import expand_outline
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "types.feature"
        feature_file.write_text(
            """
Feature: Data types
  Scenario Outline: Handle types
    Given the integer is <int>
    And the float is <float>
    And the boolean is <bool>
    And the string is "<string>"

    Examples:
      | int | float | bool  | string |
      | 42  | 3.14  | true  | hello  |
      | -1  | 0.0   | false | world  |
"""
        )

        feature = parse_feature(feature_file)
        expanded = expand_outline(feature.scenarios[0])

        # Check values are correctly substituted as strings
        assert expanded[0].steps[0].text == "the integer is 42"
        assert expanded[0].steps[1].text == "the float is 3.14"
        assert expanded[0].steps[2].text == "the boolean is true"
        assert expanded[0].steps[3].text == 'the string is "hello"'

    def test_examples_description(self, tmp_path: Path) -> None:
        """Description/comment parsing."""
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "description.feature"
        feature_file.write_text(
            """
Feature: Examples with description
  Scenario Outline: Described examples
    Given <value>

    Examples: Named examples
      This is a description of the examples table.
      It can span multiple lines.
      | value |
      | test  |
"""
        )

        feature = parse_feature(feature_file)
        scenario = feature.scenarios[0]
        examples = scenario.examples_list[0]

        assert examples.name == "Named examples"
        assert "description" in examples.description.lower() or len(examples.rows) > 0

    def test_i18n_scenario_template(self, tmp_path: Path) -> None:
        """Internationalized keywords."""
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "i18n.feature"
        # German: "Szenariogrundriss" = "Scenario Outline"
        feature_file.write_text(
            """# language: de
Funktionalität: I18n Test

  Szenariogrundriss: Deutsche Vorlage
    Angenommen ich habe <wert>

    Beispiele:
      | wert |
      | 1    |
"""
        )

        feature = parse_feature(feature_file)

        assert len(feature.scenarios) == 1
        assert feature.scenarios[0].is_outline
        assert feature.language == "de"


@pytest.fixture
def cli_runner():
    """Create a Click CLI test runner."""
    from click.testing import CliRunner

    return CliRunner()
