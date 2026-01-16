# REQ-PM-008: Milestone tracking with deadlines

## Status: NOT STARTED
## Priority: MEDIUM
## Phase: 17
## Estimated Effort: 1.5 weeks

## Description

System shall support milestone tracking with deadline dates and automated warning/critical status notifications. Milestones represent key project dates that requirements must meet, with visual indicators for approaching and overdue deadlines.

## Acceptance Criteria

- [ ] `rtmx milestone list` displays all milestones with status and dates
- [ ] `rtmx milestone create <name> --date <deadline>` creates new milestone
- [ ] `rtmx milestone assign <milestone_id> <req_id>...` assigns requirements
- [ ] `rtmx milestone unassign <milestone_id> <req_id>...` removes requirements
- [ ] `rtmx milestone show <milestone_id>` displays milestone with requirements
- [ ] Milestone status auto-calculated: `on_track`, `warning`, `critical`, `achieved`, `missed`
- [ ] Warning status triggered at configurable days before deadline (default: 7)
- [ ] Critical status triggered when deadline is reached or passed
- [ ] `rtmx status` shows milestone indicators for requirements with deadlines
- [ ] `rtmx milestone upcoming` shows milestones due within N days
- [ ] Notification hooks for warning/critical status changes
- [ ] Milestone progress bar shows completion percentage
- [ ] `rtmx milestone report` generates milestone status report
- [ ] Milestones can be linked to releases
- [ ] Missed milestones remain visible with `missed` status

## Test Cases

- `tests/test_milestone.py::test_milestone_create` - Create milestone with deadline
- `tests/test_milestone.py::test_milestone_list` - List all milestones
- `tests/test_milestone.py::test_milestone_assign` - Assign requirements to milestone
- `tests/test_milestone.py::test_milestone_unassign` - Remove requirements from milestone
- `tests/test_milestone.py::test_milestone_show` - Show milestone details
- `tests/test_milestone.py::test_status_on_track` - On-track status calculation
- `tests/test_milestone.py::test_status_warning` - Warning at 7 days
- `tests/test_milestone.py::test_status_critical` - Critical when overdue
- `tests/test_milestone.py::test_status_achieved` - Achieved when all complete
- `tests/test_milestone.py::test_upcoming_milestones` - Filter upcoming milestones
- `tests/test_milestone.py::test_milestone_report` - Generate status report

## Technical Notes

### Milestone CSV Schema

```csv
milestone_id,name,deadline,status,created_at,achieved_at,description
MS-001,Q1 Release,2024-03-31,on_track,2024-01-15T10:00:00Z,,Q1 feature freeze
MS-002,Beta Launch,2024-02-15,warning,2024-01-10T10:00:00Z,,Public beta release
MS-003,Alpha Complete,2024-01-31,achieved,2024-01-05T10:00:00Z,2024-01-30T14:00:00Z,Internal alpha
```

### Milestone-Requirement Junction

```csv
milestone_id,req_id,assigned_at
MS-001,REQ-PM-001,2024-01-15T10:00:00Z
MS-001,REQ-PM-002,2024-01-15T10:00:00Z
MS-002,REQ-PM-003,2024-01-10T10:00:00Z
```

### Status Calculation Logic

```python
from datetime import datetime, timedelta

def calculate_milestone_status(
    milestone: Milestone,
    warning_days: int = 7
) -> str:
    """Calculate milestone status based on deadline and completion."""
    children = get_milestone_requirements(milestone.milestone_id)
    all_complete = all(r.status == "COMPLETE" for r in children)

    if all_complete:
        return "achieved"

    now = datetime.now()
    deadline = milestone.deadline
    days_remaining = (deadline - now).days

    if days_remaining < 0:
        return "missed" if not all_complete else "achieved"
    elif days_remaining == 0:
        return "critical"
    elif days_remaining <= warning_days:
        return "warning"
    else:
        return "on_track"

def calculate_progress(milestone_id: str) -> float:
    """Calculate completion percentage for milestone."""
    reqs = get_milestone_requirements(milestone_id)
    if not reqs:
        return 0.0
    complete = sum(1 for r in reqs if r.status == "COMPLETE")
    return (complete / len(reqs)) * 100
```

### Configuration in rtmx.yaml

```yaml
milestones:
  warning_days: 7
  critical_days: 0
  notify_on_warning: true
  notify_on_critical: true
  date_format: "%Y-%m-%d"
  show_missed: true
```

### CLI Output Examples

```bash
$ rtmx milestone list
ID      Name            Deadline    Status      Progress  Requirements
MS-001  Q1 Release      2024-03-31  on_track    45%       12/27
MS-002  Beta Launch     2024-02-15  warning     80%       8/10
MS-003  Alpha Complete  2024-01-31  achieved    100%      5/5
MS-004  Demo Ready      2024-02-01  critical    60%       3/5

$ rtmx milestone show MS-002
Milestone: MS-002 - Beta Launch
Deadline: 2024-02-15 (5 days remaining)
Status: WARNING
Progress: [========--------] 80% (8/10 complete)

Requirements:
  REQ-PM-001  COMPLETE     Sprint entity
  REQ-PM-002  COMPLETE     Velocity calculation
  REQ-PM-003  COMPLETE     Story points
  REQ-PM-004  IN_PROGRESS  Custom workflows
  REQ-PM-005  MISSING      Epic hierarchy

$ rtmx milestone upcoming --days 14
Upcoming Milestones (next 14 days):
  MS-004  Demo Ready    2024-02-01  critical  60%
  MS-002  Beta Launch   2024-02-15  warning   80%

$ rtmx milestone report
# Milestone Status Report
Generated: 2024-01-27

## Critical (1)
- MS-004: Demo Ready - Due 2024-02-01 - 60% complete

## Warning (1)
- MS-002: Beta Launch - Due 2024-02-15 - 80% complete

## On Track (1)
- MS-001: Q1 Release - Due 2024-03-31 - 45% complete

## Achieved (1)
- MS-003: Alpha Complete - Achieved 2024-01-30
```

### Status Indicators in rtmx status

```bash
$ rtmx status
ID           Status       Points  Milestone        Description
REQ-PM-001   COMPLETE     8       MS-002 (warn)    Sprint entity
REQ-PM-002   COMPLETE     5       MS-002 (warn)    Velocity calculation
REQ-PM-004   IN_PROGRESS  5       MS-004 (crit)    Custom workflows
REQ-PM-005   MISSING      5       MS-001           Epic hierarchy
```

## Dependencies

None - this is an independent deadline tracking feature.

## Blocks

None - this is a leaf requirement for project tracking.
