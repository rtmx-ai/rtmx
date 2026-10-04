# REQ-DASH-042: Bulk Action Execution with Table Refresh and Toast Feedback

## Summary

Clicking a bulk action button in the floating action bar shall call the bulk update API
with selected IDs and the chosen update, refreshing the table and showing toast feedback.

## Acceptance Criteria

1. Clicking a bulk action button calls `POST /api/requirements/bulk` with selected IDs and chosen update
2. On success, the requirements table re-renders via htmx full partial swap
3. On success, a toast shows "Updated N requirements"
4. On error, a toast shows the error message and the selection is preserved
5. Selection is cleared after a successful bulk update

## Dependencies

- REQ-DASH-040 (floating action bar)
- REQ-DASH-041 (bulk API endpoint)
- REQ-DASH-015 (toast feedback)

## Blocks

- None

## Files to Modify

- `internal/dashboard/static/app.js`
- `internal/dashboard/templates/partials/requirements.html`

## Test Strategy

- Unit test: bulk action button sends correct POST payload
- Unit test: successful response triggers table refresh and success toast
- Unit test: error response shows error toast and preserves selection
- Integration test: end-to-end bulk status update with table refresh

## Effort

- 0.25 weeks

## Priority

- MEDIUM

## Phase

- 31
