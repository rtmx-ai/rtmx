# REQ-DASH-040: Floating Action Bar Appears When Rows Are Selected

## Summary

A floating sticky action bar shall appear above the requirements table when one or more
rows are selected, showing the selection count and bulk action buttons.

## Acceptance Criteria

1. When `selected.size > 0`, a sticky bar appears above the table showing the selection count
2. Bar contains action buttons: Set Status (dropdown), Set Assignee (input), Set Sprint (input), Set Priority (dropdown)
3. Bar disappears when all selections are cleared
4. Bar is styled with dark theme consistent with design tokens
5. Bar position is sticky so it remains visible during scroll

## Dependencies

- REQ-DASH-039 (checkbox selection column)

## Blocks

- REQ-DASH-042

## Files to Modify

- `internal/dashboard/templates/partials/requirements.html`
- `internal/dashboard/static/app.js`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: bar appears when at least one row is selected
- Unit test: bar disappears when all rows are deselected
- Unit test: bar shows correct selection count
- Unit test: each action button renders with correct type (dropdown vs input)

## Effort

- 0.25 weeks

## Priority

- MEDIUM

## Phase

- 31
