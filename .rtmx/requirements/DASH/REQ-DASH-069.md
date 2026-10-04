# REQ-DASH-069: Timeline Zoom Levels

## Summary

The timeline view supports three zoom levels that control column granularity: Sprint (one column per sprint/version), Phase (one column per phase number), and Quarter (one column per calendar quarter derived from started_date/completed_date). A toggle control above the timeline allows switching between zoom levels. Changing the zoom level triggers an htmx re-render of the timeline partial with the zoom parameter, causing the server to recompute the column layout at the requested granularity.

## Acceptance Criteria

1. A segmented toggle with three options (Sprint, Phase, Quarter) renders above the timeline SVG, below the view switcher bar.
2. The default zoom level is Sprint.
3. In Sprint zoom: columns are sprint/version values sorted by semver. This is the existing behavior from REQ-DASH-067.
4. In Phase zoom: columns are phase numbers (integers) sorted ascending. Each column aggregates all requirements in that phase regardless of sprint. Bar positioning within a phase column follows the same category grouping as Sprint zoom.
5. In Quarter zoom: columns are calendar quarters (e.g., "2026-Q1", "2026-Q2") derived from the StartedDate or CompletedDate of each requirement. Requirements with no date fields fall into an "Unscheduled" column.
6. Switching zoom level issues an htmx GET to `/partials/graph?view=timeline&zoom={level}` and swaps the timeline container.
7. The zoom parameter is persisted in the URL query string alongside the view parameter (e.g., ?view=timeline&zoom=phase).
8. The Go handler reads the zoom query parameter and builds the appropriate column structure server-side.
9. If an unrecognized zoom value is provided, the handler falls back to Sprint zoom.
10. Category row grouping and bar coloring remain consistent across all zoom levels.

## Dependencies

- REQ-DASH-067 (timeline view provides the base layout that zoom levels modify)

## Blocks

(none)

## Files to Modify

- `internal/dashboard/templates/partials/timeline.html` (add zoom toggle control)
- `internal/dashboard/static/app.js` (zoom state management, URL param sync)
- `internal/cmd/serve_dashboard.go` (zoom-aware column builder: sprint, phase, or quarter grouping)

## Test Strategy

- Unit test: Go handler with zoom=sprint produces columns sorted by semver
- Unit test: Go handler with zoom=phase produces columns as phase integers sorted ascending
- Unit test: Go handler with zoom=quarter computes correct quarter from StartedDate
- Unit test: requirements with no dates assigned to "Unscheduled" column in quarter zoom
- Unit test: unrecognized zoom value falls back to sprint
- Integration test: switch zoom from Sprint to Phase, verify column headers change to phase numbers
- Integration test: verify URL updates to include zoom parameter

## Effort

- 1.00 weeks

## Priority

- MEDIUM

## Phase

- 32
