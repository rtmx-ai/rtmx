# REQ-UX-009: CLI Design System

## Status: MISSING
## Priority: MEDIUM
## Phase: 5

## Description

RTMX CLI shall follow a consistent design system for all terminal output, ensuring visual consistency across commands, verbosity levels, and output modes.

## Rationale

- Users expect consistent visual language across all CLI commands
- Inconsistent icons, colors, and formats create confusion
- A defined design system enables maintainable, predictable output
- Centralized styling makes future theming and customization easier

## Design System Specification

### 1. Status Icons

All commands shall use this consistent icon vocabulary:

| Status | Icon | Color | Usage |
|--------|------|-------|-------|
| Complete | `✓` | Green | Requirement fully implemented and tested |
| Partial | `⚠` | Yellow | Requirement partially implemented |
| Missing | `✗` | Red | Requirement not implemented |
| Not Started | `○` | Dim | Requirement not yet begun |
| Blocked | `⊘` | Dim | Requirement blocked by dependencies |

### 2. Color Palette

Standard ANSI colors for semantic meaning:

| Color | ANSI Code | Usage |
|-------|-----------|-------|
| Green | `\033[92m` | Success, complete, positive |
| Yellow | `\033[93m` | Warning, partial, in-progress |
| Red | `\033[91m` | Error, missing, critical |
| Cyan | `\033[96m` | Headers, sections, info |
| Blue | `\033[94m` | Links, references, medium priority |
| Magenta | `\033[95m` | Special emphasis |
| Dim | `\033[2m` | Secondary info, disabled |
| Bold | `\033[1m` | Emphasis, headers |

### 3. Output Formats by Verbosity

| Level | Format | Content |
|-------|--------|---------|
| Default (v0) | Rich panels or progress bar | Summary statistics |
| `-v` | Grid table | Category breakdown |
| `-vv` | Grid table | Subcategory breakdown |
| `-vvv` | Grid table | Individual items |

### 4. Table Format

All tabular output shall:
- Use `tabulate` with `grid` format
- Include column headers
- Truncate long text with `...`
- Align columns consistently
- Adapt to terminal width (REQ-UX-008)

### 5. Header Format

Section headers shall use:
```
=============================== Title ===============================
```

With `=` characters filling to standard width (80 columns default).

### 6. Count Display Format

Status counts shall use format:
```
✓ N complete  ⚠ N partial  ✗ N missing
```

With appropriate colors applied to each segment.

### 7. Phase Display

Phase references shall show full display name:
- Full: `Phase N (Name)` - for headers, tables, Rich output
- Compact: `[PN]` - only in dense listings where space is critical

## Acceptance Criteria

- [ ] All status icons use the defined icon vocabulary
- [ ] `status.py` and `backlog.py` use identical icons for same statuses
- [ ] Colors are consistently applied per semantic meaning
- [ ] Verbosity levels follow the defined format rules
- [ ] All tables use `tabulate` with grid format
- [ ] Icon constants defined in single location (`formatting.py`)
- [ ] No hardcoded icons/colors outside `formatting.py`

## Test Cases

- `tests/test_cli_ux.py::TestDesignSystem::test_status_icons_consistent`
- `tests/test_cli_ux.py::TestDesignSystem::test_all_commands_use_formatting_module`
- `tests/test_cli_ux.py::TestDesignSystem::test_partial_icon_is_warning`

## Implementation Notes

### Centralized Constants in `formatting.py`

```python
# Single source of truth for status representation
STATUS_ICONS = {
    Status.COMPLETE: "✓",
    Status.PARTIAL: "⚠",
    Status.MISSING: "✗",
    Status.NOT_STARTED: "○",
}

STATUS_COLORS = {
    Status.COMPLETE: Colors.GREEN,
    Status.PARTIAL: Colors.YELLOW,
    Status.MISSING: Colors.RED,
    Status.NOT_STARTED: Colors.DIM,
}
```

### Migration Path

1. Define constants in `formatting.py` (if not already present)
2. Update `backlog.py` to use `status_icon()` from formatting module
3. Remove duplicate icon definitions
4. Add tests verifying consistency

## Dependencies

- REQ-UX-002: Aligned fixed-width columns (tables)
- REQ-UX-008: Width-adaptive output

## Effort

1.0 weeks
