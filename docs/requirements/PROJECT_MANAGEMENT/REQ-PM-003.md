# REQ-PM-003: Story points and effort estimation

## Status: NOT STARTED
## Priority: HIGH
## Phase: 17
## Estimated Effort: 1.5 weeks

## Description

System shall support story point estimation using Fibonacci scale alongside existing effort_weeks field. Story points provide relative complexity estimation for sprint planning and velocity calculation. Both estimation methods shall coexist, with story points recommended for agile workflows.

## Acceptance Criteria

- [ ] New `story_points` column added to RTM database schema
- [ ] Story points use Fibonacci scale: 1, 2, 3, 5, 8, 13, 21 (configurable)
- [ ] `rtmx estimate <req_id> --points <value>` sets story points
- [ ] `rtmx estimate <req_id> --effort <weeks>` sets effort_weeks (existing)
- [ ] Invalid story point values (not in scale) are rejected with helpful error
- [ ] `rtmx status` output includes story points column when present
- [ ] Story points and effort_weeks can coexist on same requirement
- [ ] `rtmx backlog` shows total story points for backlog
- [ ] Planning poker scale is configurable in `rtmx.yaml`
- [ ] Bulk estimation: `rtmx estimate --batch` reads from stdin or file
- [ ] `rtmx estimate <req_id>` without flags shows current estimates
- [ ] Schema validation allows null/empty story_points for unestimated requirements
- [ ] Export formats (JSON, CSV) include story_points field

## Test Cases

- `tests/test_estimate.py::test_set_story_points_valid` - Set valid Fibonacci value
- `tests/test_estimate.py::test_set_story_points_invalid` - Reject non-Fibonacci value
- `tests/test_estimate.py::test_set_effort_weeks` - Set effort weeks value
- `tests/test_estimate.py::test_coexist_points_and_weeks` - Both values on same requirement
- `tests/test_estimate.py::test_custom_scale` - Custom point scale from config
- `tests/test_estimate.py::test_batch_estimation` - Bulk estimate from file
- `tests/test_estimate.py::test_status_shows_points` - Points in status output
- `tests/test_estimate.py::test_backlog_total_points` - Backlog shows total points
- `tests/test_estimate.py::test_export_includes_points` - JSON export has story_points

## Technical Notes

### Schema Addition

Add to RTM CSV schema:
```csv
req_id,...,effort_weeks,story_points,...
REQ-PM-001,...,2.5,8,...
REQ-PM-002,...,2.0,5,...
```

### Fibonacci Scale Configuration

```yaml
estimation:
  scale: [1, 2, 3, 5, 8, 13, 21]  # Fibonacci
  # Alternative scales:
  # scale: [1, 2, 4, 8, 16]       # Powers of 2
  # scale: [XS, S, M, L, XL]      # T-shirt sizes
  default_unit: points  # or 'weeks'
  allow_custom_values: false
```

### CLI Examples

```bash
rtmx estimate REQ-PM-001 --points 8
rtmx estimate REQ-PM-001 --effort 2.5
rtmx estimate REQ-PM-001  # Shows: story_points=8, effort_weeks=2.5

# Bulk estimation
cat estimates.csv | rtmx estimate --batch
# estimates.csv format: req_id,story_points
# REQ-PM-001,8
# REQ-PM-002,5
```

### Estimation Display in Status

```
$ rtmx status
ID           Status    Points  Effort   Description
REQ-PM-001   MISSING   8       2.5w     Sprint entity and planning commands
REQ-PM-002   MISSING   5       2.0w     Velocity calculation and burndown charts
REQ-PM-003   PARTIAL   3       1.5w     Story points/effort estimation

Total: 3 requirements | 16 points | 6.0 weeks effort
```

### Validation Rules

```python
def validate_story_points(value: int | None, scale: list[int]) -> bool:
    """Validate story point value against configured scale."""
    if value is None:
        return True  # Unestimated is valid
    return value in scale
```

## Dependencies

None - this is a foundational requirement for the PM feature set.

## Blocks

- REQ-PM-001: Sprint entity and planning commands (sprint capacity uses points)
- REQ-PM-002: Velocity calculation and burndown charts (velocity based on points)
