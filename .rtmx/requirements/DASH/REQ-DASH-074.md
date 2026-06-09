# REQ-DASH-074: Shared View State

## Summary

All three visualization views (DAG, Timeline, Treemap) share a common filter state so that switching between views preserves active filters. The shared state includes status filter, priority filter, category filter, and search text. Filters are stored as URL query parameters, and when the view switcher changes the active view, all existing filter parameters are carried forward in the htmx request. This ensures the user can filter to a subset of requirements in one view and immediately see the same subset in another view.

## Acceptance Criteria

1. The following filter parameters are shared across all three views: status, priority, category, search (text substring match).
2. When switching views via the view switcher, all active filter query parameters are included in the htmx GET request to the new view partial.
3. The URL at all times reflects the complete state: ?view=timeline&status=MISSING&category=DASH (for example).
4. The Go handlers for all three view partials (dag, timeline, treemap) read and apply the same filter parameters, filtering the requirement set before building the view-specific data structure.
5. Filter controls (dropdowns, search input) render in a shared filter bar that persists across view switches. The filter bar is not part of the swapped partial; it is part of the graph page shell.
6. Changing a filter value triggers an htmx request that includes both the filter parameters and the current view parameter.
7. Clearing all filters (via a "Clear" button) removes filter query parameters from the URL and re-renders the current view with the full dataset.
8. If a filter combination results in zero matching requirements, the active view displays an empty state message ("No requirements match the current filters") instead of a blank visualization.
9. Filter state initialization: on page load, filters are populated from URL query parameters. If no parameters are present, no filters are applied.
10. The filter bar reuses the existing graph filter infrastructure from REQ-DASH-061, extending it with view-awareness.

## Dependencies

- REQ-DASH-066 (view switcher provides the view parameter and partial swap mechanism)
- REQ-DASH-061 (graph filters provide the existing filter dropdowns and search input)

## Blocks

(none)

## Files to Modify

- `internal/dashboard/templates/partials/graph.html` (refactor filter bar out of DAG-specific partial into shared graph page shell)
- `internal/dashboard/static/app.js` (shared filter state management, URL param aggregation for htmx requests)
- `internal/cmd/serve_dashboard.go` (extract shared filter logic into reusable function called by all three view handlers)

## Test Strategy

- Unit test: Go shared filter function correctly filters requirements by status, priority, category, and search text
- Unit test: Go shared filter function returns full dataset when no filter parameters are provided
- Unit test: each view handler (dag, timeline, treemap) calls the shared filter function before building view data
- Integration test: apply status=COMPLETE filter in DAG view, switch to Timeline, verify only COMPLETE requirements are shown
- Integration test: apply category filter in Treemap, switch to DAG, verify same category filter is active
- Integration test: verify URL preserves all filter parameters across view switches
- Integration test: clear all filters, verify all requirements appear and URL filter params are removed
- Integration test: apply filters that match zero requirements, verify empty state message appears

## Effort

- 1.00 weeks

## Priority

- HIGH

## Phase

- 32
