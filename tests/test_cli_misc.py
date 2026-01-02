"""Tests for miscellaneous RTMX CLI commands.

This module tests:
- run_init: Initialize RTM structure in a project
- run_install: Install RTM-aware prompts into AI agent configs
- run_bootstrap: Generate initial RTM from project artifacts
- generate_makefile_targets: Generate Makefile targets for RTM commands
"""

from __future__ import annotations

from pathlib import Path

import pytest

from rtmx.cli.bootstrap import run_bootstrap
from rtmx.cli.init import run_init
from rtmx.cli.install import (
    CLAUDE_PROMPT,
    COPILOT_PROMPT,
    CURSOR_PROMPT,
    detect_agent_configs,
    get_agent_prompt,
    run_install,
)
from rtmx.cli.makefile import MAKEFILE_TARGETS, run_makefile
from rtmx.config import RTMXConfig

# =============================================================================
# Tests for run_init
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_init_creates_structure(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> None:
    """Test that run_init creates the expected .rtmx/ directory structure."""
    monkeypatch.chdir(tmp_path)

    run_init(force=False)

    # Check .rtmx/ directories created
    assert (tmp_path / ".rtmx").exists()
    assert (tmp_path / ".rtmx" / "requirements").exists()
    assert (tmp_path / ".rtmx" / "requirements" / "EXAMPLE").exists()
    assert (tmp_path / ".rtmx" / "cache").exists()

    # Check files created
    assert (tmp_path / ".rtmx" / "database.csv").exists()
    assert (tmp_path / ".rtmx" / "config.yaml").exists()
    assert (tmp_path / ".rtmx" / ".gitignore").exists()
    assert (tmp_path / ".rtmx" / "requirements" / "EXAMPLE" / "REQ-EX-001.md").exists()


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_init_sample_csv_content(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> None:
    """Test that run_init creates valid CSV content."""
    monkeypatch.chdir(tmp_path)

    run_init(force=False)

    csv_path = tmp_path / ".rtmx" / "database.csv"
    content = csv_path.read_text()

    assert "req_id,category,subcategory" in content
    assert "REQ-EX-001" in content
    assert "EXAMPLE" in content


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_init_sample_requirement_content(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    """Test that run_init creates valid requirement file."""
    monkeypatch.chdir(tmp_path)

    run_init(force=False)

    req_file = tmp_path / ".rtmx" / "requirements" / "EXAMPLE" / "REQ-EX-001.md"
    content = req_file.read_text()

    assert "# REQ-EX-001" in content
    assert "## Description" in content
    assert "## Acceptance Criteria" in content


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_init_config_file_content(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> None:
    """Test that run_init creates valid config file in .rtmx/."""
    monkeypatch.chdir(tmp_path)

    run_init(force=False)

    config_file = tmp_path / ".rtmx" / "config.yaml"
    content = config_file.read_text()

    assert "rtmx:" in content
    assert "database: .rtmx/database.csv" in content
    assert "requirements_dir: .rtmx/requirements" in content


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_init_existing_files_without_force(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    """Test that run_init exits when .rtmx/ exists and force=False."""
    monkeypatch.chdir(tmp_path)

    # Create existing .rtmx directory
    (tmp_path / ".rtmx").mkdir(parents=True, exist_ok=True)

    with pytest.raises(SystemExit) as exc_info:
        run_init(force=False)

    assert exc_info.value.code == 1


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_init_existing_files_with_force(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    """Test that run_init overwrites files when force=True."""
    monkeypatch.chdir(tmp_path)

    # Create existing .rtmx directory with files
    (tmp_path / ".rtmx").mkdir(parents=True, exist_ok=True)
    (tmp_path / ".rtmx" / "database.csv").write_text("existing")

    run_init(force=True)

    # Files should be overwritten
    csv_content = (tmp_path / ".rtmx" / "database.csv").read_text()
    assert "REQ-EX-001" in csv_content
    assert "existing" not in csv_content


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_init_prints_instructions(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture
) -> None:
    """Test that run_init prints next steps."""
    monkeypatch.chdir(tmp_path)

    run_init(force=False)

    captured = capsys.readouterr()
    assert "RTM initialized successfully" in captured.out
    assert "Next steps:" in captured.out
    assert "rtmx status" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_init_legacy_mode(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> None:
    """Test that run_init with use_rtmx_dir=False creates legacy structure."""
    monkeypatch.chdir(tmp_path)

    run_init(force=False, use_rtmx_dir=False)

    # Check legacy directories created
    assert (tmp_path / "docs").exists()
    assert (tmp_path / "docs" / "requirements").exists()
    assert (tmp_path / "docs" / "rtm_database.csv").exists()
    assert (tmp_path / "rtmx.yaml").exists()

    # Check .rtmx does NOT exist
    assert not (tmp_path / ".rtmx").exists()


# =============================================================================
# Tests for detect_agent_configs
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_agent_configs_none_exist(tmp_path: Path) -> None:
    """Test detect_agent_configs when no agent configs exist."""
    result = detect_agent_configs(tmp_path)

    assert result["claude"] is None
    assert result["cursor"] is None
    assert result["copilot"] is None


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_agent_configs_claude_in_root(tmp_path: Path) -> None:
    """Test detect_agent_configs finds CLAUDE.md in root."""
    claude_file = tmp_path / "CLAUDE.md"
    claude_file.write_text("# Claude config")

    result = detect_agent_configs(tmp_path)

    assert result["claude"] == claude_file


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_agent_configs_claude_in_dotclaude(tmp_path: Path) -> None:
    """Test detect_agent_configs finds CLAUDE.md in .claude directory."""
    claude_dir = tmp_path / ".claude"
    claude_dir.mkdir()
    claude_file = claude_dir / "CLAUDE.md"
    claude_file.write_text("# Claude config")

    result = detect_agent_configs(tmp_path)

    assert result["claude"] == claude_file


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_agent_configs_cursor(tmp_path: Path) -> None:
    """Test detect_agent_configs finds .cursorrules."""
    cursor_file = tmp_path / ".cursorrules"
    cursor_file.write_text("# Cursor rules")

    result = detect_agent_configs(tmp_path)

    assert result["cursor"] == cursor_file


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_agent_configs_copilot(tmp_path: Path) -> None:
    """Test detect_agent_configs finds copilot instructions."""
    github_dir = tmp_path / ".github"
    github_dir.mkdir()
    copilot_file = github_dir / "copilot-instructions.md"
    copilot_file.write_text("# Copilot instructions")

    result = detect_agent_configs(tmp_path)

    assert result["copilot"] == copilot_file


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_agent_configs_all(tmp_path: Path) -> None:
    """Test detect_agent_configs finds all agent configs."""
    # Create all configs
    (tmp_path / "CLAUDE.md").write_text("# Claude")
    (tmp_path / ".cursorrules").write_text("# Cursor")
    (tmp_path / ".github").mkdir()
    (tmp_path / ".github" / "copilot-instructions.md").write_text("# Copilot")

    result = detect_agent_configs(tmp_path)

    assert result["claude"] is not None
    assert result["cursor"] is not None
    assert result["copilot"] is not None


# =============================================================================
# Tests for get_agent_prompt
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_get_agent_prompt_claude() -> None:
    """Test get_agent_prompt returns Claude prompt."""
    prompt = get_agent_prompt("claude")

    assert prompt == CLAUDE_PROMPT
    assert "RTMX Requirements Traceability" in prompt


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_get_agent_prompt_cursor() -> None:
    """Test get_agent_prompt returns Cursor prompt."""
    prompt = get_agent_prompt("cursor")

    assert prompt == CURSOR_PROMPT
    assert "RTMX Requirements Traceability" in prompt


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_get_agent_prompt_copilot() -> None:
    """Test get_agent_prompt returns Copilot prompt."""
    prompt = get_agent_prompt("copilot")

    assert prompt == COPILOT_PROMPT
    assert "RTMX Requirements Traceability" in prompt


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_get_agent_prompt_unknown() -> None:
    """Test get_agent_prompt with unknown agent returns empty string."""
    prompt = get_agent_prompt("unknown")

    assert prompt == ""


# =============================================================================
# Tests for run_install
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_install_dry_run(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture
) -> None:
    """Test run_install in dry run mode."""
    monkeypatch.chdir(tmp_path)
    config = RTMXConfig()

    run_install(
        dry_run=True,
        non_interactive=True,
        force=False,
        agents=["claude"],
        install_all=False,
        skip_backup=False,
        _config=config,
    )

    captured = capsys.readouterr()
    assert "DRY RUN" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_install_create_new_claude_file(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    """Test run_install creates new CLAUDE.md file."""
    monkeypatch.chdir(tmp_path)
    config = RTMXConfig()

    run_install(
        dry_run=False,
        non_interactive=True,
        force=False,
        agents=["claude"],
        install_all=False,
        skip_backup=False,
        _config=config,
    )

    claude_file = tmp_path / "CLAUDE.md"
    assert claude_file.exists()
    content = claude_file.read_text()
    assert "RTMX Requirements Traceability" in content


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_install_append_to_existing(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> None:
    """Test run_install appends to existing config file."""
    monkeypatch.chdir(tmp_path)
    config = RTMXConfig()

    # Create existing file
    claude_file = tmp_path / "CLAUDE.md"
    claude_file.write_text("# Existing content\n")

    run_install(
        dry_run=False,
        non_interactive=True,
        force=False,
        agents=["claude"],
        install_all=False,
        skip_backup=False,
        _config=config,
    )

    content = claude_file.read_text()
    assert "Existing content" in content
    assert "RTMX Requirements Traceability" in content


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_install_skip_if_exists_without_force(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture
) -> None:
    """Test run_install skips when RTMX section exists and force=False."""
    monkeypatch.chdir(tmp_path)
    config = RTMXConfig()

    # Create file with RTMX section
    claude_file = tmp_path / "CLAUDE.md"
    claude_file.write_text("## RTMX Requirements Traceability\nExisting RTMX content\n")

    run_install(
        dry_run=False,
        non_interactive=True,
        force=False,
        agents=["claude"],
        install_all=False,
        skip_backup=False,
        _config=config,
    )

    captured = capsys.readouterr()
    assert "already exists" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_install_force_overwrites(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> None:
    """Test run_install overwrites RTMX section when force=True."""
    monkeypatch.chdir(tmp_path)
    config = RTMXConfig()

    # Create file with old RTMX section
    claude_file = tmp_path / "CLAUDE.md"
    claude_file.write_text(
        "# Start\n## RTMX Requirements Traceability\nOld content\n## Other Section\nOther\n"
    )

    run_install(
        dry_run=False,
        non_interactive=True,
        force=True,
        agents=["claude"],
        install_all=False,
        skip_backup=False,
        _config=config,
    )

    content = claude_file.read_text()
    assert "Old content" not in content
    assert "RTMX Requirements Traceability" in content
    assert "Other Section" in content


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_install_creates_backup(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> None:
    """Test run_install creates backup of existing file."""
    monkeypatch.chdir(tmp_path)
    config = RTMXConfig()

    # Create existing file
    claude_file = tmp_path / "CLAUDE.md"
    claude_file.write_text("# Existing content\n")

    run_install(
        dry_run=False,
        non_interactive=True,
        force=False,
        agents=["claude"],
        install_all=False,
        skip_backup=False,
        _config=config,
    )

    # Check for backup file
    backup_files = list(tmp_path.glob("CLAUDE.rtmx-backup-*.md"))
    assert len(backup_files) == 1


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_install_skip_backup(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> None:
    """Test run_install skips backup when skip_backup=True."""
    monkeypatch.chdir(tmp_path)
    config = RTMXConfig()

    # Create existing file
    claude_file = tmp_path / "CLAUDE.md"
    claude_file.write_text("# Existing content\n")

    run_install(
        dry_run=False,
        non_interactive=True,
        force=False,
        agents=["claude"],
        install_all=False,
        skip_backup=True,
        _config=config,
    )

    # No backup should exist
    backup_files = list(tmp_path.glob("CLAUDE.rtmx-backup-*.md"))
    assert len(backup_files) == 0


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_install_multiple_agents(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> None:
    """Test run_install with multiple agents."""
    monkeypatch.chdir(tmp_path)
    config = RTMXConfig()

    run_install(
        dry_run=False,
        non_interactive=True,
        force=False,
        agents=["claude", "cursor"],
        install_all=False,
        skip_backup=False,
        _config=config,
    )

    assert (tmp_path / "CLAUDE.md").exists()
    assert (tmp_path / ".cursorrules").exists()


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_install_copilot_creates_directory(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    """Test run_install creates .github directory for copilot."""
    monkeypatch.chdir(tmp_path)
    config = RTMXConfig()

    run_install(
        dry_run=False,
        non_interactive=True,
        force=False,
        agents=["copilot"],
        install_all=False,
        skip_backup=False,
        _config=config,
    )

    assert (tmp_path / ".github").exists()
    assert (tmp_path / ".github" / "copilot-instructions.md").exists()


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_install_no_agents_selected(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch, capsys: pytest.CaptureFixture
) -> None:
    """Test run_install when no agents are selected."""
    monkeypatch.chdir(tmp_path)
    config = RTMXConfig()

    run_install(
        dry_run=False,
        non_interactive=True,
        force=False,
        agents=[],
        install_all=False,
        skip_backup=False,
        _config=config,
    )

    captured = capsys.readouterr()
    assert "No agents selected" in captured.out


# =============================================================================
# Tests for run_bootstrap
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_bootstrap_no_source_exits(capsys: pytest.CaptureFixture) -> None:
    """Test that run_bootstrap exits when no source is specified."""
    config = RTMXConfig()

    with pytest.raises(SystemExit) as exc_info:
        run_bootstrap(
            from_tests=False,
            from_github=False,
            from_jira=False,
            _merge=False,
            dry_run=False,
            prefix="REQ",
            config=config,
        )

    assert exc_info.value.code == 1
    captured = capsys.readouterr()
    assert "No source specified" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_bootstrap_from_tests(capsys: pytest.CaptureFixture) -> None:
    """Test run_bootstrap from tests (not yet implemented)."""
    config = RTMXConfig()

    run_bootstrap(
        from_tests=True,
        from_github=False,
        from_jira=False,
        _merge=False,
        dry_run=True,
        prefix="REQ",
        config=config,
    )

    captured = capsys.readouterr()
    assert "Scanning tests" in captured.out
    assert "Not yet implemented" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_bootstrap_from_github(capsys: pytest.CaptureFixture) -> None:
    """Test run_bootstrap from GitHub (not yet implemented)."""
    config = RTMXConfig()
    config.adapters.github.enabled = True
    config.adapters.github.repo = "org/repo"

    run_bootstrap(
        from_tests=False,
        from_github=True,
        from_jira=False,
        _merge=False,
        dry_run=True,
        prefix="REQ",
        config=config,
    )

    captured = capsys.readouterr()
    assert "Fetching GitHub issues" in captured.out
    assert "Not yet implemented" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_bootstrap_github_not_enabled(capsys: pytest.CaptureFixture) -> None:
    """Test run_bootstrap when GitHub adapter is not enabled."""
    config = RTMXConfig()
    config.adapters.github.enabled = False

    run_bootstrap(
        from_tests=False,
        from_github=True,
        from_jira=False,
        _merge=False,
        dry_run=True,
        prefix="REQ",
        config=config,
    )

    captured = capsys.readouterr()
    assert "GitHub adapter not enabled" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_bootstrap_github_no_repo(capsys: pytest.CaptureFixture) -> None:
    """Test run_bootstrap when GitHub repo is not configured."""
    config = RTMXConfig()
    config.adapters.github.enabled = True
    config.adapters.github.repo = ""

    run_bootstrap(
        from_tests=False,
        from_github=True,
        from_jira=False,
        _merge=False,
        dry_run=True,
        prefix="REQ",
        config=config,
    )

    captured = capsys.readouterr()
    assert "GitHub repo not configured" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_bootstrap_from_jira(capsys: pytest.CaptureFixture) -> None:
    """Test run_bootstrap from Jira (not yet implemented)."""
    config = RTMXConfig()
    config.adapters.jira.enabled = True
    config.adapters.jira.project = "PROJ"
    config.adapters.jira.server = "https://jira.example.com"

    run_bootstrap(
        from_tests=False,
        from_github=False,
        from_jira=True,
        _merge=False,
        dry_run=True,
        prefix="REQ",
        config=config,
    )

    captured = capsys.readouterr()
    assert "Fetching Jira tickets" in captured.out
    assert "Not yet implemented" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_bootstrap_jira_not_enabled(capsys: pytest.CaptureFixture) -> None:
    """Test run_bootstrap when Jira adapter is not enabled."""
    config = RTMXConfig()
    config.adapters.jira.enabled = False

    run_bootstrap(
        from_tests=False,
        from_github=False,
        from_jira=True,
        _merge=False,
        dry_run=True,
        prefix="REQ",
        config=config,
    )

    captured = capsys.readouterr()
    assert "Jira adapter not enabled" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_bootstrap_jira_no_project(capsys: pytest.CaptureFixture) -> None:
    """Test run_bootstrap when Jira project is not configured."""
    config = RTMXConfig()
    config.adapters.jira.enabled = True
    config.adapters.jira.project = ""

    run_bootstrap(
        from_tests=False,
        from_github=False,
        from_jira=True,
        _merge=False,
        dry_run=True,
        prefix="REQ",
        config=config,
    )

    captured = capsys.readouterr()
    assert "Jira project not configured" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_bootstrap_dry_run(capsys: pytest.CaptureFixture) -> None:
    """Test run_bootstrap with dry_run flag."""
    config = RTMXConfig()

    run_bootstrap(
        from_tests=True,
        from_github=False,
        from_jira=False,
        _merge=False,
        dry_run=True,
        prefix="REQ",
        config=config,
    )

    captured = capsys.readouterr()
    assert "DRY RUN" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_bootstrap_multiple_sources(capsys: pytest.CaptureFixture) -> None:
    """Test run_bootstrap with multiple sources specified."""
    config = RTMXConfig()
    config.adapters.github.enabled = True
    config.adapters.github.repo = "org/repo"

    run_bootstrap(
        from_tests=True,
        from_github=True,
        from_jira=False,
        _merge=False,
        dry_run=True,
        prefix="REQ",
        config=config,
    )

    captured = capsys.readouterr()
    assert "Scanning tests" in captured.out
    assert "Fetching GitHub issues" in captured.out


# =============================================================================
# Tests for run_makefile
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_makefile_stdout(capsys: pytest.CaptureFixture) -> None:
    """Test run_makefile outputs to stdout when no output file specified."""
    run_makefile(output=None)

    captured = capsys.readouterr()
    assert "# RTMX - Requirements Traceability Matrix targets" in captured.out
    assert ".PHONY: rtm" in captured.out
    assert "rtmx status" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_makefile_to_file(tmp_path: Path, capsys: pytest.CaptureFixture) -> None:
    """Test run_makefile writes to file when output path specified."""
    output_file = tmp_path / "rtmx.mk"

    run_makefile(output=output_file)

    assert output_file.exists()
    content = output_file.read_text()
    assert "# RTMX - Requirements Traceability Matrix targets" in content
    assert ".PHONY: rtm" in content

    captured = capsys.readouterr()
    assert f"Generated Makefile targets in {output_file}" in captured.out


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_makefile_content_has_rtm_targets() -> None:
    """Test that MAKEFILE_TARGETS constant contains expected targets."""
    assert "rtm:" in MAKEFILE_TARGETS
    assert "rtm-v:" in MAKEFILE_TARGETS
    assert "rtm-vv:" in MAKEFILE_TARGETS
    assert "rtm-vvv:" in MAKEFILE_TARGETS
    assert "backlog:" in MAKEFILE_TARGETS
    assert "cycles:" in MAKEFILE_TARGETS
    assert "reconcile:" in MAKEFILE_TARGETS


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_makefile_content_has_phony() -> None:
    """Test that MAKEFILE_TARGETS includes .PHONY declaration."""
    assert ".PHONY:" in MAKEFILE_TARGETS


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_makefile_content_has_rtmx_commands() -> None:
    """Test that MAKEFILE_TARGETS contains rtmx command invocations."""
    assert "rtmx status" in MAKEFILE_TARGETS
    assert "rtmx backlog" in MAKEFILE_TARGETS
    assert "rtmx cycles" in MAKEFILE_TARGETS
    assert "rtmx reconcile" in MAKEFILE_TARGETS
    assert "rtmx from-tests" in MAKEFILE_TARGETS


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_makefile_instructions(tmp_path: Path, capsys: pytest.CaptureFixture) -> None:
    """Test that run_makefile prints usage instructions."""
    output_file = tmp_path / "rtmx.mk"

    run_makefile(output=output_file)

    captured = capsys.readouterr()
    assert "To use, add to your Makefile:" in captured.out
    assert "include" in captured.out
    assert "Or append to existing Makefile:" in captured.out
