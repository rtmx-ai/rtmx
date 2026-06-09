# REQ-DASH-075: Progressive Rendering with Collapsed Category Clusters

## Summary

When the graph contains more than 100 nodes, the initial render shall display collapsed cluster
nodes -- one per category, showing a count badge -- instead of rendering all individual nodes.
Clicking a cluster expands it to show its member nodes in place. Dagre layout is computed only
on the currently visible nodes (clusters plus any expanded members), keeping layout cost
proportional to visible complexity rather than total database size.

## Acceptance Criteria

1. When node count exceeds 100, `renderGraph()` groups nodes by category and renders one cluster node per category with a count badge (e.g., "CLI (47)")
2. Cluster nodes are visually distinct from regular nodes: larger rectangle, dashed border, bold label with count
3. Edges between clusters represent the union of inter-category edges, with a count label showing the number of underlying edges
4. Clicking a cluster node expands it in place, replacing the cluster with its individual member nodes and re-running dagre layout on the updated visible set
5. An expanded cluster shows a collapse button (minus icon) that reverts it to the cluster node
6. Dagre layout is called only with the currently visible node set (clusters + expanded members), never with the full 500+ node set
7. When all clusters are collapsed, layout computation for a 500-node database completes in under 100ms
8. Expanding a single cluster of 50 nodes re-layouts in under 300ms
9. When node count is 100 or fewer, all nodes render individually with no clustering (existing behavior preserved)
10. The cluster threshold is configurable via a graph control (dropdown: "Auto", "Always cluster", "Never cluster")

## Dependencies

- REQ-DASH-048 (dagre layout engine must be in place)

## Blocks

- REQ-DASH-077 (screen reader support needs cluster ARIA roles)
- REQ-DASH-082 (URL state must encode expanded cluster set)

## Files to Modify

- `internal/dashboard/static/app.js` (cluster logic in renderGraph, expand/collapse handlers)
- `internal/dashboard/static/styles.css` (cluster node styles, count badge, collapse button)
- `internal/dashboard/templates/partials/graph.html` (cluster threshold control in toolbar)

## Test Strategy

- Unit test: verify clustering groups nodes by category and produces correct count badges
- Unit test: verify dagre.layout() is called with cluster node count, not full node count
- Unit test: expand cluster replaces cluster node with member nodes and re-runs layout
- Unit test: collapse restores cluster node and re-runs layout
- Performance test: 500-node database with all clusters collapsed layouts in under 100ms
- Performance test: expanding a 50-node cluster re-layouts in under 300ms
- Integration test: htmx graph partial loads with clusters when database exceeds 100 nodes
- Edge case test: category with 1 node is not clustered (rendered directly)
- Edge case test: all categories expanded renders identically to non-clustered mode

## Effort

- 1.50 weeks

## Priority

- HIGH

## Phase

- 32
