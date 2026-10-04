# REQ-DASH-052: Node Border Ring Encodes Priority

## Summary

Graph nodes shall display a colored border ring whose thickness and color encode the requirement's priority level. This provides a second visual dimension on each node (in addition to fill color for status) without conflicting visually. The mapping is: P0 = red 3px ring, HIGH = orange 2px ring, MEDIUM = blue 1px ring, LOW = no ring. Colors reference design token values.

## Acceptance Criteria

1. `renderGraph()` reads the `priority` field from each node in the enriched JSON payload.
2. Priority-to-ring mapping:
   - `P0` -> stroke: `var(--rtmx-status-missing)` (#ef4444), stroke-width: 3px
   - `HIGH` -> stroke: `var(--rtmx-status-partial)` (#f59e0b), stroke-width: 2px
   - `MEDIUM` -> stroke: #3b82f6 (blue-500), stroke-width: 1px
   - `LOW` -> stroke: none (no visible border ring)
3. Nodes with null, undefined, or empty priority are treated as LOW (no ring).
4. The ring is rendered as the SVG circle's `stroke` and `stroke-width` attributes, outside the fill area so it does not reduce the visible fill region.
5. Ring width is added to dagre node dimensions to prevent overlap (node width = 2 * radius + 2 * stroke-width).
6. The ring is visible regardless of node size (effort) and does not conflict with the dashed border used for blocked nodes (REQ-DASH-054).
7. Priority ring CSS classes are defined in `styles.css`: `.priority-p0`, `.priority-high`, `.priority-medium`, `.priority-low`.
8. The legend panel includes a "Border = Priority" entry showing ring examples for P0, HIGH, MEDIUM, and LOW.

## Dependencies

- REQ-DASH-044 (enriched graph JSON provides priority per node)
- REQ-DASH-048 (dagre layout engine)
- REQ-DASH-010 (design tokens for red and orange colors)

## Blocks

- REQ-DASH-057 (legend must reference the priority ring encoding)

## Files to Modify

- `internal/dashboard/static/app.js` (apply priority-based stroke to each node circle in renderGraph, adjust dagre node dimensions)
- `internal/dashboard/static/styles.css` (define .priority-p0, .priority-high, .priority-medium, .priority-low classes)

## Test Strategy

- Unit test (JS): verify each priority value maps to the correct CSS class and stroke-width
- Unit test (JS): verify null/empty priority defaults to no ring
- Unit test (JS): verify dagre node dimensions account for stroke-width
- Visual test: screenshot comparison showing distinct ring colors and widths on nodes with different priorities
- Combination test: verify ring is visible on nodes of all sizes (small 8px and large 28px)
- Interaction test: priority ring does not interfere with blocked dashed border (both can be active on a P0 blocked node)

## Effort

- 0.75 weeks

## Priority

- HIGH

## Phase

- 32
