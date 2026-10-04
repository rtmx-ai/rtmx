# REQ-ORCH-012: Parallel Group Assignment

## Metadata
- **Category**: ORCH
- **Subcategory**: ParallelOps
- **Priority**: HIGH
- **Phase**: 32
- **Status**: MISSING
- **Dependencies**: REQ-ORCH-010, REQ-ORCH-007
- **Blocks**: REQ-ORCH-013, REQ-ORCH-014, REQ-ORCH-017, REQ-ORCH-018

## Summary

RTMX shall partition work webs into parallel execution groups via a new function `Graph.ParallelGroups(webs []Web, overlaps []WebOverlap) [][]int`. Each group contains web indices that can safely execute concurrently: no cross-web dependencies between them and no file surface overlaps. Groups are ordered by merge priority (group 0 executes first, group 1 after group 0 merges, etc.). Within each group, webs are sorted by total effort descending.

## Acceptance Criteria

1. `ParallelGroups` returns groups where no two webs in the same group have a cross-web dependency.
2. `ParallelGroups` returns groups where no two webs in the same group have a file surface overlap.
3. Groups are ordered such that all dependencies of webs in group N are satisfied by webs in groups 0..N-1.
4. Within each group, webs are sorted by total effort descending.
5. The function produces the minimum number of groups necessary (maximizes parallelism).
6. Empty input returns empty slice.
7. Fully independent, non-overlapping webs all appear in group 0.
8. A chain of N dependent webs produces N groups of one web each.

## Dependencies
- REQ-ORCH-010 (cross-web dependency graph determines group membership)
- REQ-ORCH-007 (file overlap data prevents unsafe parallel execution)

## Blocks
- REQ-ORCH-013 (webs command displays parallel group assignment)
- REQ-ORCH-014 (plan-parallel uses parallel groups for agent assignment)
- REQ-ORCH-017 (dashboard visualizes parallel groups as swimlanes)
- REQ-ORCH-018 (MCP parallel-plan returns group assignments)

## Files to Modify
- internal/graph/web.go (add ParallelGroups function)
- internal/graph/web_test.go (add TestParallelGroups table-driven tests)

## Test Strategy
- Table-driven unit tests covering: empty input, single web, all independent webs, linear chain, diamond dependency, overlapping webs forced into separate groups, mixed dependency and overlap constraints.
- Property test: for any output, (a) no intra-group dependency or overlap edges exist, (b) all inter-group dependencies point from lower to higher group index.

## Effort
- 1.0 weeks

## Priority
- HIGH
