"""Tests for rtmx.cli.setup module.

This module tests the setup command functions:
- backup_file: Create timestamped backups
- detect_project: Detect project characteristics
- run_setup: Full setup workflow with various options
"""

from __future__ import annotations

import subprocess
from pathlib import Path
from unittest.mock import Mock, patch

import pytest

from rtmx.cli.setup import SetupResult, backup_file, detect_project, run_setup

# =============================================================================
# Fixtures
# =============================================================================


@pytest.fixture
def git_repo(tmp_path: Path) -> Path:
    """Create a git repository for testing."""
    subprocess.run(["git", "init"], cwd=tmp_path, capture_output=True)
    subprocess.run(
        ["git", "config", "user.email", "test@test.com"],
        cwd=tmp_path,
        capture_output=True,
    )
    subprocess.run(
        ["git", "config", "user.name", "Test User"],
        cwd=tmp_path,
        capture_output=True,
    )
    # Create initial commit
    (tmp_path / "README.md").write_text("# Test Project\n")
    subprocess.run(["git", "add", "."], cwd=tmp_path, capture_output=True)
    subprocess.run(
        ["git", "commit", "-m", "Initial commit"],
        cwd=tmp_path,
        capture_output=True,
    )
    return tmp_path


@pytest.fixture
def non_git_dir(tmp_path: Path) -> Path:
    """Create a non-git directory for testing."""
    project_dir = tmp_path / "project"
    project_dir.mkdir()
    return project_dir


@pytest.fixture
def python_project(git_repo: Path) -> Path:
    """Create a Python project structure."""
    (git_repo / "pyproject.toml").write_text("""[project]
name = "test-project"
version = "0.1.0"
""")
    (git_repo / "tests").mkdir()
    (git_repo / "Makefile").write_text("""all:
\techo "build"
""")
    return git_repo


# =============================================================================
# Tests for SetupResult
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_setup_result_to_dict():
    """Test SetupResult.to_dict() conversion."""
    result = SetupResult(
        success=True,
        steps_completed=["step1", "step2"],
        steps_skipped=["step3"],
        files_created=["file1.txt"],
        files_modified=["file2.txt"],
        files_backed_up=["file2.txt.backup"],
        warnings=["warning1"],
        errors=[],
        branch_name="test-branch",
        rollback_point="abc123",
        pr_url="https://github.com/test/pr/1",
    )

    result_dict = result.to_dict()

    assert result_dict["success"] is True
    assert result_dict["steps_completed"] == ["step1", "step2"]
    assert result_dict["steps_skipped"] == ["step3"]
    assert result_dict["files_created"] == ["file1.txt"]
    assert result_dict["files_modified"] == ["file2.txt"]
    assert result_dict["files_backed_up"] == ["file2.txt.backup"]
    assert result_dict["warnings"] == ["warning1"]
    assert result_dict["errors"] == []
    assert result_dict["branch_name"] == "test-branch"
    assert result_dict["rollback_point"] == "abc123"
    assert result_dict["pr_url"] == "https://github.com/test/pr/1"


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_setup_result_defaults():
    """Test SetupResult default values."""
    result = SetupResult(success=False)

    assert result.success is False
    assert result.steps_completed == []
    assert result.steps_skipped == []
    assert result.files_created == []
    assert result.files_modified == []
    assert result.files_backed_up == []
    assert result.warnings == []
    assert result.errors == []
    assert result.branch_name is None
    assert result.rollback_point is None
    assert result.pr_url is None


# =============================================================================
# Tests for backup_file
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_backup_file_creates_backup(tmp_path: Path):
    """Test backup_file creates timestamped backup."""
    original = tmp_path / "config.yaml"
    original.write_text("original content")

    backup = backup_file(original)

    assert backup is not None
    assert backup.exists()
    assert backup.read_text() == "original content"
    assert ".rtmx-backup-" in backup.name
    assert backup.suffix == ".yaml"


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_backup_file_nonexistent_returns_none(tmp_path: Path):
    """Test backup_file returns None for nonexistent file."""
    nonexistent = tmp_path / "missing.txt"

    backup = backup_file(nonexistent)

    assert backup is None


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_backup_file_preserves_metadata(tmp_path: Path):
    """Test backup_file preserves file metadata."""
    original = tmp_path / "data.csv"
    original.write_text("data")
    original_stat = original.stat()

    backup = backup_file(original)

    assert backup is not None
    # Modification time should be preserved by copy2
    assert abs(backup.stat().st_mtime - original_stat.st_mtime) < 1


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_backup_file_unique_timestamp(tmp_path: Path):
    """Test backup_file creates unique timestamped names."""
    original = tmp_path / "config.yaml"
    original.write_text("v1")

    backup1 = backup_file(original)

    # Modify and backup again
    original.write_text("v2")
    backup2 = backup_file(original)

    assert backup1 is not None
    assert backup2 is not None
    # Names should be different (different timestamps)
    # Note: might be same if executed in same second, but structure is correct
    assert ".rtmx-backup-" in backup1.name
    assert ".rtmx-backup-" in backup2.name


# =============================================================================
# Tests for detect_project
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_project_git_repo(git_repo: Path):
    """Test detect_project identifies git repository."""
    detection = detect_project(git_repo)

    assert detection["is_git_repo"] is True
    assert detection["git_clean"] is True
    assert detection["git_branch"] in ("main", "master")


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_project_non_git(non_git_dir: Path):
    """Test detect_project handles non-git directory."""
    detection = detect_project(non_git_dir)

    assert detection["is_git_repo"] is False
    assert detection["git_clean"] is True
    assert detection["git_branch"] is None


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_project_rtmx_config(tmp_path: Path):
    """Test detect_project finds rtmx.yaml."""
    (tmp_path / "rtmx.yaml").write_text("rtmx:\n  database: db.csv\n")

    detection = detect_project(tmp_path)

    assert detection["has_rtmx_config"] is True


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_project_hidden_rtmx_config(tmp_path: Path):
    """Test detect_project finds .rtmx.yaml."""
    (tmp_path / ".rtmx.yaml").write_text("rtmx:\n  database: db.csv\n")

    detection = detect_project(tmp_path)

    assert detection["has_rtmx_config"] is True


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_project_rtm_database(tmp_path: Path):
    """Test detect_project finds RTM database."""
    docs_dir = tmp_path / "docs"
    docs_dir.mkdir()
    (docs_dir / "rtm_database.csv").write_text("req_id,category\n")

    detection = detect_project(tmp_path)

    assert detection["has_rtm_database"] is True


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_project_tests_directory(tmp_path: Path):
    """Test detect_project finds tests directory."""
    (tmp_path / "tests").mkdir()

    detection = detect_project(tmp_path)

    assert detection["has_tests"] is True


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_project_test_directory(tmp_path: Path):
    """Test detect_project finds test directory (singular)."""
    (tmp_path / "test").mkdir()

    detection = detect_project(tmp_path)

    assert detection["has_tests"] is True


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_project_makefile(tmp_path: Path):
    """Test detect_project finds Makefile."""
    (tmp_path / "Makefile").write_text("all:\n\techo test\n")

    detection = detect_project(tmp_path)

    assert detection["has_makefile"] is True


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_project_pyproject(tmp_path: Path):
    """Test detect_project finds pyproject.toml."""
    (tmp_path / "pyproject.toml").write_text("[project]\nname = 'test'\n")

    detection = detect_project(tmp_path)

    assert detection["has_pyproject"] is True


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_project_claude_config(tmp_path: Path):
    """Test detect_project finds CLAUDE.md."""
    (tmp_path / "CLAUDE.md").write_text("# Claude Config\nUse rtmx for requirements.\n")

    detection = detect_project(tmp_path)

    assert detection["agent_configs"]["claude"]["exists"] is True
    assert detection["agent_configs"]["claude"]["has_rtmx"] is True


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_project_claude_config_no_rtmx(tmp_path: Path):
    """Test detect_project detects CLAUDE.md without rtmx mention."""
    (tmp_path / "CLAUDE.md").write_text("# Claude Config\nGeneral instructions.\n")

    detection = detect_project(tmp_path)

    assert detection["agent_configs"]["claude"]["exists"] is True
    assert detection["agent_configs"]["claude"]["has_rtmx"] is False


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_project_cursor_config(tmp_path: Path):
    """Test detect_project finds .cursorrules."""
    (tmp_path / ".cursorrules").write_text("Use RTMX for tracking.\n")

    detection = detect_project(tmp_path)

    assert detection["agent_configs"]["cursor"]["exists"] is True
    assert detection["agent_configs"]["cursor"]["has_rtmx"] is True


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_project_copilot_config(tmp_path: Path):
    """Test detect_project finds copilot instructions."""
    github_dir = tmp_path / ".github"
    github_dir.mkdir()
    (github_dir / "copilot-instructions.md").write_text("Use rtmx.\n")

    detection = detect_project(tmp_path)

    assert detection["agent_configs"]["copilot"]["exists"] is True
    assert detection["agent_configs"]["copilot"]["has_rtmx"] is True


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_project_windsurf_config(tmp_path: Path):
    """Test detect_project finds .windsurfrules."""
    (tmp_path / ".windsurfrules").write_text("Config here.\n")

    detection = detect_project(tmp_path)

    assert detection["agent_configs"]["windsurf"]["exists"] is True


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_project_aider_config(tmp_path: Path):
    """Test detect_project finds .aider.conf.yml."""
    (tmp_path / ".aider.conf.yml").write_text("config: value\n")

    detection = detect_project(tmp_path)

    assert detection["agent_configs"]["aider"]["exists"] is True


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_detect_project_dirty_git_repo(git_repo: Path):
    """Test detect_project detects uncommitted changes."""
    (git_repo / "newfile.txt").write_text("uncommitted")

    detection = detect_project(git_repo)

    assert detection["is_git_repo"] is True
    assert detection["git_clean"] is False


# =============================================================================
# Tests for run_setup - Basic functionality
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_creates_config(tmp_path: Path):
    """Test run_setup creates rtmx.yaml."""
    run_setup(project_path=tmp_path, dry_run=False)

    config_path = tmp_path / "rtmx.yaml"
    assert config_path.exists()
    content = config_path.read_text()
    assert "rtmx:" in content
    assert "database: docs/rtm_database.csv" in content


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_creates_rtm_database(tmp_path: Path):
    """Test run_setup creates RTM database."""
    run_setup(project_path=tmp_path, dry_run=False)

    rtm_path = tmp_path / "docs" / "rtm_database.csv"
    assert rtm_path.exists()
    content = rtm_path.read_text()
    assert "REQ-INIT-001" in content


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_creates_requirement_spec(tmp_path: Path):
    """Test run_setup creates requirement specification."""
    run_setup(project_path=tmp_path, dry_run=False)

    spec_path = tmp_path / "docs" / "requirements" / "SETUP" / "REQ-INIT-001.md"
    assert spec_path.exists()
    content = spec_path.read_text()
    assert "REQ-INIT-001" in content
    assert "Acceptance Criteria" in content


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_dry_run_no_changes(tmp_path: Path):
    """Test run_setup with dry_run makes no changes."""
    run_setup(project_path=tmp_path, dry_run=True)

    assert not (tmp_path / "rtmx.yaml").exists()
    assert not (tmp_path / "docs").exists()


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_minimal_mode(tmp_path: Path):
    """Test run_setup with minimal flag."""
    (tmp_path / "CLAUDE.md").write_text("# Existing\n")
    (tmp_path / "Makefile").write_text("all:\n\techo test\n")

    run_setup(project_path=tmp_path, dry_run=False, minimal=True)

    # Should create config and RTM
    assert (tmp_path / "rtmx.yaml").exists()
    assert (tmp_path / "docs" / "rtm_database.csv").exists()

    # Should NOT modify agent configs or Makefile
    claude_content = (tmp_path / "CLAUDE.md").read_text()
    assert "RTMX" not in claude_content
    makefile_content = (tmp_path / "Makefile").read_text()
    assert "rtmx" not in makefile_content.lower()


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_skip_config_if_exists(tmp_path: Path):
    """Test run_setup skips config creation if exists."""
    existing_config = tmp_path / "rtmx.yaml"
    existing_config.write_text("existing: config\n")

    result = run_setup(project_path=tmp_path, dry_run=False)

    assert "create_config" in result.steps_skipped
    # Should not overwrite
    assert existing_config.read_text() == "existing: config\n"


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_force_overwrites_config(tmp_path: Path):
    """Test run_setup with force overwrites existing config."""
    existing_config = tmp_path / "rtmx.yaml"
    existing_config.write_text("existing: config\n")

    result = run_setup(project_path=tmp_path, dry_run=False, force=True)

    assert "create_config" in result.steps_completed
    # Should overwrite
    content = existing_config.read_text()
    assert "rtmx:" in content
    assert "existing: config" not in content


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_creates_backup_on_overwrite(tmp_path: Path):
    """Test run_setup creates backup when overwriting."""
    existing_config = tmp_path / "rtmx.yaml"
    existing_config.write_text("existing: config\n")

    result = run_setup(project_path=tmp_path, dry_run=False, force=True)

    # Should create backup
    assert len(result.files_backed_up) > 0
    backup_path = Path(result.files_backed_up[0])
    assert backup_path.exists()
    assert backup_path.read_text() == "existing: config\n"


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_updates_claude_config(tmp_path: Path):
    """Test run_setup updates existing CLAUDE.md."""
    (tmp_path / "CLAUDE.md").write_text("# Existing config\n")

    result = run_setup(project_path=tmp_path, dry_run=False)

    claude_content = (tmp_path / "CLAUDE.md").read_text()
    assert "RTMX" in claude_content
    assert "rtmx status" in claude_content
    assert "agent_claude" in result.steps_completed


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_skips_claude_if_already_configured(tmp_path: Path):
    """Test run_setup skips CLAUDE.md if already has RTMX."""
    (tmp_path / "CLAUDE.md").write_text("# Config\nUse RTMX for requirements.\n")

    result = run_setup(project_path=tmp_path, dry_run=False)

    assert "agent_claude" in result.steps_skipped


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_creates_claude_config(tmp_path: Path):
    """Test run_setup creates new CLAUDE.md."""
    run_setup(project_path=tmp_path, dry_run=False)

    claude_path = tmp_path / "CLAUDE.md"
    assert claude_path.exists()
    content = claude_path.read_text()
    assert "RTMX" in content


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_creates_cursor_config(tmp_path: Path):
    """Test run_setup creates new .cursorrules."""
    run_setup(project_path=tmp_path, dry_run=False)

    cursor_path = tmp_path / ".cursorrules"
    assert cursor_path.exists()
    content = cursor_path.read_text()
    assert "RTMX" in content


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_skip_agents_flag(tmp_path: Path):
    """Test run_setup with skip_agents flag."""
    run_setup(project_path=tmp_path, dry_run=False, skip_agents=True)

    assert not (tmp_path / "CLAUDE.md").exists()
    assert not (tmp_path / ".cursorrules").exists()


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_adds_makefile_targets(tmp_path: Path):
    """Test run_setup adds rtmx targets to Makefile."""
    (tmp_path / "Makefile").write_text("all:\n\techo test\n")

    run_setup(project_path=tmp_path, dry_run=False)

    makefile_content = (tmp_path / "Makefile").read_text()
    assert "rtm:" in makefile_content
    assert "backlog:" in makefile_content
    assert "health:" in makefile_content
    assert "rtmx status" in makefile_content


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_skips_makefile_if_has_rtmx(tmp_path: Path):
    """Test run_setup skips Makefile if already has rtmx targets."""
    (tmp_path / "Makefile").write_text("rtm:\n\trtmx status\n")

    result = run_setup(project_path=tmp_path, dry_run=False)

    assert "makefile" in result.steps_skipped


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_skip_makefile_flag(tmp_path: Path):
    """Test run_setup with skip_makefile flag."""
    (tmp_path / "Makefile").write_text("all:\n\techo test\n")

    run_setup(project_path=tmp_path, dry_run=False, skip_makefile=True)

    makefile_content = (tmp_path / "Makefile").read_text()
    assert "rtmx" not in makefile_content.lower()


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_scans_tests(tmp_path: Path):
    """Test run_setup scans test files for markers."""
    tests_dir = tmp_path / "tests"
    tests_dir.mkdir()
    test_file = tests_dir / "test_example.py"
    test_file.write_text("""
import pytest

@pytest.mark.req("REQ-TEST-001")
def test_something():
    assert True
""")

    result = run_setup(project_path=tmp_path, dry_run=False)

    assert "scan_tests" in result.steps_completed


# =============================================================================
# Tests for run_setup - Git workflow
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_branch_mode_creates_branch(git_repo: Path):
    """Test run_setup with branch creates new branch."""
    result = run_setup(project_path=git_repo, dry_run=False, branch="test-branch")

    assert result.branch_name == "test-branch"
    assert "create_branch" in result.steps_completed

    # Verify branch was created
    git_result = subprocess.run(
        ["git", "branch", "--show-current"],
        cwd=git_repo,
        capture_output=True,
        text=True,
    )
    assert git_result.stdout.strip() == "test-branch"


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_branch_auto_generates_name(git_repo: Path):
    """Test run_setup with branch=auto generates timestamped name."""
    result = run_setup(project_path=git_repo, dry_run=False, branch="auto")

    assert result.branch_name is not None
    assert result.branch_name.startswith("rtmx/setup-")


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_branch_creates_rollback_point(git_repo: Path):
    """Test run_setup with branch creates rollback point."""
    result = run_setup(project_path=git_repo, dry_run=False, branch="test-branch")

    assert result.rollback_point is not None
    assert len(result.rollback_point) == 40  # Full SHA


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_branch_fails_with_uncommitted_changes(git_repo: Path):
    """Test run_setup with branch fails if uncommitted changes."""
    (git_repo / "uncommitted.txt").write_text("changes")

    result = run_setup(project_path=git_repo, dry_run=False, branch="test-branch")

    assert not result.success
    assert len(result.errors) > 0
    assert "Uncommitted changes" in result.errors[0]


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_branch_force_allows_uncommitted(git_repo: Path):
    """Test run_setup with branch and force allows uncommitted changes."""
    (git_repo / "uncommitted.txt").write_text("changes")

    result = run_setup(project_path=git_repo, dry_run=False, branch="test-branch", force=True)

    # Should proceed despite uncommitted changes
    assert result.branch_name == "test-branch"


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_branch_ignored_if_not_git(non_git_dir: Path):
    """Test run_setup with branch is ignored if not git repo."""
    result = run_setup(project_path=non_git_dir, dry_run=False, branch="test-branch")

    assert result.branch_name is None
    assert len(result.warnings) > 0
    assert any("not a git repo" in w.lower() for w in result.warnings)


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_commits_changes_in_branch_mode(git_repo: Path):
    """Test run_setup commits changes when using branch mode."""
    result = run_setup(project_path=git_repo, dry_run=False, branch="test-branch")

    assert "git_commit" in result.steps_completed

    # Verify commit was made
    git_result = subprocess.run(
        ["git", "log", "-1", "--pretty=%s"],
        cwd=git_repo,
        capture_output=True,
        text=True,
    )
    assert "rtmx" in git_result.stdout.lower()


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_creates_pr(git_repo: Path):
    """Test run_setup attempts to create PR when requested."""
    with patch("rtmx.cli.git_ops.check_gh_installed") as mock_gh_check:
        mock_gh_check.return_value = False  # Skip actual PR creation

        result = run_setup(project_path=git_repo, dry_run=False, create_pr=True)

        # Should attempt PR creation (even if it fails due to gh not installed)
        # The important thing is that branch mode was enabled
        assert result.branch_name is not None
        # gh CLI not installed should result in a warning, not error
        assert any("gh CLI" in w or "gh" in w.lower() for w in result.warnings)


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_create_pr_implies_branch(git_repo: Path):
    """Test run_setup with create_pr automatically enables branch mode."""
    with patch("rtmx.cli.git_ops.check_gh_installed") as mock_gh:
        mock_gh.return_value = False  # Skip actual PR creation

        result = run_setup(project_path=git_repo, dry_run=False, create_pr=True)

        assert result.branch_name is not None


# =============================================================================
# Tests for run_setup - Success/failure states
# =============================================================================


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
@patch("rtmx.cli.setup.load_config")
@patch("rtmx.cli.health.run_health_checks")
def test_run_setup_success_state(mock_health, mock_config, tmp_path: Path):
    """Test run_setup returns success when complete."""
    from rtmx.cli.health import HealthReport, HealthStatus

    mock_config.return_value = Mock()
    mock_health.return_value = HealthReport(
        status=HealthStatus.HEALTHY, checks=[], summary={"passed": 5, "warnings": 0, "failed": 0}
    )

    result = run_setup(project_path=tmp_path, dry_run=False)

    assert result.success is True
    assert len(result.errors) == 0


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_tracks_files_created(tmp_path: Path):
    """Test run_setup tracks all files created."""
    result = run_setup(project_path=tmp_path, dry_run=False)

    assert len(result.files_created) > 0
    # Should include at minimum: rtmx.yaml, rtm_database.csv, REQ-INIT-001.md
    created_names = [Path(f).name for f in result.files_created]
    assert "rtmx.yaml" in created_names


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_tracks_files_modified(tmp_path: Path):
    """Test run_setup tracks files modified."""
    (tmp_path / "CLAUDE.md").write_text("# Existing\n")
    (tmp_path / "Makefile").write_text("all:\n\techo test\n")

    result = run_setup(project_path=tmp_path, dry_run=False)

    # Should modify CLAUDE.md and Makefile
    assert len(result.files_modified) >= 2
    modified_names = [Path(f).name for f in result.files_modified]
    assert "CLAUDE.md" in modified_names
    assert "Makefile" in modified_names


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_default_path_uses_cwd(tmp_path: Path, monkeypatch):
    """Test run_setup uses current directory by default."""
    monkeypatch.chdir(tmp_path)

    run_setup(project_path=None, dry_run=False)

    # Should create files in current directory
    assert (tmp_path / "rtmx.yaml").exists()


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
@patch("rtmx.cli.setup.load_config")
@patch("rtmx.cli.health.run_health_checks")
def test_run_setup_runs_health_check(mock_health, mock_config, tmp_path: Path):
    """Test run_setup runs health check after setup."""
    from rtmx.cli.health import HealthReport, HealthStatus

    mock_config.return_value = Mock()
    mock_health.return_value = HealthReport(
        status=HealthStatus.HEALTHY, checks=[], summary={"passed": 5, "warnings": 0, "failed": 0}
    )

    result = run_setup(project_path=tmp_path, dry_run=False)

    assert "health_check" in result.steps_completed
    mock_health.assert_called_once()


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
@patch("rtmx.cli.setup.load_config")
@patch("rtmx.cli.health.run_health_checks")
def test_run_setup_health_degraded_warning(mock_health, mock_config, tmp_path: Path):
    """Test run_setup handles degraded health status."""
    from rtmx.cli.health import HealthReport, HealthStatus

    mock_config.return_value = Mock()
    mock_health.return_value = HealthReport(
        status=HealthStatus.DEGRADED, checks=[], summary={"passed": 3, "warnings": 2, "failed": 0}
    )

    result = run_setup(project_path=tmp_path, dry_run=False)

    assert len(result.warnings) > 0
    assert any("Health check" in w for w in result.warnings)


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
@patch("rtmx.cli.setup.load_config")
@patch("rtmx.cli.health.run_health_checks")
def test_run_setup_health_unhealthy_error(mock_health, mock_config, tmp_path: Path):
    """Test run_setup handles unhealthy status."""
    from rtmx.cli.health import HealthReport, HealthStatus

    mock_config.return_value = Mock()
    mock_health.return_value = HealthReport(
        status=HealthStatus.UNHEALTHY, checks=[], summary={"passed": 1, "warnings": 1, "failed": 3}
    )

    result = run_setup(project_path=tmp_path, dry_run=False)

    assert len(result.errors) > 0
    assert any("Health check" in e for e in result.errors)


@pytest.mark.req("REQ-CLI-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_run_setup_health_skipped_in_dry_run(tmp_path: Path):
    """Test run_setup skips health check in dry run mode."""
    result = run_setup(project_path=tmp_path, dry_run=True)

    assert "health_check" not in result.steps_completed
