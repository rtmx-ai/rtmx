# REQ-DASH-019: Inline Edit Escape Cancels and Loading Indicator During PATCH

## Summary

Inline edit interactions shall support Escape to cancel without issuing a PATCH,
display a loading indicator while PATCH is in flight, and revert on error with
toast feedback.

## Acceptance Criteria

1. Pressing Escape while editing reverts the cell to its original value without issuing a PATCH
2. While a PATCH is in flight, the cell shows reduced opacity or a subtle spinner
3. On PATCH error, the cell reverts to its original value
4. On PATCH error, a toast notification shows the error message (via REQ-DASH-015)

## Dependencies

- REQ-DASH-017 (status/priority inline edit)
- REQ-DASH-018 (assignee/sprint inline edit)
- REQ-DASH-015 (toast feedback)

## Blocks

- None

## Files to Modify

- `internal/dashboard/templates/partials/requirements.html`
- `internal/dashboard/static/app.js`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: Escape key does not trigger PATCH and restores original value
- Unit test: loading class applied during PATCH in flight
- Unit test: error response reverts cell and triggers error toast
- Integration test: full cancel and error flow in browser

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 31
