# REQ-CLAUDE-002: Native Claude Code Skills

## Status: NOT_STARTED
## Priority: HIGH
## Phase: 19
## Effort: 2.0 weeks

## Description

RTMX shall provide native Claude Code skills (slash commands) that enable quick RTM interactions without leaving the conversational flow. Skills provide a faster, more discoverable interface than MCP tools for common operations.

## Rationale

While MCP tools are powerful, they require Claude to decide when to use them. Skills provide:
- **Explicit user intent** - User types `/rtmx status`, Claude executes immediately
- **Discoverability** - Skills appear in Claude Code's `/` menu
- **Speed** - Direct execution without tool-use reasoning overhead

## Acceptance Criteria

- [ ] `/rtmx` skill group registered with Claude Code
- [ ] `/rtmx status` shows completion summary inline
- [ ] `/rtmx backlog` shows top 5 prioritized items
- [ ] `/rtmx req <ID>` shows full requirement details
- [ ] `/rtmx claim <ID>` marks requirement as in-progress with current user
- [ ] `/rtmx verify` runs tests and shows requirement coverage
- [ ] Skills work in both Claude Code CLI and Claude Code in IDEs
- [ ] Skill output is formatted for terminal display (colors, tables)
- [ ] Skills respect rtmx.yaml configuration
- [ ] `rtmx install --skills` registers skills with Claude Code

## Technical Notes

### Claude Code Skills Registration

Skills are registered via `~/.claude/skills/` directory:

```
~/.claude/skills/
└── rtmx/
    ├── manifest.json       # Skill metadata and commands
    ├── status.sh           # /rtmx status implementation
    ├── backlog.sh          # /rtmx backlog implementation
    ├── req.sh              # /rtmx req <ID> implementation
    ├── claim.sh            # /rtmx claim <ID> implementation
    └── verify.sh           # /rtmx verify implementation
```

### Manifest Structure

```json
{
  "name": "rtmx",
  "description": "Requirements Traceability Matrix toolkit",
  "version": "0.1.0",
  "commands": [
    {
      "name": "status",
      "description": "Show RTM completion status",
      "usage": "/rtmx status [-v|-vv|-vvv]"
    },
    {
      "name": "backlog",
      "description": "Show prioritized backlog",
      "usage": "/rtmx backlog [--limit N] [--phase N]"
    },
    {
      "name": "req",
      "description": "Show requirement details",
      "usage": "/rtmx req <REQ-ID>"
    },
    {
      "name": "claim",
      "description": "Claim requirement for current session",
      "usage": "/rtmx claim <REQ-ID>"
    },
    {
      "name": "verify",
      "description": "Run tests and update RTM status",
      "usage": "/rtmx verify [--update]"
    }
  ]
}
```

### Skill Scripts

Each skill is a shell script that invokes rtmx:

```bash
#!/bin/bash
# ~/.claude/skills/rtmx/status.sh
exec rtmx status "$@"
```

## Gherkin Specification

```gherkin
@REQ-CLAUDE-002 @scope_system @technique_nominal
Feature: Native Claude Code Skills
  As a developer using Claude Code
  I want slash commands for common RTM operations
  So that I can quickly check status without verbose prompts

  Background:
    Given RTMX skills are installed in Claude Code
    And an RTMX-enabled project is open

  Scenario: /rtmx status shows completion
    When I type "/rtmx status"
    Then Claude executes the rtmx status command
    And displays the completion percentage
    And shows phase breakdown

  Scenario: /rtmx backlog shows priorities
    When I type "/rtmx backlog"
    Then Claude shows the top 5 prioritized requirements
    And each item shows ID, description, effort, and blockers

  Scenario: /rtmx req shows details
    When I type "/rtmx req REQ-AUTH-001"
    Then Claude shows the full requirement specification
    And includes acceptance criteria
    And shows linked tests

  Scenario: /rtmx claim assigns work
    Given REQ-AUTH-001 has no assignee
    When I type "/rtmx claim REQ-AUTH-001"
    Then the requirement is marked with current user
    And status changes to "in progress" in RTM

  Scenario: Skills appear in autocomplete
    When I type "/" in Claude Code
    Then "rtmx" appears in the skills menu
    And subcommands are listed
```

## Test Cases

1. `tests/test_claude_skills.py::test_skill_manifest_valid_json`
2. `tests/test_claude_skills.py::test_skill_status_execution`
3. `tests/test_claude_skills.py::test_skill_backlog_execution`
4. `tests/test_claude_skills.py::test_skill_req_lookup`
5. `tests/test_claude_skills.py::test_skill_claim_updates_rtm`
6. `tests/test_claude_skills.py::test_skill_installation`

## Files to Create/Modify

- `src/rtmx/skills/` (new directory) - Skill templates
- `src/rtmx/skills/manifest.json` - Skill metadata
- `src/rtmx/skills/*.sh` - Skill scripts
- `src/rtmx/cli/install.py` - Add `--skills` option
- `tests/test_claude_skills.py` (new) - Skill tests

## Dependencies

- REQ-CLAUDE-001: Claude Code hooks (shares installation logic)

## Blocks

- None (leaf requirement)
