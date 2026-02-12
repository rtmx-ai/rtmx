"""Tests for step definition discovery.

REQ-BDD-002: Step Definition Discovery Across Languages
"""

from __future__ import annotations

import json
from pathlib import Path
from textwrap import dedent

import pytest


@pytest.mark.req("REQ-BDD-002")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestStepDefinitionDiscovery:
    """Test suite for step definition discovery."""

    def test_discover_behave_python_steps(self, tmp_path: Path) -> None:
        """Parse behave decorators."""
        from rtmx.bdd.steps import discover_step_definitions

        step_file = tmp_path / "steps.py"
        step_file.write_text(
            dedent("""
            from behave import given, when, then

            @given(r"a user named {name}")
            def step_impl_given(context, name):
                pass

            @when(r"they log in")
            def step_impl_when(context):
                pass

            @then(r"they should see {count:d} items")
            def step_impl_then(context, count):
                pass
            """).strip()
        )

        definitions = discover_step_definitions(tmp_path)

        assert len(definitions) == 3
        given_step = next(d for d in definitions if d.keyword == "given")
        assert "user named" in given_step.pattern
        assert given_step.file_path == str(step_file)

    def test_discover_pytest_bdd_steps(self, tmp_path: Path) -> None:
        """Parse pytest-bdd decorators."""
        from rtmx.bdd.steps import discover_step_definitions

        step_file = tmp_path / "conftest.py"
        step_file.write_text(
            dedent("""
            from pytest_bdd import given, when, then

            @given("an initialized database")
            def step_database():
                pass

            @when("the user submits the form")
            def step_submit():
                pass

            @then("the record is created")
            def step_created():
                pass
            """).strip()
        )

        definitions = discover_step_definitions(tmp_path)

        assert len(definitions) == 3
        when_step = next(d for d in definitions if d.keyword == "when")
        assert "submits the form" in when_step.pattern

    def test_match_step_to_definition(self, tmp_path: Path) -> None:
        """Regex matching algorithm."""
        from rtmx.bdd.steps import StepDefinition, match_step_to_definition

        definitions = [
            StepDefinition(
                keyword="given",
                pattern=r"a user named (.+)",
                function_name="step_given",
                file_path="steps.py",
                line=10,
            ),
            StepDefinition(
                keyword="when",
                pattern=r"they log in",
                function_name="step_when",
                file_path="steps.py",
                line=15,
            ),
        ]

        # Test matching
        match = match_step_to_definition("a user named Alice", definitions)
        assert match is not None
        assert match.function_name == "step_given"
        assert match.parameters == ["Alice"]

        # Test non-matching
        no_match = match_step_to_definition("something else", definitions)
        assert no_match is None

    def test_report_unimplemented_steps(self, tmp_path: Path) -> None:
        """Identify missing implementations."""
        from rtmx.bdd.parser import parse_feature
        from rtmx.bdd.steps import find_unimplemented_steps

        # Create feature file
        feature_file = tmp_path / "test.feature"
        feature_file.write_text(
            dedent("""
            Feature: Test

              Scenario: Has unimplemented steps
                Given a configured system
                When the user clicks submit
                Then they see a confirmation
            """).strip()
        )

        # Create partial step definitions
        step_file = tmp_path / "steps.py"
        step_file.write_text(
            dedent("""
            from pytest_bdd import given

            @given("a configured system")
            def step_configured():
                pass
            """).strip()
        )

        feature = parse_feature(feature_file)
        unimplemented = find_unimplemented_steps(feature, tmp_path)

        # "When" and "Then" steps are unimplemented
        assert len(unimplemented) == 2
        assert any("clicks submit" in step.text for step in unimplemented)

    def test_report_ambiguous_steps(self, tmp_path: Path) -> None:
        """Detect multiple matches."""
        from rtmx.bdd.steps import StepDefinition, find_ambiguous_matches

        definitions = [
            StepDefinition(
                keyword="given",
                pattern=r"a user named (.+)",
                function_name="step1",
                file_path="steps1.py",
                line=10,
            ),
            StepDefinition(
                keyword="given",
                pattern=r"a user named Alice",
                function_name="step2",
                file_path="steps2.py",
                line=20,
            ),
        ]

        ambiguous = find_ambiguous_matches("a user named Alice", definitions)
        assert len(ambiguous) == 2

    def test_extract_step_parameters(self) -> None:
        """Capture group extraction."""
        from rtmx.bdd.steps import extract_parameters

        pattern = r"a user with (\d+) items and name (.+)"
        text = "a user with 5 items and name Bob"

        params = extract_parameters(pattern, text)
        assert params == ["5", "Bob"]

    def test_cucumber_expressions(self) -> None:
        """Parse cucumber expression syntax."""
        from rtmx.bdd.steps import cucumber_to_regex

        # Basic types
        assert cucumber_to_regex("{int}") == r"(-?\d+)"
        assert cucumber_to_regex("{float}") == r"(-?\d+\.?\d*)"
        assert cucumber_to_regex("{string}") == r'"([^"]*)"'
        assert cucumber_to_regex("{word}") == r"(\w+)"

        # Full pattern
        pattern = cucumber_to_regex("a user with {int} items and name {string}")
        import re

        match = re.match(pattern, 'a user with 5 items and name "Bob"')
        assert match is not None
        assert match.groups() == ("5", "Bob")

    def test_discover_steps_cli_output(self, tmp_path: Path, cli_runner: pytest.fixture) -> None:
        """CLI JSON output format."""
        from click.testing import CliRunner

        from rtmx.cli.main import main

        # Create step file
        step_file = tmp_path / "steps.py"
        step_file.write_text(
            dedent("""
            from pytest_bdd import given, when

            @given("a test step")
            def step_test():
                pass

            @when("another step")
            def step_another():
                pass
            """).strip()
        )

        runner = CliRunner()
        result = runner.invoke(main, ["discover-steps", str(tmp_path), "--json"])

        assert result.exit_code == 0
        data = json.loads(result.output)
        assert len(data) == 2

    def test_handle_multiline_patterns(self, tmp_path: Path) -> None:
        """Handle step patterns spanning multiple lines."""
        from rtmx.bdd.steps import discover_step_definitions

        step_file = tmp_path / "steps.py"
        step_file.write_text(
            dedent("""
            from pytest_bdd import given

            @given(
                "a long step that spans "
                "multiple lines"
            )
            def step_long():
                pass
            """).strip()
        )

        definitions = discover_step_definitions(tmp_path)

        assert len(definitions) == 1
        assert "long step" in definitions[0].pattern

    def test_handle_regex_patterns(self, tmp_path: Path) -> None:
        """Handle raw regex patterns."""
        from rtmx.bdd.steps import discover_step_definitions

        step_file = tmp_path / "steps.py"
        step_file.write_text(
            dedent("""
            from behave import given

            @given(r"a number (\\d+)")
            def step_number(context, num):
                pass
            """).strip()
        )

        definitions = discover_step_definitions(tmp_path)

        assert len(definitions) == 1
        # Pattern should be usable as regex
        import re

        assert re.match(definitions[0].pattern, "a number 42")


@pytest.fixture
def cli_runner():
    """Create a Click CLI test runner."""
    from click.testing import CliRunner

    return CliRunner()
