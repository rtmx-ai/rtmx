# REQ-PLUGIN-003a: Cursor IDE Integration via MCP

## Metadata
- **Category**: PLUGIN
- **Subcategory**: Cursor
- **Priority**: HIGH
- **Phase**: 20
- **Status**: MISSING
- **Dependencies**: REQ-MCP-003
- **Blocks**: (none)

## Requirement
`rtmx install --cursor` shall write MCP server configuration to .cursor/mcp.json so Cursor discovers RTMX tools automatically.

## Acceptance Criteria
1. rtmx install --cursor writes valid .cursor/mcp.json.
2. Cursor discovers all RTMX MCP tools.
3. Idempotent (safe to run multiple times).

## Files to Create/Modify
- internal/cmd/install.go
- templates/cursor/mcp.json
