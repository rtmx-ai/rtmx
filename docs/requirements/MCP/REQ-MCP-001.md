# REQ-MCP-001: Claude Code MCP Plugin for RTMX Commands

## Status: MISSING
## Priority: HIGH
## Phase: 11

## Description
RTMX shall provide a Model Context Protocol (MCP) server that integrates with Claude Code, enabling users to query RTM status, view requirement specifications, and manage requirements directly from within Claude Code sessions.

## Acceptance Criteria
- [ ] MCP server is installable via `pip install rtmx[mcp]`
- [ ] Claude Code configuration documented in README with `.mcp.json` example
- [ ] `rtmx_get_spec` tool reads full requirement specification markdown files
- [ ] `rtmx_status` tool returns RTM completion summary
- [ ] `rtmx_backlog` tool returns prioritized backlog items
- [ ] `rtmx_get_requirement` tool returns database fields for a requirement
- [ ] `rtmx_search` tool searches requirements by text
- [ ] `rtmx_deps` tool shows dependency graph for a requirement
- [ ] All tools return well-formatted JSON that Claude can present clearly
- [ ] Error messages are user-friendly and actionable

## Technical Notes
- MCP server implemented in `src/rtmx/adapters/mcp/`
- Server runs via `python -m rtmx.mcp` or `rtmx mcp serve`
- Configuration auto-discovery from `rtmx.yaml` or `.rtmx/config.yaml`
- New `rtmx_get_spec` tool reads `requirement_file` path and returns markdown content
- Tools should cache database to avoid repeated disk reads within a session

## Claude Code Configuration

Add to `.mcp.json` in project root:

```json
{
  "mcpServers": {
    "rtmx": {
      "command": "python",
      "args": ["-m", "rtmx.mcp"],
      "env": {}
    }
  }
}
```

Or with uvx for isolated environment:

```json
{
  "mcpServers": {
    "rtmx": {
      "command": "uvx",
      "args": ["rtmx", "mcp", "serve"],
      "env": {}
    }
  }
}
```

## MCP Tools

| Tool | Description | Example Use |
|------|-------------|-------------|
| `rtmx_status` | Get RTM completion summary | "What's the project status?" |
| `rtmx_backlog` | Get prioritized backlog | "What should I work on next?" |
| `rtmx_get_requirement` | Get requirement database fields | "Show me REQ-CLI-001 details" |
| `rtmx_get_spec` | Read full specification file | "Show me the spec for REQ-LANG-007" |
| `rtmx_search` | Search requirements by text | "Find requirements about authentication" |
| `rtmx_deps` | Show dependency relationships | "What does REQ-LANG-007 block?" |
| `rtmx_update_status` | Update requirement status | "Mark REQ-CLI-001 as complete" |

## Test Cases
1. `tests/test_mcp.py::test_mcp_server_startup` - Server starts without errors
2. `tests/test_mcp.py::test_rtmx_status_tool` - Status tool returns valid JSON
3. `tests/test_mcp.py::test_rtmx_backlog_tool` - Backlog tool filters correctly
4. `tests/test_mcp.py::test_rtmx_get_spec_tool` - Spec tool reads markdown file
5. `tests/test_mcp.py::test_rtmx_get_spec_missing_file` - Graceful error for missing spec
6. `tests/test_mcp.py::test_rtmx_search_tool` - Search returns relevant results
7. `tests/test_mcp.py::test_mcp_integration_claude_code` - End-to-end with Claude Code mock

## Dependencies
- None

## Blocks
- None (enables AI-assisted RTM workflows)

## Effort
1.5 weeks

## Notes
The MCP server infrastructure already exists in `src/rtmx/adapters/mcp/`. This requirement focuses on:
1. Adding the `rtmx_get_spec` tool to read full specification files
2. Documenting Claude Code integration setup
3. Ensuring all tools are well-tested and documented
