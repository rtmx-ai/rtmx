# REQ-DASH-082: Graph State Persistence via URL Query Parameters

## Summary

All graph view settings shall be encoded in URL query parameters, making graph configurations
bookmarkable and shareable. Browser back/forward navigation preserves and restores graph state.
Settings include view mode, group-by, layout direction, zoom level, active filters, and
expanded clusters.

## Acceptance Criteria

1. The following graph settings are encoded as URL query parameters: `view` (graph/table), `groupBy` (category/status/phase/none), `dir` (TB/LR), `zoom` (float), `panX` (float), `panY` (float), `filter` (comma-separated status values), `expanded` (comma-separated category names)
2. Changing any graph setting calls `history.pushState()` with the updated query string
3. Loading a URL with graph query parameters restores all encoded settings on page load
4. Browser back button restores the previous graph state (view, groupBy, filters, zoom, etc.)
5. Browser forward button re-applies the state that was navigated away from
6. Parameters with default values are omitted from the URL to keep it clean (e.g., `dir=TB` is default and not shown)
7. Invalid parameter values are silently ignored and the default is used (no error, no crash)
8. The URL is human-readable: `/dashboard/graph?groupBy=category&filter=BLOCKED,IN_PROGRESS&expanded=CLI,DASH`
9. Copying the URL and opening it in a new tab reproduces the exact graph state
10. URL state integrates with the existing pushState pattern from REQ-DASH-029 (detail panel URL takes precedence when panel is open)

## Dependencies

- REQ-DASH-048 (dagre layout provides the graph settings to persist)
- REQ-DASH-029 (existing pushState/popstate URL pattern for detail panel)

## Blocks

- None

## Files to Modify

- `internal/dashboard/static/app.js` (URL serialization/deserialization, pushState on setting change, popstate handler for graph state)
- `internal/dashboard/templates/partials/graph.html` (read URL params on partial load to initialize settings)

## Test Strategy

- Unit test: changing groupBy calls pushState with updated query parameter
- Unit test: changing filter calls pushState with comma-separated filter values
- Unit test: changing zoom/pan calls pushState with float values (debounced, not on every pixel)
- Unit test: expanding a cluster adds it to the `expanded` query parameter
- Unit test: loading URL with `?groupBy=status&filter=BLOCKED` initializes graph with those settings
- Unit test: default values are not included in the URL string
- Unit test: invalid parameter values (e.g., `groupBy=invalid`) fall back to defaults without error
- Unit test: popstate event restores graph settings from the popped URL
- Integration test: full round-trip -- set state, copy URL, load in new context, verify identical graph
- Integration test: back/forward navigation cycles through multiple graph state changes
- Edge case test: URL with no query parameters loads graph with all defaults

## Effort

- 1.00 weeks

## Priority

- MEDIUM

## Phase

- 32
