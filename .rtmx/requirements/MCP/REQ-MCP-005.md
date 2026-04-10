# REQ-MCP-005: MCP Server Mutation Tools for Claim/Verify/Release

## Metadata
- **Category**: MCP
- **Subcategory**: Server
- **Priority**: HIGH
- **Phase**: 20
- **Status**: MISSING
- **Dependencies**: REQ-MCP-003, REQ-ORCH-005
- **Blocks**: REQ-PLUGIN-001

## Requirement

The RTMX MCP server shall expose mutation tools (claim, release, heartbeat, verify --update) behind an authorization model, allowing agents to modify RTM state through the MCP protocol.

## Rationale

Read-only MCP tools (REQ-MCP-003) let agents see the roadmap. Mutation tools let agents act on it -- claim requirements, run verification, update status. These are higher-risk operations and need authorization controls to prevent accidental or malicious state changes.

## Design

### Mutation Tools

| Tool Name | Description | Parameters | Side Effects |
|-----------|-------------|------------|-------------|
| rtmx_claim | Claim a requirement | req_id (string), agent (string, optional) | Writes to .rtmx/claims.json |
| rtmx_release | Release a claim | req_id (string) | Removes from .rtmx/claims.json |
| rtmx_heartbeat | Update claim heartbeat | agent (string, optional) | Updates timestamps in claims.json |
| rtmx_verify | Run verification and update status | results_file (string, optional), update (bool) | Modifies database.csv |
| rtmx_next_claim | Claim next available requirement or web | mode (string: "one" or "batch"), agent (string), max_effort (string, optional) | Claims + returns plan |

### Authorization Model

Mutation tools require the MCP client to provide an agent identity. The server validates:
1. Agent name is non-empty.
2. For release/heartbeat: the agent must own the claim.
3. For verify --update: the agent must have a claim on at least one requirement being verified.

Authorization is lightweight (identity-based, not token-based) since the MCP server runs locally. Remote HTTP/SSE transport should require an API key configured in rtmx.yaml.

### Concurrency

Mutation tools acquire an exclusive write lock on the relevant resource:
- claims.json: file lock for claim/release/heartbeat
- database.csv: file lock for verify --update

Read tools continue to use shared read locks so they do not block during writes.

### Dry Run

All mutation tools accept a `dry_run` (bool) parameter. When true, the tool returns what would happen without making changes.

## Acceptance Criteria

1. All 5 mutation tools are registered and discoverable.
2. Agent identity is required for mutation operations.
3. Unauthorized operations return clear error messages.
4. Write locks prevent concurrent mutations on the same resource.
5. Dry run mode works for all mutation tools.
6. HTTP/SSE transport requires API key for mutations.
7. All mutations are idempotent (re-claiming your own claim is a no-op).

## Files to Create/Modify

- `internal/adapters/mcp/tools_mutation.go` -- Mutation tool handlers
- `internal/adapters/mcp/tools_mutation_test.go` -- Authorization and concurrency tests
- `internal/adapters/mcp/auth.go` -- Agent identity validation

## Test Strategy

- Table-driven tests for each mutation tool covering success, missing agent, unauthorized, and concurrent access
- Mock file system tests verifying claims.json and database.csv locking behavior
- Dry run tests confirming no side effects for all 5 tools
- Authorization tests: agent identity validation, ownership checks for release/heartbeat
- Concurrency tests: multiple goroutines invoking mutation tools simultaneously to verify lock correctness
- Integration test: end-to-end claim/verify/release cycle through MCP protocol
- Golden file tests for tool discovery (tools/list) output including mutation tools
