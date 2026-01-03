# REQ-UX-003: rtmx status --live shall auto-refresh on file changes

## Status: MISSING
## Priority: MEDIUM
## Phase: 5

## Description
rtmx status --live shall auto-refresh on file changes

## Acceptance Criteria
- [ ] Updates within 1s

## Test Cases
- `tests/test_cli_ux.py::test_live_refresh`


## Notes
Watch database file and re-render on change
