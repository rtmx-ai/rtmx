# REQ-PM-001: Sprint entity and planning commands

## Status: NOT STARTED
## Priority: HIGH
## Phase: 17
## Estimated Effort: 2.5 weeks

## Description

System shall provide a Sprint entity with full lifecycle management through CLI commands. Sprints enable time-boxed requirement grouping for iterative development planning. Sprint configuration is managed in `rtmx.yaml` and sprint data is persisted alongside the RTM database.

## Acceptance Criteria

- [ ] Sprint entity has fields: `sprint_id`, `name`, `start_date`, `end_date`, `status`, `goal`
- [ ] Sprint status follows lifecycle: `planned` -> `active` -> `completed`
- [ ] Only one sprint can be in `active` status at a time
- [ ] `rtmx sprint list` displays all sprints with status, dates, and requirement counts
- [ ] `rtmx sprint create <name> --start <date> --end <date> --goal <goal>` creates new sprint
- [ ] `rtmx sprint assign <sprint_id> <req_id>...` assigns requirements to a sprint
- [ ] `rtmx sprint unassign <sprint_id> <req_id>...` removes requirements from a sprint
- [ ] `rtmx sprint start <sprint_id>` activates a sprint (transitions from planned to active)
- [ ] `rtmx sprint close <sprint_id>` completes a sprint (transitions from active to completed)
- [ ] Sprint configuration (default duration, naming convention) stored in `rtmx.yaml`
- [ ] Sprint data persisted in `docs/sprints.csv` or similar CSV format
- [ ] Requirements can only be assigned to one active sprint at a time
- [ ] `rtmx sprint show <sprint_id>` displays sprint details with assigned requirements
- [ ] Sprint list output shows total story points per sprint (when story points are available)

## Test Cases

- `tests/test_sprint.py::test_sprint_create` - Create new sprint with valid parameters
- `tests/test_sprint.py::test_sprint_create_invalid_dates` - Reject sprint with end before start
- `tests/test_sprint.py::test_sprint_list` - List all sprints with correct counts
- `tests/test_sprint.py::test_sprint_assign_requirements` - Assign requirements to sprint
- `tests/test_sprint.py::test_sprint_unassign_requirements` - Unassign requirements from sprint
- `tests/test_sprint.py::test_sprint_start` - Activate planned sprint
- `tests/test_sprint.py::test_sprint_start_conflict` - Reject starting when another sprint active
- `tests/test_sprint.py::test_sprint_close` - Complete active sprint
- `tests/test_sprint.py::test_sprint_status_transitions` - Validate status state machine
- `tests/test_sprint.py::test_sprint_single_assignment` - Requirement only in one active sprint

## Technical Notes

### Sprint Configuration in rtmx.yaml

```yaml
sprints:
  default_duration_days: 14
  naming_convention: "Sprint {number}"
  allow_carryover: true
  auto_increment: true
```

### Sprint CSV Schema

```csv
sprint_id,name,start_date,end_date,status,goal,created_at,completed_at
SPRINT-001,Sprint 1,2024-01-15,2024-01-29,completed,Complete core features,2024-01-14T10:00:00Z,2024-01-29T17:00:00Z
SPRINT-002,Sprint 2,2024-01-29,2024-02-12,active,API stabilization,2024-01-28T10:00:00Z,
```

### Sprint-Requirement Assignment

Stored in RTM database as new column `sprint_id` or separate junction table `sprint_requirements.csv`:

```csv
sprint_id,req_id,assigned_at,completed_in_sprint
SPRINT-002,REQ-CLI-001,2024-01-29T10:00:00Z,true
SPRINT-002,REQ-TEST-003,2024-01-29T10:00:00Z,false
```

### CLI Examples

```bash
rtmx sprint create "Sprint 3" --start 2024-02-12 --end 2024-02-26 --goal "Phase 17 PM features"
rtmx sprint assign SPRINT-003 REQ-PM-001 REQ-PM-002 REQ-PM-003
rtmx sprint list
rtmx sprint start SPRINT-003
rtmx sprint show SPRINT-003
rtmx sprint close SPRINT-003
```

## Dependencies

- REQ-PM-003: Story points/effort estimation (for sprint capacity planning)

## Blocks

- REQ-PM-002: Velocity calculation and burndown charts (requires sprint history)
