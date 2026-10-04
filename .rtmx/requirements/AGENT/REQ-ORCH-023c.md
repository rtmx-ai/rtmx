# REQ-ORCH-023c: Scientific MCP tool surface

## Metadata
- **Category**: ORCH
- **Subcategory**: AgentLoop
- **Priority**: P0
- **Phase**: 37
- **Status**: COMPLETE
- **Dependencies**: REQ-ORCH-023b
- **Blocks**: REQ-ORCH-023d
- **ADR**:

## Requirement

The MCP server SHALL expose scientific-workflow tools so agents can
complete a delivery cycle without shelling arbitrary CLI: `loop_tick`,
`decompose`, `hygiene`, `cycles`, `webs`, `context`, `delivery_check`,
plus trade tools from 023d. No generic `run_cli` tool.

## Acceptance Criteria

1. [ ] `tools/list` includes `loop_tick`, `decompose`, `hygiene`, `cycles`, `webs`, `context`, `delivery_check`.
2. [ ] Each tool has a tested happy path returning JSON (or structured content).
3. [ ] Mutations (`loop_tick`, `decompose` when writing) require `agent_id` where applicable.
4. [ ] Tool descriptions stay short; filters support `limit` where lists are large.
5. [ ] Tests bind `REQ-ORCH-023c`. Docs note auth/sync/serve stay CLI-only.

## Files

- `internal/adapters/mcp/server.go`
- `internal/adapters/mcp/server_test.go` / delivery tests
