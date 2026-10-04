# REQ-ORCH-010: Cross-Web Dependency Graph

## Metadata
- **Category**: ORCH
- **Subcategory**: ParallelOps
- **Priority**: HIGH
- **Phase**: 32
- **Status**: MISSING
- **Dependencies**: REQ-ORCH-001, REQ-ORCH-007
- **Blocks**: REQ-ORCH-011, REQ-ORCH-012, REQ-ORCH-013, REQ-ORCH-015, REQ-ORCH-018

## Summary

RTMX shall compute cross-web dependencies by analyzing inter-requirement dependencies that span web boundaries. A new function `Graph.WebDependencies(webs []Web) map[int][]int` returns a directed graph of web indices where web B depends on web A if any requirement in B has a dependency on any requirement in A. This enables merge ordering and parallel grouping to respect logical dependency chains that cross web boundaries.

## Acceptance Criteria

1. `WebDependencies` returns an empty map when no cross-web dependencies exist.
2. `WebDependencies` correctly identifies that web B depends on web A when a requirement in B lists a dependency on a requirement in A.
3. Transitive cross-web dependencies are represented (if web C depends on B and B depends on A, both B->A and C->B appear in the map).
4. Self-dependencies (a web depending on itself) are excluded from the output.
5. The function handles the degenerate case of a single web (returns empty map).
6. The function handles webs with multiple cross-web dependency edges, deduplicating the result.
7. Performance: O(W^2 * R) where W is web count and R is max requirements per web.

## Dependencies
- REQ-ORCH-001 (provides DetectWebs and the Web type)
- REQ-ORCH-007 (provides DetectOverlaps and the WebOverlap type)

## Blocks
- REQ-ORCH-011 (merge order needs cross-web dependency graph)
- REQ-ORCH-012 (parallel groups need cross-web dependency graph)
- REQ-ORCH-013 (webs command displays dependency info)
- REQ-ORCH-015 (merge-gate checks upstream web completion)
- REQ-ORCH-018 (MCP parallel-plan returns dependency info)

## Files to Modify
- internal/graph/web.go (add WebDependencies function)
- internal/graph/web_test.go (add TestWebDependencies table-driven tests)

## Test Strategy
- Table-driven unit tests covering: no webs, single web, two independent webs, two dependent webs, three webs with chain dependency, webs with multiple cross-edges, webs with bidirectional dependencies (cycle detection).
- Property test: for any set of webs, every edge in the web dependency graph corresponds to at least one concrete requirement-level dependency.

## Effort
- 0.75 weeks

## Priority
- HIGH
