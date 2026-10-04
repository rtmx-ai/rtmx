# REQ-DASH-054: Blocked Node Visual Treatment

## Summary

Nodes representing blocked requirements shall be visually distinguished with a dashed border and reduced opacity, signaling that the requirement cannot currently proceed. On hover, the tooltip shows the specific blocking dependencies. This encoding is orthogonal to status fill color and priority ring -- a blocked node retains its status color and priority ring but gains a dashed stroke pattern and semi-transparent appearance.

## Acceptance Criteria

1. `renderGraph()` reads the `is_blocked` field (boolean) from each node in the enriched JSON payload.
2. Blocked nodes receive these visual treatments:
   - `stroke-dasharray: 5 3` on the node circle, creating a dashed border effect
   - `opacity: 0.7` on the entire node group (circle + label), reducing visual prominence
3. The dashed border is applied in addition to the priority ring stroke. For a blocked P0 node, the red 3px ring becomes a dashed red 3px ring.
4. Non-blocked nodes have no stroke-dasharray (solid border or no border per priority).
5. Hovering over a blocked node shows a tooltip listing the blocking requirement IDs. The blocking dependencies are derived from the enriched JSON (the backend provides blocking_ids or they are computed from the dependency graph in the frontend).
6. The tooltip uses design token colors: `var(--rtmx-bg-elevated)` background, `var(--rtmx-text)` foreground, `var(--rtmx-border)` border.
7. The reduced opacity does not affect edge visibility -- edges connected to blocked nodes remain at full opacity.
8. CSS class `.node-blocked` is defined in `styles.css` with the dasharray and opacity rules.
9. The legend panel includes a "Dashed = Blocked" entry showing a dashed-border circle vs. solid circle.

## Dependencies

- REQ-DASH-044 (enriched graph JSON provides is_blocked per node)
- REQ-DASH-048 (dagre layout engine)
- REQ-DASH-010 (design tokens for tooltip styling)

## Blocks

- REQ-DASH-057 (legend must reference the blocked node encoding)

## Files to Modify

- `internal/dashboard/static/app.js` (apply blocked styling in renderGraph, add tooltip with blocking deps)
- `internal/dashboard/static/styles.css` (define .node-blocked class with stroke-dasharray and opacity)

## Test Strategy

- Unit test (JS): verify nodes with is_blocked=true receive the .node-blocked CSS class
- Unit test (JS): verify nodes with is_blocked=false do not have dashed border or reduced opacity
- Unit test (CSS): verify .node-blocked defines stroke-dasharray: 5 3 and opacity: 0.7
- Combination test: blocked P0 node has both dashed border and red stroke color
- Tooltip test: hover over blocked node, verify tooltip appears with correct blocking dependency IDs
- Visual test: screenshot comparison showing blocked vs. unblocked nodes side by side
- Edge case test: node that is both blocked and a cycle member (both visual treatments apply)

## Effort

- 0.75 weeks

## Priority

- HIGH

## Phase

- 32
