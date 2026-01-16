# REQ-SITE-007: Live Roadmap via RTMX Sync WebSocket

## Status: MISSING
## Priority: HIGH
## Phase: 10
## Effort: 1.5 weeks

## Description

The website roadmap page shall display live RTM status updates via the RTMX Sync WebSocket API. This serves as the first public demonstration of RTMX real-time collaboration capabilities.

When the RTMX Sync server (REQ-COLLAB-001) is operational, the roadmap page at rtmx.ai/roadmap will subscribe to RTM changes and update phase completion percentages, requirement statuses, and progress indicators in real-time without page refresh.

## Acceptance Criteria

- [ ] Roadmap page establishes WebSocket connection to RTMX Sync server
- [ ] Phase completion percentages update live when requirements change status
- [ ] Connection status indicator shows sync state (connected, disconnected, reconnecting)
- [ ] Graceful fallback to static data when sync server is unavailable
- [ ] Updates propagate in <1 second from change to display
- [ ] Page handles reconnection automatically on connection loss

## Test Cases

- `tests/test_website.py::test_roadmap_live_sync` - WebSocket connection established
- `tests/test_website.py::test_roadmap_status_indicator` - Connection status displayed
- `tests/test_website.py::test_roadmap_fallback` - Static fallback when offline

## Technical Notes

Implementation approach:
1. Client-side JavaScript WebSocket client connecting to RTMX Sync endpoint
2. Subscribe to Y.Doc changes for requirements status
3. Re-render phase progress bars on CRDT update events
4. Display connection status badge in roadmap header
5. Cache last known state for graceful degradation

This demonstrates the same sync technology that powers team collaboration features.

## Dependencies

- REQ-COLLAB-001: CRDT sync server operational

## Blocks

None
