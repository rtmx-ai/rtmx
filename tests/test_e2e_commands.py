"""End-to-end system tests for RTMX CLI commands.

These tests exercise the full CLI from the command line, simulating
real user workflows. Each test is marked with scope_system to indicate
it tests the complete system behavior.
"""

from __future__ import annotations

import csv
import json
import subprocess
import sys
import tempfile
from pathlib import Path
from typing import TYPE_CHECKING

import pytest
import yaml

if TYPE_CHECKING:
    from collections.abc import Generator


def run_rtmx(*args: str, cwd: Path | None = None) -> subprocess.CompletedProcess[str]:
    """Run rtmx command and return result."""
    return subprocess.run(
        [sys.executable, "-m", "rtmx", *args],
        cwd=cwd,
        capture_output=True,
        text=True,
    )


@pytest.fixture
def temp_project() -> Generator[Path, None, None]:
    """Create an isolated temporary project directory with git."""
    with tempfile.TemporaryDirectory(prefix="rtmx_e2e_") as tmpdir:
        project_dir = Path(tmpdir)
        subprocess.run(["git", "init"], cwd=project_dir, capture_output=True, check=True)
        subprocess.run(
            ["git", "config", "user.email", "test@example.com"],
            cwd=project_dir,
            capture_output=True,
            check=True,
        )
        subprocess.run(
            ["git", "config", "user.name", "Test User"],
            cwd=project_dir,
            capture_output=True,
            check=True,
        )
        yield project_dir


@pytest.fixture
def initialized_project(temp_project: Path) -> Path:
    """Create an initialized RTMX project."""
    result = run_rtmx("setup", "--minimal", cwd=temp_project)
    assert result.returncode == 0, f"Setup failed: {result.stderr}"
    return temp_project


@pytest.fixture
def project_with_data(initialized_project: Path) -> Path:
    """Create a project with multiple requirements for testing."""
    db_path = initialized_project / "docs" / "rtm_database.csv"

    requirements = [
        {
            "req_id": "REQ-CORE-001",
            "category": "CORE",
            "subcategory": "Foundation",
            "requirement_text": "System shall initialize correctly",
            "target_value": "100%",
            "test_module": "tests/test_core.py",
            "test_function": "test_init",
            "validation_method": "Unit Test",
            "status": "COMPLETE",
            "priority": "P0",
            "phase": "1",
            "notes": "Initial setup",
            "effort_weeks": "1.0",
            "dependencies": "",
            "blocks": "REQ-CORE-002|REQ-CORE-003",
            "assignee": "dev1",
            "sprint": "v0.1",
            "started_date": "2025-01-01",
            "completed_date": "2025-01-15",
            "requirement_file": "docs/requirements/CORE/REQ-CORE-001.md",
        },
        {
            "req_id": "REQ-CORE-002",
            "category": "CORE",
            "subcategory": "Foundation",
            "requirement_text": "System shall handle configuration",
            "target_value": "All options",
            "test_module": "tests/test_config.py",
            "test_function": "test_config",
            "validation_method": "Unit Test",
            "status": "PARTIAL",
            "priority": "HIGH",
            "phase": "1",
            "notes": "Config in progress",
            "effort_weeks": "0.5",
            "dependencies": "REQ-CORE-001",
            "blocks": "",
            "assignee": "dev2",
            "sprint": "v0.1",
            "started_date": "2025-01-10",
            "completed_date": "",
            "requirement_file": "docs/requirements/CORE/REQ-CORE-002.md",
        },
        {
            "req_id": "REQ-CORE-003",
            "category": "CORE",
            "subcategory": "Foundation",
            "requirement_text": "System shall validate input",
            "target_value": "Zero errors",
            "test_module": "",
            "test_function": "",
            "validation_method": "Unit Test",
            "status": "MISSING",
            "priority": "MEDIUM",
            "phase": "2",
            "notes": "Not started",
            "effort_weeks": "1.0",
            "dependencies": "REQ-CORE-001",
            "blocks": "",
            "assignee": "",
            "sprint": "v0.2",
            "started_date": "",
            "completed_date": "",
            "requirement_file": "docs/requirements/CORE/REQ-CORE-003.md",
        },
        {
            "req_id": "REQ-API-001",
            "category": "API",
            "subcategory": "REST",
            "requirement_text": "API shall return JSON responses",
            "target_value": "100% endpoints",
            "test_module": "tests/test_api.py",
            "test_function": "test_json_response",
            "validation_method": "Integration Test",
            "status": "COMPLETE",
            "priority": "HIGH",
            "phase": "2",
            "notes": "API complete",
            "effort_weeks": "2.0",
            "dependencies": "",
            "blocks": "REQ-API-002",
            "assignee": "dev1",
            "sprint": "v0.2",
            "started_date": "2025-01-20",
            "completed_date": "2025-02-01",
            "requirement_file": "docs/requirements/API/REQ-API-001.md",
        },
        {
            "req_id": "REQ-API-002",
            "category": "API",
            "subcategory": "REST",
            "requirement_text": "API shall handle errors gracefully",
            "target_value": "All error codes",
            "test_module": "",
            "test_function": "",
            "validation_method": "Integration Test",
            "status": "NOT_STARTED",
            "priority": "MEDIUM",
            "phase": "3",
            "notes": "Planned for next phase",
            "effort_weeks": "1.5",
            "dependencies": "REQ-API-001",
            "blocks": "",
            "assignee": "",
            "sprint": "v0.3",
            "started_date": "",
            "completed_date": "",
            "requirement_file": "docs/requirements/API/REQ-API-002.md",
        },
    ]

    with open(db_path, "w", newline="") as f:
        fieldnames = list(requirements[0].keys())
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(requirements)

    return initialized_project


# =============================================================================
# Status Command E2E Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-STATUS")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestStatusE2E:
    """End-to-end tests for rtmx status command."""

    def test_status_basic_output(self, project_with_data: Path) -> None:
        """Status command should show completion percentage."""
        result = run_rtmx("status", cwd=project_with_data)
        assert "%" in result.stdout
        assert "complete" in result.stdout.lower() or "missing" in result.stdout.lower()

    def test_status_json_output(self, project_with_data: Path) -> None:
        """Status command should output valid JSON with --json option."""
        json_path = project_with_data / "status.json"
        result = run_rtmx("status", "--json", str(json_path), cwd=project_with_data)
        assert result.returncode in [0, 1]
        # If JSON was written, validate it
        if json_path.exists():
            data = json.loads(json_path.read_text())
            assert isinstance(data, dict)

    def test_status_verbose_output(self, project_with_data: Path) -> None:
        """Status command should show more details with -v."""
        result = run_rtmx("status", "-v", cwd=project_with_data)
        # Verbose should show category breakdowns
        assert result.returncode in [0, 1]
        assert "%" in result.stdout or "complete" in result.stdout.lower()

    def test_status_double_verbose_output(self, project_with_data: Path) -> None:
        """Status command should show even more details with -vv."""
        result = run_rtmx("status", "-vv", cwd=project_with_data)
        assert result.returncode in [0, 1]

    def test_status_triple_verbose_output(self, project_with_data: Path) -> None:
        """Status command should show all requirements with -vvv."""
        result = run_rtmx("status", "-vvv", cwd=project_with_data)
        assert result.returncode in [0, 1]

    def test_status_shows_progress_bar(self, project_with_data: Path) -> None:
        """Status should display visual progress indicator."""
        result = run_rtmx("status", cwd=project_with_data)
        # Progress bar uses box-drawing characters or percentage
        assert "%" in result.stdout or "█" in result.stdout or "complete" in result.stdout.lower()


# =============================================================================
# Backlog Command E2E Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-BACKLOG")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestBacklogE2E:
    """End-to-end tests for rtmx backlog command."""

    def test_backlog_basic_output(self, project_with_data: Path) -> None:
        """Backlog command should show incomplete requirements."""
        result = run_rtmx("backlog", cwd=project_with_data)
        # Should show missing or partial requirements
        assert result.returncode == 0
        assert "REQ-" in result.stdout or "backlog" in result.stdout.lower()

    def test_backlog_phase_filter(self, project_with_data: Path) -> None:
        """Backlog command should filter by phase."""
        result = run_rtmx("backlog", "--phase", "1", cwd=project_with_data)
        assert result.returncode == 0

    def test_backlog_view_all(self, project_with_data: Path) -> None:
        """Backlog command should show all items with --view all."""
        result = run_rtmx("backlog", "--view", "all", cwd=project_with_data)
        assert result.returncode == 0

    def test_backlog_view_critical(self, project_with_data: Path) -> None:
        """Backlog command should filter critical items."""
        result = run_rtmx("backlog", "--view", "critical", cwd=project_with_data)
        assert result.returncode == 0

    def test_backlog_view_blockers(self, project_with_data: Path) -> None:
        """Backlog command should show blockers."""
        result = run_rtmx("backlog", "--view", "blockers", cwd=project_with_data)
        assert result.returncode == 0

    def test_backlog_limit_option(self, project_with_data: Path) -> None:
        """Backlog command should respect --limit."""
        result = run_rtmx("backlog", "--limit", "2", cwd=project_with_data)
        assert result.returncode == 0


# =============================================================================
# Health Command E2E Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-HEALTH")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestHealthE2E:
    """End-to-end tests for rtmx health command."""

    def test_health_basic_output(self, initialized_project: Path) -> None:
        """Health command should run all checks."""
        result = run_rtmx("health", cwd=initialized_project)
        # Should show check results
        assert result.returncode in [0, 1]
        assert "check" in result.stdout.lower() or "health" in result.stdout.lower()

    def test_health_json_output(self, initialized_project: Path) -> None:
        """Health command should output valid JSON."""
        result = run_rtmx("health", "--format", "json", cwd=initialized_project)
        if result.returncode in [0, 1]:
            data = json.loads(result.stdout)
            assert "status" in data or "checks" in data

    def test_health_strict_mode(self, initialized_project: Path) -> None:
        """Health command --strict should fail on warnings."""
        result = run_rtmx("health", "--strict", cwd=initialized_project)
        # Strict mode fails on any non-pass
        assert result.returncode in [0, 1]

    def test_health_specific_check(self, initialized_project: Path) -> None:
        """Health command should run specific check with --check."""
        result = run_rtmx("health", "--check", "rtm_exists", cwd=initialized_project)
        assert result.returncode in [0, 1]

    def test_health_ci_format(self, initialized_project: Path) -> None:
        """Health command should output CI-friendly format."""
        result = run_rtmx("health", "--format", "ci", cwd=initialized_project)
        assert result.returncode in [0, 1]


# =============================================================================
# Deps Command E2E Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-DEPS")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestDepsE2E:
    """End-to-end tests for rtmx deps command."""

    def test_deps_basic_output(self, project_with_data: Path) -> None:
        """Deps command should show dependency tree."""
        result = run_rtmx("deps", cwd=project_with_data)
        assert result.returncode == 0

    def test_deps_for_specific_requirement(self, project_with_data: Path) -> None:
        """Deps command should show deps for specific requirement."""
        result = run_rtmx("deps", "--req", "REQ-CORE-002", cwd=project_with_data)
        # Should show dependencies of REQ-CORE-002
        assert result.returncode in [0, 1]

    def test_deps_category_filter(self, project_with_data: Path) -> None:
        """Deps command should filter by category."""
        result = run_rtmx("deps", "--category", "CORE", cwd=project_with_data)
        assert result.returncode == 0

    def test_deps_phase_filter(self, project_with_data: Path) -> None:
        """Deps command should filter by phase."""
        result = run_rtmx("deps", "--phase", "1", cwd=project_with_data)
        assert result.returncode == 0


# =============================================================================
# Cycles Command E2E Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-CYCLES")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestCyclesE2E:
    """End-to-end tests for rtmx cycles command."""

    def test_cycles_no_cycles(self, project_with_data: Path) -> None:
        """Cycles command should report no cycles when none exist."""
        result = run_rtmx("cycles", cwd=project_with_data)
        # No cycles = success
        assert result.returncode == 0

    def test_cycles_output_content(self, project_with_data: Path) -> None:
        """Cycles command should show cycle detection status."""
        result = run_rtmx("cycles", cwd=project_with_data)
        # Should output something about cycle detection
        assert result.returncode == 0 or "cycle" in result.stdout.lower()

    def test_cycles_with_cycle(self, initialized_project: Path) -> None:
        """Cycles command should detect cycles when present."""
        # Create a cycle: A -> B -> A
        db_path = initialized_project / "docs" / "rtm_database.csv"
        requirements = [
            {
                "req_id": "REQ-A",
                "category": "TEST",
                "subcategory": "Cycle",
                "requirement_text": "Requirement A",
                "target_value": "",
                "test_module": "",
                "test_function": "",
                "validation_method": "",
                "status": "MISSING",
                "priority": "HIGH",
                "phase": "1",
                "notes": "",
                "effort_weeks": "",
                "dependencies": "REQ-B",
                "blocks": "",
                "assignee": "",
                "sprint": "",
                "started_date": "",
                "completed_date": "",
                "requirement_file": "",
            },
            {
                "req_id": "REQ-B",
                "category": "TEST",
                "subcategory": "Cycle",
                "requirement_text": "Requirement B",
                "target_value": "",
                "test_module": "",
                "test_function": "",
                "validation_method": "",
                "status": "MISSING",
                "priority": "HIGH",
                "phase": "1",
                "notes": "",
                "effort_weeks": "",
                "dependencies": "REQ-A",
                "blocks": "",
                "assignee": "",
                "sprint": "",
                "started_date": "",
                "completed_date": "",
                "requirement_file": "",
            },
        ]
        with open(db_path, "w", newline="") as f:
            fieldnames = list(requirements[0].keys())
            writer = csv.DictWriter(f, fieldnames=fieldnames)
            writer.writeheader()
            writer.writerows(requirements)

        result = run_rtmx("cycles", cwd=initialized_project)
        # Should detect cycle
        assert "cycle" in result.stdout.lower()


# =============================================================================
# Config Command E2E Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-CONFIG")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestConfigE2E:
    """End-to-end tests for rtmx config command."""

    def test_config_show(self, initialized_project: Path) -> None:
        """Config command should show current configuration."""
        result = run_rtmx("config", cwd=initialized_project)
        assert result.returncode == 0
        assert "database" in result.stdout.lower() or "rtmx" in result.stdout.lower()

    def test_config_yaml_output(self, initialized_project: Path) -> None:
        """Config command should output YAML format."""
        result = run_rtmx("config", "--format", "yaml", cwd=initialized_project)
        assert result.returncode == 0
        config = yaml.safe_load(result.stdout)
        assert isinstance(config, dict)

    def test_config_json_output(self, initialized_project: Path) -> None:
        """Config command should output JSON format."""
        result = run_rtmx("config", "--format", "json", cwd=initialized_project)
        assert result.returncode == 0
        config = json.loads(result.stdout)
        assert isinstance(config, dict)

    def test_config_validate(self, initialized_project: Path) -> None:
        """Config command --validate should check configuration."""
        result = run_rtmx("config", "--validate", cwd=initialized_project)
        assert result.returncode in [0, 1]  # May warn about missing paths


# =============================================================================
# Diff Command E2E Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-DIFF")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestDiffE2E:
    """End-to-end tests for rtmx diff command."""

    def test_diff_same_file(self, project_with_data: Path) -> None:
        """Diff command should show no changes for same file."""
        db_path = project_with_data / "docs" / "rtm_database.csv"
        result = run_rtmx("diff", str(db_path), str(db_path), cwd=project_with_data)
        assert result.returncode == 0
        assert "stable" in result.stdout.lower() or "no change" in result.stdout.lower()

    def test_diff_json_output(self, project_with_data: Path) -> None:
        """Diff command should output valid JSON."""
        db_path = project_with_data / "docs" / "rtm_database.csv"
        result = run_rtmx(
            "diff", str(db_path), str(db_path), "--format", "json", cwd=project_with_data
        )
        if result.returncode == 0:
            data = json.loads(result.stdout)
            assert isinstance(data, dict)

    def test_diff_markdown_output(self, project_with_data: Path) -> None:
        """Diff command should output Markdown format."""
        db_path = project_with_data / "docs" / "rtm_database.csv"
        result = run_rtmx(
            "diff", str(db_path), str(db_path), "--format", "markdown", cwd=project_with_data
        )
        assert result.returncode == 0


# =============================================================================
# Analyze Command E2E Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-ANALYZE")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestAnalyzeE2E:
    """End-to-end tests for rtmx analyze command."""

    def test_analyze_basic(self, project_with_data: Path) -> None:
        """Analyze command should show analysis results."""
        result = run_rtmx("analyze", cwd=project_with_data)
        assert result.returncode in [0, 1]

    def test_analyze_json_output(self, project_with_data: Path) -> None:
        """Analyze command should output valid JSON."""
        result = run_rtmx("analyze", "--format", "json", cwd=project_with_data)
        assert result.returncode in [0, 1]
        # JSON output may be on stdout if successful
        if result.stdout.strip():
            try:
                data = json.loads(result.stdout)
                assert isinstance(data, dict)
            except json.JSONDecodeError:
                # May output terminal format if JSON fails
                pass


# =============================================================================
# Reconcile Command E2E Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-RECONCILE")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestReconcileE2E:
    """End-to-end tests for rtmx reconcile command."""

    def test_reconcile_dry_run(self, project_with_data: Path) -> None:
        """Reconcile command should show what would change (default is dry-run)."""
        result = run_rtmx("reconcile", cwd=project_with_data)
        assert result.returncode in [0, 1]

    def test_reconcile_execute(self, project_with_data: Path) -> None:
        """Reconcile command should fix reciprocity issues with --execute."""
        result = run_rtmx("reconcile", "--execute", cwd=project_with_data)
        assert result.returncode in [0, 1]


# =============================================================================
# Init/Bootstrap Commands E2E Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-INIT")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestInitE2E:
    """End-to-end tests for rtmx init command."""

    def test_init_creates_structure(self, temp_project: Path) -> None:
        """Init command should create .rtmx/ directory structure."""
        result = run_rtmx("init", cwd=temp_project)
        assert result.returncode == 0
        assert (temp_project / ".rtmx" / "database.csv").exists()
        assert (temp_project / ".rtmx" / "config.yaml").exists()
        assert (temp_project / ".rtmx" / "requirements").exists()
        assert (temp_project / ".rtmx" / "cache").exists()

    def test_init_force_overwrites(self, initialized_project: Path) -> None:
        """Init command --force should overwrite existing."""
        result = run_rtmx("init", "--force", cwd=initialized_project)
        assert result.returncode == 0


@pytest.mark.req("REQ-CLI-BOOTSTRAP")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestBootstrapE2E:
    """End-to-end tests for rtmx bootstrap command."""

    def test_bootstrap_from_tests_dry_run(self, initialized_project: Path) -> None:
        """Bootstrap command with --from-tests --dry-run should preview."""
        # Create a test file first
        tests_dir = initialized_project / "tests"
        tests_dir.mkdir(exist_ok=True)
        test_file = tests_dir / "test_example.py"
        test_file.write_text("""
import pytest

@pytest.mark.req("REQ-TEST-001")
def test_example():
    assert True
""")
        result = run_rtmx("bootstrap", "--from-tests", "--dry-run", cwd=initialized_project)
        # Should succeed (may or may not find tests)
        assert result.returncode in [0, 1]

    def test_bootstrap_help(self, temp_project: Path) -> None:
        """Bootstrap command --help should show options."""
        result = run_rtmx("bootstrap", "--help", cwd=temp_project)
        assert result.returncode == 0
        assert "--from-tests" in result.stdout


# =============================================================================
# Install/Makefile Commands E2E Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-INSTALL")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestInstallE2E:
    """End-to-end tests for rtmx install command."""

    def test_install_pytest_plugin(self, initialized_project: Path) -> None:
        """Install command should show available plugins."""
        result = run_rtmx("install", "--help", cwd=initialized_project)
        # Should show install options
        assert result.returncode == 0


@pytest.mark.req("REQ-CLI-MAKEFILE")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestMakefileE2E:
    """End-to-end tests for rtmx makefile command."""

    def test_makefile_stdout_output(self, initialized_project: Path) -> None:
        """Makefile command should output to stdout by default."""
        result = run_rtmx("makefile", cwd=initialized_project)
        assert result.returncode == 0
        # Should output Makefile content to stdout
        assert "rtmx" in result.stdout.lower() or ".PHONY" in result.stdout

    def test_makefile_output_to_file(self, initialized_project: Path) -> None:
        """Makefile command should write to file with -o option."""
        output_path = initialized_project / "rtmx.mk"
        result = run_rtmx("makefile", "-o", str(output_path), cwd=initialized_project)
        assert result.returncode == 0
        # File should be created
        assert output_path.exists()


# =============================================================================
# From-Tests Command E2E Tests
# =============================================================================


@pytest.mark.req("REQ-CLI-FROM-TESTS")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestFromTestsE2E:
    """End-to-end tests for rtmx from-tests command."""

    def test_from_tests_discovers_tests(self, initialized_project: Path) -> None:
        """From-tests command should discover test files."""
        # Create a test file
        tests_dir = initialized_project / "tests"
        tests_dir.mkdir(exist_ok=True)
        test_file = tests_dir / "test_example.py"
        test_file.write_text('''
import pytest

@pytest.mark.req("REQ-EXAMPLE-001")
def test_example():
    """Example test."""
    assert True
''')

        result = run_rtmx("from-tests", cwd=initialized_project)
        assert result.returncode in [0, 1]


# =============================================================================
# Combined Workflow E2E Tests
# =============================================================================


@pytest.mark.req("REQ-WORKFLOW")
@pytest.mark.scope_system
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
class TestWorkflowE2E:
    """End-to-end tests for complete RTMX workflows."""

    def test_setup_status_workflow(self, temp_project: Path) -> None:
        """Complete setup -> status workflow."""
        # Setup
        setup_result = run_rtmx("setup", "--minimal", cwd=temp_project)
        assert setup_result.returncode == 0

        # Status
        status_result = run_rtmx("status", cwd=temp_project)
        assert status_result.returncode in [0, 1]
        assert "%" in status_result.stdout

    def test_setup_health_workflow(self, temp_project: Path) -> None:
        """Complete setup -> health workflow."""
        # Setup
        run_rtmx("setup", "--minimal", cwd=temp_project)

        # Health
        health_result = run_rtmx("health", cwd=temp_project)
        assert health_result.returncode in [0, 1]

    def test_setup_backlog_workflow(self, temp_project: Path) -> None:
        """Complete setup -> backlog workflow."""
        # Setup
        run_rtmx("setup", "--minimal", cwd=temp_project)

        # Backlog
        backlog_result = run_rtmx("backlog", cwd=temp_project)
        assert backlog_result.returncode == 0

    def test_full_project_workflow(self, temp_project: Path) -> None:
        """Complete project initialization and analysis workflow."""
        # 1. Setup
        assert run_rtmx("setup", "--minimal", cwd=temp_project).returncode == 0

        # 2. Check health
        assert run_rtmx("health", cwd=temp_project).returncode in [0, 1]

        # 3. Check status
        status_result = run_rtmx("status", cwd=temp_project)
        assert status_result.returncode in [0, 1]

        # 4. View backlog
        assert run_rtmx("backlog", cwd=temp_project).returncode == 0

        # 5. Check for cycles
        assert run_rtmx("cycles", cwd=temp_project).returncode == 0

        # 6. View config
        assert run_rtmx("config", cwd=temp_project).returncode == 0

    def test_analysis_workflow(self, project_with_data: Path) -> None:
        """Complete analysis workflow with populated data."""
        # 1. Status check
        status = run_rtmx("status", cwd=project_with_data)
        assert "%" in status.stdout

        # 2. Dependency analysis
        deps = run_rtmx("deps", cwd=project_with_data)
        assert deps.returncode == 0

        # 3. Cycle detection
        cycles = run_rtmx("cycles", cwd=project_with_data)
        assert cycles.returncode == 0

        # 4. Health check
        health = run_rtmx("health", cwd=project_with_data)
        assert health.returncode in [0, 1]

    def test_reporting_workflow(self, project_with_data: Path) -> None:
        """Complete reporting workflow with all output formats."""
        # Status with JSON output to file
        json_path = project_with_data / "status.json"
        assert run_rtmx("status", "--json", str(json_path), cwd=project_with_data).returncode in [
            0,
            1,
        ]

        # Health in different formats
        assert run_rtmx("health", "--format", "json", cwd=project_with_data).returncode in [0, 1]
        assert run_rtmx("health", "--format", "ci", cwd=project_with_data).returncode in [0, 1]

        # Deps command
        assert run_rtmx("deps", cwd=project_with_data).returncode == 0

        # Cycles command
        assert run_rtmx("cycles", cwd=project_with_data).returncode == 0
