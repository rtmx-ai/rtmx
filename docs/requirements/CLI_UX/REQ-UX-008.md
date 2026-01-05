# REQ-UX-008: Width-Adaptive CLI Output

## Status: MISSING

## Description

CLI output shall adapt column widths to terminal size, detecting available width and truncating or adjusting columns dynamically to fit without horizontal overflow.

## Rationale

- Real terminals have varying widths (80, 120, 200+ columns)
- Long descriptions and requirement IDs can overflow narrow terminals
- Website terminal animations need to display well at different viewport sizes
- Users expect CLI tools to respect terminal boundaries

## Acceptance Criteria

- [ ] CLI commands detect terminal width via `shutil.get_terminal_size()`
- [ ] Tables auto-adjust column widths based on available space
- [ ] Long text fields truncate with `...` when space is limited
- [ ] Minimum column widths preserved for readability
- [ ] Output remains aligned and readable at 80-column minimum
- [ ] Rich library integration respects console width
- [ ] Website animations can calculate char capacity from container width

## Technical Approach

### CLI Implementation
```python
import shutil

def get_max_width() -> int:
    """Get terminal width with sensible default."""
    return shutil.get_terminal_size().columns or 120

def truncate_text(text: str, max_len: int) -> str:
    """Truncate text with ellipsis if too long."""
    if len(text) <= max_len:
        return text
    return text[:max_len - 3] + "..."
```

### Website Animation (Future)
```javascript
const charWidth = 8.4; // monospace char width at 14px
const containerWidth = terminal.offsetWidth - 40;
const maxChars = Math.floor(containerWidth / charWidth);

lines.forEach(line => {
  if (line.text.length > maxChars) {
    line.text = line.text.slice(0, maxChars - 3) + '...';
  }
});
```

## Test Cases

1. Output at 80 columns - all tables fit
2. Output at 120 columns - uses extra space
3. Output at 60 columns - graceful truncation
4. Long requirement IDs truncated appropriately
5. Long descriptions truncated with ellipsis
6. Progress bars scale to available width

## Dependencies

- REQ-UX-002: All tabular output shall use aligned fixed-width columns

## Effort

1.5 weeks

## Priority

MEDIUM

## Phase

5 (CLI UX)
