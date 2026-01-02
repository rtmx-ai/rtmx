# REQ-TEST-008: MCP Server E2E Tests

## Status: NOT_STARTED
## Priority: MEDIUM
## Phase: 4
## Effort: 0.75 weeks

## Description

Add E2E tests for MCP server startup, tool invocation, and shutdown.

## Acceptance Criteria

- [ ] E2E test for server startup on available port
- [ ] E2E test for server startup with port conflict
- [ ] E2E test for daemon mode with PID file
- [ ] E2E test for SIGTERM graceful shutdown
- [ ] E2E test for rtmx_status tool invocation
- [ ] E2E test for rtmx_backlog tool invocation
- [ ] E2E test for concurrent client connections
- [ ] At least 8 new scope_system tests

## Test Scenarios

### Server Lifecycle
1. Start server on random available port
2. Handle port already in use
3. Daemon mode creates PID file
4. SIGTERM causes graceful shutdown
5. SIGINT causes graceful shutdown

### Tool Invocation
1. rtmx_status returns valid JSON
2. rtmx_backlog returns requirement list
3. rtmx_get_requirement returns specific requirement
4. Invalid tool name returns error

### Error Handling
1. Server handles malformed requests
2. Concurrent connections don't deadlock

## Files to Create

- `tests/test_mcp_e2e.py`

## Notes

Tests should use subprocess to spawn server and MCP client library to connect.
