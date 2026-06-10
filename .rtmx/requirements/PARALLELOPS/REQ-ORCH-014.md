# REQ-ORCH-014: CLI rtmx plan-parallel Command

## Metadata
- **Category**: ORCH
- **Subcategory**: ParallelOps
- **Priority**: HIGH
- **Phase**: 32
- **Status**: MISSING
- **Dependencies**: REQ-ORCH-011, REQ-ORCH-012
- **Blocks**: (none)

## Summary

RTMX shall provide a new CLI command `rtmx plan-parallel [--agents N]` that produces a concrete execution plan for parallel workstream orchestration. The plan assigns webs to agents across parallel groups, computes merge gates between groups, and estimates critical path duration. The `--agents` flag (default: number of webs in the largest parallel group) controls how many concurrent agents are available. Output shows a structured plan with phases, agent assignments, merge gates, and timing estimates. The `--json` flag outputs structured JSON for programmatic consumption.

## Acceptance Criteria

1. `rtmx plan-parallel` outputs a structured execution plan with phases corresponding to parallel groups.
2. Each phase shows which webs are assigned to which agent (agent-0, agent-1, etc.).
3. When `--agents N` is less than webs in a group, webs are scheduled across rounds within the phase, with the highest-effort webs first.
4. Merge gates between phases list the preconditions (all webs in previous phase must be merged).
5. Critical path duration is estimated as the sum of the longest web effort in each phase.
6. `rtmx plan-parallel --json` outputs JSON with: phases (array of {group, webs, agents, merge_gate}), critical_path_weeks, total_effort_weeks, parallelism_ratio (total_effort / critical_path).
7. When no incomplete requirements exist, the command prints a message and exits 0.
8. When `--agents 0` or negative, the command returns an error.
9. Exit code 0 on success.

## Dependencies
- REQ-ORCH-011 (merge order determines phase sequencing)
- REQ-ORCH-012 (parallel groups determine phase composition)

## Blocks
- (none)

## Files to Modify
- internal/cmd/plan_parallel.go (new command implementation)
- internal/cmd/plan_parallel_test.go (new test file)
- internal/cmd/root.go (register plan-parallel command)

## Test Strategy
- Unit tests using CommandContext with mock database covering: empty database, single web, multiple independent webs (all in one phase), dependent webs (multiple phases), agent constraint limiting parallelism, JSON output validation.
- Golden file tests for text output formatting.
- Edge case: more agents than webs, one agent, webs with zero effort.

## Effort
- 1.25 weeks

## Priority
- HIGH
