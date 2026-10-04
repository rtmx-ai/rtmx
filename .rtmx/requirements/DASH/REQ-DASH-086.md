# REQ-DASH-086: Timeline D3 SVG Rendering (067b)

## Summary

Client-side D3 rendering of the timeline view. Reads the JSON data structure produced by REQ-DASH-067 and renders horizontal bars in an SVG element within the timeline partial. The X-axis shows sprint columns, the Y-axis groups by category, bar width is proportional to effort_weeks, and bar color reflects status via design tokens. Hovering a bar shows a tooltip with requirement details. The SVG viewBox scales to fit data and the container is horizontally scrollable.

## Acceptance Criteria

1. The timeline renders as an SVG element within the htmx swap target when view=timeline is selected.
2. The X-axis displays one column per unique Sprint value, ordered by semver as provided by the JSON.
3. The Y-axis groups requirements by Category. Each category label appears once as a row header. Requirements within a category are stacked vertically.
4. Each requirement renders as a horizontal bar. Bar width is proportional to effort_weeks relative to the column width. Minimum bar width is 20px for visibility.
5. Bar fill color uses status design tokens: --rtmx-status-complete, --rtmx-status-partial, --rtmx-status-missing, --rtmx-status-not-started.
6. Hovering a bar shows a tooltip with: requirement ID, title, status, effort_weeks, assignee, and sprint.
7. The SVG viewBox scales to fit the data; the container is horizontally scrollable if the timeline exceeds the viewport width.
8. The timeline renders correctly for 1 to 300 requirements without layout overflow or label collision.

## Dependencies

- REQ-DASH-067 (Go handler provides the timeline JSON data structure)
- REQ-DASH-092 (timeline partial template provides the SVG container)
- REQ-DASH-091 (reusable tooltip component for hover)
- REQ-DASH-010 (design tokens for status colors)

## Blocks

- REQ-DASH-068 (dependency arrows overlay on timeline)
- REQ-DASH-069 (zoom levels for timeline)

## Files to Modify

- `internal/dashboard/static/app.js` (timeline D3 rendering function)
- `internal/dashboard/templates/partials/timeline.html` (D3 script block for rendering)

## Test Strategy

- Integration test: load timeline view, verify SVG element renders with correct number of bars matching requirement count
- Integration test: verify bar colors match status values from the database
- Visual regression test: screenshot timeline with test fixture data, compare against golden image
- Edge case test: single requirement, category with one item, 300 requirements

## Effort

- 0.75 weeks

## Priority

- HIGH

## Phase

- 32
