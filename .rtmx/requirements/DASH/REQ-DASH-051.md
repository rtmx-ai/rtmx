# REQ-DASH-051: Node Fill Color Encodes Status

## Summary

Graph nodes shall use fill color to encode requirement status, using the four status colors already defined in the design token system. This formalizes the existing behavior into a well-defined mapping with consistent token references, ensuring the graph uses the same status palette as the table view and stat bar. The mapping is: COMPLETE=#22c55e, PARTIAL=#f59e0b, MISSING=#ef4444, NOT_STARTED=#6b7280.

## Acceptance Criteria

1. `renderGraph()` sets each node circle's `fill` attribute based on the `status` field from the JSON payload.
2. Status-to-color mapping uses CSS custom properties from design tokens:
   - `COMPLETE` -> `var(--rtmx-status-complete)` (#22c55e)
   - `PARTIAL` -> `var(--rtmx-status-partial)` (#f59e0b)
   - `MISSING` -> `var(--rtmx-status-missing)` (#ef4444)
   - `NOT_STARTED` -> `var(--rtmx-status-not-started)` (#6b7280)
3. Unknown or missing status values default to `var(--rtmx-text-dim)` (#6b7280).
4. Fill colors are applied via CSS classes (e.g., `.node-complete`, `.node-partial`) rather than inline style attributes, enabling future theme overrides.
5. The status color mapping is defined once in `styles.css` and referenced by both the graph nodes and the legend panel.
6. Fill color is independent of node size, border, and opacity -- those are governed by other visual encoding requirements.
7. The legend panel includes a "Fill = Status" entry showing four colored circles with status labels.

## Dependencies

- REQ-DASH-044 (enriched graph JSON provides status per node)
- REQ-DASH-048 (dagre layout engine)
- REQ-DASH-010 (design tokens define the four status colors)

## Blocks

- REQ-DASH-057 (legend must reference the status color encoding)

## Files to Modify

- `internal/dashboard/static/app.js` (apply status-based CSS class to each node circle in renderGraph)
- `internal/dashboard/static/styles.css` (define .node-complete, .node-partial, .node-missing, .node-not-started classes)

## Test Strategy

- Unit test (JS): verify each status value maps to the correct CSS class
- Unit test (JS): verify unknown status defaults to not-started color
- Unit test (CSS): parse styles.css and verify all four .node-* classes reference the correct design token variables
- Visual test: screenshot comparison showing four distinct node colors for a database with all four status values
- Regression test: existing table and stat bar status colors unchanged

## Effort

- 0.50 weeks

## Priority

- HIGH

## Phase

- 32
