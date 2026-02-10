# REQ-CLAUDE-001: Claude Code Hooks Integration

## Status: NOT_STARTED
## Priority: HIGH
## Phase: 19
## Effort: 2.5 weeks

## Description

RTMX shall provide native Claude Code hooks that automatically inject requirements context into AI coding sessions. The hooks shall leverage Claude Code's `~/.claude/hooks/` system to provide pre-task context, post-task validation, and real-time RTM awareness without requiring explicit user prompts.

## Rationale

Current RTMX integration relies on:
1. Manual `rtmx status` / `rtmx backlog` commands
2. MCP server queries (requires explicit tool calls)
3. Static prompts in CLAUDE.md

Native hooks provide **automatic context injection**, reducing cognitive load and ensuring Claude always has relevant requirements context. This is especially valuable for:
- Large codebases where relevant requirements aren't obvious
- Teams where multiple developers work on interconnected requirements
- Compliance scenarios requiring traceability evidence

## Acceptance Criteria

- [ ] Pre-prompt hook (`PrePromptSubmit`) injects relevant RTM context based on:
  - Files mentioned in the user's prompt
  - Current working directory and recent git changes
  - Active sprint/phase from RTM configuration
- [ ] Post-task hook (`PostToolUse`) validates that code changes align with claimed requirements
- [ ] Context injection is token-efficient (summarized, not full RTM dump)
- [ ] Hook configuration via `rtmx.yaml` allows customization:
  - Enable/disable specific hooks
  - Context verbosity level (minimal/standard/verbose)
  - Requirement filtering by phase, priority, or category
- [ ] `rtmx install --hooks --claude` installs Claude Code hooks
- [ ] Hooks gracefully degrade when RTMX is not configured in project
- [ ] Performance: hook execution < 100ms to avoid UX lag

## Technical Notes

### Claude Code Hook System

Claude Code hooks are shell scripts in `~/.claude/hooks/`:

```
~/.claude/hooks/
├── PrePromptSubmit.sh    # Before user prompt is sent
├── PostToolUse.sh        # After each tool execution
├── Stop.sh               # When conversation ends
└── SubagentSpawn.sh      # When spawning subagents
```

### Hook Implementation

```bash
#!/bin/bash
# ~/.claude/hooks/PrePromptSubmit.sh
# RTMX context injection hook for Claude Code

# Check if rtmx is available and project has rtmx.yaml
if command -v rtmx &> /dev/null && [ -f "rtmx.yaml" ]; then
    # Get compact context (JSON for parsing)
    CONTEXT=$(rtmx context --format json --compact 2>/dev/null)

    if [ -n "$CONTEXT" ]; then
        echo "<rtmx-context>"
        echo "$CONTEXT"
        echo "</rtmx-context>"
    fi
fi
```

### New CLI Command: `rtmx context`

```bash
rtmx context [OPTIONS]

Options:
  --format [text|json|markdown]  Output format (default: text)
  --compact                      Minimal token-efficient output
  --files FILE...               Focus on requirements related to specific files
  --phase N                      Filter to specific phase
  --verbose                      Include full requirement text
```

Output structure (JSON):
```json
{
  "project": "rtmx",
  "completion": 52.8,
  "active_phase": 10,
  "current_sprint": "v0.1.0",
  "relevant_requirements": [
    {
      "id": "REQ-COLLAB-001",
      "summary": "Cross-repo dependency tracking",
      "status": "MISSING",
      "blocks": ["REQ-COLLAB-002", "REQ-COLLAB-003"]
    }
  ],
  "blockers": ["REQ-ZT-001"],
  "quick_wins": ["REQ-GIT-003"]
}
```

### File-to-Requirement Mapping

The hook should map modified files to relevant requirements:

1. Parse `test_module` and `test_function` columns from RTM
2. Check `requirement_file` paths
3. Use directory heuristics (e.g., `src/auth/` → AUTH requirements)
4. Cache mapping for performance

## Gherkin Specification

```gherkin
@REQ-CLAUDE-001 @scope_system @technique_nominal
Feature: Claude Code Hooks Integration
  As a developer using Claude Code with RTMX
  I want automatic requirements context injection
  So that Claude understands my project's requirements without explicit commands

  Background:
    Given an RTMX-enabled project with rtmx.yaml
    And Claude Code hooks are installed

  @happy-path
  Scenario: Pre-prompt hook injects relevant context
    Given the RTM database has 50 requirements
    And 10 requirements are related to "src/auth/"
    When I ask Claude to "fix the login bug in src/auth/login.py"
    Then the PrePromptSubmit hook executes
    And Claude receives context about the 10 auth-related requirements
    And the context includes current completion status
    And the context is under 500 tokens

  @filtering
  Scenario: Context respects phase filtering
    Given rtmx.yaml has "hooks.context_filter.phase: 10"
    And requirements exist in phases 1-18
    When the PrePromptSubmit hook executes
    Then only Phase 10 requirements are included in context

  @performance
  Scenario: Hook execution is fast
    Given a large RTM database with 500 requirements
    When the PrePromptSubmit hook executes
    Then execution completes in under 100ms

  @graceful-degradation
  Scenario: Hook handles missing RTMX gracefully
    Given a project without rtmx.yaml
    When the PrePromptSubmit hook executes
    Then no error is raised
    And no context is injected

  @post-task-validation
  Scenario: Post-task hook validates requirement claims
    Given a test file contains "@pytest.mark.req('REQ-AUTH-001')"
    When Claude modifies the test file
    And the PostToolUse hook executes
    Then the hook verifies REQ-AUTH-001 exists in the RTM
    And warns if the requirement is already COMPLETE
```

## Test Cases

1. `tests/test_claude_hooks.py::test_preprompt_hook_generates_context`
2. `tests/test_claude_hooks.py::test_context_command_json_output`
3. `tests/test_claude_hooks.py::test_context_command_file_filtering`
4. `tests/test_claude_hooks.py::test_context_token_efficiency`
5. `tests/test_claude_hooks.py::test_hook_installation`
6. `tests/test_claude_hooks.py::test_hook_graceful_degradation`
7. `tests/test_claude_hooks.py::test_posttool_hook_validation`
8. `tests/test_claude_hooks.py::test_hook_performance_under_100ms`

## Files to Create/Modify

- `src/rtmx/cli/context.py` (new) - Context generation command
- `src/rtmx/cli/main.py` - Register context command
- `src/rtmx/cli/install.py` - Add `--claude` option for hook installation
- `src/rtmx/hooks/` (new directory) - Hook script templates
- `src/rtmx/hooks/claude_preprompt.sh` - Pre-prompt hook template
- `src/rtmx/hooks/claude_posttool.sh` - Post-tool hook template
- `tests/test_claude_hooks.py` (new) - Hook tests

## Dependencies

- REQ-MCP-001: MCP server (COMPLETE) - Shares context generation logic

## Blocks

- REQ-CLAUDE-002: Native Claude Code skills
- REQ-CLAUDE-003: Claude Cowork RTM sharing
