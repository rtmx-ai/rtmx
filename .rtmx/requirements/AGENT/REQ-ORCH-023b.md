# REQ-ORCH-023b: Agent-callable loop tick with learning re-decompose

## Metadata
- **Category**: ORCH
- **Subcategory**: AgentLoop
- **Priority**: P0
- **Phase**: 37
- **Status**: COMPLETE
- **Dependencies**: REQ-ORCH-023a|REQ-ORCH-019c|REQ-ORCH-020
- **Blocks**: REQ-ORCH-023c|REQ-ORCH-023d
- **ADR**:

## Requirement

`rtmx loop tick --agent-id X [--json]` SHALL select the next unblocked
requirement (open-PR skip applies), claim it for the agent, decompose
if coarse, and optionally re-decompose web descendants listed in prior
delivery notes. MCP `loop_tick` SHALL return the same JSON plan.

## Acceptance Criteria

1. [ ] `rtmx loop tick --agent-id X` claims at most one requirement and returns a plan.
2. [ ] Plan includes `req_id`, `claimed`, `decomposed`, `children`, `redecomposed`, `pr_policy`, `commit_policy`, `trade_required`, `ready_to_implement`.
3. [ ] If `.rtmx/delivery/<prior>.md` has `## Follow-on decomposition` listing REQ IDs (or YAML `redecompose:`), those IDs are decomposed and listed under `redecomposed`.
4. [ ] Idle backlog returns `{"idle":true}` without mutating claims.
5. [ ] MCP `loop_tick` mirrors CLI JSON (requires `agent_id`).
6. [ ] Tests bind `REQ-ORCH-023b`.

## Files

- `internal/cmd/loop_tick.go` (extend)
- `internal/cmd/loop.go` (subcommand)
- `internal/adapters/mcp/server.go`

## Out of Scope

- Hosting the coding agent.
