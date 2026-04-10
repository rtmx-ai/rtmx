# REQ-PLUGIN-003c: OpenAI Codex Integration via CLI

## Metadata
- **Category**: PLUGIN
- **Subcategory**: Codex
- **Priority**: MEDIUM
- **Phase**: 21
- **Status**: MISSING
- **Dependencies**: REQ-MCP-003
- **Blocks**: (none)

## Requirement
`rtmx install --codex` shall generate a Codex-compatible tool definition enabling Codex to invoke rtmx --json commands.

## Acceptance Criteria
1. rtmx install --codex generates tool definition.
2. Codex can discover and invoke rtmx commands.
3. All commands use --json for structured output.

## Files to Create/Modify
- internal/cmd/install.go
- templates/codex/
