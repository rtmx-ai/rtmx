# REQ-DASH-018: Click-to-Edit Assignee and Sprint Cells with Text Input

## Summary

Assignee and sprint cells in the requirements table shall be editable inline via
text inputs, committing changes on blur or Enter through hx-patch with row partial swap.

## Acceptance Criteria

1. Clicking an assignee cell shows an `<input type="text">` pre-filled with the current value
2. Clicking a sprint cell shows an `<input type="text">` pre-filled with the current value
3. Input commits on blur or Enter key via `hx-patch`
4. The PATCH response targets the closest `<tr>` via row partial swap

## Dependencies

- REQ-DASH-016 (row partial endpoint)
- REQ-DASH-010 (design tokens)

## Blocks

- REQ-DASH-019

## Files to Modify

- `internal/dashboard/templates/partials/requirements.html`
- `internal/dashboard/static/app.js`

## Test Strategy

- Unit test: clicking assignee cell renders input with current value
- Unit test: blur event triggers PATCH with updated value
- Unit test: Enter key triggers PATCH with updated value
- Integration test: row updates in place after assignee change

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 31
