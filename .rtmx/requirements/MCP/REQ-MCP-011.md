# REQ-MCP-011: MCP set_status Tool (Status Writeback, Not COMPLETE)

## Metadata
- **Category**: MCP
- **Subcategory**: Server
- **Priority**: P1
- **Phase**: 29
- **Status**: MISSING
- **Dependencies**: REQ-MCP-005
- **Blocks**: (none)

## Requirement

The RTMX MCP server shall expose a `set_status` mutation tool that sets a
requirement's status with provenance (`agent_id`, optional `reason`). It shall
**reject COMPLETE**: completion is verify-driven and may only come from the
`verify` tool running tests. All other valid statuses (PARTIAL, MISSING,
NOT_STARTED) are permitted.

## Rationale

The MCP surface exposes `claim`/`release` (orchestration) and `verify`
(closed-loop completion) but no way to set a requirement's status — e.g. to
reopen a regressed requirement, mark work as started, or park a blocked one.
Consumers (like aegis-cli's orchestrator) currently work around this by editing
`.rtmx/database.csv` directly, which means a second writer to the database and
duplicated CSV logic. A first-class `set_status` tool keeps rtmx the single
writer of its own data.

The COMPLETE guard is the load-bearing constraint: allowing an agent to mark a
requirement COMPLETE via a mutation tool would reintroduce a self-grading loop
(an agent declaring its own work done without tests). Completion must remain a
function of `verify` running the acceptance tests.

## Design

### Tool Schema

`set_status` accepts:
- `req_id` (required) — the requirement to update.
- `status` (required) — the new status; parsed via `database.ParseStatus`.
  `COMPLETE` is rejected with an error.
- `agent_id` (required) — the acting agent (mutation authorization, as with
  `claim`/`release`).
- `reason` (optional) — provenance for the change.

### Behavior

On success it sets `req.Status`, saves the database, and returns
`{req_id, status, previous, agent_id, reason}`. Unknown `req_id`, missing
required args, an unparseable status, or `COMPLETE` each return an error result.

## Acceptance Criteria

- A valid non-COMPLETE status is applied and persisted.
- `COMPLETE` is refused.
- Missing `agent_id` (or `req_id`/`status`) is refused.
- An unknown requirement returns an error.

## Test

`internal/adapters/mcp/set_status_test.go::TestMCPSetStatus`
