# REQ-DASH-022: Per-Column Filter Dropdowns for Status, Priority, and Category

## Summary

Filterable column headers shall have filter icons that open Alpine-driven dropdown
panels with checkboxes, enabling multi-select filtering of the requirements table
via htmx partial re-requests.

## Acceptance Criteria

1. Status, priority, and category column headers have a filter icon
2. Clicking the filter icon opens an Alpine dropdown with checkbox options for that column
3. Checking an option re-requests the partial with the appropriate filter param
4. Multiple selections are supported (comma-separated filter values)
5. Toast feedback confirms filter application (via REQ-DASH-015)

## Dependencies

- REQ-DASH-020 (sort/filter query parameters)
- REQ-DASH-015 (toast feedback)

## Blocks

- REQ-DASH-023

## Files to Modify

- `internal/dashboard/templates/partials/requirements.html`
- `internal/dashboard/static/app.js`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: filter icon click opens dropdown with correct options
- Unit test: checking an option triggers partial re-request with filter param
- Unit test: multiple selections produce comma-separated filter value
- Integration test: table updates to show only filtered rows

## Effort

- 0.5 weeks

## Priority

- HIGH

## Phase

- 31
