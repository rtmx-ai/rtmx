# REQ-MCP-006: Stdio Transport for MCP Server

## Status: COMPLETE
## Priority: P0
## Phase: 26

## Requirement

RTMX MCP server shall support stdio transport for seamless integration
with Claude Code, Cursor, and other MCP clients that use the standard
stdin/stdout protocol.

## Rationale

MCP clients (Claude Code, Cursor, Codex CLI) spawn local MCP servers
as subprocesses and communicate via JSON-RPC 2.0 over stdin/stdout.
Without stdio support, users must manage an HTTP server process
separately, adding friction to onboarding and preventing standard
one-command setup workflows.

## Acceptance Criteria

1. `rtmx mcp-server --stdio` reads JSON-RPC requests from stdin (one per line)
   and writes responses to stdout (one per line)
2. All 10 tools (7 read + 3 mutation) work identically over stdio and HTTP
3. Notifications (e.g., `notifications/initialized`) produce no response
4. Empty lines in input are silently ignored
5. Parse errors return proper JSON-RPC error responses
6. Diagnostic output goes to stderr, keeping stdout clean for the protocol
7. `claude mcp add rtmx -- rtmx mcp-server --stdio` integrates with Claude Code
8. `.cursor/mcp.json` config with `["mcp-server", "--stdio"]` works with Cursor

## Dependencies

- REQ-MCP-003: Read-only tools (provides the tool implementations)
- REQ-MCP-005: Mutation tools (provides claim/release/assign)

## Test

- `internal/adapters/mcp/server_test.go::TestMCPStdio`
- Subtests: initialize_and_tools_list, tools_call_via_stdio,
  notification_no_response, empty_lines_ignored, parse_error

## Files Modified

- `internal/adapters/mcp/server.go` -- Added StartStdio(), processRPC(), refactored handleRPC
- `internal/cmd/mcp.go` -- Added --stdio flag and stdio execution path
- `internal/adapters/mcp/server_test.go` -- Added TestMCPStdio with 5 subtests
