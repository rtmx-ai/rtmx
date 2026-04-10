# REQ-PLUGIN-003b: Coder Workspace Integration via HTTP MCP

## Metadata
- **Category**: PLUGIN
- **Subcategory**: Coder
- **Priority**: MEDIUM
- **Phase**: 21
- **Status**: MISSING
- **Dependencies**: REQ-MCP-003, REQ-MCP-005
- **Blocks**: (none)

## Requirement
`rtmx install --coder` shall generate a workspace template that starts the RTMX MCP server on HTTP/SSE for remote agent access.

## Acceptance Criteria
1. rtmx install --coder generates workspace setup script.
2. MCP server starts on workspace creation.
3. HTTP/SSE transport accessible from Coder agents.

## Files to Create/Modify
- internal/cmd/install.go
- templates/coder/
