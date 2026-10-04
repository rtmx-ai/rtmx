# REQ-DASH-071: Treemap Go Handler and JSON Hierarchy (071a)

## Summary

The Go backend builds a hierarchical JSON data structure for the D3 treemap visualization. The hierarchy is two levels: root > categories > requirements. Each requirement entry includes effort_weeks (with a minimum of 0.25 for zero/null values), status, category, assignee, and ID. Empty categories are excluded. The JSON is embedded as a script block in the treemap partial for client-side rendering.

## Acceptance Criteria

1. The Go handler builds a two-level hierarchy: root > categories > requirements.
2. Each requirement entry includes: id, title, status, effort_weeks, category, and assignee.
3. Requirements with effort_weeks=0 or null are assigned a minimum value of 0.25 in the JSON output.
4. Empty categories (no requirements after filtering) are excluded from the hierarchy.
5. The JSON structure is embedded as a `<script type="application/json">` block in the treemap partial.
6. The handler is registered at the `/partials/treemap` route and returns an HTML fragment.
7. The treemap JSON renders correctly for databases with 1 to 300 requirements.

## Dependencies

- REQ-DASH-066 (view switcher provides partial swap mechanism)
- REQ-DASH-044 (enriched JSON provides effort_weeks, category, status per node)
- REQ-DASH-093 (treemap partial template skeleton must exist)

## Blocks

- REQ-DASH-072 (treemap color modes)
- REQ-DASH-073 (treemap drill-down)
- REQ-DASH-087 (D3 squarify rendering consumes this JSON)

## Files to Modify

- `internal/cmd/serve_dashboard.go` (build hierarchical JSON for treemap: root > categories > requirements)

## Test Strategy

- Unit test: Go handler produces correct hierarchical JSON with categories as children of root and requirements as children of categories
- Unit test: requirements with effort_weeks=0 are assigned minimum value of 0.25 in JSON output
- Unit test: empty categories (no requirements after filtering) are excluded from hierarchy
- Unit test: JSON structure matches expected schema for consumer (REQ-DASH-087)

## Effort

- 0.50 weeks

## Priority

- HIGH

## Phase

- 32
