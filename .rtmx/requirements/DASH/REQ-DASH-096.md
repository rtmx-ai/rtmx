# REQ-DASH-096: Add cycle_path Data to Enriched Graph JSON

## Summary

The enriched graph JSON includes a `cycles` field: an array of arrays, where each inner array lists the requirement IDs forming a dependency cycle in traversal order. This enables the cycle indicator (REQ-DASH-055) to highlight cycle members and show cycle paths. Cycle detection reuses the existing Tarjan SCC implementation in the graph package.

## Acceptance Criteria

1. The graph JSON response includes a top-level `cycles` field of type `[][]string`.
2. Each inner array contains the requirement IDs forming one cycle, listed in traversal order.
3. If no cycles exist, `cycles` is an empty array `[]`, not null.
4. Cycle detection uses the existing `graph.TarjanSCC()` implementation -- no new cycle detection algorithm.
5. Only SCCs of size >= 2 are included (single-node trivial SCCs are excluded).
6. The JSON schema change is backward-compatible: existing consumers ignore unknown fields.
7. Unit tests verify correct cycle detection for acyclic graphs, single cycles, and multiple overlapping cycles.

## Dependencies

- REQ-DASH-044 (enriched graph JSON structure to extend)

## Blocks

- REQ-DASH-055 (cycle indicator needs cycle_path data to identify cycle members)

## Files to Modify

- `internal/cmd/serve_dashboard.go` (add cycles computation to graph enrichment using graph.TarjanSCC)

## Test Strategy

- Unit test: acyclic graph produces empty cycles array
- Unit test: graph with one 3-node cycle produces one cycle entry with 3 IDs
- Unit test: graph with two independent cycles produces two cycle entries
- Unit test: trivial SCCs (single nodes) are excluded
- Integration test: GET /api/graph response includes cycles field

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 32
