# REQ-DASH-032: Palette Search Input with Debounced API Call and Result List

## Summary

The command palette search input shall query the requirements API with debounced input,
rendering results as a styled list with requirement ID, status badge, and truncated description.

## Acceptance Criteria

1. Typing queries `GET /api/requirements?search=<query>&per_page=10` after 150ms debounce
2. Results render as list items with: req ID (monospace emerald), status badge, truncated description
3. Empty query shows "Type to search..." placeholder text
4. Loading state shown while API request is in flight
5. No results state shows appropriate message

## Dependencies

- REQ-DASH-031 (command palette modal)

## Blocks

- REQ-DASH-033, REQ-DASH-034

## Files to Modify

- `internal/dashboard/static/app.js`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: typing triggers API call after 150ms debounce
- Unit test: results render with correct structure (ID, badge, description)
- Unit test: empty query shows placeholder text
- Integration test: search returns and displays matching requirements

## Effort

- 0.5 weeks

## Priority

- HIGH

## Phase

- 31
