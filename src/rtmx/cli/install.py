"""RTMX install command.

Inject RTM-aware prompts into AI agent configurations.
"""

from __future__ import annotations

import shutil
from datetime import datetime
from pathlib import Path

from rtmx.config import RTMXConfig
from rtmx.formatting import Colors

# Agent prompt templates (to be moved to Jinja2 templates in Phase 3)
CLAUDE_PROMPT = """
## RTMX Requirements Traceability

This project uses RTMX for requirements traceability management.

### Quick Commands
- `rtmx status` - Completion status (-v/-vv/-vvv for detail)
- `rtmx backlog` - Prioritized incomplete requirements
- `rtmx backlog --phase N` - Filter by phase
- `make rtm` / `make backlog` - Makefile shortcuts (if available)

### Development Workflow
1. Check if requirement exists before implementing (`rtmx status -vvv | grep "feature"`)
2. Link tests with `@pytest.mark.req("REQ-XX-NNN")`
3. Update RTM status when completing features
4. Run `rtmx from-tests --update` to sync test info

### Test Markers
| Marker | Purpose |
|--------|---------|
| `@pytest.mark.req("ID")` | Link to requirement |
| `@pytest.mark.scope_unit` | Single component |
| `@pytest.mark.scope_integration` | Multi-component |
| `@pytest.mark.technique_nominal` | Happy path |
| `@pytest.mark.technique_stress` | Edge cases |
"""

CURSOR_PROMPT = """# RTMX Requirements Traceability

## Context Commands
- rtmx status -v      # Category-level completion
- rtmx backlog        # What needs work
- rtmx deps --req ID  # Requirement dependencies

## Test Generation Rules
When generating tests, add @pytest.mark.req("REQ-XX-NNN") markers.
Include scope markers (scope_unit, scope_integration, scope_system).
Reference: docs/requirements/ for requirement details.
"""

COPILOT_PROMPT = """# RTMX Requirements Traceability

This project uses RTMX for requirements traceability.

## Test Markers
- @pytest.mark.req("REQ-XX-NNN") - Links test to requirement
- @pytest.mark.scope_unit/integration/system - Test scope

## Commands
- rtmx status - Check completion status
- rtmx backlog - See incomplete requirements
"""


def detect_agent_configs(cwd: Path) -> dict[str, Path | None]:
    """Detect existing AI agent configuration files.

    Args:
        cwd: Current working directory

    Returns:
        Dictionary mapping agent name to config path (None if not found)
    """
    agents: dict[str, Path | None] = {}

    # Claude Code
    claude_paths = [
        cwd / "CLAUDE.md",
        cwd / ".claude" / "CLAUDE.md",
    ]
    agents["claude"] = next((p for p in claude_paths if p.exists()), None)

    # Cursor
    cursor_path = cwd / ".cursorrules"
    agents["cursor"] = cursor_path if cursor_path.exists() else None

    # GitHub Copilot
    copilot_path = cwd / ".github" / "copilot-instructions.md"
    agents["copilot"] = copilot_path if copilot_path.exists() else None

    return agents


def get_agent_prompt(agent: str) -> str:
    """Get the prompt template for an agent.

    Args:
        agent: Agent name (claude, cursor, copilot)

    Returns:
        Prompt template string
    """
    prompts = {
        "claude": CLAUDE_PROMPT,
        "cursor": CURSOR_PROMPT,
        "copilot": COPILOT_PROMPT,
    }
    return prompts.get(agent, "")


def run_install(
    dry_run: bool,
    non_interactive: bool,
    force: bool,
    agents: list[str] | None,
    install_all: bool,
    skip_backup: bool,
    _config: RTMXConfig,
) -> None:
    """Run install command.

    Inject RTM-aware prompts into AI agent configurations.

    Args:
        dry_run: Preview changes without writing
        non_interactive: Skip confirmation prompts
        force: Overwrite existing rtmx sections
        agents: List of agents to install (None = detect)
        install_all: Install to all detected agents
        skip_backup: Don't backup modified files
        config: RTMX configuration
    """
    cwd = Path.cwd()
    detected = detect_agent_configs(cwd)

    print("=== RTMX Agent Installation ===")
    print()

    if dry_run:
        print(f"{Colors.YELLOW}DRY RUN - no files will be written{Colors.RESET}")
        print()

    # Show detected configs
    print(f"{Colors.BOLD}Detected agent configurations:{Colors.RESET}")
    for agent, path in detected.items():
        if path:
            print(f"  {Colors.GREEN}✓{Colors.RESET} {agent}: {path}")
        else:
            print(f"  {Colors.DIM}○ {agent}: not found{Colors.RESET}")
    print()

    # Determine which agents to install
    if agents:
        target_agents = agents
    elif install_all:
        target_agents = list(detected.keys())
    else:
        # Interactive selection
        if non_interactive:
            target_agents = [a for a, p in detected.items() if p]
        else:
            print(f"{Colors.BOLD}Select agents to configure:{Colors.RESET}")
            target_agents = []
            for agent, path in detected.items():
                if path:
                    response = input(f"  Install to {agent} ({path})? [Y/n]: ").strip().lower()
                    if response in ("", "y", "yes"):
                        target_agents.append(agent)
                else:
                    response = input(f"  Create {agent} config? [y/N]: ").strip().lower()
                    if response in ("y", "yes"):
                        target_agents.append(agent)
            print()

    if not target_agents:
        print(f"{Colors.YELLOW}No agents selected{Colors.RESET}")
        return

    # Install to each agent
    for agent in target_agents:
        print(f"{Colors.BOLD}Installing to {agent}...{Colors.RESET}")

        path = detected.get(agent)
        prompt = get_agent_prompt(agent)

        if path and path.exists():
            # Check if rtmx section already exists
            content = path.read_text()
            if "RTMX Requirements Traceability" in content and not force:
                print(f"  {Colors.YELLOW}RTMX section already exists (use --force to overwrite){Colors.RESET}")
                continue

            if not skip_backup and not dry_run:
                # Create backup
                timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
                backup_path = path.with_suffix(f".rtmx-backup-{timestamp}{path.suffix}")
                shutil.copy2(path, backup_path)
                print(f"  {Colors.DIM}Backup: {backup_path}{Colors.RESET}")

            # Append rtmx section
            if force and "RTMX Requirements Traceability" in content:
                # Remove existing section
                lines = content.split("\n")
                new_lines = []
                in_rtmx_section = False
                for line in lines:
                    if "## RTMX Requirements Traceability" in line or "# RTMX Requirements Traceability" in line:
                        in_rtmx_section = True
                        continue
                    if in_rtmx_section and line.startswith("## "):
                        in_rtmx_section = False
                    if not in_rtmx_section:
                        new_lines.append(line)
                content = "\n".join(new_lines)

            new_content = content.rstrip() + "\n" + prompt.strip() + "\n"

            if dry_run:
                print(f"  Would append {len(prompt)} characters")
            else:
                path.write_text(new_content)
                print(f"  {Colors.GREEN}✓{Colors.RESET} Updated {path}")
        else:
            # Create new file
            if agent == "claude":
                new_path = cwd / "CLAUDE.md"
            elif agent == "cursor":
                new_path = cwd / ".cursorrules"
            elif agent == "copilot":
                new_path = cwd / ".github" / "copilot-instructions.md"
                new_path.parent.mkdir(parents=True, exist_ok=True)
            else:
                print(f"  {Colors.RED}Unknown agent: {agent}{Colors.RESET}")
                continue

            if dry_run:
                print(f"  Would create {new_path}")
            else:
                new_path.write_text(prompt.strip() + "\n")
                print(f"  {Colors.GREEN}✓{Colors.RESET} Created {new_path}")

    print()
    print(f"{Colors.GREEN}✓ Installation complete{Colors.RESET}")
