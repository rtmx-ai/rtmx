# REQ-DASH-072: Treemap Color Modes

## Summary

The treemap view supports three color modes that re-color rectangles without recomputing the layout: Status (default, 4-color categorical), Priority (4-color categorical by P0-P3), and Completion (continuous green-to-red gradient based on the percentage of dependencies completed). An Alpine.js toggle switches between modes, and D3 applies the new color scale by updating fill attributes on existing rectangles without a full re-render.

## Acceptance Criteria

1. A segmented toggle with three options (Status, Priority, Completion) renders above the treemap SVG.
2. The default color mode is Status.
3. Status mode: rectangles use --rtmx-status-complete (COMPLETE), --rtmx-status-partial (PARTIAL), --rtmx-status-missing (MISSING), --rtmx-status-not-started (NOT_STARTED). This is the existing behavior from REQ-DASH-071.
4. Priority mode: rectangles use four distinct colors: P0=#ef4444 (red/critical), P1=#f59e0b (amber/high), P2=#3b82f6 (blue/medium), P3=#6b7280 (gray/low).
5. Completion mode: each requirement's completion percentage is calculated as (number of COMPLETE dependencies / total dependencies). Requirements with zero dependencies show 100%. Color is a continuous gradient from #ef4444 (0%) through #f59e0b (50%) to #22c55e (100%).
6. Switching color mode updates rectangle fills via D3 transition (300ms duration) without re-running the treemap layout algorithm.
7. A color legend renders below the toggle showing the mapping for the active color mode.
8. The color mode is stored as an Alpine.js reactive property, not in the URL (it is a client-side visual preference, not a data query).
9. The tooltip continues to display correct data regardless of active color mode.
10. Category group borders and labels are unaffected by color mode changes.

## Dependencies

- REQ-DASH-071 (treemap view provides the rectangle layout and SVG structure)

## Blocks

(none)

## Files to Modify

- `internal/dashboard/templates/partials/treemap.html` (add color mode toggle and legend)
- `internal/dashboard/static/app.js` (color scale definitions for each mode, D3 transition logic, completion percentage calculation)

## Test Strategy

- Unit test: completion percentage calculation returns correct values (0 deps = 100%, 3/5 complete = 60%)
- Unit test: priority color scale maps P0-P3 to correct hex values
- Unit test: status color scale maps each status to correct design token value
- Integration test: load treemap, switch to Priority mode, verify rectangles change color
- Integration test: switch to Completion mode, verify gradient colors appear
- Integration test: verify color transitions animate (duration > 0) rather than snap
- Integration test: verify legend updates to reflect active color mode

## Effort

- 0.75 weeks

## Priority

- MEDIUM

## Phase

- 32
