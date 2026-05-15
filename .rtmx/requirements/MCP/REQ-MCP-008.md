# REQ-MCP-008: Filtering Parameters on MCP Tools

## Status: MISSING
## Priority: P0
## Phase: 27

## Requirement

RTMX MCP tools that return collections (backlog, verify, markers, deps)
shall accept optional filtering parameters -- category, status, and limit
-- so agents retrieve only the data they need instead of the full database.

## Rationale

Without filtering, the four collection tools dump the entire database on
every call. Measured on a 234-requirement project:

| Tool     | Unfiltered bytes | Unfiltered tokens |
|----------|----------------:|------------------:|
| markers  | 26,158          | 6,540             |
| deps     | 33,670          | 8,418             |
| verify   | 25,077          | 6,269             |
| backlog  | variable        | up to 12,000      |

At 1,000 requirements these grow 4x, consuming 25,000-35,000 tokens per
call -- a significant fraction of the context window. An agent checking
"are my AUTH tests passing?" should not receive all 1,000 rows.

Filtering is the highest-impact token efficiency improvement because it
reduces response size multiplicatively rather than additively.

## Acceptance Criteria

1. `backlog` accepts optional arguments:
   - `category` (string): filter to requirements in this category
   - `status` (string): filter to requirements with this status
   - `limit` (integer): return at most N items (default: all)

2. `verify` accepts optional arguments:
   - `category` (string): filter to requirements in this category
   - `status` (string): filter to requirements with this status
   - `limit` (integer): return at most N items

3. `markers` accepts optional arguments:
   - `category` (string): filter to requirements in this category
   - `status` (string): filter to requirements with this status
   - `limit` (integer): return at most N items

4. `deps` accepts optional arguments (overview mode only):
   - `category` (string): filter to requirements in this category
   - `limit` (integer): return at most N items
   (specific-requirement mode is already filtered by req_id)

5. All filters are additive (AND logic): `category=AUTH` + `status=MISSING`
   returns only AUTH requirements with status MISSING

6. Omitting all filter arguments produces the same output as today
   (backward compatible -- existing agent code continues to work)

7. Tool descriptions in tools/list include the new arguments in
   inputSchema so agents can discover them via MCP introspection

8. Response includes a `filtered` boolean and `total_unfiltered` count
   so agents know when they received a subset

## Dependencies

- REQ-MCP-003: Read-only tools (tool implementations to extend)
- REQ-MCP-007: Response size logging (log filtered vs unfiltered sizes)

## Test

- `internal/adapters/mcp/server_test.go::TestMCPToolFiltering`
- Subtests: backlog_filter_category, backlog_filter_status, backlog_limit,
  verify_filter_category, markers_filter_combined, deps_overview_limit,
  no_filters_backward_compatible, filtered_metadata_present

## Files to Create/Modify

- `internal/adapters/mcp/server.go`:
  - Update handleToolsList() inputSchema for backlog, verify, markers, deps
    to declare category/status/limit properties
  - Update handleToolsCall() to extract filter arguments from call.Arguments
  - Update toolBacklog(), toolVerify(), toolMarkers(), toolDeps() signatures
    to accept filter parameters
  - Add filterRequirements() helper that applies category/status/limit
  - Add filtered/total_unfiltered fields to collection response wrappers
- `internal/adapters/mcp/server_test.go` -- Add TestMCPToolFiltering

## Design Notes

The filter should be applied after the tool gathers its full result set
but before JSON marshaling. This keeps tool logic clean -- each tool
builds its natural output, then a shared filter trims it.

The `limit` parameter is applied last (after category and status filters)
so it acts as a "top N from the filtered set."

Category matching should be case-insensitive to handle agents that send
"auth" vs "AUTH" vs "Auth".
