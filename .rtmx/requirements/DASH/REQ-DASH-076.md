# REQ-DASH-076: Graph Keyboard Navigation via Edge Connections

## Summary

The graph view shall support arrow-key navigation that moves focus between nodes following
the dependency graph structure. Right and Down arrow keys follow dependency edges (move to
nodes this requirement depends on), Left and Up arrow keys follow blocker edges (move to
nodes that depend on this requirement). Enter opens the detail panel for the focused node.
Tab cycles through graph toolbar controls. A visible focus ring highlights the active node.

## Acceptance Criteria

1. Pressing Tab from the page focuses the graph SVG container, then cycles through graph toolbar controls (zoom, group-by, direction, export buttons)
2. When the graph SVG is focused, arrow keys move focus between nodes: Right/Down follows outgoing dependency edges, Left/Up follows incoming blocker edges
3. If multiple edges exist in the arrow direction, focus moves to the nearest node (by dagre x/y position in the arrow direction)
4. If no edge exists in the arrow direction, focus does not move (no wrap-around)
5. The focused node displays a 2px emerald focus ring using `var(--rtmx-emerald)` and the ring is visible at all zoom levels
6. The focused node auto-pans into view if it is outside the current viewport
7. Enter on a focused node opens the detail panel for that requirement
8. Escape returns focus from the detail panel back to the previously focused graph node
9. Home key focuses the root node (node with no incoming edges, or first node in topological order)
10. Focus state is announced to screen readers via aria-activedescendant on the SVG container

## Dependencies

- REQ-DASH-048 (dagre layout provides edge structure and node positions for navigation)
- REQ-DASH-035 (keyboard framework patterns: j/k selection model, focus management conventions)

## Blocks

- REQ-DASH-077 (screen reader support builds on keyboard focus infrastructure)

## Files to Modify

- `internal/dashboard/static/app.js` (keyboard event handlers, focus management, auto-pan logic)
- `internal/dashboard/static/styles.css` (focus ring styles for graph nodes)
- `internal/dashboard/templates/partials/graph.html` (tabindex, aria-activedescendant on SVG)

## Test Strategy

- Unit test: Right arrow from node A follows dependency edge to node B
- Unit test: Left arrow from node B follows blocker edge back to node A
- Unit test: multiple outgoing edges selects nearest node by position
- Unit test: no edge in direction keeps focus on current node
- Unit test: Enter on focused node triggers detail panel open
- Unit test: Escape from detail panel returns focus to graph node
- Unit test: Home key focuses the root node
- Unit test: focused node has emerald focus ring class applied
- Integration test: Tab cycles through toolbar controls then into graph
- Integration test: auto-pan scrolls viewport to keep focused node visible

## Effort

- 1.25 weeks

## Priority

- HIGH

## Phase

- 32
