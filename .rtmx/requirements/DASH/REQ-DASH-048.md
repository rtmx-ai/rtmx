# REQ-DASH-048: Replace Force-Directed Graph with Dagre Layered Layout

## Summary

Replace the `renderGraph()` function in `app.js` that uses D3 force simulation with a dagre-based layered (Sugiyama) layout. Dagre computes node positions using layer assignment (from REQ-DASH-046) and crossing minimization, producing a clean top-to-bottom DAG visualization. D3 remains the rendering engine (SVG elements, zoom, pan), but node coordinates come from dagre instead of force simulation. Edges are rendered as orthogonal or spline paths instead of straight lines.

## Acceptance Criteria

1. `renderGraph()` in `app.js` no longer creates a `d3.forceSimulation`.
2. `renderGraph()` creates a `dagre.graphlib.Graph`, adds nodes and edges from the JSON payload, and calls `dagre.layout()` to compute positions.
3. Nodes are positioned with y-coordinate derived from their layer (layer 0 at top, increasing downward) and x-coordinate from dagre's crossing minimization.
4. Node spacing is configurable: default `ranksep: 80` (vertical gap between layers), `nodesep: 40` (horizontal gap between nodes in same layer).
5. Edges are rendered as SVG `<path>` elements using dagre's computed edge points, producing orthogonal or polyline paths with rounded corners -- not straight lines.
6. Arrow markers remain on edge endpoints, pointing in the correct direction along the path.
7. Zoom and pan (D3 zoom behavior) continue to work.
8. Node drag is removed (layered layout positions are deterministic, dragging would break the layout).
9. Hover behavior (label on mouseover, node enlargement) continues to work.
10. The SVG viewBox auto-sizes to fit the computed layout with padding, rather than using a fixed width/height.
11. Graph renders correctly for 1 node, 10 nodes, 100 nodes, and 300+ nodes without layout errors.
12. Layout computation for 300 nodes completes in under 500ms in a modern browser.

## Dependencies

- REQ-DASH-043 (dagre.js must be vendored and available)
- REQ-DASH-046 (layer data in JSON payload drives vertical positioning)
- REQ-DASH-001 (SPA framework)
- REQ-DASH-010 (design tokens)

## Blocks

- REQ-DASH-049 (direction toggle requires the layout engine to support direction parameter)

## Files to Modify

- `internal/dashboard/static/app.js` (rewrite `renderGraph()` function)
- `internal/dashboard/templates/partials/graph.html` (adjust SVG container sizing if needed)

## Test Strategy

- Unit test (JS): mock dagre and D3, verify renderGraph() calls dagre.layout() with correct node/edge data
- Visual test: screenshot comparison of graph output for testdata/golden database before and after (layout should be strictly better)
- Edge case test: empty graph (0 nodes), single node, two disconnected components, cycle-containing graph
- Performance test: 300-node graph layout completes in under 500ms (browser timing)
- Regression test: zoom, pan, hover label, arrow markers all functional
- Integration test: load graph partial via htmx swap, verify dagre layout is applied (not force simulation)

## Effort

- 1.50 weeks

## Priority

- HIGH

## Phase

- 32
