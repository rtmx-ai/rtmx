# REQ-DASH-084: Dagre Interaction Layer: Zoom, Pan, Hover Preservation (048c)

## Summary

Preserve the existing D3 zoom/pan behavior and node hover interactions (label on mouseover, node enlargement) after the migration from force-directed to dagre layout. Zoom and pan operate on the SVG group containing all nodes and edges. Hover behavior shows the full label on mouseover and applies a subtle scale transform to the hovered node.

## Acceptance Criteria

1. D3 zoom behavior (scroll to zoom, drag to pan) continues to work on the dagre-rendered graph.
2. Zoom and pan operate on the SVG `<g>` container, not on individual nodes.
3. Hover over a node shows the full label text (requirement ID and title) via mouseover event.
4. Hover over a node applies a subtle scale enlargement (1.1x) to the node group.
5. Mouse leave restores the node to its original scale.
6. Node drag is not available (dagre positions are deterministic; this is explicitly removed).
7. Double-click to reset zoom continues to work.

## Dependencies

- REQ-DASH-048 (dagre layout must produce the node/edge SVG structure to attach interactions to)

## Blocks

- None currently identified

## Files to Modify

- `internal/dashboard/static/app.js` (zoom/pan setup, hover event handlers in renderGraph)

## Test Strategy

- Integration test: verify scroll-to-zoom changes the SVG transform scale
- Integration test: verify drag-to-pan changes the SVG transform translate
- Integration test: verify mouseover on a node shows the label text
- Integration test: verify mouseleave restores original node scale
- Regression test: double-click resets zoom to fit the layout

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 32
