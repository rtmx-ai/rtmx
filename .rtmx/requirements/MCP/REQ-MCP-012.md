# REQ-MCP-012: MCP next returns the delivery plan

## Metadata
- **Category**: MCP
- **Subcategory**: Server
- **Priority**: MEDIUM
- **Phase**: 36
- **Status**: COMPLETE
- **Dependencies**: REQ-ORCH-019c
- **Blocks**:
- **ADR**:

## Requirement

The MCP `next` tool SHALL return the same delivery plan as one loop
tick: the selected requirement, whether it was decomposed, the child
IDs, and the policies one PR per requirement and one commit per
acceptance criterion. Rationale text is not added.

## Rationale

Agents that drive RTMX through MCP never see `rtmx loop` stdout. If
the plan exists only on the CLI, those agents skip decomposition and
the delivery rule.

## Acceptance Criteria

1. [ ] `next` response gains `delivery` when the server can run the tick planner: `req_id`, `decomposed`, `children`, `pr_policy`, `commit_policy`, `skipped`.
2. [ ] `pr_policy` is `one_pr_per_requirement`. `commit_policy` is `one_commit_per_ac`.
3. [ ] Idle backlog returns `delivery.idle = true` and does not claim.
4. [ ] Existing `next` fields (`webs`, `top_item`) remain. New fields are additive.
5. [ ] Tests bind `REQ-MCP-012`. No GitHub calls.

## Files

- `internal/adapters/mcp/server.go`
- `internal/adapters/mcp/server_test.go`

## Out of Scope

- A separate MCP tool named `loop`. `next` carries the plan; `rtmx loop` remains the deployed watcher.
