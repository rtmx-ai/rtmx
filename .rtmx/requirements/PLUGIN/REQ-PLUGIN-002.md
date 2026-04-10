# REQ-PLUGIN-002: gemini-cli Extension for RTMX

## Metadata
- **Category**: PLUGIN
- **Subcategory**: GeminiCLI
- **Priority**: HIGH
- **Phase**: 20
- **Status**: MISSING
- **Dependencies**: REQ-MCP-003
- **Blocks**: (none)

## Requirement

RTMX shall provide a gemini-cli extension that surfaces RTM status, backlog, and agent coordination commands inline in the Gemini CLI TUI, using the same underlying MCP server or CLI integration as the Claude Code skill pack.

## Rationale

gemini-cli recently added extension support. As a competing agent CLI, first-class RTMX integration expands the user base and validates the platform-agnostic design. The integration should reuse the MCP server (REQ-MCP-003) so both platforms get identical data.

## Design

### Extension Registration

gemini-cli extensions are registered via configuration. The RTMX extension:
1. Registers as an MCP tool provider (if gemini-cli supports MCP).
2. Or registers as a CLI extension that shells out to rtmx commands.

### Commands

Same surface as the Claude Code skill pack:
- Status, backlog, next, claim, verify
- Rendered inline in the gemini-cli conversation

### Installation

```
rtmx install --gemini-cli
```

Detects gemini-cli config directory and writes the extension configuration.

### Research Required

Before implementation, investigate:
1. Does gemini-cli support MCP natively? (Check their extension API docs)
2. What is the extension registration format?
3. How does gemini-cli render tool responses? (Markdown? Structured data?)
4. Are there gemini-cli extension examples to follow?

## Acceptance Criteria

1. RTMX commands work inline in gemini-cli.
2. Installation via `rtmx install --gemini-cli`.
3. Uses MCP server if gemini-cli supports it, CLI fallback otherwise.
4. Output formatting matches gemini-cli conventions.
5. Same data as Claude Code integration (both hit the same MCP server or CLI).

## Files to Create/Modify

- `internal/cmd/install.go` -- Add --gemini-cli flag
- `templates/gemini-cli/` -- Extension configuration templates
- `docs/integrations/gemini-cli.md` -- Setup guide

## Test Strategy

- Unit tests for install command with --gemini-cli flag
- Integration test: verify extension configuration is written correctly
- Golden file tests: extension config output matches expected format
- Manual validation: invoke RTMX tools from within gemini-cli session
