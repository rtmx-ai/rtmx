# REQ-MCP-003: Production-Grade MCP Server for Read-Only RTM Operations

## Metadata
- **Category**: MCP
- **Subcategory**: Server
- **Priority**: P0
- **Phase**: 19
- **Status**: MISSING
- **Dependencies**: (none)
- **Blocks**: REQ-MCP-005, REQ-PLUGIN-001, REQ-PLUGIN-002, REQ-PLUGIN-003

## Requirement

The RTMX MCP server shall expose production-grade, read-only tools for RTM operations (status, backlog, deps, health, next --show) that any MCP-compatible agent (Claude Code, Cursor, gemini-cli) can discover and invoke. The server must handle concurrent requests, validate inputs, and return structured JSON responses.

## Rationale

The MCP server is the universal integration point. Rather than building bespoke plugins for each agent platform, a robust MCP server lets any MCP-speaking client get RTMX for free. The existing server in internal/adapters/mcp/ needs hardening for production use: better error handling, input validation, concurrent safety, and comprehensive tool coverage.

## Design

### Tools to Expose

| Tool Name | Description | Parameters | Returns |
|-----------|-------------|------------|---------|
| rtmx_status | RTM completion summary | --fail-under (int, optional) | Status JSON with total, complete, partial, missing, percentage |
| rtmx_backlog | Prioritized list of open requirements | --category (string, optional), --limit (int, optional) | Array of requirement objects with priority, effort, dependencies |
| rtmx_deps | Dependency graph for a requirement | req_id (string, required) | Upstream and downstream dependencies with status |
| rtmx_health | Project health assessment | (none) | Health status, warnings, scores |
| rtmx_next_show | Available work webs | --category (string, optional), --max-effort (string, optional) | Web array from REQ-AGENT-002 |
| rtmx_requirement | Read a single requirement spec | req_id (string, required) | Full requirement text and metadata |
| rtmx_claims | Active claims | (none) | Claims array from REQ-AGENT-005 |

### Server Architecture

- stdio transport (default for Claude Code / Cursor)
- HTTP/SSE transport (for remote agents, Gastown)
- Request validation with JSON Schema per tool
- Concurrent request handling with read locks on database
- Graceful error responses (not panics)

### Tool Registration

Each tool is registered with a JSON Schema describing its parameters and return type, following the MCP tool specification. Agents discover available tools via the standard MCP `tools/list` method.

## Acceptance Criteria

1. All 7 tools are registered and discoverable via MCP tools/list.
2. Each tool validates inputs against its JSON Schema.
3. Concurrent requests do not cause data races (verified with -race flag).
4. Error responses include actionable messages.
5. stdio and HTTP/SSE transports both work.
6. Tools return identical data to their CLI equivalents with --json.
7. Server starts in <100ms.
8. Integration test: Claude Code can discover and invoke all tools.

## Files to Create/Modify

- `internal/adapters/mcp/server.go` -- Harden existing server
- `internal/adapters/mcp/tools.go` -- Tool registration and schemas
- `internal/adapters/mcp/tools_test.go` -- Per-tool tests
- `internal/adapters/mcp/server_test.go` -- Concurrency tests

## Test Strategy

- Unit test per tool: verify JSON Schema validation and correct output
- Concurrency tests: parallel tool invocations with -race flag
- Golden file tests: tool outputs match CLI --json equivalents
- Transport tests: verify both stdio and HTTP/SSE transports
- Error handling tests: invalid inputs, missing requirements, malformed requests
