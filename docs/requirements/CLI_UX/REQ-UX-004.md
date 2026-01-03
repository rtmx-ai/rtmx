# REQ-UX-004: rtmx tui command shall launch interactive dashboard

## Status: COMPLETE
## Priority: MEDIUM
## Phase: 5

## Description
rtmx tui command shall launch an interactive terminal dashboard using the textual library, providing a split-pane view with requirements list and detail panels.

## Acceptance Criteria
- [x] `rtmx tui` command launches interactive dashboard
- [x] Split-pane layout: requirements list on left, details on right
- [x] Requirements list shows status, ID, and description
- [x] Detail pane shows full requirement information
- [x] Status bar shows summary statistics
- [x] Graceful exit with 'q' key
- [x] textual is optional dependency (`rtmx[tui]`)

## Test Cases
- `tests/test_cli_ux.py::TestTUI::test_tui_command_exists`
- `tests/test_cli_ux.py::TestTUI::test_tui_app_creates`

## Implementation Notes
- Use textual library for TUI framework
- Create `src/rtmx/cli/tui.py` for TUI implementation
- Add `textual` to optional dependencies
- Show helpful error if textual not installed

## Visual Design
```
┌─ Requirements ──────────────────────┬─ Details ────────────────────────────┐
│ ✓ REQ-CORE-001  Core requirement    │ REQ-CORE-001                         │
│ ✓ REQ-CORE-002  Another core req    │ Status: COMPLETE                     │
│ ✗ REQ-UX-001    Rich progress bars  │ Priority: HIGH                       │
│ > REQ-UX-002    Aligned columns     │ Phase: 5                             │
│ ✗ REQ-UX-003    Live refresh        │                                      │
│                                     │ Description:                         │
│                                     │ All tabular output shall use         │
│                                     │ aligned fixed-width columns...       │
│                                     │                                      │
│                                     │ Dependencies: REQ-UX-001             │
│                                     │ Blocks: REQ-UX-004                   │
├─────────────────────────────────────┴──────────────────────────────────────┤
│ 32/73 complete (43.8%) | Phase 5: 5/7 | q:quit j/k:navigate Enter:select   │
└────────────────────────────────────────────────────────────────────────────┘
```
