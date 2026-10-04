# REQ-DASH-035: Table Row j/k Selection with Emerald Highlight

## Summary

The requirements table shall support j/k keyboard navigation to move a selection highlight
through rows, with Enter or o opening the detail panel for the selected row.

## Acceptance Criteria

1. j key moves `.row-selected` class to the next table row
2. k key moves `.row-selected` class to the previous table row
3. Selected row has an emerald left border using `var(--rtmx-emerald)`
4. Selected row auto-scrolls into view
5. Enter or o on a selected row opens the detail panel

## Dependencies

- REQ-DASH-001 (SPA framework)
- REQ-DASH-010 (design tokens)

## Blocks

- REQ-DASH-038, REQ-DASH-039

## Files to Modify

- `internal/dashboard/static/app.js`
- `internal/dashboard/static/styles.css`
- `internal/dashboard/templates/partials/requirements.html` (data-req-id attributes)

## Test Strategy

- Unit test: j key moves selection to next row
- Unit test: k key moves selection to previous row
- Unit test: selected row has emerald left border class
- Unit test: Enter on selected row opens detail panel
- Integration test: auto-scroll behavior when selection moves off screen

## Effort

- 0.5 weeks

## Priority

- MEDIUM

## Phase

- 31
