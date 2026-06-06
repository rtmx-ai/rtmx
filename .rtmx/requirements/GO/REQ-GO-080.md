# REQ-GO-080: MCP Server Command Tests

## Metadata
- **Category**: GO
- **Subcategory**: CLI
- **Priority**: HIGH
- **Phase**: 29
- **Status**: MISSING
- **Dependencies**: REQ-MCP-001
- **Blocks**: (none)

## Requirement

The `mcp-server` command shall have command-level tests covering flag
parsing (--port, --host, --stdio, --quiet), config fallback for port/host,
and error handling when config or database is missing. The MCP adapter
layer is well-tested but the command wiring has zero coverage.

## Acceptance Criteria

1. Test --port flag overrides config port.
2. Test --host flag overrides config host.
3. Test --stdio flag selects stdio transport.
4. Test --quiet flag suppresses startup banner.
5. Test missing database returns error.
6. Test default port/host from config when flags omitted.

## Files to Create/Modify

- `internal/cmd/mcp_test.go` -- Command-level MCP server tests

## Effort Estimate

0.25 weeks
