# REQ-DASH-064: Zoom Controls and Fit-to-View Toolbar

## Summary

The graph view shall include a zoom control toolbar with zoom-in, zoom-out, and
fit-to-view buttons, accompanied by a zoom level indicator. Keyboard shortcuts
(+/- for zoom, 0 for fit-to-view) provide quick access. The fit-to-view action
calculates the bounding box of all visible nodes and adjusts the viewport transform
to display the entire graph with appropriate padding.

## Acceptance Criteria

1. A toolbar in the graph view contains three buttons: zoom-in (+), zoom-out (-), and fit-all
2. Zoom-in increments zoom by 25% of current level; zoom-out decrements by 25%
3. Zoom is clamped between 10% and 400%
4. Fit-all calculates the bounding box of all visible (non-filtered) nodes and sets the transform to fit with 40px padding
5. A zoom level indicator displays the current percentage (e.g., "100%") adjacent to the toolbar
6. Keyboard shortcut `+` or `=` zooms in; `-` zooms out; `0` fits to view
7. Keyboard shortcuts are disabled when an input field, textarea, or the command palette has focus
8. Zoom transitions are animated with 200ms ease-out timing
9. Toolbar buttons use design token icon styling and hover states
10. Zoom level persists during the session but resets to fit-all on page load

## Dependencies

- REQ-DASH-048 (dagre layout engine provides node positions for bounding box calculation)

## Blocks

- None

## Files to Modify

- `internal/dashboard/static/graph.js`
- `internal/dashboard/templates/partials/graph.html`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: zoom-in button increases zoom by 25% of current level
- Unit test: zoom-out button decreases zoom by 25% of current level
- Unit test: zoom clamps at 10% minimum and 400% maximum
- Unit test: fit-all computes correct transform from node bounding box with 40px padding
- Unit test: keyboard shortcuts +/-/0 trigger corresponding zoom actions
- Unit test: keyboard shortcuts are suppressed when input elements have focus
- Integration test: zoom level indicator updates in real time during zoom operations

## Effort

- 0.50 weeks

## Priority

- HIGH

## Phase

- 32
