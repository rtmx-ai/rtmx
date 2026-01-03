# REQ-UX-007: rtmx backlog --view list shall show all requirements for phase

## Status: MISSING
## Priority: HIGH
## Phase: 5

## Description
Add a new `--view list` option to the backlog command that shows ALL requirements for a given phase, not just filtered subsets (critical path, quick wins, blockers).

## Acceptance Criteria
- [ ] `rtmx backlog --view list` shows all incomplete requirements
- [ ] `rtmx backlog --view list --phase N` shows all requirements for phase N (complete and incomplete)
- [ ] Output includes status, requirement ID, description, effort, and dependencies
- [ ] Requirements sorted by status (missing first), then by ID
- [ ] Summary shows total count and completion percentage for filtered view

## Example Output

```
$ rtmx backlog --view list --phase 5

======================== Phase 5: CLI UX (All Requirements) ========================

Total: 7 requirements | 2 complete (28.6%) | 5 incomplete

+-----+----------+---------------+-------------------------------------+----------+-------------+
|   # | Status   | Requirement   | Description                         | Effort   | Depends On  |
+=====+==========+===============+=====================================+==========+=============+
|   1 | ✗        | REQ-UX-002    | All tabular output shall use ali... | 1.0w     | REQ-UX-001  |
|   2 | ✗        | REQ-UX-003    | rtmx status --live shall auto-re... | 1.5w     | REQ-UX-001  |
|   3 | ✗        | REQ-UX-004    | rtmx tui command shall launch in... | 2.0w     | REQ-UX-001  |
|   4 | ✗        | REQ-UX-005    | TUI shall support vim-style keyb... | 1.0w     | REQ-UX-004  |
|   5 | ✗        | REQ-UX-007    | rtmx backlog --view list shall s... | 0.5w     | -           |
|   6 | ✓        | REQ-UX-001    | rtmx status shall display rich p... | 1.0w     | -           |
|   7 | ✓        | REQ-UX-006    | rich library shall be optional d... | 0.5w     | REQ-UX-001  |
+-----+----------+---------------+-------------------------------------+----------+-------------+
```

## Test Cases
- `tests/test_cli_ux.py::test_backlog_list_view`
- `tests/test_cli_ux.py::test_backlog_list_view_with_phase`
- `tests/test_cli_ux.py::test_backlog_list_view_shows_complete`

## Implementation Notes
- Add `list` to BacklogView enum
- Implement `_show_list()` function in backlog.py
- When `--phase` is specified with `--view list`, show ALL requirements (not just incomplete)
- Sort: incomplete first (by priority), then complete (by ID)
