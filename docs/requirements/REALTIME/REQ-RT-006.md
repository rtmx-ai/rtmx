# REQ-RT-006: WebSocket shall auto-reconnect on disconnect

## Status: MISSING
## Priority: MEDIUM
## Phase: 7

## Description
WebSocket shall auto-reconnect on disconnect

## Acceptance Criteria
- [ ] Reconnects within 5s

## Test Cases
- `tests/test_realtime.py::test_auto_reconnect`


## Notes
Exponential backoff reconnection with state sync
