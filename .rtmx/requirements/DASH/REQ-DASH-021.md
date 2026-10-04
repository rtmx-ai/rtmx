# REQ-DASH-021: Clickable Column Headers with Sort Direction Indicators

## Summary

Requirements table column headers shall be clickable links that trigger htmx-powered
sorting with visual indicators showing the current sort direction.

## Acceptance Criteria

1. Column headers are `<a>` elements with `hx-get` including `sort=field&order=dir` params
2. An arrow icon (CSS triangle) shows the current sort direction on the active column
3. Clicking the same header toggles between ascending and descending order
4. Clicking a different header sorts by that column in ascending order

## Dependencies

- REQ-DASH-020 (sort/filter query parameters)

## Blocks

- None

## Files to Modify

- `internal/dashboard/templates/partials/requirements.html`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: column header renders with correct hx-get URL and sort params
- Unit test: active sort column shows direction indicator
- Visual test: sort indicator arrow renders correctly in both directions

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 31
