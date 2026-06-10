# REQ-DASH-067: Timeline Go Handler and JSON Structure (067a)

## Summary

The Go backend builds a timeline JSON data structure grouped by sprint/category with semver-ordered sprint columns. The handler produces a JSON payload containing sprints (sorted by semver), categories, and per-requirement entries with effort_weeks, status, assignee, and sprint assignment. Requirements with empty Sprint are placed in an "Unscheduled" group. The JSON is embedded as a script block in the timeline partial for client-side rendering.

## Acceptance Criteria

1. The Go handler builds a timeline data structure with sprints sorted by semver (e.g., v0.1.0 < v0.2.0 < v1.0.0).
2. Requirements with empty Sprint field are grouped in an "Unscheduled" column positioned at the far right.
3. Requirements are grouped by Category on the Y-axis. Each category appears once.
4. Each requirement entry includes: id, title, status, effort_weeks, assignee, sprint, and category.
5. Empty categories (all requirements filtered out) are excluded from the output.
6. The JSON structure is embedded as a `<script type="application/json">` block in the timeline partial.
7. The handler is registered at the `/partials/timeline` route and returns an HTML fragment.
8. The timeline JSON renders correctly for databases with 1 to 300 requirements.

## Dependencies

- REQ-DASH-066 (view switcher provides partial swap mechanism)
- REQ-DASH-044 (enriched JSON provides effort_weeks, sprint, assignee, category per node)
- REQ-DASH-092 (timeline partial template skeleton must exist)

## Blocks

- REQ-DASH-068 (dependency arrows overlay on timeline)
- REQ-DASH-069 (zoom levels for timeline)
- REQ-DASH-070 (assignee swimlanes for timeline)
- REQ-DASH-086 (D3 SVG rendering consumes this JSON)

## Files to Modify

- `internal/cmd/serve_dashboard.go` (build timeline data structure, serve as JSON in partial)

## Test Strategy

- Unit test: Go handler produces correct timeline JSON with sprints sorted by semver, categories grouped, effort values populated
- Unit test: requirements with empty Sprint field appear in "Unscheduled" column
- Unit test: categories with zero matching requirements are excluded from output
- Unit test: JSON structure matches expected schema for consumer (REQ-DASH-086)

## Effort

- 0.75 weeks

## Priority

- HIGH

## Phase

- 32
