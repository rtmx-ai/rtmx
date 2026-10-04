# REQ-DASH-045: Enrich Graph Edge JSON with Critical Path Flag

## Summary

The graph JSON edge objects currently contain only source and target fields. This requirement adds an `is_critical_path` boolean field to each edge, derived from the `graph.CriticalPath()` analysis. An edge is on the critical path if both its source and target are consecutive nodes in the critical path sequence. This enables the frontend to render critical path edges with distinct styling (thicker, highlighted) in future visual requirements.

## Acceptance Criteria

1. The `gEdge` JSON struct includes:
   - `source` (string) -- existing
   - `target` (string) -- existing
   - `is_critical_path` (bool) -- true if this edge connects consecutive nodes on the critical path
2. CriticalPath() result is computed once and shared with REQ-DASH-044 node enrichment (no redundant computation).
3. An edge (A -> B) is marked `is_critical_path: true` if and only if A and B are adjacent in the CriticalPath() sequence and the edge direction matches.
4. When no critical path exists (all complete, or disconnected graph), all edges have `is_critical_path: false`.
5. Existing frontend `renderGraph()` continues to work because the new field is additive.

## Dependencies

- REQ-DASH-001 (SPA framework)
- REQ-DASH-010 (design tokens)

## Blocks

- REQ-DASH-050+ (visual encoding of critical path edges)

## Files to Modify

- `internal/cmd/serve_dashboard.go` (expand gEdge struct, compute critical path edge set, populate flag)

## Test Strategy

- Unit test: linear chain A->B->C where all are on critical path; verify edges A->B and B->C have is_critical_path=true
- Unit test: diamond graph where only one path is critical; verify non-critical edges are false
- Unit test: all-complete database produces no critical path edges
- Unit test: cycle-containing graph handles critical path gracefully

## Effort

- 0.50 weeks

## Priority

- HIGH

## Phase

- 32
