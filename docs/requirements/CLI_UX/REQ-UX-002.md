# REQ-UX-002: All tabular output shall use aligned fixed-width columns

## Status: COMPLETE
## Priority: HIGH
## Phase: 5

## Description
All tabular output shall use aligned fixed-width columns using tabulate with grid format.

## Acceptance Criteria
- [x] Tables have consistent column widths
- [x] No misalignment due to varying content length
- [x] Use tabulate grid format for consistent display
- [x] Column headers present
- [x] Truncation with ellipsis for long content

## Commands Affected
- `rtmx backlog` (all views)
- `rtmx status -v/-vv/-vvv`
- `rtmx health`
- `rtmx deps`

## Visual Design

```
┏━━━━━━━━━━━━━━┳━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━┓
┃ Requirement  ┃ Status   ┃ Description                  ┃ Effort   ┃
┡━━━━━━━━━━━━━━╇━━━━━━━━━━╇━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━┩
│ REQ-UX-001   │ ✓        │ Rich progress bars for st... │ 1.0w     │
│ REQ-UX-002   │ ✗        │ Aligned fixed-width colum... │ 1.0w     │
│ REQ-UX-003   │ ✗        │ Live auto-refresh on file... │ 1.5w     │
└──────────────┴──────────┴──────────────────────────────┴──────────┘
```

## Test Cases
- `tests/test_cli_ux.py::TestAlignedTables::test_format_table_with_rich`
- `tests/test_cli_ux.py::TestAlignedTables::test_format_table_fallback_to_tabulate`
- `tests/test_cli_ux.py::TestAlignedTables::test_format_table_consistent_column_widths`
- `tests/test_cli_ux.py::TestAlignedTables::test_format_table_handles_rich_text_objects`
- `tests/test_cli_ux.py::TestAlignedTables::test_format_table_auto_detects_rich`

## Implementation Notes
- Created `format_table()` helper in formatting.py
- Uses tabulate with grid format for consistent display
- Handles rich Text objects by extracting plain text
- Match existing color scheme (green/yellow/red for status)
