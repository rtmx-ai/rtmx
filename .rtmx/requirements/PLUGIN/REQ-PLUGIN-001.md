# REQ-PLUGIN-001: Claude Code Skill Pack (/status, /backlog, /next, /claim, /verify)

## Metadata
- **Category**: PLUGIN
- **Subcategory**: ClaudeCode
- **Priority**: P0
- **Phase**: 19
- **Status**: MISSING
- **Dependencies**: REQ-MCP-003
- **Blocks**: REQ-PLUGIN-004

## Requirement

RTMX shall provide a Claude Code skill pack that registers slash commands (/status, /backlog, /next, /claim, /verify) rendered inline in the conversation TUI. Users shall not need a separate terminal to view RTM state.

## Rationale

The strongest UX pain point is context-switching between the agent conversation and a terminal running rtmx commands. Claude Code supports MCP servers and slash commands. An RTMX skill pack surfaces the roadmap where the work happens.

## Design

### Slash Commands

| Command | Maps to | Display |
|---------|---------|---------|
| /status | rtmx status --json | Compact inline status bar: completion %, category breakdown |
| /backlog | rtmx backlog --json --limit 10 | Top 10 open requirements with priority and effort |
| /next | rtmx next --show --json | Available work webs with stats |
| /claim REQ-XXX | rtmx claim REQ-XXX --agent claude | Claim confirmation with spec path and implementation hints |
| /verify | rtmx verify --dry-run --json | Verification results inline |

### Integration Method

Two options (implement both, user chooses):

1. **MCP Server mode**: RTMX runs as an MCP server. Claude Code discovers tools via MCP. Skills invoke the MCP tools and format the response as markdown.

2. **CLI mode**: Skills shell out to `rtmx <command> --json` and format the response. Simpler, no server process needed.

The skill pack includes a `.claude/skills/rtmx.md` or equivalent that registers the slash commands and their rendering logic.

### Rendering

Each command produces compact, inline-friendly markdown:

/status output:
```
RTMX  82% complete  138/168 reqs
COMPLETE 138  PARTIAL 12  MISSING 18
Next unclaimed: REQ-VERIFY-004 (HIGH, 1wk, unblocked)
```

/backlog output:
```
# Open Requirements (18 remaining)

| # | Req ID | Priority | Effort | Category | Status |
|---|--------|----------|--------|----------|--------|
| 1 | REQ-SYNC-003 | P0 | 2wk | SYNC | MISSING |
| 2 | REQ-MCP-003 | P0 | 1wk | MCP | MISSING |
...
```

### Installation

```
rtmx install --claude-code
```

This command:
1. Detects the Claude Code config directory (~/.claude/ or project .claude/).
2. Writes the MCP server configuration to claude_desktop_config.json or .mcp.json.
3. Writes skill definitions if using skill mode.
4. Confirms installation.

## Acceptance Criteria

1. /status, /backlog, /next, /claim, /verify work as slash commands in Claude Code.
2. Output is rendered inline in the conversation, not requiring a separate terminal.
3. MCP server mode and CLI mode both work.
4. `rtmx install --claude-code` sets up the integration automatically.
5. Output respects the project's no-emoji convention.
6. Commands complete in <2 seconds.
7. Errors produce helpful messages (e.g., "No rtmx.yaml found -- run rtmx init first").

## Files to Create/Modify

- `internal/cmd/install.go` -- Add --claude-code flag
- `internal/cmd/install_test.go` -- Installation tests
- `templates/claude-code/mcp.json` -- MCP server config template
- `templates/claude-code/skills/` -- Skill definitions
- `docs/integrations/claude-code.md` -- Setup guide

## Effort Estimate

2 weeks (skill definitions + MCP integration + installation command)

## Test Strategy

- Unit test: verify skill rendering for each command produces expected markdown
- Unit test: verify MCP server config template generation
- Integration test: `rtmx install --claude-code --dry-run` shows correct output
- Idempotency test: install twice, verify no duplication of config entries
- E2E test: slash commands produce inline output in Claude Code session
