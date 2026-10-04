# REQ-ORCH-023d: Trade analysis checkpoints

## Metadata
- **Category**: ORCH
- **Subcategory**: AgentLoop
- **Priority**: P0
- **Phase**: 37
- **Status**: COMPLETE
- **Dependencies**: REQ-ORCH-023b|REQ-ORCH-023c
- **Blocks**:
- **ADR**:

## Requirement

RTMX SHALL support trade-analysis checkpoints as Markdown artifacts
under `.rtmx/trades/`. Agents open/list/resolve trades via CLI and MCP.
`loop_tick` sets `trade_required: true` and `ready_to_implement: false`
while an open trade exists for the selected requirement (or ambiguity
markers are present). Warn-first globally; `--strict` on loop tick fails
when trade is required but open.

## Acceptance Criteria

1. [ ] `rtmx trade open --req ID --title …` writes `.rtmx/trades/TRADE-<req>-<slug>.md` with status open.
2. [ ] `rtmx trade list [--req ID]` and `rtmx trade resolve TRADE-ID --choice …` work.
3. [ ] MCP tools `trade_open`, `trade_list`, `trade_resolve` mirror CLI.
4. [ ] `loop_tick` plan includes `trade_required` when an open trade exists for the req or the requirement text matches ambiguity markers (`TBD`, `TODO(decision)`, `either/or`, `options:`), or `needs_trade: true`.
5. [ ] When `trade_required`, `ready_to_implement` is false.
6. [ ] Tests bind `REQ-ORCH-023d`.

## Files

- `internal/cmd/trade.go`
- `internal/cmd/trade_test.go`
- `internal/adapters/mcp/server.go`
- `internal/orchestration/trade` (or package under cmd)

## Out of Scope

- Requiring ADR files; agents may write ADRs after resolve by convention.
