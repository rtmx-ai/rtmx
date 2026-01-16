# REQ-PM-002: Velocity calculation and burndown charts

## Status: NOT STARTED
## Priority: HIGH
## Phase: 17
## Estimated Effort: 2.0 weeks

## Description

System shall calculate team velocity from completed sprints and generate ASCII burndown charts for sprint progress visualization. Velocity metrics enable evidence-based sprint planning by analyzing historical completion rates.

## Acceptance Criteria

- [ ] `rtmx velocity` displays current sprint velocity and rolling average
- [ ] `rtmx velocity --history` shows velocity for all completed sprints
- [ ] Velocity calculated as sum of story points completed per sprint
- [ ] Rolling average velocity computed over configurable window (default: 3 sprints)
- [ ] `rtmx burndown` displays ASCII burndown chart for active sprint
- [ ] `rtmx burndown <sprint_id>` displays burndown for specific sprint
- [ ] Burndown chart shows ideal line vs actual progress
- [ ] Burndown X-axis shows days in sprint, Y-axis shows remaining story points
- [ ] Burndown handles weekends/non-working days based on configuration
- [ ] Velocity calculation excludes incomplete/carried-over requirements
- [ ] `rtmx sprint show <sprint_id>` includes velocity summary when sprint is completed
- [ ] Export velocity data as JSON for external analysis (`rtmx velocity --format json`)
- [ ] Burndown chart updates in real-time as requirements are completed

## Test Cases

- `tests/test_velocity.py::test_velocity_calculation` - Calculate velocity from completed sprint
- `tests/test_velocity.py::test_velocity_rolling_average` - Rolling average over multiple sprints
- `tests/test_velocity.py::test_velocity_empty_history` - Handle no completed sprints gracefully
- `tests/test_velocity.py::test_velocity_history` - Show all historical velocities
- `tests/test_burndown.py::test_burndown_ascii_chart` - Generate valid ASCII burndown
- `tests/test_burndown.py::test_burndown_ideal_line` - Ideal line calculation is correct
- `tests/test_burndown.py::test_burndown_actual_progress` - Actual progress tracks completions
- `tests/test_burndown.py::test_burndown_weekend_handling` - Non-working days handled correctly
- `tests/test_velocity.py::test_velocity_json_export` - JSON export format is valid

## Technical Notes

### Velocity Calculation

```python
def calculate_velocity(sprint_id: str) -> float:
    """Sum of story points for all COMPLETE requirements in sprint."""
    completed = [r for r in sprint.requirements if r.status == "COMPLETE"]
    return sum(r.story_points or 0 for r in completed)

def rolling_average(window: int = 3) -> float:
    """Average velocity over last N completed sprints."""
    completed_sprints = sorted(
        [s for s in sprints if s.status == "completed"],
        key=lambda s: s.completed_at,
        reverse=True
    )[:window]
    return sum(s.velocity for s in completed_sprints) / len(completed_sprints)
```

### ASCII Burndown Chart Format

```
Sprint 3 Burndown (Feb 12 - Feb 26)
Points
  42 |*
  36 |  *  .
  30 |    *  .
  24 |      *  .
  18 |        *  .
  12 |          *  .
   6 |            *  .
   0 +--+--+--+--+--+--+--+
     D1 D2 D3 D4 D5 D6 D7

Legend: * = Actual  . = Ideal
Remaining: 18 points | Velocity: 4.2/day | Projected: 3 days over
```

### Configuration in rtmx.yaml

```yaml
velocity:
  rolling_window: 3
  working_days: [Mon, Tue, Wed, Thu, Fri]
  chart_width: 60
  chart_height: 15
```

### CLI Examples

```bash
rtmx velocity                    # Current sprint velocity
rtmx velocity --history          # All sprints history
rtmx velocity --format json      # Export as JSON
rtmx burndown                    # Active sprint burndown
rtmx burndown SPRINT-002         # Specific sprint burndown
```

## Dependencies

- REQ-PM-001: Sprint entity and planning commands (requires sprint data)
- REQ-PM-003: Story points/effort estimation (requires point values)

## Blocks

None - this is a leaf requirement in the PM dependency chain.
