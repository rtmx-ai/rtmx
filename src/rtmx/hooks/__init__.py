"""Claude Code hooks for RTMX integration.

REQ-CLAUDE-001: Claude Code Hooks Integration

Provides automatic requirements context injection into Claude Code sessions.
"""

from __future__ import annotations

from textwrap import dedent


def generate_preprompt_hook() -> str:
    """Generate the PrePromptSubmit hook script.

    This hook injects RTMX requirements context before each user prompt.

    Returns:
        Bash script content for the hook.
    """
    return dedent("""
        #!/bin/bash
        # RTMX Pre-Prompt Hook for Claude Code
        # REQ-CLAUDE-001: Automatic requirements context injection
        #
        # This hook runs before each user prompt is sent to Claude.
        # It injects relevant RTM context to help Claude understand
        # the project's requirements without explicit commands.

        # Check if rtmx is available
        if ! command -v rtmx &> /dev/null; then
            exit 0
        fi

        # Check if project has RTMX configuration
        if [ ! -f "rtmx.yaml" ] && [ ! -f ".rtmx/config.yaml" ]; then
            exit 0
        fi

        # Generate compact context (suppress errors)
        CONTEXT=$(rtmx context --format json --compact 2>/dev/null)

        # Only output if we got valid context
        if [ -n "$CONTEXT" ] && [ "$CONTEXT" != "{}" ]; then
            echo "<rtmx-context>"
            echo "$CONTEXT"
            echo "</rtmx-context>"
        fi
    """).strip()


def generate_posttool_hook() -> str:
    """Generate the PostToolUse hook script.

    This hook validates requirement markers after code changes.

    Returns:
        Bash script content for the hook.
    """
    return dedent("""
        #!/bin/bash
        # RTMX Post-Tool Hook for Claude Code
        # REQ-CLAUDE-001: Requirement validation after code changes
        #
        # This hook runs after Claude uses a tool (like Edit or Write).
        # It validates that any requirement markers are valid.

        # Get the tool name and result from environment
        TOOL_NAME="${CLAUDE_TOOL_NAME:-}"
        TOOL_INPUT="${CLAUDE_TOOL_INPUT:-}"

        # Only process file modification tools
        case "$TOOL_NAME" in
            Edit|Write|NotebookEdit)
                ;;
            *)
                exit 0
                ;;
        esac

        # Check if rtmx is available
        if ! command -v rtmx &> /dev/null; then
            exit 0
        fi

        # Check if project has RTMX configuration
        if [ ! -f "rtmx.yaml" ] && [ ! -f ".rtmx/config.yaml" ]; then
            exit 0
        fi

        # If the tool input contains requirement markers, validate them
        if echo "$TOOL_INPUT" | grep -qE '@pytest\\.mark\\.req|rtmx\\.Req|@REQ-[A-Z]+-[0-9]+'; then
            # Extract requirement IDs
            REQ_IDS=$(echo "$TOOL_INPUT" | grep -oE 'REQ-[A-Z]+-[0-9]+' | sort -u)

            if [ -n "$REQ_IDS" ]; then
                # Validate each requirement exists
                for REQ_ID in $REQ_IDS; do
                    if ! rtmx validate --req "$REQ_ID" --quiet 2>/dev/null; then
                        echo "<rtmx-warning>"
                        echo "Requirement $REQ_ID referenced but not found in RTM database."
                        echo "Consider adding it with: rtmx bootstrap --from-tests"
                        echo "</rtmx-warning>"
                    fi
                done
            fi
        fi
    """).strip()


def generate_stop_hook() -> str:
    """Generate the Stop hook script.

    This hook runs when a Claude Code conversation ends.

    Returns:
        Bash script content for the hook.
    """
    return dedent("""
        #!/bin/bash
        # RTMX Stop Hook for Claude Code
        # REQ-CLAUDE-001: Session summary on conversation end
        #
        # This hook runs when the conversation ends.
        # It can log session activity or update RTM status.

        # Check if rtmx is available
        if ! command -v rtmx &> /dev/null; then
            exit 0
        fi

        # Check if project has RTMX configuration
        if [ ! -f "rtmx.yaml" ] && [ ! -f ".rtmx/config.yaml" ]; then
            exit 0
        fi

        # Optional: Log session end or sync changes
        # rtmx sync --auto 2>/dev/null || true
    """).strip()


def install_claude_hooks(
    hooks_dir: str | None = None,
    dry_run: bool = False,
    force: bool = False,
) -> dict[str, str]:
    """Install RTMX hooks for Claude Code.

    Args:
        hooks_dir: Custom hooks directory (default: ~/.claude/hooks)
        dry_run: Preview without writing files
        force: Overwrite existing hooks

    Returns:
        Dictionary mapping hook names to their paths.
    """
    import os
    from pathlib import Path

    if hooks_dir is None:
        home = os.environ.get("HOME", os.path.expanduser("~"))
        hooks_dir = os.path.join(home, ".claude", "hooks")

    hooks_path = Path(hooks_dir)
    installed = {}

    hooks = {
        "PrePromptSubmit.sh": generate_preprompt_hook(),
        "PostToolUse.sh": generate_posttool_hook(),
        "Stop.sh": generate_stop_hook(),
    }

    for name, content in hooks.items():
        hook_path = hooks_path / name

        if hook_path.exists() and not force:
            # Check if it's already an RTMX hook
            existing = hook_path.read_text()
            if "RTMX" not in existing:
                continue  # Skip non-RTMX hooks

        if not dry_run:
            hooks_path.mkdir(parents=True, exist_ok=True)
            hook_path.write_text(content)
            # Make executable
            hook_path.chmod(0o755)

        installed[name] = str(hook_path)

    return installed


def uninstall_claude_hooks(hooks_dir: str | None = None) -> list[str]:
    """Uninstall RTMX hooks for Claude Code.

    Args:
        hooks_dir: Custom hooks directory (default: ~/.claude/hooks)

    Returns:
        List of removed hook paths.
    """
    import os
    from pathlib import Path

    if hooks_dir is None:
        home = os.environ.get("HOME", os.path.expanduser("~"))
        hooks_dir = os.path.join(home, ".claude", "hooks")

    hooks_path = Path(hooks_dir)
    removed = []

    hook_names = ["PrePromptSubmit.sh", "PostToolUse.sh", "Stop.sh"]

    for name in hook_names:
        hook_path = hooks_path / name
        if hook_path.exists():
            content = hook_path.read_text()
            if "RTMX" in content:
                hook_path.unlink()
                removed.append(str(hook_path))

    return removed
