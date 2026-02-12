"""Tests for the Gherkin parser.

REQ-BDD-001: Gherkin Parser for Feature Files
"""

from __future__ import annotations

import json
from pathlib import Path
from textwrap import dedent

import pytest


@pytest.mark.req("REQ-BDD-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestGherkinParser:
    """Test suite for Gherkin feature file parsing."""

    def test_parse_simple_feature(self, tmp_path: Path) -> None:
        """Parse a minimal feature file."""
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "simple.feature"
        feature_file.write_text(
            dedent("""
            Feature: Simple feature
              As a user
              I want to test something

              Scenario: Basic scenario
                Given a precondition
                When an action is performed
                Then an outcome is expected
            """).strip()
        )

        result = parse_feature(feature_file)

        assert result is not None
        assert result.name == "Simple feature"
        assert len(result.scenarios) == 1
        assert result.scenarios[0].name == "Basic scenario"

    def test_extract_feature_metadata(self, tmp_path: Path) -> None:
        """Validate feature name/description extraction."""
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "metadata.feature"
        feature_file.write_text(
            dedent("""
            @important @smoke
            Feature: Feature with metadata
              This is the feature description.
              It can span multiple lines.

              Scenario: Test scenario
                Given something
            """).strip()
        )

        result = parse_feature(feature_file)

        assert result.name == "Feature with metadata"
        assert "This is the feature description" in result.description
        assert "important" in result.tags
        assert "smoke" in result.tags
        assert str(feature_file) == result.file_path

    def test_extract_scenario_metadata(self, tmp_path: Path) -> None:
        """Validate scenario extraction."""
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "scenarios.feature"
        feature_file.write_text(
            dedent("""
            Feature: Multiple scenarios

              @smoke
              Scenario: First scenario
                Given first step

              @regression @critical
              Scenario: Second scenario
                Given second step
            """).strip()
        )

        result = parse_feature(feature_file)

        assert len(result.scenarios) == 2

        first = result.scenarios[0]
        assert first.name == "First scenario"
        assert "smoke" in first.tags
        assert first.line > 0

        second = result.scenarios[1]
        assert second.name == "Second scenario"
        assert "regression" in second.tags
        assert "critical" in second.tags

    def test_extract_step_definitions(self, tmp_path: Path) -> None:
        """Validate Given/When/Then parsing."""
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "steps.feature"
        feature_file.write_text(
            dedent("""
            Feature: Step definitions

              Scenario: Full steps
                Given a precondition is met
                And another precondition
                When the user performs an action
                But not this action
                Then the result is visible
            """).strip()
        )

        result = parse_feature(feature_file)
        steps = result.scenarios[0].steps

        assert len(steps) == 5
        assert steps[0].keyword == "Given "
        assert steps[0].text == "a precondition is met"
        assert steps[1].keyword == "And "
        assert steps[2].keyword == "When "
        assert steps[3].keyword == "But "
        assert steps[4].keyword == "Then "

    def test_parse_data_tables(self, tmp_path: Path) -> None:
        """Handle step data tables."""
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "tables.feature"
        feature_file.write_text(
            dedent("""
            Feature: Data tables

              Scenario: With data table
                Given the following users exist:
                  | name  | email           |
                  | Alice | alice@test.com  |
                  | Bob   | bob@test.com    |
                When I list users
                Then I see 2 users
            """).strip()
        )

        result = parse_feature(feature_file)
        steps = result.scenarios[0].steps

        assert steps[0].data_table is not None
        table = steps[0].data_table
        assert len(table.rows) == 3  # header + 2 data rows
        assert table.rows[0] == ["name", "email"]
        assert table.rows[1] == ["Alice", "alice@test.com"]

    def test_extract_requirement_tags(self, tmp_path: Path) -> None:
        """Parse @REQ-XXX-NNN tags."""
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "requirements.feature"
        feature_file.write_text(
            dedent("""
            @REQ-AUTH-001 @REQ-AUTH-002
            Feature: Authentication

              @REQ-LOGIN-001
              Scenario: User login
                Given a registered user
                When they enter credentials
                Then they are logged in

              @smoke @REQ-LOGOUT-001
              Scenario: User logout
                Given a logged in user
                When they click logout
                Then they are logged out
            """).strip()
        )

        result = parse_feature(feature_file)

        # Feature-level requirement tags
        assert "REQ-AUTH-001" in result.requirement_tags
        assert "REQ-AUTH-002" in result.requirement_tags

        # Scenario-level requirement tags
        login = result.scenarios[0]
        assert "REQ-LOGIN-001" in login.requirement_tags
        # Inherited from feature
        assert "REQ-AUTH-001" in login.inherited_requirement_tags

        logout = result.scenarios[1]
        assert "REQ-LOGOUT-001" in logout.requirement_tags

    def test_i18n_keywords_french(self, tmp_path: Path) -> None:
        """Parse French Gherkin keywords."""
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "french.feature"
        feature_file.write_text(
            dedent("""
            # language: fr
            Fonctionnalité: Authentification
              En tant qu'utilisateur
              Je veux me connecter

              Scénario: Connexion réussie
                Soit un utilisateur enregistré
                Quand il entre ses identifiants
                Alors il est connecté
            """).strip()
        )

        result = parse_feature(feature_file)

        assert result.name == "Authentification"
        assert len(result.scenarios) == 1
        assert result.scenarios[0].name == "Connexion réussie"
        assert result.language == "fr"

    def test_i18n_keywords_german(self, tmp_path: Path) -> None:
        """Parse German Gherkin keywords."""
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "german.feature"
        feature_file.write_text(
            dedent("""
            # language: de
            Funktionalität: Authentifizierung
              Als Benutzer
              Möchte ich mich anmelden

              Szenario: Erfolgreiche Anmeldung
                Angenommen ein registrierter Benutzer
                Wenn er seine Anmeldedaten eingibt
                Dann ist er angemeldet
            """).strip()
        )

        result = parse_feature(feature_file)

        assert result.name == "Authentifizierung"
        assert len(result.scenarios) == 1
        assert result.language == "de"

    def test_malformed_feature_error_handling(self, tmp_path: Path) -> None:
        """Graceful error handling for malformed files."""
        from rtmx.bdd.parser import GherkinParseError, parse_feature

        feature_file = tmp_path / "malformed.feature"
        feature_file.write_text(
            dedent("""
            This is not valid Gherkin
            It has no Feature keyword
            """).strip()
        )

        with pytest.raises(GherkinParseError) as exc_info:
            parse_feature(feature_file)

        assert "malformed.feature" in str(exc_info.value)

    def test_parse_feature_cli_json_output(
        self, tmp_path: Path, cli_runner: pytest.fixture
    ) -> None:
        """CLI JSON output format."""
        from click.testing import CliRunner

        from rtmx.cli.main import main

        feature_file = tmp_path / "cli_test.feature"
        feature_file.write_text(
            dedent("""
            @REQ-CLI-001
            Feature: CLI Test

              Scenario: Test scenario
                Given a step
            """).strip()
        )

        runner = CliRunner()
        result = runner.invoke(main, ["parse-feature", str(feature_file), "--json"])

        assert result.exit_code == 0
        data = json.loads(result.output)
        # Output is wrapped in {"features": [...]}
        assert len(data["features"]) == 1
        feature = data["features"][0]
        assert feature["name"] == "CLI Test"
        assert "REQ-CLI-001" in feature["requirement_tags"]

    def test_recursive_feature_discovery(self, tmp_path: Path) -> None:
        """Glob pattern scanning."""
        from rtmx.bdd.parser import discover_features

        # Create nested directory structure
        (tmp_path / "features" / "auth").mkdir(parents=True)
        (tmp_path / "features" / "api").mkdir(parents=True)

        (tmp_path / "features" / "auth" / "login.feature").write_text(
            "Feature: Login\n  Scenario: Test\n    Given step"
        )
        (tmp_path / "features" / "auth" / "logout.feature").write_text(
            "Feature: Logout\n  Scenario: Test\n    Given step"
        )
        (tmp_path / "features" / "api" / "rest.feature").write_text(
            "Feature: REST API\n  Scenario: Test\n    Given step"
        )

        features = discover_features(tmp_path / "features")

        assert len(features) == 3
        names = {f.name for f in features}
        assert names == {"Login", "Logout", "REST API"}

    def test_scenario_outline_support(self, tmp_path: Path) -> None:
        """Parse Scenario Outline with Examples."""
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "outline.feature"
        feature_file.write_text(
            dedent("""
            Feature: Scenario Outline

              Scenario Outline: Login with different users
                Given a user "<username>"
                When they enter password "<password>"
                Then login should "<result>"

                Examples:
                  | username | password | result  |
                  | alice    | pass123  | succeed |
                  | bob      | wrong    | fail    |
            """).strip()
        )

        result = parse_feature(feature_file)

        assert len(result.scenarios) == 1
        scenario = result.scenarios[0]
        assert scenario.is_outline
        assert scenario.examples is not None
        assert len(scenario.examples.rows) == 3  # header + 2 examples

    def test_background_support(self, tmp_path: Path) -> None:
        """Parse Background section."""
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "background.feature"
        feature_file.write_text(
            dedent("""
            Feature: Background support

              Background:
                Given the system is initialized
                And the database is seeded

              Scenario: First test
                When I perform an action
                Then something happens

              Scenario: Second test
                When I do something else
                Then another thing happens
            """).strip()
        )

        result = parse_feature(feature_file)

        assert result.background is not None
        assert len(result.background.steps) == 2
        assert result.background.steps[0].text == "the system is initialized"

    def test_doc_string_support(self, tmp_path: Path) -> None:
        """Parse doc strings in steps."""
        from rtmx.bdd.parser import parse_feature

        feature_file = tmp_path / "docstring.feature"
        feature_file.write_text(
            dedent('''
            Feature: Doc strings

              Scenario: With doc string
                Given the following JSON:
                  """json
                  {
                    "name": "test",
                    "value": 123
                  }
                  """
                Then it should be parsed
            ''').strip()
        )

        result = parse_feature(feature_file)
        step = result.scenarios[0].steps[0]

        assert step.doc_string is not None
        assert "json" in step.doc_string.media_type or step.doc_string.media_type == "json"
        assert '"name": "test"' in step.doc_string.content


@pytest.fixture
def cli_runner():
    """Create a Click CLI test runner."""
    from click.testing import CliRunner

    return CliRunner()
