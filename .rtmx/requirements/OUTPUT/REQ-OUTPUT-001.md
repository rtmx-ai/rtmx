# REQ-OUTPUT-001: Capped Progress Bar Width and Aligned Status Columns

## Metadata
- **Category**: OUTPUT
- **Subcategory**: Formatting
- **Priority**: HIGH
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: (none)

## Requirement

`rtmx status -v` and `rtmx status -vv` shall cap progress bar width
at a configurable maximum (default 60 characters) regardless of terminal
width, and shall right-align the percentage and requirement count
columns for consistent vertical alignment.

## Rationale

On wide terminals (120+ columns), progress bars stretch to fill the
entire width, creating a wall of color with no additional information
density. The percentage labels float at the far right edge, separated
from the bars by empty space. The percentage column is left-aligned,
causing `0.0%` and `100.0%` to have different widths, which misaligns
the `(N reqs)` counts.

Screenshot evidence shows on a ~170-column terminal:
- Bars span ~140 characters, far past useful visual resolution
- `100.0% (18 reqs)` is pushed to the far right margin
- `0.0%` and `100.0%` don't align, cascading into misaligned counts

## Design

### Progress Bar Cap

Add `MaxBarWidth` constant (default 60) to `internal/output/terminal.go`.
All progress bar width calculations clamp to `min(computed, MaxBarWidth)`.

```go
const MaxBarWidth = 60

func ClampBarWidth(computed int) int {
    if computed > MaxBarWidth {
        return MaxBarWidth
    }
    if computed < 10 {
        return 10
    }
    return computed
}
```

### Right-Aligned Percentage

Change `FormatPercent` to produce a fixed-width string (6 chars: `100.0%`).
Use `%6s` formatting so `0.0%` becomes `  0.0%` and aligns with `100.0%`.

### Fixed-Width Count Column

Format `(N reqs)` with right-aligned count: `(%3d reqs)` so
`(4 reqs)` becomes `(  4 reqs)` and aligns with `(218 reqs)`.

### Layout

Before:
```
AGENT        : [============================...==================] 100.0% (18 reqs)
HITL         : [                                                 ] 0.0% (4 reqs)
```

After:
```
AGENT        : [============================================================] 100.0%  (18 reqs)
HITL         : [                                                            ]   0.0%   (4 reqs)
```

The bar stops at 60 chars. Percentage is right-aligned in a 6-char field.
Count is right-aligned in a fixed-width field.

### Configuration

The max bar width can be overridden via config:

```yaml
rtmx:
  output:
    max_bar_width: 60  # default
```

This allows narrow-terminal users to reduce and wide-terminal users
to increase if they prefer.

## Acceptance Criteria

1. Progress bars never exceed MaxBarWidth characters (default 60)
2. Percentage column is right-aligned (6 chars: `100.0%`, `  0.0%`)
3. Count column is right-aligned (`(  4 reqs)`, `( 18 reqs)`)
4. Layout is consistent regardless of terminal width (80 to 300 cols)
5. Existing tests pass with updated formatting expectations
6. Works in both `-v` (category) and `-vv` (phase+category) views
7. `--by-version` view also uses capped bars

## Files to Create/Modify

- `internal/output/terminal.go` -- Add MaxBarWidth, ClampBarWidth
- `internal/output/format.go` -- Update FormatPercent to fixed-width
- `internal/cmd/status.go` -- Use ClampBarWidth in all display functions
- `internal/cmd/status_test.go` -- Test alignment properties

## Effort Estimate

0.5 weeks
