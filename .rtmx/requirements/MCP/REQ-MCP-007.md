# REQ-MCP-007: Response Size Logging

## Status: MISSING
## Priority: P0
## Phase: 27

## Requirement

RTMX MCP server shall log response byte count and estimated token count
to stderr on every tools/call invocation, providing zero-cost observability
into agent token consumption.

## Rationale

MCP tool responses vary from 38 tokens (next --one) to 8,400+ tokens
(deps overview) depending on database size and tool. Without instrumentation,
operators and agent developers have no visibility into how much context
window each tool call consumes. Logging to stderr is free -- agents read
stdout only -- and enables measurement without code changes on the client.

## Acceptance Criteria

1. Every successful tools/call response logs to stderr:
   `[rtmx-mcp] tool=%s bytes=%d tokens=%d` where tokens = ceil(bytes/4)
2. Error responses are also logged with `error=true` appended
3. Logging occurs on both stdio and HTTP transports
4. Logging does not affect stdout or HTTP response payloads
5. Token estimate uses the 4-bytes-per-token heuristic (industry standard
   for JSON-dense content; accurate within 20% for Claude/GPT tokenizers)
6. Log output is machine-parseable (structured key=value format)
7. A `--quiet` flag on `mcp-server` suppresses response size logging

## Dependencies

- REQ-MCP-003: Read-only tools (tool implementations that produce responses)
- REQ-MCP-006: Stdio transport (stderr logging must not interfere with stdio)

## Test

- `internal/adapters/mcp/server_test.go::TestMCPResponseSizeLogging`
- Subtests: logs_bytes_and_tokens, logs_on_error, quiet_flag_suppresses,
  does_not_affect_stdout, http_transport_logs

## Files to Create/Modify

- `internal/adapters/mcp/server.go` -- Add logging in handleToolsCall after
  JSON marshaling (line ~424-433); add logger field to Server struct
- `internal/cmd/mcp.go` -- Add --quiet flag, wire to Server option
- `internal/adapters/mcp/server_test.go` -- Add TestMCPResponseSizeLogging

## Design Notes

The logging hook point is handleToolsCall (server.go:424-433), after
json.Marshal(data) produces jsonBytes. At that point both the tool name
and response size are known. The log line is written before the toolResult
wrapper is constructed, so it measures payload size, not envelope overhead.

For HTTP transport, the same handleToolsCall path is used via processRPC,
so logging is transport-agnostic.
