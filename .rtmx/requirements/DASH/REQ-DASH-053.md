# REQ-DASH-053: Critical Path Edge Rendering

## Summary

Edges that lie on the critical path shall be visually distinguished from normal dependency edges using bold width, emerald accent color, and an animated dash pattern. This highlights the longest dependency chain in the roadmap, drawing immediate attention to the sequence of requirements that determines the project's minimum completion time. Normal edges remain thin and gray.

## Acceptance Criteria

1. `renderGraph()` reads the `is_critical_path` boolean from each edge in the enriched JSON payload.
2. Critical path edges are rendered with:
   - stroke-width: 3px (normal edges: 1px)
   - stroke color: `var(--rtmx-emerald)` (#10b981)
   - stroke-dasharray: `8 4` with a CSS animation that offsets the dash pattern, creating a flowing/marching-ants effect
   - Arrow marker fill matches `var(--rtmx-emerald)`
3. Normal (non-critical) edges retain:
   - stroke-width: 1px
   - stroke color: `var(--rtmx-border)` (#262626)
   - No dash animation
   - Arrow marker fill matches `var(--rtmx-border)`
4. The CSS animation uses `@keyframes` with `stroke-dashoffset` and runs at a subtle speed (duration: 2s, linear, infinite).
5. Critical path edges render on top of normal edges (higher SVG z-order) so they are never occluded.
6. The animation can be paused via a `prefers-reduced-motion` media query for accessibility.
7. The legend panel includes an "Edge = Critical Path" entry showing bold emerald animated line vs. thin gray line.

## Dependencies

- REQ-DASH-044 (enriched graph JSON provides is_critical_path per edge)
- REQ-DASH-048 (dagre layout engine computes edge paths)
- REQ-DASH-010 (design tokens for emerald and border colors)

## Blocks

- REQ-DASH-057 (legend must reference the critical path edge encoding)

## Files to Modify

- `internal/dashboard/static/app.js` (apply critical path styling to edges in renderGraph, set SVG render order)
- `internal/dashboard/static/styles.css` (define .edge-critical class with animation keyframes, prefers-reduced-motion override)

## Test Strategy

- Unit test (JS): verify edges with is_critical_path=true receive the .edge-critical CSS class
- Unit test (JS): verify edges with is_critical_path=false receive default edge styling
- Unit test (CSS): parse styles.css and verify .edge-critical defines stroke-width, stroke color, and animation
- Visual test: screenshot comparison showing emerald bold edges vs. gray thin edges
- Accessibility test: verify animation is paused when prefers-reduced-motion is active
- SVG order test: verify critical path edge elements appear after normal edges in the DOM (higher z-order)
- Edge case test: graph with no critical path (all edges are normal), graph where all edges are critical

## Effort

- 0.75 weeks

## Priority

- HIGH

## Phase

- 32
