# REQ-DASH-039: Checkbox Selection Column on Requirements Table Rows

## Summary

The requirements table shall include a checkbox column for row selection, with a header
checkbox for select-all, enabling multi-row selection for bulk operations.

## Acceptance Criteria

1. First column of the requirements table has a checkbox input per row
2. Header row has a checkbox that toggles select-all for the current page
3. Selection state is managed by an Alpine.js Set in the requirements table `x-data`
4. Selecting/deselecting individual rows updates the Set and header checkbox state
5. Header checkbox shows indeterminate state when some but not all rows are selected

## Dependencies

- REQ-DASH-035 (j/k table row selection)

## Blocks

- REQ-DASH-040

## Files to Modify

- `internal/dashboard/templates/partials/requirements.html`
- `internal/dashboard/static/app.js`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: each row renders a checkbox with correct data-req-id
- Unit test: header checkbox toggles all row checkboxes
- Unit test: Alpine Set tracks selected requirement IDs correctly
- Unit test: header checkbox shows indeterminate when partially selected

## Effort

- 0.25 weeks

## Priority

- MEDIUM

## Phase

- 31
