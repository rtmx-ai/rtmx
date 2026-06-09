# REQ-DASH-071: Treemap View Partial

## Summary

A D3 treemap visualization displays all requirements in a space-filling layout where rectangle area represents effort_weeks and color represents status. Requirements are nested within their category, creating a two-level hierarchy: the outer rectangles are categories, and the inner rectangles are individual requirements. The treemap provides an immediate visual sense of where effort is concentrated and what the status distribution looks like. Hover shows a tooltip with requirement details, and clicking a requirement opens the detail panel.

## Acceptance Criteria

1. The treemap partial renders as an SVG element within the htmx swap target when view=treemap is selected.
2. The treemap uses a two-level hierarchy: root > categories > requirements.
3. Rectangle area is proportional to effort_weeks. Requirements with effort_weeks=0 or null are assigned a minimum value of 0.25 for visibility.
4. Rectangle fill color uses status design tokens: --rtmx-status-complete (green), --rtmx-status-partial (amber), --rtmx-status-missing (red), --rtmx-status-not-started (gray).
5. Category group boundaries are shown with a 2px border in --rtmx-border color. Category labels render at the top of each group.
6. Individual requirement rectangles display the requirement ID as text if the rectangle is large enough (width > 60px and height > 20px). Smaller rectangles show no text.
7. Hovering a requirement rectangle shows a tooltip with: ID, title, status, effort_weeks, category, and assignee.
8. Clicking a requirement rectangle navigates to the detail view or opens the detail panel (consistent with existing graph node click behavior).
9. The D3 treemap uses the squarify tiling algorithm (d3.treemapSquarify) for optimal aspect ratios.
10. The treemap fills the available container width and uses a 16:9 aspect ratio for height, with a minimum height of 400px.
11. The Go handler builds the hierarchical JSON data structure server-side (root > categories > requirements with effort values).
12. The treemap renders correctly for databases with 1 to 300 requirements without text overlap or invisible rectangles.

## Dependencies

- REQ-DASH-066 (view switcher provides partial swap mechanism)
- REQ-DASH-044 (enriched JSON provides effort_weeks, category, status per node)

## Blocks

- REQ-DASH-072 (treemap color modes)
- REQ-DASH-073 (treemap drill-down)

## Files to Modify

- `internal/dashboard/templates/partials/treemap.html` (new partial with SVG container and D3 rendering script)
- `internal/dashboard/static/app.js` (treemap rendering function using D3 hierarchy and treemap layout)
- `internal/cmd/serve_dashboard.go` (build hierarchical JSON for treemap: root > categories > requirements)

## Test Strategy

- Unit test: Go handler produces correct hierarchical JSON with categories as children of root and requirements as children of categories
- Unit test: requirements with effort_weeks=0 are assigned minimum value of 0.25 in JSON output
- Unit test: empty categories (no requirements after filtering) are excluded from hierarchy
- Integration test: load treemap view, verify SVG element renders with correct number of leaf rectangles
- Integration test: verify rectangle colors match status values from database
- Integration test: hover a rectangle, verify tooltip appears with correct data
- Integration test: click a rectangle, verify navigation to detail view
- Performance test: treemap renders 300 requirements in under 200ms

## Effort

- 1.25 weeks

## Priority

- HIGH

## Phase

- 32
