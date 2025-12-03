"""RTMX analyze command.

Analyze project for requirements artifacts.
"""

from __future__ import annotations

from pathlib import Path

from rtmx.config import RTMXConfig
from rtmx.formatting import Colors


def run_analyze(
    path: Path | None,
    _output: Path | None,
    _output_format: str,
    _deep: bool,
    config: RTMXConfig,
) -> None:
    """Run analyze command.

    Discovers tests, issues, documentation that could become requirements.

    Args:
        path: Path to analyze (defaults to cwd)
        output: Output file path for report
        output_format: Output format (terminal, json, markdown)
        deep: Include source code analysis
        config: RTMX configuration
    """
    target_path = path or Path.cwd()

    print("=== RTMX Project Analysis ===")
    print()
    print(f"Analyzing: {target_path}")
    print()

    # Discover test files
    print(f"{Colors.BOLD}Test Files:{Colors.RESET}")
    test_files = list(target_path.rglob("test_*.py"))
    if test_files:
        from rtmx.cli.from_tests import extract_markers_from_file

        unmarked_count = 0
        for test_file in test_files[:10]:  # Limit to first 10
            markers = extract_markers_from_file(test_file)
            if not markers:
                print(f"  {Colors.YELLOW}○{Colors.RESET} {test_file.relative_to(target_path)}")
                unmarked_count += 1
            else:
                print(
                    f"  {Colors.GREEN}✓{Colors.RESET} {test_file.relative_to(target_path)} ({len(markers)} req markers)"
                )

        if len(test_files) > 10:
            print(f"  ... and {len(test_files) - 10} more test files")
        print()
        print(f"  Tests without req markers: {unmarked_count}")
    else:
        print(f"  {Colors.DIM}No test files found{Colors.RESET}")
    print()

    # Check GitHub adapter config
    if config.adapters.github.enabled and config.adapters.github.repo:
        print(f"{Colors.BOLD}GitHub Issues:{Colors.RESET}")
        print(f"  Repository: {config.adapters.github.repo}")
        print(f"  {Colors.YELLOW}(Run 'rtmx bootstrap --from-github' to import){Colors.RESET}")
    else:
        print(f"{Colors.BOLD}GitHub Issues:{Colors.RESET}")
        print(f"  {Colors.DIM}Not configured (add adapters.github to rtmx.yaml){Colors.RESET}")
    print()

    # Check Jira adapter config
    if config.adapters.jira.enabled and config.adapters.jira.project:
        print(f"{Colors.BOLD}Jira Tickets:{Colors.RESET}")
        print(f"  Project: {config.adapters.jira.project}")
        print(f"  {Colors.YELLOW}(Run 'rtmx bootstrap --from-jira' to import){Colors.RESET}")
    else:
        print(f"{Colors.BOLD}Jira Tickets:{Colors.RESET}")
        print(f"  {Colors.DIM}Not configured (add adapters.jira to rtmx.yaml){Colors.RESET}")
    print()

    # Check for existing RTM
    rtm_path = target_path / config.database
    if rtm_path.exists():
        print(f"{Colors.BOLD}Existing RTM:{Colors.RESET}")
        print(f"  {Colors.GREEN}✓{Colors.RESET} Found at {rtm_path}")
    else:
        print(f"{Colors.BOLD}Existing RTM:{Colors.RESET}")
        print(f"  {Colors.YELLOW}○{Colors.RESET} Not found (run 'rtmx init' to create)")
    print()

    # Recommendations
    print(f"{Colors.BOLD}Recommendations:{Colors.RESET}")
    recommendations = []
    if not rtm_path.exists():
        recommendations.append("Run 'rtmx init' to create RTM structure")
    if test_files and unmarked_count > 0:
        recommendations.append(
            "Run 'rtmx bootstrap --from-tests' to generate requirements from tests"
        )
    if config.adapters.github.enabled and config.adapters.github.repo:
        recommendations.append("Run 'rtmx sync github --import' to import GitHub issues")
    if config.adapters.jira.enabled and config.adapters.jira.project:
        recommendations.append("Run 'rtmx sync jira --import' to import Jira tickets")
    recommendations.append("Run 'rtmx install' to add RTM prompts to AI agent configs")

    for i, rec in enumerate(recommendations, 1):
        print(f"  {i}. {rec}")
