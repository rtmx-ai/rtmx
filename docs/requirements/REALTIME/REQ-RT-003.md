# REQ-RT-003: Updates shall include only changed requirements

## Status: COMPLETE
## Priority: MEDIUM
## Phase: 7

## Description
Updates shall include only changed requirements

## Acceptance Criteria
- [ ] Delta not full refresh

## Test Cases
- `tests/test_realtime.py::test_delta_updates`


## Notes
Diff previous and current state to send minimal payload
