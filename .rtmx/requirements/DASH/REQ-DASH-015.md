# REQ-DASH-015: PATCH Responses Include HX-Trigger Header for Toast Feedback

## Summary

PATCH responses from the requirements API shall include an HX-Trigger header containing
toast notification data, enabling htmx to automatically dispatch showToast events for
success and error feedback without custom JavaScript.

## Acceptance Criteria

1. PATCH `/api/requirements/:id` returns `HX-Trigger: {"showToast": {"type":"success","message":"Saved"}}` on successful update
2. On validation error, PATCH returns `HX-Trigger` with `type: "error"` and descriptive message
3. htmx `htmx:trigger` event dispatches to the Alpine `showToast` function
4. Toast appears automatically after any inline edit PATCH completes

## Dependencies

- REQ-DASH-014 (toast container and store)

## Blocks

- REQ-DASH-022, REQ-DASH-042

## Files to Modify

- `internal/cmd/serve_api.go`

## Test Strategy

- Unit test: PATCH handler returns correct HX-Trigger header on success
- Unit test: PATCH handler returns error HX-Trigger on validation failure
- Integration test: inline edit triggers toast notification in browser

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 31
