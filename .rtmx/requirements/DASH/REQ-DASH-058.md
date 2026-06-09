# REQ-DASH-058: Click Graph Node to Open Detail Panel

## Summary

Clicking a node in the dependency graph shall open the existing detail slide-in panel
populated with that requirement's full information. This reuses the REQ-DASH-025-030
detail panel infrastructure by dispatching an htmx GET to `/partials/detail/{id}` and
swapping the response into `#detail-panel-body`, providing seamless navigation from the
visual graph to structured requirement data without leaving the graph view.

## Acceptance Criteria

1. Each rendered SVG node element has a click event listener registered during graph initialization
2. Clicking a node dispatches `hx-get /partials/detail/{id}` targeting `#detail-panel-body` with `hx-swap="innerHTML"`
3. The detail panel opens (Alpine `panelOpen = true`, `panelReqId = id`) with the same transition as table-row clicks
4. Clicking a different node while the panel is open swaps the panel content without closing and reopening
5. Clicking the same node that is already displayed in the panel closes the panel (toggle behavior)
6. The clicked node receives a visual selected state (2px emerald border or stroke) that clears when the panel closes
7. Panel close (Escape or close button) clears the node selected state
8. Works correctly regardless of current zoom/pan transform applied to the SVG

## Dependencies

- REQ-DASH-048 (dagre layout engine renders clickable node elements)
- REQ-DASH-025 (detail panel container with Alpine toggle state)

## Blocks

- None

## Files to Modify

- `internal/dashboard/static/graph.js`
- `internal/dashboard/static/app.js`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: clicking a node element dispatches htmx request with correct requirement ID
- Unit test: clicking the same node toggles panel closed
- Unit test: clicking a different node swaps content without close/reopen cycle
- Unit test: node selected state applies and clears correctly on panel open/close
- Integration test: full click-to-panel flow with mock server returning detail partial

## Effort

- 0.50 weeks

## Priority

- HIGH

## Phase

- 32
