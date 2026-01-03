# REQ-UX-001: rtmx status shall display rich progress bars per phase

## Status: MISSING
## Priority: HIGH
## Phase: 5

## Description
rtmx status shall display rich progress bars per phase using the `rich` library for enhanced terminal visualization.

## Acceptance Criteria
- [ ] `rich` is available as optional dependency (`rtmx[rich]`)
- [ ] Progress bars render with colored segments (green=complete, yellow=partial, red=missing)
- [ ] Each phase displays its own progress bar
- [ ] Overall progress displays at top with percentage
- [ ] Box-drawing characters create visual panels
- [ ] Graceful fallback to current output when `rich` not installed
- [ ] `--rich` flag forces rich output (error if not installed)
- [ ] `--no-rich` flag forces plain output

## Visual Design

```
╭─ RTMX Status ───────────────────────────────────────────╮
│ Overall: ████████████████████████████████░░░░░░░░  85%  │
│                                                         │
│ ✓ 27 complete  ⚠ 0 partial  ✗ 43 missing               │
╰─────────────────────────────────────────────────────────╯

╭─ Phase Progress ────────────────────────────────────────╮
│ Phase 1:  ████████████████████████████████████████ 100% │
│ Phase 2:  ████████████████████████████████████████ 100% │
│ Phase 3:  ████████████████████████████████████████ 100% │
│ Phase 4:  ████████████████████████████████████████ 100% │
│ Phase 5:  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░   0% │
│ Phase 6:  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░   0% │
╰─────────────────────────────────────────────────────────╯
```

## Test Cases
- `tests/test_cli_ux.py::test_rich_progress_bars`
- `tests/test_cli_ux.py::test_rich_fallback_without_library`
- `tests/test_cli_ux.py::test_rich_flag_forces_output`
- `tests/test_cli_ux.py::test_no_rich_flag_forces_plain`

## Implementation Notes
- Use `rich.panel.Panel` for bordered sections
- Use `rich.progress.Progress` with custom columns
- Use `rich.console.Console` for output
- Detect `rich` availability with try/except import
- Store preference in formatting module
