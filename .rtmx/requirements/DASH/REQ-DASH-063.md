# REQ-DASH-063: Minimap Overview Panel

## Summary

A minimap panel shall render in the bottom-right corner of the graph view, showing
a scaled-down overview of the entire graph with a viewport rectangle indicating the
currently visible area. Clicking or dragging on the minimap pans the main graph view
to the corresponding position. This aids navigation in large graphs where the full
extent is not visible at the current zoom level.

## Acceptance Criteria

1. A minimap panel renders in the bottom-right corner of the graph container, 200x150px with `--color-surface-secondary` background
2. The minimap displays a proportionally scaled rendering of all graph nodes and edges
3. A semi-transparent viewport rectangle (`--color-emerald-500` at 0.3 opacity, 1px border) shows the currently visible area
4. The viewport rectangle updates in real time as the user pans or zooms the main graph
5. Clicking a position on the minimap pans the main graph to center on that position
6. Dragging on the minimap continuously pans the main graph to follow the drag
7. The minimap is collapsible via a toggle button; collapsed state persists in localStorage
8. The minimap re-renders when the graph data changes (filtering, grouping, layout changes)
9. The minimap does not render when the full graph fits within the viewport at the current zoom level
10. Minimap rendering uses simplified shapes (rectangles only, no labels) for performance

## Dependencies

- REQ-DASH-048 (dagre layout engine provides graph extent and node positions for minimap rendering)

## Blocks

- None

## Files to Modify

- `internal/dashboard/static/graph.js`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: minimap renders proportionally scaled node positions matching the main graph
- Unit test: viewport rectangle dimensions and position match the visible area of the main graph
- Unit test: clicking minimap computes correct pan target and applies transform to main graph
- Unit test: minimap hides when full graph fits in viewport
- Unit test: collapsed state persists in localStorage across page loads
- Integration test: minimap updates in sync with pan, zoom, filter, and group-by changes

## Effort

- 0.75 weeks

## Priority

- LOW

## Phase

- 32
