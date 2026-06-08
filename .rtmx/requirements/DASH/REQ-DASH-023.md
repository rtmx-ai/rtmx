# REQ-DASH-023: Active Filter Chips with Clear-All Control

## Summary

Active filters shall be displayed as removable pill-shaped chips below the table header,
each with an X button to remove that filter, and a "Clear all" link to remove all filters.

## Acceptance Criteria

1. Active filters are displayed as removable pill elements below the table header
2. Each chip shows the filter field and value with an X button to remove it
3. Clicking X on a chip removes that filter and re-requests the partial
4. A "Clear all" link appears when any filters are active, removing all filters on click

## Dependencies

- REQ-DASH-022 (per-column filter dropdowns)

## Blocks

- None

## Files to Modify

- `internal/dashboard/templates/partials/requirements.html`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: active filters render as chip elements with correct labels
- Unit test: clicking X on a chip removes that filter param from request
- Unit test: "Clear all" removes all filter params
- Integration test: chip removal updates table content

## Effort

- 0.25 weeks

## Priority

- MEDIUM

## Phase

- 31
