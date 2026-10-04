# REQ-DASH-091: Reusable Tooltip Component for Graph and Timeline

## Summary

A standalone Alpine.js tooltip component used across multiple views: graph node hover (REQ-DASH-059), blocked node tooltip (REQ-DASH-054), and timeline bar hover (REQ-DASH-086). The tooltip positions itself near the cursor, avoids viewport edges with automatic repositioning, and uses design token styling. It accepts structured content (key-value pairs) and renders consistently across all consumers.

## Acceptance Criteria

1. The tooltip is implemented as a reusable Alpine.js component (`x-data="tooltip"`) that can be invoked from any view.
2. The tooltip accepts structured content: an array of key-value pairs (e.g., [{label: "ID", value: "REQ-DASH-048"}, ...]).
3. The tooltip positions itself near the cursor (offset by 12px right and 12px below).
4. When the tooltip would overflow the viewport on the right or bottom, it repositions to the opposite side of the cursor.
5. The tooltip uses design token styling: `var(--rtmx-bg-card)` background, `var(--rtmx-border)` border, `var(--rtmx-text)` text, `var(--rtmx-shadow)` box-shadow.
6. The tooltip appears on mouseover and disappears on mouseleave with no delay.
7. The tooltip has a max-width of 320px and wraps long text values.
8. The tooltip is accessible: it has `role="tooltip"` and the triggering element references it via `aria-describedby`.

## Dependencies

- REQ-DASH-010 (design tokens for styling)

## Blocks

- REQ-DASH-059 (graph node hover tooltip consumes this component)

## Files to Modify

- `internal/dashboard/static/app.js` (Alpine.js tooltip component definition)
- `internal/dashboard/static/styles.css` (.tooltip styles with design tokens)

## Test Strategy

- Unit test (JS): tooltip renders with provided key-value content
- Unit test (JS): tooltip repositions when it would overflow viewport edges
- Unit test (JS): tooltip appears on mouseover and disappears on mouseleave
- Unit test (CSS): tooltip uses design token custom properties
- Accessibility test: tooltip has role="tooltip" and aria-describedby linkage

## Effort

- 0.50 weeks

## Priority

- HIGH

## Phase

- 32
