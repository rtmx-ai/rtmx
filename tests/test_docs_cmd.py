"""Tests for rtmx docs command (REQ-DX-004).

This module tests the docs command for auto-generating documentation:
- rtmx docs schema: Generate schema.md from rtmx.schema module
- rtmx docs config: Generate config reference from RTMXConfig
- Output to .rtmx/cache/ by default
- --output flag for custom location
- Generated docs include version and timestamp
"""

from __future__ import annotations

from datetime import datetime
from pathlib import Path

import pytest

from rtmx import __version__

# =============================================================================
# Tests for docs schema subcommand
# =============================================================================


@pytest.mark.req("REQ-DX-004")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestDocsSchemaCommand:
    """Tests for rtmx docs schema command."""

    def test_docs_schema_generates_markdown(self, tmp_path: Path) -> None:
        """docs schema should generate schema.md file."""
        from rtmx.cli.docs import run_docs_schema

        output_path = tmp_path / "schema.md"
        run_docs_schema(output=output_path)

        assert output_path.exists()
        content = output_path.read_text()
        assert content.strip()  # Not empty

    def test_docs_schema_contains_core_schema(self, tmp_path: Path) -> None:
        """docs schema should document the core schema."""
        from rtmx.cli.docs import run_docs_schema

        output_path = tmp_path / "schema.md"
        run_docs_schema(output=output_path)

        content = output_path.read_text()
        assert "core" in content.lower()
        assert "req_id" in content
        assert "category" in content
        assert "status" in content

    def test_docs_schema_contains_phoenix_schema(self, tmp_path: Path) -> None:
        """docs schema should document the phoenix schema extension."""
        from rtmx.cli.docs import run_docs_schema

        output_path = tmp_path / "schema.md"
        run_docs_schema(output=output_path)

        content = output_path.read_text()
        assert "phoenix" in content.lower()
        assert "scope_unit" in content
        assert "technique_nominal" in content

    def test_docs_schema_contains_column_types(self, tmp_path: Path) -> None:
        """docs schema should document column types."""
        from rtmx.cli.docs import run_docs_schema

        output_path = tmp_path / "schema.md"
        run_docs_schema(output=output_path)

        content = output_path.read_text()
        assert "string" in content.lower()
        assert "integer" in content.lower()
        assert "boolean" in content.lower()

    def test_docs_schema_contains_version(self, tmp_path: Path) -> None:
        """docs schema should include rtmx version."""
        from rtmx.cli.docs import run_docs_schema

        output_path = tmp_path / "schema.md"
        run_docs_schema(output=output_path)

        content = output_path.read_text()
        assert __version__ in content

    def test_docs_schema_contains_timestamp(self, tmp_path: Path) -> None:
        """docs schema should include generation timestamp."""
        from rtmx.cli.docs import run_docs_schema

        output_path = tmp_path / "schema.md"
        run_docs_schema(output=output_path)

        content = output_path.read_text()
        # Should contain a date in YYYY-MM-DD format
        today = datetime.now().strftime("%Y-%m-%d")
        assert today in content

    def test_docs_schema_default_output_location(self, tmp_path: Path) -> None:
        """docs schema without --output should write to .rtmx/cache/."""
        import os

        from rtmx.cli.docs import run_docs_schema

        os.chdir(tmp_path)
        cache_dir = tmp_path / ".rtmx" / "cache"
        cache_dir.mkdir(parents=True, exist_ok=True)

        run_docs_schema(output=None)

        expected_path = cache_dir / "schema.md"
        assert expected_path.exists()

    def test_docs_schema_creates_cache_dir(self, tmp_path: Path) -> None:
        """docs schema should create .rtmx/cache/ if it doesn't exist."""
        import os

        from rtmx.cli.docs import run_docs_schema

        os.chdir(tmp_path)

        run_docs_schema(output=None)

        cache_dir = tmp_path / ".rtmx" / "cache"
        assert cache_dir.exists()
        assert (cache_dir / "schema.md").exists()

    def test_docs_schema_documents_required_columns(self, tmp_path: Path) -> None:
        """docs schema should indicate which columns are required."""
        from rtmx.cli.docs import run_docs_schema

        output_path = tmp_path / "schema.md"
        run_docs_schema(output=output_path)

        content = output_path.read_text()
        # Should indicate required status somehow
        assert "required" in content.lower()

    def test_docs_schema_includes_descriptions(self, tmp_path: Path) -> None:
        """docs schema should include column descriptions."""
        from rtmx.cli.docs import run_docs_schema

        output_path = tmp_path / "schema.md"
        run_docs_schema(output=output_path)

        content = output_path.read_text()
        # Should include known descriptions from schema.py
        assert "Unique requirement identifier" in content


# =============================================================================
# Tests for docs config subcommand
# =============================================================================


@pytest.mark.req("REQ-DX-004")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestDocsConfigCommand:
    """Tests for rtmx docs config command."""

    def test_docs_config_generates_markdown(self, tmp_path: Path) -> None:
        """docs config should generate config.md file."""
        from rtmx.cli.docs import run_docs_config

        output_path = tmp_path / "config.md"
        run_docs_config(output=output_path)

        assert output_path.exists()
        content = output_path.read_text()
        assert content.strip()  # Not empty

    def test_docs_config_contains_database_setting(self, tmp_path: Path) -> None:
        """docs config should document database configuration."""
        from rtmx.cli.docs import run_docs_config

        output_path = tmp_path / "config.md"
        run_docs_config(output=output_path)

        content = output_path.read_text()
        assert "database" in content.lower()
        assert "docs/rtm_database.csv" in content  # Default value

    def test_docs_config_contains_agents_section(self, tmp_path: Path) -> None:
        """docs config should document agents configuration."""
        from rtmx.cli.docs import run_docs_config

        output_path = tmp_path / "config.md"
        run_docs_config(output=output_path)

        content = output_path.read_text()
        assert "agents" in content.lower()
        assert "claude" in content.lower()
        assert "cursor" in content.lower()
        assert "copilot" in content.lower()

    def test_docs_config_contains_adapters_section(self, tmp_path: Path) -> None:
        """docs config should document adapters configuration."""
        from rtmx.cli.docs import run_docs_config

        output_path = tmp_path / "config.md"
        run_docs_config(output=output_path)

        content = output_path.read_text()
        assert "adapters" in content.lower()
        assert "github" in content.lower()
        assert "jira" in content.lower()

    def test_docs_config_contains_mcp_section(self, tmp_path: Path) -> None:
        """docs config should document MCP server configuration."""
        from rtmx.cli.docs import run_docs_config

        output_path = tmp_path / "config.md"
        run_docs_config(output=output_path)

        content = output_path.read_text()
        assert "mcp" in content.lower()
        assert "port" in content.lower()
        assert "3000" in content  # Default port

    def test_docs_config_contains_version(self, tmp_path: Path) -> None:
        """docs config should include rtmx version."""
        from rtmx.cli.docs import run_docs_config

        output_path = tmp_path / "config.md"
        run_docs_config(output=output_path)

        content = output_path.read_text()
        assert __version__ in content

    def test_docs_config_contains_timestamp(self, tmp_path: Path) -> None:
        """docs config should include generation timestamp."""
        from rtmx.cli.docs import run_docs_config

        output_path = tmp_path / "config.md"
        run_docs_config(output=output_path)

        content = output_path.read_text()
        today = datetime.now().strftime("%Y-%m-%d")
        assert today in content

    def test_docs_config_default_output_location(self, tmp_path: Path) -> None:
        """docs config without --output should write to .rtmx/cache/."""
        import os

        from rtmx.cli.docs import run_docs_config

        os.chdir(tmp_path)
        cache_dir = tmp_path / ".rtmx" / "cache"
        cache_dir.mkdir(parents=True, exist_ok=True)

        run_docs_config(output=None)

        expected_path = cache_dir / "config.md"
        assert expected_path.exists()

    def test_docs_config_shows_default_values(self, tmp_path: Path) -> None:
        """docs config should show default values for settings."""
        from rtmx.cli.docs import run_docs_config

        output_path = tmp_path / "config.md"
        run_docs_config(output=output_path)

        content = output_path.read_text()
        # Should show default values
        assert "docs/requirements" in content  # Default requirements_dir
        assert "core" in content  # Default schema

    def test_docs_config_documents_yaml_structure(self, tmp_path: Path) -> None:
        """docs config should show example YAML configuration."""
        from rtmx.cli.docs import run_docs_config

        output_path = tmp_path / "config.md"
        run_docs_config(output=output_path)

        content = output_path.read_text()
        # Should contain YAML example
        assert "rtmx:" in content or "yaml" in content.lower()


# =============================================================================
# Tests for CLI integration
# =============================================================================


@pytest.mark.req("REQ-DX-004")
@pytest.mark.scope_integration
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestDocsCommandCLI:
    """Integration tests for docs CLI command."""

    def test_docs_command_group_exists(self) -> None:
        """docs command group should be registered in CLI."""
        from click.testing import CliRunner

        from rtmx.cli.main import main

        runner = CliRunner()
        result = runner.invoke(main, ["docs", "--help"])

        assert result.exit_code == 0
        assert "schema" in result.output
        assert "config" in result.output

    def test_docs_schema_cli(self, tmp_path: Path) -> None:
        """docs schema should work via CLI."""
        from click.testing import CliRunner

        from rtmx.cli.main import main

        output_path = tmp_path / "schema.md"
        runner = CliRunner()
        result = runner.invoke(main, ["docs", "schema", "--output", str(output_path)])

        assert result.exit_code == 0
        assert output_path.exists()

    def test_docs_config_cli(self, tmp_path: Path) -> None:
        """docs config should work via CLI."""
        from click.testing import CliRunner

        from rtmx.cli.main import main

        output_path = tmp_path / "config.md"
        runner = CliRunner()
        result = runner.invoke(main, ["docs", "config", "--output", str(output_path)])

        assert result.exit_code == 0
        assert output_path.exists()

    def test_docs_schema_prints_path_on_success(self, tmp_path: Path) -> None:
        """docs schema should print output path on success."""
        from click.testing import CliRunner

        from rtmx.cli.main import main

        output_path = tmp_path / "schema.md"
        runner = CliRunner()
        result = runner.invoke(main, ["docs", "schema", "--output", str(output_path)])

        assert result.exit_code == 0
        assert str(output_path) in result.output or "schema.md" in result.output

    def test_docs_config_prints_path_on_success(self, tmp_path: Path) -> None:
        """docs config should print output path on success."""
        from click.testing import CliRunner

        from rtmx.cli.main import main

        output_path = tmp_path / "config.md"
        runner = CliRunner()
        result = runner.invoke(main, ["docs", "config", "--output", str(output_path)])

        assert result.exit_code == 0
        assert str(output_path) in result.output or "config.md" in result.output


# =============================================================================
# Tests for edge cases and error handling
# =============================================================================


@pytest.mark.req("REQ-DX-004")
@pytest.mark.scope_unit
@pytest.mark.technique_stress
@pytest.mark.env_simulation
class TestDocsCommandEdgeCases:
    """Edge case tests for docs command."""

    def test_docs_schema_overwrites_existing(self, tmp_path: Path) -> None:
        """docs schema should overwrite existing file."""
        from rtmx.cli.docs import run_docs_schema

        output_path = tmp_path / "schema.md"
        output_path.write_text("old content")

        run_docs_schema(output=output_path)

        content = output_path.read_text()
        assert "old content" not in content
        assert "req_id" in content

    def test_docs_config_overwrites_existing(self, tmp_path: Path) -> None:
        """docs config should overwrite existing file."""
        from rtmx.cli.docs import run_docs_config

        output_path = tmp_path / "config.md"
        output_path.write_text("old content")

        run_docs_config(output=output_path)

        content = output_path.read_text()
        assert "old content" not in content
        assert "database" in content.lower()

    def test_docs_schema_creates_parent_dirs(self, tmp_path: Path) -> None:
        """docs schema should create parent directories if needed."""
        from rtmx.cli.docs import run_docs_schema

        output_path = tmp_path / "nested" / "dir" / "schema.md"
        run_docs_schema(output=output_path)

        assert output_path.exists()

    def test_docs_config_creates_parent_dirs(self, tmp_path: Path) -> None:
        """docs config should create parent directories if needed."""
        from rtmx.cli.docs import run_docs_config

        output_path = tmp_path / "nested" / "dir" / "config.md"
        run_docs_config(output=output_path)

        assert output_path.exists()

    def test_docs_schema_handles_unicode(self, tmp_path: Path) -> None:
        """docs schema output should be valid UTF-8."""
        from rtmx.cli.docs import run_docs_schema

        output_path = tmp_path / "schema.md"
        run_docs_schema(output=output_path)

        # Should not raise encoding errors
        content = output_path.read_text(encoding="utf-8")
        assert content

    def test_docs_config_handles_unicode(self, tmp_path: Path) -> None:
        """docs config output should be valid UTF-8."""
        from rtmx.cli.docs import run_docs_config

        output_path = tmp_path / "config.md"
        run_docs_config(output=output_path)

        # Should not raise encoding errors
        content = output_path.read_text(encoding="utf-8")
        assert content
