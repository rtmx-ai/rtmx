# REQ-UX-003: rtmx status --live shall auto-refresh on file changes

## Status: COMPLETE
## Priority: MEDIUM
## Phase: 5

## Description
rtmx status --live shall auto-refresh the display when the RTM database file changes, providing real-time visibility into requirement status.

## Acceptance Criteria
- [x] --live flag added to status command
- [x] Watches RTM database file for changes
- [x] Clears terminal and re-renders on file change
- [x] Updates within 1s of file modification (0.5s poll interval)
- [x] Graceful exit on Ctrl+C
- [x] Shows timestamp of last refresh

## Test Cases
- `tests/test_cli_ux.py::TestLiveRefresh::test_live_flag_exists`
- `tests/test_cli_ux.py::TestLiveRefresh::test_live_detects_file_change`

## Implementation Notes
- Use watchdog library for cross-platform file watching
- Clear terminal with ANSI escape codes
- Poll-based fallback if watchdog not available
- Show "Watching for changes... (Ctrl+C to exit)" message

## Visual Design
```
=============================== RTM Status Check ===============================

Requirements: [██████████████████████████████████████████████████]  42.5%

✓ 31 complete  ⚠ 0 partial  ✗ 42 missing
(73 total)

Last updated: 2026-01-03 14:32:15
Watching for changes... (Ctrl+C to exit)
```
