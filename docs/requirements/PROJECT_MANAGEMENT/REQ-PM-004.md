# REQ-PM-004: Custom status workflows with state machine

## Status: NOT STARTED
## Priority: HIGH
## Phase: 17
## Estimated Effort: 2.0 weeks

## Description

System shall support custom status workflows defined as state machines in `rtmx.yaml`. Custom workflows enable teams to define their own requirement lifecycle stages and valid transitions, replacing or extending the default MISSING/PARTIAL/COMPLETE statuses.

## Acceptance Criteria

- [ ] Workflow state machine defined in `rtmx.yaml` under `workflows` key
- [ ] Each workflow has named states with allowed transitions
- [ ] Default workflow (MISSING -> PARTIAL -> COMPLETE) remains available
- [ ] `rtmx transition <req_id> <status>` transitions requirement to new status
- [ ] Invalid transitions are rejected with error showing allowed transitions
- [ ] `rtmx workflow list` shows all configured workflows
- [ ] `rtmx workflow show <workflow_name>` displays states and transitions
- [ ] `rtmx workflow validate` checks all requirements have valid statuses
- [ ] Requirements track status history with timestamps
- [ ] Transition hooks can trigger actions (e.g., notify on BLOCKED)
- [ ] Workflows support initial state and terminal states configuration
- [ ] `rtmx status` respects workflow when displaying requirements
- [ ] Multiple workflows can coexist (e.g., different workflows per category)
- [ ] Workflow visualization as ASCII state diagram

## Test Cases

- `tests/test_workflow.py::test_valid_transition` - Transition between allowed states
- `tests/test_workflow.py::test_invalid_transition` - Reject disallowed transition
- `tests/test_workflow.py::test_custom_workflow_config` - Load workflow from yaml
- `tests/test_workflow.py::test_default_workflow` - Default workflow when none configured
- `tests/test_workflow.py::test_workflow_list` - List all workflows
- `tests/test_workflow.py::test_workflow_show` - Display workflow details
- `tests/test_workflow.py::test_workflow_validate` - Validate all statuses are valid
- `tests/test_workflow.py::test_status_history` - Track transition history
- `tests/test_workflow.py::test_terminal_state` - No transitions from terminal states
- `tests/test_workflow.py::test_workflow_per_category` - Different workflows per category

## Technical Notes

### Workflow Configuration in rtmx.yaml

```yaml
workflows:
  default:
    initial: BACKLOG
    terminal: [COMPLETE, WONTFIX]
    states:
      BACKLOG:
        transitions: [TODO, WONTFIX]
        description: "Requirement captured but not scheduled"
      TODO:
        transitions: [IN_PROGRESS, BLOCKED, BACKLOG]
        description: "Scheduled for current sprint"
      IN_PROGRESS:
        transitions: [REVIEW, BLOCKED, TODO]
        description: "Actively being worked on"
      BLOCKED:
        transitions: [TODO, IN_PROGRESS]
        description: "Blocked by external dependency"
        on_enter: notify_blocked  # Hook for notifications
      REVIEW:
        transitions: [COMPLETE, IN_PROGRESS]
        description: "In review/verification"
      COMPLETE:
        transitions: []
        description: "Done and verified"
      WONTFIX:
        transitions: []
        description: "Intentionally not implementing"

  # Simple workflow for smaller teams
  simple:
    initial: MISSING
    terminal: [COMPLETE]
    states:
      MISSING:
        transitions: [PARTIAL, COMPLETE]
      PARTIAL:
        transitions: [COMPLETE, MISSING]
      COMPLETE:
        transitions: []
```

### State Machine Implementation

```python
@dataclass
class WorkflowState:
    name: str
    transitions: list[str]
    description: str = ""
    on_enter: str | None = None
    on_exit: str | None = None

@dataclass
class Workflow:
    name: str
    initial: str
    terminal: list[str]
    states: dict[str, WorkflowState]

    def can_transition(self, from_state: str, to_state: str) -> bool:
        """Check if transition is valid."""
        if from_state not in self.states:
            return False
        return to_state in self.states[from_state].transitions
```

### CLI Examples

```bash
rtmx transition REQ-PM-001 IN_PROGRESS
# Output: REQ-PM-001: TODO -> IN_PROGRESS

rtmx transition REQ-PM-001 COMPLETE
# Error: Invalid transition IN_PROGRESS -> COMPLETE
# Allowed transitions: REVIEW, BLOCKED, TODO

rtmx workflow show default
# Output: ASCII state diagram

rtmx workflow validate
# Output: All 47 requirements have valid statuses
```

### Status History Storage

```csv
req_id,from_status,to_status,transitioned_at,transitioned_by
REQ-PM-001,BACKLOG,TODO,2024-01-15T10:00:00Z,user@example.com
REQ-PM-001,TODO,IN_PROGRESS,2024-01-16T09:00:00Z,user@example.com
```

### ASCII Workflow Diagram

```
$ rtmx workflow show default

  [BACKLOG] ─────────────────────────────────────────┐
      │                                               │
      ▼                                               ▼
   [TODO] ◄──────────────┐                       [WONTFIX]
      │                  │                         (terminal)
      ▼                  │
[IN_PROGRESS] ◄─────► [BLOCKED]
      │
      ▼
  [REVIEW]
      │
      ▼
 [COMPLETE]
  (terminal)
```

## Dependencies

None - this is an independent enhancement to the status system.

## Blocks

None - other features work with any status workflow.
