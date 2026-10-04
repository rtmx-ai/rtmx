# REQ-DASH-049: Graph Layout Direction Toggle (Top-to-Bottom / Left-to-Right)

## Summary

The dagre layered graph layout shall support two direction modes: top-to-bottom (TB) and left-to-right (LR), selectable via a UI toggle in the graph partial. The default is TB (roots at top, leaves at bottom). LR mode places roots at left and leaves at right, which is preferred for wide/shallow graphs. The toggle persists in Alpine.js client-side state so the preference survives htmx partial reloads within the same session.

## Acceptance Criteria

1. The graph partial template includes a toggle control (button or select) with options "Top-to-Bottom" and "Left-to-Right".
2. Clicking the toggle re-renders the graph with the selected direction by calling `renderGraph()` with a direction parameter.
3. `renderGraph(data, options)` accepts an `options.rankdir` parameter ('TB' or 'LR') and passes it to `dagre.layout()` via `graph.setGraph({ rankdir: ... })`.
4. In TB mode: layer 0 nodes are at the top, dependency arrows point downward.
5. In LR mode: layer 0 nodes are at the left, dependency arrows point rightward.
6. The SVG viewBox adjusts to the layout dimensions for each direction (TB graphs are taller, LR graphs are wider).
7. The selected direction is stored in Alpine.js component state and survives htmx content swaps within the same page session.
8. Default direction is TB when no preference has been set.
9. Arrow markers orient correctly for both directions.
10. The legend and stat bar remain functional in both directions.

## Dependencies

- REQ-DASH-048 (dagre layout must be implemented before direction can be toggled)
- REQ-DASH-001 (SPA framework)
- REQ-DASH-010 (design tokens)

## Blocks

- None currently identified

## Files to Modify

- `internal/dashboard/static/app.js` (add direction parameter to renderGraph, adjust viewBox logic)
- `internal/dashboard/templates/partials/graph.html` (add direction toggle UI element)

## Test Strategy

- Unit test (JS): verify renderGraph with rankdir='LR' passes 'LR' to dagre graph config
- Unit test (JS): verify renderGraph with rankdir='TB' (or default) passes 'TB' to dagre graph config
- Visual test: screenshot comparison of same graph in TB vs LR mode, verify root positions differ
- Integration test: toggle direction, verify graph re-renders without page reload
- State test: toggle to LR, trigger htmx category filter reload, verify LR direction persists

## Effort

- 0.50 weeks

## Priority

- MEDIUM

## Phase

- 32
