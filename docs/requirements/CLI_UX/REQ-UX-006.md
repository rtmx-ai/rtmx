# REQ-UX-006: rich library shall be optional dependency

## Status: MISSING
## Priority: LOW
## Phase: 5

## Description
rich library shall be optional dependency

## Acceptance Criteria
- [ ] Works without rich

## Test Cases
- `tests/test_cli_ux.py::test_graceful_degradation`


## Notes
Graceful fallback to basic output if rich not installed
