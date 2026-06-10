# REQ-DASH-083: Dagre Edge Path Rendering (048b)

## Summary

Render edges as SVG `<path>` elements using dagre's computed edge points, producing polyline paths with rounded corners instead of straight lines. Each edge path follows the control points output by dagre layout, with arrow markers at the endpoint pointing in the correct direction along the path. Edge styling distinguishes critical path edges from regular edges.

## Acceptance Criteria

1. Edges are rendered as SVG `<path>` elements using dagre's computed edge points array.
2. Edge paths use polyline interpolation with rounded corners (corner radius of 5px).
3. Arrow markers remain on edge endpoints, pointing in the correct direction along the final path segment.
4. Edges are not rendered as straight lines between node centers.
5. Critical path edges are visually distinct (bold, emerald, animated) per existing encoding (REQ-DASH-053).
6. Regular edges use thin gray styling per design tokens.
7. Edge paths render correctly for edges spanning multiple layers and for edges between nodes in the same layer.

## Dependencies

- REQ-DASH-048 (dagre layout must compute edge points before rendering)
- REQ-DASH-053 (critical path edge encoding defines styling)
- REQ-DASH-010 (design tokens)

## Blocks

- None currently identified

## Files to Modify

- `internal/dashboard/static/app.js` (edge path rendering function in renderGraph)
- `internal/dashboard/static/styles.css` (edge path styles, arrow marker styles)

## Test Strategy

- Unit test (JS): verify edge paths are constructed from dagre edge point arrays, not as straight lines
- Unit test (JS): verify arrow markers point in the correct direction on the final segment
- Visual test: screenshot comparison of edge rendering for test fixture data
- Edge case test: self-referencing edge, edge spanning 5+ layers, edge between adjacent nodes

## Effort

- 0.50 weeks

## Priority

- HIGH

## Phase

- 32
