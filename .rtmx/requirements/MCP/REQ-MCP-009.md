# REQ-MCP-009: Token Budget Awareness in Tool Descriptions

## Status: MISSING
## Priority: HIGH
## Phase: 27

## Requirement

RTMX MCP tool descriptions shall advertise approximate response sizes
so agents and orchestrators can make informed decisions about which
tools to call and when to apply filters.

## Rationale

MCP tool descriptions are the primary interface through which agents
decide which tools to invoke. Currently, all 10 tools have terse
descriptions that give no indication of response size. An agent cannot
distinguish between `status` (424 tokens) and `deps` (8,418 tokens)
without calling both.

By embedding size hints in descriptions, agents (and the humans
prompting them) can make token-aware choices:
- Call `status` for a quick overview instead of `verify` for full detail
- Use `deps` with a specific req_id (350 tokens) instead of overview mode (8,400 tokens)
- Apply filters on `markers` when only one category is relevant

This is especially important for orchestrators that manage multiple
agents with shared token budgets.

## Acceptance Criteria

1. Each tool description in handleToolsList() includes a size hint suffix:
   `(~N tokens)` for fixed-size tools, `(~N tokens per requirement)` for
   collection tools

2. Size hints are derived from measured response sizes, not hardcoded
   guesses. The build or test suite validates that advertised sizes are
   within 50% of actual sizes on the project's own database.

3. Collection tools that support filtering (REQ-MCP-008) note filtering
   availability in their description, e.g.:
   "Show verification status (~27 tokens/req, filterable by category/status)"

4. The `deps` tool description distinguishes between specific mode
   (~50 tokens + deps) and overview mode (~36 tokens/req)

5. Descriptions remain concise -- the size hint is appended, not a
   separate paragraph. Total description length stays under 120 characters.

6. `next` tool description notes that response size scales with number
   of incomplete requirements and work webs

## Dependencies

- REQ-MCP-003: Read-only tools (descriptions to update)
- REQ-MCP-007: Response size logging (provides measurement data)
- REQ-MCP-008: Filtering parameters (descriptions reference filtering)

## Test

- `internal/adapters/mcp/server_test.go::TestMCPToolDescriptions`
- Subtests: descriptions_include_size_hints, size_hints_within_tolerance,
  collection_tools_mention_filtering

## Files to Create/Modify

- `internal/adapters/mcp/server.go`:
  - Update tool descriptions in handleToolsList() (lines 296-377)
  - Add size hint strings to each toolDef Description field
- `internal/adapters/mcp/server_test.go` -- Add TestMCPToolDescriptions
  that calls tools/list and validates descriptions contain size hints
  and that the hints are within tolerance of actual response sizes

## Design Notes

Size hints should be expressed as approximate tokens, not bytes, since
tokens are the unit agents and orchestrators reason about.

For collection tools, the per-requirement rate is more useful than an
absolute number because it lets agents estimate cost for any database
size: "I have 500 requirements, markers costs ~27 tokens/req, so ~13,500
tokens unfiltered vs ~1,350 for a single category of 50 requirements."

The tolerance check (AC2) should measure actual tool responses against
the current database and compare to the advertised per-req rate. This
prevents size hints from going stale as response format evolves.
