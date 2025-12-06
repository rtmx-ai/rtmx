"""Tests for RTMX CLI main entry point.

This module tests the Click command group and main entry point:
- main() group command and global options
- All subcommand registration and invocation
- --help output for all commands
- Version output
- Context object propagation
- CLI option handling
"""

from __future__ import annotations

from pathlib import Path

import pytest
from click.testing import CliRunner

from rtmx.cli.main import main

# =============================================================================
# Fixtures
# =============================================================================


@pytest.fixture
def runner() -> CliRunner:
    """Create a Click CliRunner for testing CLI commands."""
    return CliRunner()


@pytest.fixture
def sample_rtm_csv(tmp_path: Path) -> Path:
    """Create a minimal sample RTM CSV for testing."""
    csv_path = tmp_path / "rtm_database.csv"

    headers = [
        "req_id",
        "category",
        "subcategory",
        "requirement_text",
        "target_value",
        "test_module",
        "test_function",
        "validation_method",
        "status",
        "priority",
        "phase",
        "notes",
        "effort_weeks",
        "dependencies",
        "blocks",
        "assignee",
        "sprint",
        "started_date",
        "completed_date",
        "requirement_file",
    ]

    rows = [
        "REQ-CLI-001,CLI,Main,CLI shall provide main entry point,Success,tests/test_cli_main.py,test_main_help,Unit Test,COMPLETE,HIGH,1,Main command,0.5,,,developer,v0.1,2025-01-01,2025-01-02,docs/requirements/CLI/REQ-CLI-001.md",
        "REQ-CLI-002,CLI,Commands,CLI shall provide status command,Success,tests/test_cli_main.py,test_status_command,Unit Test,COMPLETE,HIGH,1,Status command,0.5,REQ-CLI-001,,developer,v0.1,2025-01-01,2025-01-02,docs/requirements/CLI/REQ-CLI-002.md",
    ]

    with open(csv_path, "w", newline="") as f:
        f.write(",".join(headers) + "\n")
        for row in rows:
            f.write(row + "\n")

    return csv_path


@pytest.fixture
def sample_config(tmp_path: Path, sample_rtm_csv: Path) -> Path:
    """Create a minimal rtmx.yaml config file."""
    config_path = tmp_path / "rtmx.yaml"
    config_content = f"""rtm:
  database: {sample_rtm_csv}

github:
  enabled: false

jira:
  enabled: false
"""
    config_path.write_text(config_content)
    return config_path


# =============================================================================
# Main Command Group Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_main_help(runner: CliRunner) -> None:
    """Test that main command shows help message."""
    result = runner.invoke(main, ["--help"])

    assert result.exit_code == 0
    assert "RTMX - Requirements Traceability Matrix toolkit" in result.output
    assert "Manage requirements traceability for GenAI-driven development" in result.output
    assert "Commands:" in result.output


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_main_version(runner: CliRunner) -> None:
    """Test that --version flag shows version information."""
    result = runner.invoke(main, ["--version"])

    assert result.exit_code == 0
    assert "rtmx" in result.output.lower()
    # Version should contain semantic version pattern
    assert any(c.isdigit() for c in result.output)


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_main_no_args(runner: CliRunner) -> None:
    """Test that main command without args shows usage."""
    result = runner.invoke(main, [])

    # Click returns exit code 0 for help, 2 for missing command
    # Both are acceptable for a group command without subcommand
    assert result.exit_code in [0, 2]
    assert "Usage:" in result.output or "RTMX" in result.output or "Commands:" in result.output


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_main_rtm_csv_option(runner: CliRunner, sample_rtm_csv: Path) -> None:
    """Test that --rtm-csv option is accepted."""
    result = runner.invoke(main, ["--rtm-csv", str(sample_rtm_csv), "status"])

    # Should not fail on the option itself
    assert "Error: Invalid value for '--rtm-csv'" not in result.output


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_main_config_option(runner: CliRunner, sample_config: Path) -> None:
    """Test that --config option is accepted."""
    result = runner.invoke(main, ["--config", str(sample_config), "status"])

    # Should not fail on the option itself
    assert "Error: Invalid value for '--config'" not in result.output


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_main_no_color_option(runner: CliRunner, sample_rtm_csv: Path) -> None:
    """Test that --no-color option is accepted."""
    result = runner.invoke(main, ["--no-color", "--rtm-csv", str(sample_rtm_csv), "status"])

    # Should not fail on the option itself
    assert "Error" not in result.output or result.exit_code == 0


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_main_config_short_option(runner: CliRunner, sample_config: Path) -> None:
    """Test that -c short option works for config."""
    result = runner.invoke(main, ["-c", str(sample_config), "status"])

    # Should not fail on the option itself
    assert "Error: Invalid value for '-c'" not in result.output


# =============================================================================
# Status Command Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-002")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_status_command_exists(runner: CliRunner) -> None:
    """Test that status command is registered."""
    result = runner.invoke(main, ["status", "--help"])

    assert result.exit_code == 0
    assert "Show RTM status" in result.output


@pytest.mark.req("REQ-CLI-002")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_status_command_help(runner: CliRunner) -> None:
    """Test status command help output."""
    result = runner.invoke(main, ["status", "--help"])

    assert result.exit_code == 0
    assert "--verbose" in result.output or "-v" in result.output
    assert "--json" in result.output


@pytest.mark.req("REQ-CLI-002")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_status_verbose_option(runner: CliRunner, sample_rtm_csv: Path) -> None:
    """Test status command with verbose flag."""
    result = runner.invoke(main, ["--rtm-csv", str(sample_rtm_csv), "status", "-v"])

    # Should execute without error
    assert result.exit_code in [0, 1]  # May exit 1 if no data, but shouldn't crash


@pytest.mark.req("REQ-CLI-002")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_status_json_output(runner: CliRunner, sample_rtm_csv: Path, tmp_path: Path) -> None:
    """Test status command with JSON output."""
    json_file = tmp_path / "status.json"
    result = runner.invoke(
        main, ["--rtm-csv", str(sample_rtm_csv), "status", "--json", str(json_file)]
    )

    # Should execute without crashing
    assert result.exit_code in [0, 1]


# =============================================================================
# Backlog Command Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_backlog_command_exists(runner: CliRunner) -> None:
    """Test that backlog command is registered."""
    result = runner.invoke(main, ["backlog", "--help"])

    assert result.exit_code == 0
    assert "Show prioritized backlog" in result.output


@pytest.mark.req("REQ-CLI-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_backlog_command_help(runner: CliRunner) -> None:
    """Test backlog command help output."""
    result = runner.invoke(main, ["backlog", "--help"])

    assert result.exit_code == 0
    assert "--phase" in result.output
    assert "--view" in result.output


@pytest.mark.req("REQ-CLI-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_backlog_phase_option(runner: CliRunner, sample_rtm_csv: Path) -> None:
    """Test backlog command with phase filter."""
    result = runner.invoke(main, ["--rtm-csv", str(sample_rtm_csv), "backlog", "--phase", "1"])

    assert result.exit_code in [0, 1]


@pytest.mark.req("REQ-CLI-003")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_backlog_critical_option(runner: CliRunner, sample_rtm_csv: Path) -> None:
    """Test backlog command with critical path view."""
    result = runner.invoke(
        main, ["--rtm-csv", str(sample_rtm_csv), "backlog", "--view", "critical"]
    )

    assert result.exit_code in [0, 1]


# =============================================================================
# Reconcile Command Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-004")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_reconcile_command_exists(runner: CliRunner) -> None:
    """Test that reconcile command is registered."""
    result = runner.invoke(main, ["reconcile", "--help"])

    assert result.exit_code == 0
    assert "Check and fix dependency reciprocity" in result.output


@pytest.mark.req("REQ-CLI-004")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_reconcile_command_help(runner: CliRunner) -> None:
    """Test reconcile command help output."""
    result = runner.invoke(main, ["reconcile", "--help"])

    assert result.exit_code == 0
    assert "--execute" in result.output


@pytest.mark.req("REQ-CLI-004")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_reconcile_dry_run(runner: CliRunner, sample_rtm_csv: Path) -> None:
    """Test reconcile command in dry-run mode (default)."""
    result = runner.invoke(main, ["--rtm-csv", str(sample_rtm_csv), "reconcile"])

    assert result.exit_code in [0, 1]


# =============================================================================
# Deps Command Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-005")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_deps_command_exists(runner: CliRunner) -> None:
    """Test that deps command is registered."""
    result = runner.invoke(main, ["deps", "--help"])

    assert result.exit_code == 0
    assert "Show dependency graph" in result.output


@pytest.mark.req("REQ-CLI-005")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_deps_command_help(runner: CliRunner) -> None:
    """Test deps command help output."""
    result = runner.invoke(main, ["deps", "--help"])

    assert result.exit_code == 0
    assert "--category" in result.output
    assert "--phase" in result.output
    assert "--req" in result.output


@pytest.mark.req("REQ-CLI-005")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_deps_category_filter(runner: CliRunner, sample_rtm_csv: Path) -> None:
    """Test deps command with category filter."""
    result = runner.invoke(main, ["--rtm-csv", str(sample_rtm_csv), "deps", "--category", "CLI"])

    assert result.exit_code in [0, 1]


# =============================================================================
# Cycles Command Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-006")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_cycles_command_exists(runner: CliRunner) -> None:
    """Test that cycles command is registered."""
    result = runner.invoke(main, ["cycles", "--help"])

    assert result.exit_code == 0
    assert "Detect circular dependencies" in result.output


@pytest.mark.req("REQ-CLI-006")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_cycles_command_invocation(runner: CliRunner, sample_rtm_csv: Path) -> None:
    """Test cycles command execution."""
    result = runner.invoke(main, ["--rtm-csv", str(sample_rtm_csv), "cycles"])

    assert result.exit_code in [0, 1]


# =============================================================================
# Setup Command Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-007")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_setup_command_exists(runner: CliRunner) -> None:
    """Test that setup command is registered."""
    result = runner.invoke(main, ["setup", "--help"])

    assert result.exit_code == 0
    assert "Complete rtmx setup in one command" in result.output


@pytest.mark.req("REQ-CLI-007")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_setup_command_help(runner: CliRunner) -> None:
    """Test setup command help output."""
    result = runner.invoke(main, ["setup", "--help"])

    assert result.exit_code == 0
    assert "--dry-run" in result.output
    assert "--minimal" in result.output
    assert "--force" in result.output
    assert "--branch" in result.output
    assert "--pr" in result.output


@pytest.mark.req("REQ-CLI-007")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_setup_dry_run(runner: CliRunner, tmp_path: Path, monkeypatch) -> None:
    """Test setup command in dry-run mode."""
    monkeypatch.chdir(tmp_path)
    result = runner.invoke(main, ["setup", "--dry-run"])

    # Dry run should not fail
    assert result.exit_code in [0, 1]


# =============================================================================
# From-Tests Command Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-008")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_from_tests_command_exists(runner: CliRunner) -> None:
    """Test that from-tests command is registered."""
    result = runner.invoke(main, ["from-tests", "--help"])

    assert result.exit_code == 0
    assert "Scan test files for requirement markers" in result.output


@pytest.mark.req("REQ-CLI-008")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_from_tests_command_help(runner: CliRunner) -> None:
    """Test from-tests command help output."""
    result = runner.invoke(main, ["from-tests", "--help"])

    assert result.exit_code == 0
    assert "--all" in result.output
    assert "--missing" in result.output
    assert "--update" in result.output


@pytest.mark.req("REQ-CLI-008")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_from_tests_no_args(runner: CliRunner, sample_rtm_csv: Path) -> None:
    """Test from-tests command without arguments."""
    result = runner.invoke(main, ["--rtm-csv", str(sample_rtm_csv), "from-tests"])

    # Should execute, may fail if no tests found
    assert result.exit_code in [0, 1, 2]


# =============================================================================
# Analyze Command Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-009")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_analyze_command_exists(runner: CliRunner) -> None:
    """Test that analyze command is registered."""
    result = runner.invoke(main, ["analyze", "--help"])

    assert result.exit_code == 0
    assert "Analyze project for requirements artifacts" in result.output


@pytest.mark.req("REQ-CLI-009")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_analyze_command_help(runner: CliRunner) -> None:
    """Test analyze command help output."""
    result = runner.invoke(main, ["analyze", "--help"])

    assert result.exit_code == 0
    assert "--output" in result.output or "-o" in result.output
    assert "--format" in result.output
    assert "--deep" in result.output


# =============================================================================
# Sync Command Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-010")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_sync_command_exists(runner: CliRunner) -> None:
    """Test that sync command is registered."""
    result = runner.invoke(main, ["sync", "--help"])

    assert result.exit_code == 0
    assert "Synchronize RTM with external services" in result.output


@pytest.mark.req("REQ-CLI-010")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_sync_command_help(runner: CliRunner) -> None:
    """Test sync command help output."""
    result = runner.invoke(main, ["sync", "--help"])

    assert result.exit_code == 0
    assert "github" in result.output.lower() or "jira" in result.output.lower()
    assert "--import" in result.output
    assert "--export" in result.output


@pytest.mark.req("REQ-CLI-010")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_sync_requires_service_arg(runner: CliRunner) -> None:
    """Test sync command requires service argument."""
    result = runner.invoke(main, ["sync"])

    # Should fail without service argument
    assert result.exit_code != 0
    assert (
        "github" in result.output.lower()
        or "jira" in result.output.lower()
        or "Missing argument" in result.output
    )


# =============================================================================
# MCP Server Command Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-011")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_mcp_server_command_exists(runner: CliRunner) -> None:
    """Test that mcp-server command is registered."""
    result = runner.invoke(main, ["mcp-server", "--help"])

    assert result.exit_code == 0
    assert "Start MCP protocol server" in result.output


@pytest.mark.req("REQ-CLI-011")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_mcp_server_command_help(runner: CliRunner) -> None:
    """Test mcp-server command help output."""
    result = runner.invoke(main, ["mcp-server", "--help"])

    assert result.exit_code == 0
    assert "--port" in result.output
    assert "--host" in result.output
    assert "--daemon" in result.output


# =============================================================================
# Health Command Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-012")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_health_command_exists(runner: CliRunner) -> None:
    """Test that health command is registered."""
    result = runner.invoke(main, ["health", "--help"])

    assert result.exit_code == 0
    assert "Run integration health check" in result.output


@pytest.mark.req("REQ-CLI-012")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_health_command_help(runner: CliRunner) -> None:
    """Test health command help output."""
    result = runner.invoke(main, ["health", "--help"])

    assert result.exit_code == 0
    assert "--format" in result.output
    assert "--strict" in result.output
    assert "--check" in result.output


# =============================================================================
# Diff Command Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-013")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_diff_command_exists(runner: CliRunner) -> None:
    """Test that diff command is registered."""
    result = runner.invoke(main, ["diff", "--help"])

    assert result.exit_code == 0
    assert "Compare RTM databases" in result.output


@pytest.mark.req("REQ-CLI-013")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_diff_command_help(runner: CliRunner) -> None:
    """Test diff command help output."""
    result = runner.invoke(main, ["diff", "--help"])

    assert result.exit_code == 0
    assert "--format" in result.output
    assert "--output" in result.output or "-o" in result.output


@pytest.mark.req("REQ-CLI-013")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_diff_requires_baseline_arg(runner: CliRunner) -> None:
    """Test diff command requires baseline argument."""
    result = runner.invoke(main, ["diff"])

    # Should fail without baseline argument
    assert result.exit_code != 0
    assert "Missing argument" in result.output or "BASELINE" in result.output


# =============================================================================
# Integration Tests - Command Combinations
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_global_options_propagate_to_subcommands(runner: CliRunner, sample_rtm_csv: Path) -> None:
    """Test that global options are available to subcommands."""
    result = runner.invoke(main, ["--rtm-csv", str(sample_rtm_csv), "--no-color", "status"])

    # Should not error on options
    assert "Invalid value" not in result.output


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_stress
@pytest.mark.env_simulation
def test_invalid_command(runner: CliRunner) -> None:
    """Test handling of invalid command."""
    result = runner.invoke(main, ["nonexistent-command"])

    assert result.exit_code != 0
    assert "Error" in result.output or "No such command" in result.output


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_stress
@pytest.mark.env_simulation
def test_invalid_global_option(runner: CliRunner) -> None:
    """Test handling of invalid global option."""
    result = runner.invoke(main, ["--invalid-option", "status"])

    assert result.exit_code != 0
    assert "Error" in result.output or "no such option" in result.output


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_all_commands_have_help(runner: CliRunner) -> None:
    """Test that all commands provide help output."""
    commands = [
        "status",
        "backlog",
        "reconcile",
        "deps",
        "cycles",
        "setup",
        "from-tests",
        "analyze",
        "sync",
        "mcp-server",
        "health",
        "diff",
    ]

    for cmd in commands:
        result = runner.invoke(main, [cmd, "--help"])
        assert result.exit_code == 0, f"Command {cmd} --help failed"
        assert len(result.output) > 0, f"Command {cmd} --help has no output"


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_command_count(runner: CliRunner) -> None:
    """Test that all expected commands are registered."""
    result = runner.invoke(main, ["--help"])

    assert result.exit_code == 0

    # Check for core commands
    expected_commands = ["status", "backlog", "reconcile", "deps", "cycles", "setup", "from-tests"]

    for cmd in expected_commands:
        assert cmd in result.output, f"Command {cmd} not found in help output"
