"""RTMX integrate command.

Orchestrates the full integration workflow for existing projects.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from enum import Enum
from pathlib import Path
from typing import Any

from rtmx.cli.git_ops import (
    GitError,
    check_gh_installed,
    create_branch,
    create_pr,
    create_rollback_point,
    create_worktree,
    generate_branch_name,
    get_git_status,
    is_git_repo,
    print_git_status,
)
from rtmx.cli.health import HealthStatus, run_health_checks
from rtmx.comparison import capture_baseline, compare_databases
from rtmx.config import RTMXConfig, load_config
from rtmx.formatting import Colors, header


class IntegrationMode(str, Enum):
    """Integration mode."""

    VALIDATE = "validate"  # Only validate, no changes
    PREVIEW = "preview"  # Show what would happen
    EXECUTE = "execute"  # Actually perform integration


class GitStrategy(str, Enum):
    """Git isolation strategy."""

    WORKTREE = "worktree"  # Create separate worktree
    BRANCH = "branch"  # Create branch in same repo


@dataclass
class IntegrationResult:
    """Result of integration attempt."""

    success: bool
    mode: IntegrationMode
    git_strategy: GitStrategy | None
    branch_name: str | None
    worktree_path: Path | None
    rollback_point: str | None
    baseline_captured: bool
    health_status: HealthStatus | None
    comparison_status: str | None
    pr_url: str | None
    errors: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)

    def to_dict(self) -> dict[str, Any]:
        """Convert to dictionary."""
        return {
            "success": self.success,
            "mode": self.mode.value,
            "git_strategy": self.git_strategy.value if self.git_strategy else None,
            "branch_name": self.branch_name,
            "worktree_path": str(self.worktree_path) if self.worktree_path else None,
            "rollback_point": self.rollback_point,
            "baseline_captured": self.baseline_captured,
            "health_status": self.health_status.value if self.health_status else None,
            "comparison_status": self.comparison_status,
            "pr_url": self.pr_url,
            "errors": self.errors,
            "warnings": self.warnings,
        }


def run_integrate(
    project_path: Path | None = None,
    mode: IntegrationMode = IntegrationMode.VALIDATE,
    git_strategy: GitStrategy = GitStrategy.WORKTREE,
    branch_name: str | None = None,
    create_pr_flag: bool = False,
    non_interactive: bool = False,  # noqa: ARG001 Reserved for future interactive mode
    config: RTMXConfig | None = None,
) -> IntegrationResult:
    """Run integration workflow.

    Args:
        project_path: Path to project (default: current directory)
        mode: Integration mode (validate, preview, execute)
        git_strategy: Git isolation strategy
        branch_name: Custom branch name (auto-generated if None)
        create_pr_flag: Whether to create a PR after integration
        non_interactive: Skip interactive prompts
        config: Pre-loaded config

    Returns:
        IntegrationResult with status and details
    """
    project_path = project_path or Path.cwd()

    result = IntegrationResult(
        success=False,
        mode=mode,
        git_strategy=git_strategy if mode == IntegrationMode.EXECUTE else None,
        branch_name=None,
        worktree_path=None,
        rollback_point=None,
        baseline_captured=False,
        health_status=None,
        comparison_status=None,
        pr_url=None,
    )

    print(header("RTMX Integration", "="))
    print()
    print(f"Project: {project_path}")
    print(f"Mode: {mode.value}")
    if mode == IntegrationMode.EXECUTE:
        print(f"Git Strategy: {git_strategy.value}")
    print()

    # Phase 1: Pre-flight checks
    print(header("Phase 1: Pre-flight Checks", "-"))

    # Check if git repo
    if not is_git_repo(project_path):
        result.errors.append("Not a git repository")
        print(f"{Colors.RED}  [FAIL] Not a git repository{Colors.RESET}")
        return result
    print(f"{Colors.GREEN}  [PASS] Git repository detected{Colors.RESET}")

    # Check git status
    try:
        git_status = get_git_status(project_path)
        print_git_status(git_status)

        if not git_status.is_clean:
            if mode == IntegrationMode.EXECUTE:
                result.errors.append("Uncommitted changes detected")
                print(f"{Colors.RED}  [FAIL] Cannot execute with uncommitted changes{Colors.RESET}")
                return result
            else:
                result.warnings.append("Uncommitted changes detected")
                print(f"{Colors.YELLOW}  [WARN] Uncommitted changes (would block execute mode){Colors.RESET}")
    except GitError as e:
        result.errors.append(str(e))
        print(f"{Colors.RED}  [FAIL] Git error: {e}{Colors.RESET}")
        return result

    # Create rollback point
    result.rollback_point = create_rollback_point(project_path)
    print(f"  Rollback point: {result.rollback_point[:8]}")
    print()

    # Phase 2: Baseline capture
    print(header("Phase 2: Baseline Capture", "-"))

    rtm_path = project_path / "docs" / "rtm_database.csv"
    if rtm_path.exists():
        try:
            baseline = capture_baseline(rtm_path)
            result.baseline_captured = True
            print(f"{Colors.GREEN}  [PASS] Baseline captured{Colors.RESET}")
            print(f"    Requirements: {baseline['req_count']}")
            print(f"    Completion: {baseline['completion']:.1f}%")
            print(f"    Cycles: {baseline['cycles']}")
        except Exception as e:
            result.warnings.append(f"Baseline capture failed: {e}")
            print(f"{Colors.YELLOW}  [WARN] Baseline capture failed: {e}{Colors.RESET}")
    else:
        print("  [SKIP] No existing RTM database found")
    print()

    # Phase 3: Health check
    print(header("Phase 3: Health Check", "-"))

    if config is None:
        config = load_config()

    health_report = run_health_checks(config)
    result.health_status = health_report.status

    if health_report.status == HealthStatus.HEALTHY:
        print(f"{Colors.GREEN}  [PASS] Health: HEALTHY{Colors.RESET}")
    elif health_report.status == HealthStatus.DEGRADED:
        print(f"{Colors.YELLOW}  [WARN] Health: DEGRADED{Colors.RESET}")
        for check in health_report.checks:
            if check.result.value == "warn":
                result.warnings.append(check.message)
    else:
        print(f"{Colors.RED}  [FAIL] Health: UNHEALTHY{Colors.RESET}")
        for check in health_report.checks:
            if check.result.value == "fail" and check.blocking:
                result.errors.append(check.message)

    print(
        f"    Checks: {health_report.summary['passed']} passed, "
        f"{health_report.summary['warnings']} warnings, "
        f"{health_report.summary['failed']} failed"
    )
    print()

    # Validate mode stops here
    if mode == IntegrationMode.VALIDATE:
        result.success = health_report.status != HealthStatus.UNHEALTHY
        print(header("Validation Complete", "="))
        if result.success:
            print(f"{Colors.GREEN}Validation PASSED{Colors.RESET}")
        else:
            print(f"{Colors.RED}Validation FAILED{Colors.RESET}")
        return result

    # Phase 4: Preview/Execute integration
    print(header(f"Phase 4: {'Preview' if mode == IntegrationMode.PREVIEW else 'Execute'} Integration", "-"))

    # Generate branch name
    if branch_name is None:
        branch_name = generate_branch_name()
    result.branch_name = branch_name
    print(f"  Branch: {branch_name}")

    if mode == IntegrationMode.PREVIEW:
        print()
        print("Preview mode - showing what would happen:")
        print()

        if git_strategy == GitStrategy.WORKTREE:
            worktree_path = project_path.parent / f"{project_path.name}-rtmx-integration"
            print(f"  1. Create worktree at: {worktree_path}")
            print(f"  2. Create branch: {branch_name}")
        else:
            print(f"  1. Create branch: {branch_name}")
            print("  2. Checkout branch")

        print("  3. Run `rtmx init` (if needed)")
        print("  4. Run `rtmx install --all`")
        print("  5. Run health validation")
        print("  6. Generate comparison report")

        if create_pr_flag:
            print("  7. Create pull request")

        result.success = True
        print()
        print(f"{Colors.GREEN}Preview complete. Run with --mode execute to proceed.{Colors.RESET}")
        return result

    # Execute mode
    if git_strategy == GitStrategy.WORKTREE:
        worktree_path = project_path.parent / f"{project_path.name}-rtmx-integration"
        result.worktree_path = worktree_path

        try:
            print(f"  Creating worktree at: {worktree_path}")
            create_worktree(project_path, worktree_path, branch_name)
            print(f"{Colors.GREEN}  [PASS] Worktree created{Colors.RESET}")
        except GitError as e:
            result.errors.append(f"Failed to create worktree: {e}")
            print(f"{Colors.RED}  [FAIL] Worktree creation failed: {e}{Colors.RESET}")
            return result

        # Work in worktree
        work_path = worktree_path
    else:
        # Branch strategy - work in place
        try:
            print(f"  Creating branch: {branch_name}")
            create_branch(project_path, branch_name)
            print(f"{Colors.GREEN}  [PASS] Branch created{Colors.RESET}")
        except GitError as e:
            result.errors.append(f"Failed to create branch: {e}")
            print(f"{Colors.RED}  [FAIL] Branch creation failed: {e}{Colors.RESET}")
            return result

        work_path = project_path

    print()

    # Run rtmx init if needed
    rtmx_yaml = work_path / "rtmx.yaml"
    if not rtmx_yaml.exists():
        print("  Running rtmx init...")
        from rtmx.cli.init import run_init

        try:
            # Save current directory and change to work path
            import os

            original_cwd = os.getcwd()
            os.chdir(work_path)
            run_init(force=False)
            os.chdir(original_cwd)
            print(f"{Colors.GREEN}  [PASS] rtmx init completed{Colors.RESET}")
        except Exception as e:
            result.warnings.append(f"rtmx init warning: {e}")
            print(f"{Colors.YELLOW}  [WARN] rtmx init: {e}{Colors.RESET}")

    # Run rtmx install
    print("  Running rtmx install --all...")
    from rtmx.cli.install import run_install

    try:
        import os

        original_cwd = os.getcwd()
        os.chdir(work_path)
        run_install(
            dry_run=False,
            non_interactive=True,
            force=False,
            agents=None,
            install_all=True,
            skip_backup=False,
            config=config,
        )
        os.chdir(original_cwd)
        print(f"{Colors.GREEN}  [PASS] rtmx install completed{Colors.RESET}")
    except Exception as e:
        result.warnings.append(f"rtmx install warning: {e}")
        print(f"{Colors.YELLOW}  [WARN] rtmx install: {e}{Colors.RESET}")

    print()

    # Phase 5: Post-integration validation
    print(header("Phase 5: Post-Integration Validation", "-"))

    # Reload config from work path
    work_config = load_config(work_path / "rtmx.yaml" if (work_path / "rtmx.yaml").exists() else None)

    post_health = run_health_checks(work_config)

    if post_health.status == HealthStatus.HEALTHY:
        print(f"{Colors.GREEN}  [PASS] Post-integration health: HEALTHY{Colors.RESET}")
    elif post_health.status == HealthStatus.DEGRADED:
        print(f"{Colors.YELLOW}  [WARN] Post-integration health: DEGRADED{Colors.RESET}")
    else:
        print(f"{Colors.RED}  [FAIL] Post-integration health: UNHEALTHY{Colors.RESET}")
        result.errors.append("Post-integration health check failed")

    # Comparison
    if result.baseline_captured:
        new_rtm_path = work_path / "docs" / "rtm_database.csv"
        if new_rtm_path.exists():
            try:
                comparison = compare_databases(rtm_path, new_rtm_path)
                result.comparison_status = comparison.summary_status
                print(f"  Comparison: {comparison.summary_status}")
                print(f"    Req count: {comparison.baseline_req_count} -> {comparison.current_req_count}")
                print(f"    Completion: {comparison.baseline_completion:.1f}% -> {comparison.current_completion:.1f}%")
            except Exception as e:
                result.warnings.append(f"Comparison failed: {e}")

    print()

    # Phase 6: Create PR if requested
    if create_pr_flag:
        print(header("Phase 6: Create Pull Request", "-"))

        if not check_gh_installed():
            result.warnings.append("GitHub CLI not installed - skipping PR creation")
            print(f"{Colors.YELLOW}  [SKIP] GitHub CLI (gh) not installed{Colors.RESET}")
        else:
            pr_title = "chore: Integrate rtmx for requirements traceability"
            pr_body = f"""## Summary
- Added rtmx configuration (rtmx.yaml)
- Updated agent prompts with RTMX guidance
- Integration validated with rtmx health check

## Validation Results
- Health Status: {post_health.status.value}
- Checks Passed: {post_health.summary['passed']}
- Warnings: {post_health.summary['warnings']}

## Rollback
If needed, rollback with:
```bash
git revert -m 1 <merge-commit-sha>
```

---
Generated by `rtmx integrate`
"""
            try:
                pr_url = create_pr(work_path, pr_title, pr_body, base="main")
                if pr_url:
                    result.pr_url = pr_url
                    print(f"{Colors.GREEN}  [PASS] PR created: {pr_url}{Colors.RESET}")
                else:
                    result.warnings.append("PR creation returned no URL")
            except GitError as e:
                result.warnings.append(f"PR creation failed: {e}")
                print(f"{Colors.YELLOW}  [WARN] PR creation failed: {e}{Colors.RESET}")

    print()

    # Final summary
    print(header("Integration Summary", "="))

    if result.errors:
        result.success = False
        print(f"{Colors.RED}Integration FAILED{Colors.RESET}")
        print()
        print("Errors:")
        for error in result.errors:
            print(f"  - {error}")
    else:
        result.success = True
        print(f"{Colors.GREEN}Integration SUCCESSFUL{Colors.RESET}")

    if result.warnings:
        print()
        print("Warnings:")
        for warning in result.warnings:
            print(f"  - {warning}")

    print()
    print("Next steps:")
    if result.pr_url:
        print(f"  1. Review PR: {result.pr_url}")
        print("  2. Merge when ready: gh pr merge --squash")
    elif result.branch_name:
        print(f"  1. Review changes in branch: {result.branch_name}")
        print("  2. Create PR: gh pr create --title 'chore: Integrate rtmx'")
        print("  3. Merge when ready")

    if result.worktree_path:
        print(f"  4. Cleanup worktree: git worktree remove {result.worktree_path}")

    print()
    print("Rollback if needed:")
    print(f"  git reset --hard {result.rollback_point}")

    return result
