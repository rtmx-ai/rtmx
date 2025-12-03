"""RTMX bootstrap command.

Generate initial RTM from project artifacts.
"""

from __future__ import annotations

import sys

from rtmx.config import RTMXConfig
from rtmx.formatting import Colors


def run_bootstrap(
    from_tests: bool,
    from_github: bool,
    from_jira: bool,
    _merge: bool,
    dry_run: bool,
    prefix: str,
    config: RTMXConfig,
) -> None:
    """Run bootstrap command.

    Generate initial RTM from tests, GitHub issues, or Jira tickets.

    Args:
        from_tests: Generate requirements from test functions
        from_github: Import from GitHub issues
        from_jira: Import from Jira tickets
        merge: Merge with existing RTM
        dry_run: Preview without writing files
        prefix: Requirement ID prefix
        config: RTMX configuration
    """
    if not any([from_tests, from_github, from_jira]):
        print(
            f"{Colors.YELLOW}No source specified. Use --from-tests, --from-github, or --from-jira{Colors.RESET}"
        )
        sys.exit(1)

    print("=== RTMX Bootstrap ===")
    print()

    if dry_run:
        print(f"{Colors.YELLOW}DRY RUN - no files will be written{Colors.RESET}")
        print()

    requirements_to_create: list[dict] = []

    # Bootstrap from tests
    if from_tests:
        print(f"{Colors.BOLD}Scanning tests...{Colors.RESET}")
        # TODO: Implement test scanning and requirement generation
        print(f"  {Colors.DIM}Not yet implemented - coming in Phase 2{Colors.RESET}")
        print()

    # Bootstrap from GitHub
    if from_github:
        print(f"{Colors.BOLD}Fetching GitHub issues...{Colors.RESET}")
        if not config.adapters.github.enabled:
            print(f"  {Colors.RED}GitHub adapter not enabled in rtmx.yaml{Colors.RESET}")
        elif not config.adapters.github.repo:
            print(f"  {Colors.RED}GitHub repo not configured in rtmx.yaml{Colors.RESET}")
        else:
            # TODO: Implement GitHub issue fetching
            print(f"  Repository: {config.adapters.github.repo}")
            print(f"  Labels: {', '.join(config.adapters.github.labels)}")
            print(f"  {Colors.DIM}Not yet implemented - coming in Phase 4{Colors.RESET}")
        print()

    # Bootstrap from Jira
    if from_jira:
        print(f"{Colors.BOLD}Fetching Jira tickets...{Colors.RESET}")
        if not config.adapters.jira.enabled:
            print(f"  {Colors.RED}Jira adapter not enabled in rtmx.yaml{Colors.RESET}")
        elif not config.adapters.jira.project:
            print(f"  {Colors.RED}Jira project not configured in rtmx.yaml{Colors.RESET}")
        else:
            # TODO: Implement Jira ticket fetching
            print(f"  Server: {config.adapters.jira.server}")
            print(f"  Project: {config.adapters.jira.project}")
            print(f"  {Colors.DIM}Not yet implemented - coming in Phase 4{Colors.RESET}")
        print()

    if requirements_to_create:
        print(f"{Colors.BOLD}Requirements to create:{Colors.RESET}")
        for req in requirements_to_create:
            print(f"  {prefix}-{req['id']}: {req['text'][:60]}...")

        if not dry_run:
            # TODO: Write requirements to RTM
            pass
    else:
        print(f"{Colors.DIM}No requirements to create{Colors.RESET}")
