# REQ-DASH-081: Graph Loading State and Layout Error Handling

## Summary

The graph view shall display a loading state (skeleton or spinner) while the dagre layout
computes, and shall gracefully handle layout failures by falling back to a simple list view
with an error toast. A 5-second timeout prevents the UI from hanging on pathologically large
or malformed graphs.

## Acceptance Criteria

1. When the graph partial loads, a skeleton placeholder (gray rectangles mimicking a graph layout) is shown while dagre layout computes
2. The skeleton is replaced by the rendered graph once layout completes
3. If dagre layout takes longer than 5 seconds, layout is aborted and the view falls back to a simple requirement list (table format)
4. On timeout, an error toast displays: "Graph layout timed out. Showing list view instead."
5. If dagre.layout() throws an error (e.g., invalid graph input), the exception is caught, the view falls back to list view, and an error toast displays: "Graph layout failed: {error message}"
6. The fallback list view shows requirement ID, title, status, and dependency count in a simple table
7. The fallback list view includes a "Retry graph" button that re-attempts the dagre layout
8. The loading skeleton animates with a subtle pulse effect to indicate activity
9. Layout computation runs asynchronously (via requestAnimationFrame or setTimeout(0)) to avoid blocking the main thread
10. The loading/error state is logged to the browser console with timing information for debugging

## Dependencies

- REQ-DASH-048 (dagre layout is the operation being monitored)
- REQ-DASH-014 (toast notification system for error display)

## Blocks

- None

## Files to Modify

- `internal/dashboard/static/app.js` (async layout wrapper, timeout logic, error handling, fallback rendering)
- `internal/dashboard/static/styles.css` (skeleton placeholder styles, pulse animation)
- `internal/dashboard/templates/partials/graph.html` (skeleton markup, fallback list container, retry button)

## Test Strategy

- Unit test: skeleton placeholder is visible before dagre layout completes
- Unit test: skeleton is removed and graph is visible after layout completes
- Unit test: layout exceeding 5 seconds triggers timeout, shows list fallback and error toast
- Unit test: dagre.layout() exception is caught, shows list fallback and error toast with message
- Unit test: retry button re-invokes renderGraph()
- Unit test: layout computation is called via requestAnimationFrame or setTimeout (non-blocking)
- Integration test: simulated slow layout (mock dagre with delay) triggers timeout fallback
- Integration test: simulated error (mock dagre throws) triggers error fallback
- Edge case test: empty graph (0 nodes) skips layout and shows empty state, no error

## Effort

- 0.75 weeks

## Priority

- HIGH

## Phase

- 32
