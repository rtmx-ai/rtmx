# REQ-DASH-075: Cluster Rendering for Graphs Over 100 Nodes (075a)

## Summary

When the graph contains more than 100 nodes, the initial render shall display collapsed cluster nodes -- one per category, showing a count badge -- instead of rendering all individual nodes. Cluster nodes are visually distinct from regular nodes (larger rectangle, dashed border, bold label with count). Edges between clusters represent the union of inter-category edges with a count label. Dagre layout is computed only on the currently visible nodes (clusters plus any expanded members), keeping layout cost proportional to visible complexity.

## Acceptance Criteria

1. When node count exceeds 100, `renderGraph()` groups nodes by category and renders one cluster node per category with a count badge (e.g., "CLI (47)").
2. Cluster nodes are visually distinct from regular nodes: larger rectangle, dashed border, bold label with count.
3. Edges between clusters represent the union of inter-category edges, with a count label showing the number of underlying edges.
4. Dagre layout is called only with the currently visible node set (clusters + expanded members), never with the full 500+ node set.
5. When all clusters are collapsed, layout computation for a 500-node database completes in under 100ms.
6. When node count is 100 or fewer, all nodes render individually with no clustering (existing behavior preserved).
7. Category with 1 node is not clustered (rendered directly).

## Dependencies

- REQ-DASH-048 (dagre layout engine must be in place)

## Blocks

- REQ-DASH-077 (screen reader support needs cluster ARIA roles)
- REQ-DASH-082 (URL state must encode expanded cluster set)
- REQ-DASH-088 (expand/collapse interaction depends on cluster rendering)
- REQ-DASH-089 (threshold control depends on cluster rendering)

## Files to Modify

- `internal/dashboard/static/app.js` (cluster logic in renderGraph, cluster node rendering)
- `internal/dashboard/static/styles.css` (cluster node styles, count badge)

## Test Strategy

- Unit test: verify clustering groups nodes by category and produces correct count badges
- Unit test: verify dagre.layout() is called with cluster node count, not full node count
- Performance test: 500-node database with all clusters collapsed layouts in under 100ms
- Integration test: htmx graph partial loads with clusters when database exceeds 100 nodes
- Edge case test: category with 1 node is not clustered (rendered directly)

## Effort

- 0.75 weeks

## Priority

- HIGH

## Phase

- 32
