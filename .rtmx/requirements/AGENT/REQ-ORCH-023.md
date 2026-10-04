# REQ-ORCH-023: Agent Delivery Science

## Metadata
- **Category**: ORCH
- **Subcategory**: AgentLoop
- **Priority**: P0
- **Phase**: 37
- **Status**: COMPLETE
- **Dependencies**: REQ-ORCH-019c|REQ-ORCH-020|REQ-ORCH-021|REQ-ORCH-022
- **Blocks**: REQ-ORCH-023a|REQ-ORCH-023b|REQ-ORCH-023c|REQ-ORCH-023d
- **ADR**:

## Requirement

RTMX SHALL provide warn-first mechanical delivery checks (one PR per
requirement, one commit per acceptance criterion), an agent-callable
loop tick that next+claims and re-decomposes from delivery learnings,
a scientific-workflow MCP tool surface, and trade-analysis checkpoints
so agents capture human intent before coding and verify it after.

Parent COMPLETE only when 023a–d are COMPLETE.

## Children

| ID | Focus |
|----|--------|
| REQ-ORCH-023a | `rtmx delivery-check` warn-first (+ `--strict`) |
| REQ-ORCH-023b | `rtmx loop tick` / MCP `loop_tick` + learning re-decompose |
| REQ-ORCH-023c | MCP scientific surface (decompose, hygiene, cycles, webs, context, delivery_check) |
| REQ-ORCH-023d | Trade analysis checkpoints |

## Effort Estimate

4.0 weeks
