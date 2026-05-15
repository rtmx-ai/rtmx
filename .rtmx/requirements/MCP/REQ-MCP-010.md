# REQ-MCP-010: MCP Verify Tool Executes Tests and Updates Requirements

## Metadata
- **Category**: MCP
- **Subcategory**: Server
- **Priority**: P0
- **Phase**: 27
- **Status**: MISSING
- **Dependencies**: REQ-MCP-005
- **Blocks**: (none)

## Requirement

The RTMX MCP verify tool shall execute a test command, parse multi-format
test output, map results to requirements by test_function matching, and
update the RTM database -- matching the behavior of `rtmx verify --command`.

## Rationale

The MCP verify tool was originally read-only: it returned current database
status without running tests. This created a verify-loop trap where agents
called verify repeatedly expecting it to acknowledge passing tests, wasting
30-40% of token budget. Closed-loop verification requires the MCP tool to
actually execute tests and update requirement status, just as the CLI does.

## Design

### Tool Schema

The verify tool accepts an optional `command` parameter:

```json
{
  "name": "verify",
  "inputSchema": {
    "type": "object",
    "properties": {
      "command": {
        "type": "string",
        "description": "Test command to run. Auto-detects from project files if omitted."
      }
    }
  }
}
```

### Test Command Detection

When `command` is omitted, auto-detect from project files in the database
directory (same heuristic as CLI `DetectTestCommand`):

| File Present | Command |
|-------------|---------|
| Cargo.toml | `cargo test --workspace` |
| package.json | `npm test` |
| pyproject.toml / setup.py / requirements.txt | `python3 -m pytest -v` |
| build.gradle(.kts) | `gradle test` |
| Makefile | `make test` |
| (default) | `go test -json ./...` |

### Multi-Format Test Output Parsing

Parse stdout line-by-line for these formats:

1. **Go test JSON** -- `{"Test":"...", "Action":"pass|fail|skip"}`
2. **Cargo** -- `test module::name ... ok|FAILED|ignored`
3. **pytest verbose** -- `file.py::test_name PASSED|FAILED|SKIPPED`
4. **Node TAP** -- `ok|not ok N - description`

### Requirement Mapping

Match parsed test names to requirement `test_function` fields using
suffix matching at path boundaries (`::`, `.`, `/`) and substring
containment, mirroring `matchTestFunction` from the CLI verify command.

### Database Update

- Test passed: set status to COMPLETE, set started/completed dates
- Test failed and was COMPLETE: downgrade to PARTIAL
- Save database only if any status changed

### Response

```json
{
  "total": 10,
  "complete": 7,
  "verified": 8,
  "updated": 2,
  "command": "go test -json ./...",
  "items": [
    {
      "req_id": "REQ-GO-001",
      "status": "COMPLETE",
      "previous_status": "MISSING",
      "has_test": true,
      "test_function": "TestRootCommandHelp",
      "test_passed": true,
      "updated": true
    }
  ]
}
```

## Acceptance Criteria

1. Verify tool executes the provided or auto-detected test command.
2. Verify tool parses Go JSON, cargo, pytest, and TAP test output formats.
3. Verify tool maps test results to requirements by test_function matching.
4. Verify tool updates requirement status and dates in the database.
5. Verify tool returns complete/verified/updated counts and per-item detail.
6. When no command is provided, auto-detection selects the correct runner.
7. Database is saved only when status changes occur.

## Files to Create/Modify

- `internal/adapters/mcp/server.go` -- Enhanced toolVerify, new helper functions
- `internal/adapters/mcp/server_test.go` -- Tests for verify execution pipeline

## Test Strategy

- Table-driven tests with mock test command output for each supported format
- Test auto-detection with temp directories containing marker files
- Test requirement mapping with exact, suffix, and substring matches
- Test database update: MISSING->COMPLETE on pass, COMPLETE->PARTIAL on fail
- Test that database is not saved when no changes occur
- Integration test: end-to-end verify with a real test command
