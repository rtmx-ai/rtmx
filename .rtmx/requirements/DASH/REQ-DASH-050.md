# REQ-DASH-050: Node Size Proportional to Effort

## Summary

Graph nodes shall be rendered as circles with radius proportional to the requirement's `effort_weeks` value, providing an immediate visual indication of relative effort across the roadmap. The radius scales linearly from 8px (for effort <= 0.25 weeks) to 28px (for effort >= 4 weeks), with a default of 12px when effort is unset. A corresponding legend entry shows the size scale with labeled reference circles.

## Acceptance Criteria

1. `renderGraph()` reads `effort_weeks` from each node in the enriched JSON payload.
2. Node circle radius is computed as: `clamp(8, 8 + (effort_weeks - 0.25) * (20 / 3.75), 28)` -- linear interpolation from 8px at 0.25w to 28px at 4w.
3. Nodes with `effort_weeks` of 0, null, or undefined render at a default radius of 12px.
4. Nodes with `effort_weeks` below 0.25 render at the minimum radius of 8px.
5. Nodes with `effort_weeks` above 4.0 render at the maximum radius of 28px.
6. Dagre node dimensions are set to match the computed radius (width = height = 2 * radius) so the layout engine allocates correct spacing.
7. The legend panel includes a "Size = Effort" entry showing three reference circles: small (0.5w), medium (2w), and large (4w) with labels.
8. Node size does not affect fill color, border, or label positioning -- those are governed by other visual encoding requirements.

## Dependencies

- REQ-DASH-044 (enriched graph JSON provides effort_weeks per node)
- REQ-DASH-048 (dagre layout engine must be in place for node dimension integration)
- REQ-DASH-010 (design tokens for legend styling)

## Blocks

- REQ-DASH-057 (legend must reference the size encoding)

## Files to Modify

- `internal/dashboard/static/app.js` (add effort-to-radius mapping in renderGraph, set dagre node dimensions)

## Test Strategy

- Unit test (JS): verify radius computation for boundary values: 0, 0.25, 1.0, 2.0, 4.0, 8.0, null, undefined
- Unit test (JS): verify dagre node width/height matches 2 * computed radius
- Visual test: screenshot comparison showing visually distinct node sizes for a database with varied effort values
- Regression test: hover, zoom, pan behaviors remain functional with variable-size nodes
- Edge case test: all nodes with same effort render at same size, layout remains clean

## Effort

- 0.75 weeks

## Priority

- HIGH

## Phase

- 32
