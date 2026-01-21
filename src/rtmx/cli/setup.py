"""RTMX setup command.

Single command to fully integrate rtmx into a project.
Idempotent, non-destructive, with smart defaults.
Supports git branch isolation and PR workflow.
"""

from __future__ import annotations

import shutil
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Any

from rtmx.config import load_config
from rtmx.formatting import Colors, header


@dataclass
class SetupResult:
    """Result of setup operation."""

    success: bool
    steps_completed: list[str] = field(default_factory=list)
    steps_skipped: list[str] = field(default_factory=list)
    files_created: list[str] = field(default_factory=list)
    files_modified: list[str] = field(default_factory=list)
    files_backed_up: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)
    errors: list[str] = field(default_factory=list)
    # Git workflow fields
    branch_name: str | None = None
    rollback_point: str | None = None
    pr_url: str | None = None

    def to_dict(self) -> dict[str, Any]:
        """Convert to dictionary."""
        return {
            "success": self.success,
            "steps_completed": self.steps_completed,
            "steps_skipped": self.steps_skipped,
            "files_created": self.files_created,
            "files_modified": self.files_modified,
            "files_backed_up": self.files_backed_up,
            "warnings": self.warnings,
            "errors": self.errors,
            "branch_name": self.branch_name,
            "rollback_point": self.rollback_point,
            "pr_url": self.pr_url,
        }


def backup_file(path: Path) -> Path | None:
    """Create timestamped backup of a file.

    Returns:
        Path to backup file, or None if file doesn't exist.
    """
    if not path.exists():
        return None

    timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
    backup_path = path.with_suffix(f".rtmx-backup-{timestamp}{path.suffix}")
    shutil.copy2(path, backup_path)
    return backup_path


def detect_project(project_path: Path) -> dict[str, Any]:
    """Detect project characteristics.

    Returns:
        Dictionary with detection results.
    """
    from rtmx.cli.git_ops import get_git_status, is_git_repo

    detection: dict[str, Any] = {
        "is_git_repo": is_git_repo(project_path),
        "git_clean": True,
        "git_branch": None,
        "has_rtmx_config": (project_path / "rtmx.yaml").exists()
        or (project_path / ".rtmx.yaml").exists(),
        "has_rtm_database": (project_path / "docs" / "rtm_database.csv").exists(),
        "has_tests": (project_path / "tests").is_dir() or (project_path / "test").is_dir(),
        "has_makefile": (project_path / "Makefile").exists(),
        "has_pyproject": (project_path / "pyproject.toml").exists(),
        "agent_configs": {},
    }

    # Get git status if in a repo
    if detection["is_git_repo"]:
        try:
            git_status = get_git_status(project_path)
            detection["git_clean"] = git_status.is_clean
            detection["git_branch"] = git_status.branch
        except Exception:
            pass

    # Detect agent config files
    agent_files = {
        "claude": "CLAUDE.md",
        "cursor": ".cursorrules",
        "copilot": ".github/copilot-instructions.md",
        "windsurf": ".windsurfrules",
        "aider": ".aider.conf.yml",
    }

    for agent_name, filename in agent_files.items():
        agent_path = project_path / filename
        if agent_path.exists():
            content = agent_path.read_text()
            detection["agent_configs"][agent_name] = {
                "exists": True,
                "has_rtmx": "RTMX" in content or "rtmx" in content,
                "path": str(agent_path),
            }
        else:
            detection["agent_configs"][agent_name] = {
                "exists": False,
                "has_rtmx": False,
                "path": str(agent_path),
            }

    return detection


def run_setup(
    project_path: Path | None = None,
    dry_run: bool = False,
    minimal: bool = False,
    force: bool = False,
    skip_agents: bool = False,
    skip_makefile: bool = False,
    branch: str | None = None,
    create_pr: bool = False,
) -> SetupResult:
    """Run setup command.

    Performs complete rtmx integration in a single command:
    1. Optionally create git branch for isolation
    2. Detect project characteristics
    3. Create rtmx.yaml if missing
    4. Create RTM database if missing
    5. Scan tests for existing markers
    6. Inject agent prompts (with backup)
    7. Add Makefile targets
    8. Run health validation
    9. Optionally create PR

    Args:
        project_path: Path to project (default: current directory)
        dry_run: Preview without making changes
        minimal: Only create config and RTM, skip agents/makefile
        force: Overwrite existing files
        skip_agents: Skip agent config injection
        skip_makefile: Skip Makefile targets
        branch: Create git branch with this name (or auto-generate if True-ish)
        create_pr: Create pull request after setup (implies branch)

    Returns:
        SetupResult with details of what was done
    """
    project_path = project_path or Path.cwd()

    result = SetupResult(success=False)

    # If --pr specified, ensure we're using branch mode
    if create_pr and not branch:
        branch = "auto"

    print(header("RTMX Setup", "="))
    print()
    print(f"Project: {project_path}")
    if dry_run:
        print(f"{Colors.YELLOW}DRY RUN - no changes will be made{Colors.RESET}")
    if branch:
        print(f"{Colors.CYAN}Branch mode: changes will be on a new branch{Colors.RESET}")
    if create_pr:
        print(f"{Colors.CYAN}PR mode: will create pull request{Colors.RESET}")
    print()

    # Phase 1: Detection
    print(header("Phase 1: Project Detection", "-"))
    detection = detect_project(project_path)

    print(f"  Git repository: {'Yes' if detection['is_git_repo'] else 'No'}")
    if detection["is_git_repo"]:
        print(f"  Git branch: {detection['git_branch']}")
        print(f"  Working tree: {'Clean' if detection['git_clean'] else 'Has changes'}")
    print(f"  RTMX config: {'Found' if detection['has_rtmx_config'] else 'Not found'}")
    print(f"  RTM database: {'Found' if detection['has_rtm_database'] else 'Not found'}")
    print(f"  Tests directory: {'Found' if detection['has_tests'] else 'Not found'}")
    print(f"  Makefile: {'Found' if detection['has_makefile'] else 'Not found'}")

    agent_summary = []
    for agent_name, info in detection["agent_configs"].items():
        if info["exists"]:
            status = "configured" if info["has_rtmx"] else "exists"
            agent_summary.append(f"{agent_name}({status})")
    if agent_summary:
        print(f"  Agent configs: {', '.join(agent_summary)}")
    else:
        print("  Agent configs: None found")
    print()

    # Phase 1.5: Git branch setup (if requested)
    if branch and detection["is_git_repo"]:
        print(header("Phase 1.5: Git Branch Setup", "-"))

        from rtmx.cli.git_ops import (
            GitError,
            create_branch,
            create_rollback_point,
            generate_branch_name,
        )

        # Check for uncommitted changes
        if not detection["git_clean"] and not force:
            print(f"  {Colors.RED}[FAIL] Uncommitted changes detected{Colors.RESET}")
            print(f"  {Colors.DIM}Commit or stash changes first, or use --force{Colors.RESET}")
            result.errors.append("Uncommitted changes block branch creation")
            return result

        # Generate branch name if auto
        if branch == "auto":
            branch = generate_branch_name(prefix="rtmx/setup")
        result.branch_name = branch

        # Create rollback point
        try:
            result.rollback_point = create_rollback_point(project_path)
            print(f"  Rollback point: {result.rollback_point[:8]}")
        except GitError as e:
            print(f"  {Colors.YELLOW}[WARN] Could not create rollback point: {e}{Colors.RESET}")
            result.warnings.append(f"Rollback point failed: {e}")

        # Create and checkout branch
        if not dry_run:
            try:
                create_branch(project_path, branch)
                print(f"  {Colors.GREEN}[CREATE] Branch: {branch}{Colors.RESET}")
                result.steps_completed.append("create_branch")
            except GitError as e:
                print(f"  {Colors.RED}[FAIL] Could not create branch: {e}{Colors.RESET}")
                result.errors.append(f"Branch creation failed: {e}")
                return result
        else:
            print(f"  {Colors.DIM}[SKIP] Would create branch: {branch}{Colors.RESET}")

        print()
    elif branch and not detection["is_git_repo"]:
        print(
            f"{Colors.YELLOW}Warning: --branch requested but not a git repo, ignoring{Colors.RESET}"
        )
        result.warnings.append("--branch ignored: not a git repository")
        branch = None
        print()

    # Phase 2: Create config
    print(header("Phase 2: Configuration", "-"))
    config_path = project_path / "rtmx.yaml"

    if detection["has_rtmx_config"] and not force:
        print(f"  {Colors.DIM}[SKIP] rtmx.yaml already exists{Colors.RESET}")
        result.steps_skipped.append("create_config")
    else:
        config_content = """# RTMX Configuration
# See https://rtmx.ai for documentation

rtmx:
  database: docs/rtm_database.csv
  requirements_dir: docs/requirements
  schema: core
  pytest:
    marker_prefix: "req"
    register_markers: true
"""
        if not dry_run:
            if config_path.exists():
                backup = backup_file(config_path)
                if backup:
                    result.files_backed_up.append(str(backup))
            config_path.write_text(config_content)
            result.files_created.append(str(config_path))
        print(f"  {Colors.GREEN}[CREATE] rtmx.yaml{Colors.RESET}")
        result.steps_completed.append("create_config")
    print()

    # Phase 3: Create RTM database
    print(header("Phase 3: RTM Database", "-"))
    docs_dir = project_path / "docs"
    rtm_path = docs_dir / "rtm_database.csv"
    req_dir = docs_dir / "requirements"

    if detection["has_rtm_database"] and not force:
        print(f"  {Colors.DIM}[SKIP] RTM database already exists{Colors.RESET}")
        result.steps_skipped.append("create_rtm")
    else:
        rtm_content = """req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file
REQ-INIT-001,SETUP,RTMX,RTMX integration complete,Fully configured,tests/test_rtmx.py,test_rtmx_configured,Unit Test,MISSING,HIGH,1,Auto-generated by rtmx setup,0.5,,,developer,v0.1,,,docs/requirements/SETUP/REQ-INIT-001.md
"""
        if not dry_run:
            docs_dir.mkdir(parents=True, exist_ok=True)
            req_dir.mkdir(parents=True, exist_ok=True)
            if rtm_path.exists():
                backup = backup_file(rtm_path)
                if backup:
                    result.files_backed_up.append(str(backup))
            rtm_path.write_text(rtm_content)
            result.files_created.append(str(rtm_path))

            # Create sample requirement spec
            spec_dir = req_dir / "SETUP"
            spec_dir.mkdir(parents=True, exist_ok=True)
            spec_path = spec_dir / "REQ-INIT-001.md"
            spec_content = """# REQ-INIT-001: RTMX Integration Complete

## Description
RTMX has been integrated into this project for requirements traceability.

## Acceptance Criteria
- [ ] rtmx.yaml configuration exists
- [ ] RTM database initialized
- [ ] Agent configs include RTMX guidance
- [ ] `rtmx status` runs without errors

## Validation
- **Test**: tests/test_rtmx.py::test_rtmx_configured
- **Method**: Unit Test
"""
            spec_path.write_text(spec_content)
            result.files_created.append(str(spec_path))

        print(f"  {Colors.GREEN}[CREATE] docs/rtm_database.csv{Colors.RESET}")
        print(f"  {Colors.GREEN}[CREATE] docs/requirements/SETUP/REQ-INIT-001.md{Colors.RESET}")
        result.steps_completed.append("create_rtm")
    print()

    # Phase 4: Scan tests for markers
    if detection["has_tests"] and not minimal:
        print(header("Phase 4: Test Marker Scan", "-"))
        try:
            from rtmx.cli.from_tests import extract_markers_from_file

            test_dir = project_path / "tests"
            if not test_dir.exists():
                test_dir = project_path / "test"

            markers_found = 0
            test_files = list(test_dir.rglob("test_*.py"))

            for test_file in test_files:
                try:
                    file_markers = extract_markers_from_file(test_file)
                    markers_found += len(file_markers)
                except Exception:
                    pass  # Skip files that can't be parsed

            print(f"  Scanned {len(test_files)} test files")
            print(f"  Found {markers_found} requirement markers")
            result.steps_completed.append("scan_tests")
        except Exception as e:
            print(f"  {Colors.YELLOW}[WARN] Could not scan tests: {e}{Colors.RESET}")
            result.warnings.append(f"Test scan failed: {e}")
        print()

    # Phase 5: Agent configs
    if not minimal and not skip_agents:
        print(header("Phase 5: Agent Configurations", "-"))

        rtmx_section = """
## RTMX

This project uses RTMX for requirements traceability.

### Quick Commands
- `rtmx status` - Show RTM progress
- `rtmx backlog` - View prioritized backlog
- `rtmx health` - Run health checks

### When Implementing Requirements
1. Check the RTM: `rtmx status`
2. Mark tests with `@pytest.mark.req("REQ-XXX-NNN")`
3. Update status when complete

### RTM Location
- Database: `docs/rtm_database.csv`
- Specs: `docs/requirements/`
"""

        for agent_name, info in detection["agent_configs"].items():
            agent_path = Path(info["path"])

            if info["has_rtmx"]:
                print(
                    f"  {Colors.DIM}[SKIP] {agent_path.name} already has RTMX section{Colors.RESET}"
                )
                result.steps_skipped.append(f"agent_{agent_name}")
            elif info["exists"]:
                # Append to existing file
                if not dry_run:
                    backup = backup_file(agent_path)
                    if backup:
                        result.files_backed_up.append(str(backup))
                    with agent_path.open("a") as f:
                        f.write("\n" + rtmx_section)
                    result.files_modified.append(str(agent_path))
                print(f"  {Colors.GREEN}[UPDATE] {agent_path.name}{Colors.RESET}")
                result.steps_completed.append(f"agent_{agent_name}")
            else:
                # Create new file for key agents
                if agent_name in ("claude", "cursor"):
                    if not dry_run:
                        agent_path.parent.mkdir(parents=True, exist_ok=True)
                        agent_path.write_text(f"# {agent_path.name}\n{rtmx_section}")
                        result.files_created.append(str(agent_path))
                    print(f"  {Colors.GREEN}[CREATE] {agent_path.name}{Colors.RESET}")
                    result.steps_completed.append(f"agent_{agent_name}")

        print()

    # Phase 6: Makefile
    if not minimal and not skip_makefile and detection["has_makefile"]:
        print(header("Phase 6: Makefile Targets", "-"))
        makefile_path = project_path / "Makefile"

        # Check if rtmx targets already exist
        makefile_content = makefile_path.read_text()
        if "rtmx" in makefile_content.lower() and "rtm:" in makefile_content:
            print(f"  {Colors.DIM}[SKIP] Makefile already has rtmx targets{Colors.RESET}")
            result.steps_skipped.append("makefile")
        else:
            makefile_targets = """
# RTMX targets
.PHONY: rtm backlog health

rtm:
\t@rtmx status

backlog:
\t@rtmx backlog

health:
\t@rtmx health
"""
            if not dry_run:
                backup = backup_file(makefile_path)
                if backup:
                    result.files_backed_up.append(str(backup))
                with makefile_path.open("a") as f:
                    f.write(makefile_targets)
                result.files_modified.append(str(makefile_path))
            print(
                f"  {Colors.GREEN}[UPDATE] Makefile (added rtm, backlog, health targets){Colors.RESET}"
            )
            result.steps_completed.append("makefile")
        print()

    # Phase 7: Health check
    print(header("Phase 7: Validation", "-"))

    if not dry_run:
        try:
            config = load_config(config_path if config_path.exists() else None)
            from rtmx.cli.health import HealthStatus, run_health_checks

            report = run_health_checks(config)

            if report.status == HealthStatus.HEALTHY:
                print(f"  {Colors.GREEN}[PASS] Health: HEALTHY{Colors.RESET}")
            elif report.status == HealthStatus.DEGRADED:
                print(f"  {Colors.YELLOW}[WARN] Health: DEGRADED{Colors.RESET}")
                result.warnings.append("Health check shows warnings")
            else:
                print(f"  {Colors.RED}[FAIL] Health: UNHEALTHY{Colors.RESET}")
                result.errors.append("Health check failed")

            print(
                f"    Checks: {report.summary['passed']} passed, "
                f"{report.summary['warnings']} warnings, "
                f"{report.summary['failed']} failed"
            )
            result.steps_completed.append("health_check")
        except Exception as e:
            print(f"  {Colors.YELLOW}[WARN] Could not run health check: {e}{Colors.RESET}")
            result.warnings.append(f"Health check error: {e}")
    else:
        print(f"  {Colors.DIM}[SKIP] Health check (dry run){Colors.RESET}")

    print()

    # Phase 8: Git commit and PR (if branch mode)
    if branch and not dry_run and detection["is_git_repo"]:
        print(header("Phase 8: Git Commit", "-"))

        import subprocess

        try:
            # Stage all changes
            subprocess.run(
                ["git", "add", "-A"],
                cwd=project_path,
                capture_output=True,
                check=True,
            )

            # Commit
            commit_msg = "chore: Add rtmx configuration\n\nAdded by rtmx setup command."
            subprocess.run(
                ["git", "commit", "-m", commit_msg],
                cwd=project_path,
                capture_output=True,
                check=True,
            )
            print(f"  {Colors.GREEN}[COMMIT] Changes committed to {branch}{Colors.RESET}")
            result.steps_completed.append("git_commit")

            # Create PR if requested
            if create_pr:
                print()
                print(header("Phase 9: Create Pull Request", "-"))

                from rtmx.cli.git_ops import check_gh_installed
                from rtmx.cli.git_ops import create_pr as git_create_pr

                if not check_gh_installed():
                    print(f"  {Colors.YELLOW}[SKIP] GitHub CLI (gh) not installed{Colors.RESET}")
                    result.warnings.append("PR creation skipped: gh CLI not installed")
                else:
                    try:
                        # Push branch first
                        subprocess.run(
                            ["git", "push", "-u", "origin", branch],
                            cwd=project_path,
                            capture_output=True,
                            check=True,
                        )

                        pr_title = "chore: Add rtmx for requirements traceability"
                        pr_body = """## Summary
- Added rtmx configuration (`rtmx.yaml`)
- Initialized RTM database (`docs/rtm_database.csv`)
- Updated agent configurations with RTMX guidance

## Validation
Run `rtmx health` to verify setup.

---
Generated by `rtmx setup --pr`
"""
                        pr_url = git_create_pr(project_path, pr_title, pr_body)
                        if pr_url:
                            result.pr_url = pr_url
                            print(f"  {Colors.GREEN}[CREATE] PR: {pr_url}{Colors.RESET}")
                            result.steps_completed.append("create_pr")
                        else:
                            print(
                                f"  {Colors.YELLOW}[WARN] PR created but no URL returned{Colors.RESET}"
                            )
                    except Exception as e:
                        print(f"  {Colors.YELLOW}[WARN] Could not create PR: {e}{Colors.RESET}")
                        result.warnings.append(f"PR creation failed: {e}")

        except subprocess.CalledProcessError as e:
            print(f"  {Colors.YELLOW}[WARN] Git commit failed: {e}{Colors.RESET}")
            result.warnings.append(f"Git commit failed: {e}")

        print()

    # Summary
    print(header("Setup Complete", "="))

    if result.errors:
        result.success = False
        print(f"{Colors.RED}Setup completed with errors{Colors.RESET}")
    else:
        result.success = True
        print(f"{Colors.GREEN}Setup completed successfully!{Colors.RESET}")

    print()
    print(f"  Steps completed: {len(result.steps_completed)}")
    print(f"  Steps skipped: {len(result.steps_skipped)}")
    print(f"  Files created: {len(result.files_created)}")
    print(f"  Files modified: {len(result.files_modified)}")
    print(f"  Backups created: {len(result.files_backed_up)}")

    if result.branch_name:
        print(f"  Branch: {result.branch_name}")
    if result.rollback_point:
        print(f"  Rollback: git reset --hard {result.rollback_point[:8]}")
    if result.pr_url:
        print(f"  PR: {result.pr_url}")

    if result.warnings:
        print()
        print(f"{Colors.YELLOW}Warnings:{Colors.RESET}")
        for warning in result.warnings:
            print(f"  - {warning}")

    if result.errors:
        print()
        print(f"{Colors.RED}Errors:{Colors.RESET}")
        for error in result.errors:
            print(f"  - {error}")

    print()
    print("Next steps:")
    if result.pr_url:
        print(f"  1. Review PR: {result.pr_url}")
        print("  2. Merge when ready")
    elif result.branch_name:
        print(f"  1. Review changes: git diff main..{result.branch_name}")
        print("  2. Create PR: gh pr create")
        print("  3. Merge when ready")
    else:
        print("  1. Run 'rtmx status' to see your RTM")
        print("  2. Add requirements to docs/rtm_database.csv")
        print("  3. Mark tests with @pytest.mark.req('REQ-XXX-NNN')")
        if detection["has_makefile"]:
            print("  4. Use 'make rtm' for quick status checks")

        # Suggest branch mode for git repos
        if detection["is_git_repo"] and not branch:
            print()
            print(
                f"{Colors.DIM}Tip: Use 'rtmx setup --branch' for PR-based review workflow{Colors.RESET}"
            )

    return result
