# REQ-DASH-048: Dagre Node Layout Computation (048a)

## Summary

Integrate the dagre library into `renderGraph()` in `app.js` to compute node positions using a layered (Sugiyama) layout. The function creates a `dagre.graphlib.Graph`, adds nodes and edges from the JSON payload, and calls `dagre.layout()` to produce x/y coordinates for each node. Force simulation is removed. Nodes are positioned with y derived from layer assignment and x from crossing minimization. The SVG viewBox auto-sizes to fit the computed layout with padding.

## Acceptance Criteria

1. `renderGraph()` in `app.js` no longer creates a `d3.forceSimulation`.
2. `renderGraph()` creates a `dagre.graphlib.Graph`, adds nodes and edges from the JSON payload, and calls `dagre.layout()` to compute positions.
3. Nodes are positioned with y-coordinate derived from their layer (layer 0 at top, increasing downward) and x-coordinate from dagre's crossing minimization.
4. Node spacing is configurable: default `ranksep: 80` (vertical gap between layers), `nodesep: 40` (horizontal gap between nodes in same layer).
5. Node drag is removed (layered layout positions are deterministic, dragging would break the layout).
6. The SVG viewBox auto-sizes to fit the computed layout with padding, rather than using a fixed width/height.
7. Graph renders correctly for 1 node, 10 nodes, 100 nodes, and 300+ nodes without layout errors.
8. Arrow markers remain on edge endpoints, pointing in the correct direction.

## Dependencies

- REQ-DASH-043 (dagre.js must be vendored and available)
- REQ-DASH-046 (layer data in JSON payload drives vertical positioning)
- REQ-DASH-001 (SPA framework)
- REQ-DASH-010 (design tokens)

## Blocks

- REQ-DASH-049 (direction toggle requires the layout engine to support direction parameter)
- REQ-DASH-083 (edge path rendering depends on dagre layout output)
- REQ-DASH-084 (interaction layer depends on dagre layout output)
- REQ-DASH-094 (performance tests target dagre layout computation)

## Files to Modify

- `internal/dashboard/static/app.js` (rewrite `renderGraph()` function)

## Test Strategy

- Unit test (JS): mock dagre and D3, verify renderGraph() calls dagre.layout() with correct node/edge data
- Edge case test: empty graph (0 nodes), single node, two disconnected components
- Integration test: load graph partial via htmx swap, verify dagre layout is applied (not force simulation)
- Regression test: arrow markers functional after layout change

## Effort

- 0.75 weeks

## Priority

- HIGH

## Phase

- 32
