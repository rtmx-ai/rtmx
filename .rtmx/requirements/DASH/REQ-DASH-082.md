# REQ-DASH-082: URL Query Parameter Encoding for Graph Settings (082a)

## Summary

All graph view settings shall be serialized into URL query parameters, making graph configurations bookmarkable and shareable. The encoding covers view mode, group-by, layout direction, zoom level, pan position, active filters, and expanded clusters. Parameters with default values are omitted to keep URLs clean. This requirement handles the serialization (write) side only; deserialization and browser history integration are handled by REQ-DASH-090.

## Acceptance Criteria

1. The following graph settings are encoded as URL query parameters: `view` (graph/table), `groupBy` (category/status/phase/none), `dir` (TB/LR), `zoom` (float), `panX` (float), `panY` (float), `filter` (comma-separated status values), `expanded` (comma-separated category names).
2. Changing any graph setting calls `history.pushState()` with the updated query string.
3. Parameters with default values are omitted from the URL to keep it clean (e.g., `dir=TB` is default and not shown).
4. Invalid parameter values are silently ignored and the default is used (no error, no crash).
5. The URL is human-readable: `/dashboard/graph?groupBy=category&filter=BLOCKED,IN_PROGRESS&expanded=CLI,DASH`.
6. Zoom/pan pushState calls are debounced (not on every pixel movement).

## Dependencies

- REQ-DASH-048 (dagre layout provides the graph settings to persist)
- REQ-DASH-029 (existing pushState/popstate URL pattern for detail panel)

## Blocks

- REQ-DASH-090 (URL state restore depends on the encoding format defined here)

## Files to Modify

- `internal/dashboard/static/app.js` (URL serialization functions, pushState on setting change)

## Test Strategy

- Unit test: changing groupBy calls pushState with updated query parameter
- Unit test: changing filter calls pushState with comma-separated filter values
- Unit test: changing zoom/pan calls pushState with float values (debounced)
- Unit test: expanding a cluster adds it to the `expanded` query parameter
- Unit test: default values are not included in the URL string
- Unit test: invalid parameter values fall back to defaults without error

## Effort

- 0.50 weeks

## Priority

- MEDIUM

## Phase

- 32
