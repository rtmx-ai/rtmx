# REQ-ORCH-011: Merge Order Computation

## Metadata
- **Category**: ORCH
- **Subcategory**: ParallelOps
- **Priority**: HIGH
- **Phase**: 32
- **Status**: MISSING
- **Dependencies**: REQ-ORCH-010, REQ-ORCH-007
- **Blocks**: REQ-ORCH-013, REQ-ORCH-014, REQ-ORCH-015, REQ-ORCH-018

## Summary

RTMX shall compute a safe merge ordering for work webs via a new function `Graph.MergeOrder(webs []Web, overlaps []WebOverlap) ([]int, error)`. The ordering respects three rules applied in priority order: (a) dependency order -- if web B depends on web A, A appears before B; (b) file-conflict sequencing -- webs with file surface overlaps are not adjacent in the merge sequence unless no dependency constraint forces it; (c) tiebreaker by total effort descending -- among webs with no ordering constraints between them, the largest web merges first to unblock the most downstream work. Returns an error if cyclic cross-web dependencies make a valid ordering impossible.

## Acceptance Criteria

1. `MergeOrder` returns web indices in an order where every dependency is satisfied (upstream before downstream).
2. `MergeOrder` returns an error when cyclic cross-web dependencies exist, with the cycle described in the error message.
3. Among webs with no dependency relationship, overlapping webs are separated in the merge sequence when possible.
4. Among webs with no dependency or overlap constraints, the web with higher total effort appears first.
5. The function handles the empty input case (returns empty slice, no error).
6. The function handles a single web (returns that web index).
7. The output is deterministic for the same input (stable sort with effort tiebreaker).

## Dependencies
- REQ-ORCH-010 (cross-web dependency graph feeds into merge ordering)
- REQ-ORCH-007 (file overlap data feeds into conflict sequencing)

## Blocks
- REQ-ORCH-013 (webs command shows merge order)
- REQ-ORCH-014 (plan-parallel uses merge order)
- REQ-ORCH-015 (merge-gate validates merge order compliance)
- REQ-ORCH-018 (MCP parallel-plan returns merge order)

## Files to Modify
- internal/graph/web.go (add MergeOrder function)
- internal/graph/web_test.go (add TestMergeOrder table-driven tests)

## Test Strategy
- Table-driven unit tests covering: empty webs, single web, two independent webs (effort tiebreaker), two dependent webs, chain of three, diamond dependency, cyclic dependency error, overlapping webs with no dependency, overlapping webs with dependency constraint.
- Property test: for any valid merge order output, every dependency edge is respected (upstream index appears before downstream).

## Effort
- 1.0 weeks

## Priority
- HIGH
