# REQ-DASH-066: View Switcher Component

## Summary

The graph page shall display a segmented button bar at the top allowing the user to switch between DAG, Timeline, and Treemap visualizations of the same requirement data. The switcher uses htmx to swap the view partial without a full page reload. The currently selected view is persisted in the URL query parameter (?view=dag|timeline|treemap) so that bookmarks and shared links restore the correct view. Alpine.js state tracks the active view for immediate visual feedback on the button bar.

## Acceptance Criteria

1. A segmented button bar with three options (DAG, Timeline, Treemap) renders at the top of the graph page, above the visualization area.
2. The active button is visually distinguished using design tokens (--rtmx-emerald background, --rtmx-text foreground). Inactive buttons use --rtmx-bg-card background.
3. Clicking a view button issues an htmx GET to `/partials/graph?view={selected}` and swaps the visualization container via hx-target.
4. The URL query parameter `?view=` updates on each switch via `history.replaceState` without a full page navigation.
5. On page load, the view parameter is read from the URL; if absent, defaults to `dag`.
6. The Go handler for the graph partial reads the `view` query parameter and returns the appropriate partial (graph.html for dag, timeline.html for timeline, treemap.html for treemap).
7. If an unrecognized view value is provided, the handler falls back to `dag` without error.
8. The segmented button bar is keyboard-accessible: arrow keys move focus between buttons, Enter/Space activates.
9. The button bar does not re-render when the partial swaps; only the visualization container below it changes.

## Dependencies

- REQ-DASH-001 (SPA framework provides htmx partial swap infrastructure)
- REQ-DASH-010 (design tokens for button styling)

## Blocks

- REQ-DASH-067 (timeline view partial)
- REQ-DASH-071 (treemap view partial)
- REQ-DASH-074 (shared view state)

## Files to Modify

- `internal/dashboard/templates/partials/graph.html` (add view switcher bar above visualization container)
- `internal/dashboard/static/app.js` (Alpine component for view switcher state, URL param sync)
- `internal/cmd/serve_dashboard.go` (read `view` query param, route to correct partial template)
- `internal/dashboard/templates/partials/timeline.html` (new, placeholder until REQ-DASH-067)
- `internal/dashboard/templates/partials/treemap.html` (new, placeholder until REQ-DASH-071)

## Test Strategy

- Unit test: Go handler returns graph.html partial when view=dag or view is absent
- Unit test: Go handler returns timeline.html when view=timeline, treemap.html when view=treemap
- Unit test: Go handler falls back to graph.html for unrecognized view values
- Integration test: load graph page, verify segmented button bar renders with three buttons
- Integration test: click Timeline button, verify htmx swap replaces visualization container
- Integration test: verify URL query parameter updates after view switch

## Effort

- 0.75 weeks

## Priority

- HIGH

## Phase

- 32
