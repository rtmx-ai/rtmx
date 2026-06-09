# REQ-DASH-067: Timeline View Partial

## Summary

A server-rendered timeline view displays requirements as horizontal bars in a sprint-based grid. The X-axis represents sprint/version columns derived from the Sprint field on each requirement. The Y-axis groups requirements by category. Each bar is colored by status using design tokens, and bar width is proportional to the requirement's effort_weeks value. The partial is served by the Go backend as an HTML fragment containing an SVG timeline, swapped into the graph page via the view switcher.

## Acceptance Criteria

1. The timeline partial renders as an SVG element within the htmx swap target when view=timeline is selected.
2. The X-axis displays one column per unique Sprint value, ordered by version semver (e.g., v0.1.0 < v0.2.0 < v1.0.0). Requirements with empty Sprint are grouped in an "Unscheduled" column at the far right.
3. The Y-axis groups requirements by Category. Each category label appears once as a row header. Requirements within a category are stacked vertically.
4. Each requirement renders as a horizontal bar. Bar width is proportional to effort_weeks relative to the column width. Minimum bar width is 20px for visibility of small-effort items.
5. Bar fill color uses status design tokens: --rtmx-status-complete for COMPLETE, --rtmx-status-partial for PARTIAL, --rtmx-status-missing for MISSING, --rtmx-status-not-started for NOT_STARTED.
6. Hovering a bar shows a tooltip with: requirement ID, title, status, effort_weeks, assignee, and sprint.
7. The SVG viewBox scales to fit the data; the container is horizontally scrollable if the timeline exceeds the viewport width.
8. Empty categories (all requirements filtered out) are not shown.
9. The Go handler builds the timeline data structure server-side, embedding it as a JSON script block that D3 reads to render the SVG client-side.
10. The timeline renders correctly for databases with 1 to 300 requirements without layout overflow or label collision.

## Dependencies

- REQ-DASH-066 (view switcher provides partial swap mechanism)
- REQ-DASH-044 (enriched JSON provides effort_weeks, sprint, assignee, category per node)

## Blocks

- REQ-DASH-068 (dependency arrows overlay on timeline)
- REQ-DASH-069 (zoom levels for timeline)
- REQ-DASH-070 (assignee swimlanes for timeline)

## Files to Modify

- `internal/dashboard/templates/partials/timeline.html` (new partial with SVG container and D3 rendering script)
- `internal/dashboard/static/app.js` (timeline rendering function using D3)
- `internal/cmd/serve_dashboard.go` (build timeline data structure, serve as JSON in partial)

## Test Strategy

- Unit test: Go handler produces correct timeline JSON with sprints sorted by semver, categories grouped, effort values populated
- Unit test: requirements with empty Sprint field appear in "Unscheduled" column
- Unit test: categories with zero matching requirements are excluded from output
- Integration test: load timeline view, verify SVG element renders with correct number of bars matching requirement count
- Integration test: verify bar colors match status values from the database
- Visual regression test: screenshot timeline with test fixture data, compare against golden image

## Effort

- 1.50 weeks

## Priority

- HIGH

## Phase

- 32
