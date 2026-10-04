# REQ-DASH-029: Panel Open/Close Updates Browser URL via pushState

## Summary

Opening and closing the detail panel shall update the browser URL using pushState,
enabling direct linking to a specific requirement detail view.

## Acceptance Criteria

1. Opening the panel pushes `/requirements/:id` to browser history via `history.pushState`
2. Closing the panel pops back to the previous URL
3. Direct navigation to `/requirements/:id` opens the panel on page load
4. Browser back button closes the panel if it was opened via pushState

## Dependencies

- REQ-DASH-028 (panel close on Escape and backdrop)

## Blocks

- None

## Files to Modify

- `internal/dashboard/static/app.js`

## Test Strategy

- Unit test: opening panel calls pushState with correct URL
- Unit test: closing panel restores previous URL
- Integration test: direct URL navigation opens panel on load

## Effort

- 0.25 weeks

## Priority

- MEDIUM

## Phase

- 31
