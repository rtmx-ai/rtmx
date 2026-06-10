# REQ-DASH-088: Cluster Expand/Collapse Interaction and Re-layout (075b)

## Summary

Clicking a cluster node expands it in place, replacing the cluster with its individual member nodes and re-running dagre layout on the updated visible set. An expanded cluster shows a collapse button (minus icon) that reverts it to the cluster node. Layout is always recomputed on the visible node set after expand or collapse.

## Acceptance Criteria

1. Clicking a cluster node expands it in place, replacing the cluster with its individual member nodes.
2. After expansion, dagre layout is re-run on the updated visible set (clusters + expanded members).
3. An expanded cluster shows a collapse button (minus icon) in the cluster's header area.
4. Clicking the collapse button reverts the expanded cluster to a single cluster node and re-runs layout.
5. Expanding a single cluster of 50 nodes re-layouts in under 300ms.
6. All edges between the expanded members and other visible nodes are rendered correctly after re-layout.
7. All clusters expanded renders identically to the non-clustered mode for the same data.

## Dependencies

- REQ-DASH-075 (cluster rendering must produce the cluster nodes to expand/collapse)
- REQ-DASH-048 (dagre layout engine for re-layout after expand/collapse)

## Blocks

- None currently identified

## Files to Modify

- `internal/dashboard/static/app.js` (expand/collapse handlers, re-layout logic)
- `internal/dashboard/static/styles.css` (collapse button styles)

## Test Strategy

- Unit test: expand cluster replaces cluster node with member nodes and re-runs layout
- Unit test: collapse restores cluster node and re-runs layout
- Performance test: expanding a 50-node cluster re-layouts in under 300ms
- Edge case test: all categories expanded renders identically to non-clustered mode

## Effort

- 0.50 weeks

## Priority

- HIGH

## Phase

- 32
