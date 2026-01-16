# REQ-AGENT-003: System shall broadcast agent activity in real-time

## Status: MISSING
## Priority: HIGH
## Phase: 16
## Estimated Effort: 2.0 weeks

## Description
System shall broadcast agent activity in real-time enabling humans and other agents to observe what work is being performed. Activity broadcasts provide transparency into agent actions, enabling coordination and allowing intervention when needed.

## Acceptance Criteria
- [ ] Activity broadcast uses WebSocket for real-time delivery
- [ ] Activity types include: claimed, analyzing, implementing, testing, submitting, completed, failed
- [ ] Each activity includes agent_id, req_id, action_type, timestamp, and optional details
- [ ] Activity log persisted to `.rtmx/activity.log` in append-only format
- [ ] Subscribers receive activity events in under 500ms
- [ ] Activity history queryable via `rtmx activity` CLI command
- [ ] Activity can be filtered by agent, requirement, or action type
- [ ] Activity includes progress percentage for long-running operations
- [ ] Failed activities include error details for debugging
- [ ] Activity log rotates when exceeding 10MB

## Test Cases
- `tests/test_agent.py::test_activity_broadcast` - Activity events reach subscribers
- `tests/test_agent.py::test_activity_log` - Activities persisted to log file
- `tests/test_agent.py::test_activity_types` - All action types broadcast correctly
- `tests/test_agent.py::test_activity_filter_agent` - Filter by agent_id works
- `tests/test_agent.py::test_activity_filter_req` - Filter by req_id works
- `tests/test_agent.py::test_activity_latency` - Events delivered under 500ms
- `tests/test_agent.py::test_activity_log_rotation` - Log rotates at size limit

## Technical Notes
Activity event format:
```json
{
  "event_id": "evt_abc123",
  "agent_id": "claude:abc123",
  "req_id": "REQ-AGENT-001",
  "action": "implementing",
  "timestamp": "2024-01-15T10:35:00Z",
  "progress": 45,
  "details": {
    "file": "src/rtmx/agent.py",
    "operation": "writing test cases"
  }
}
```

Action state machine:
```
claimed -> analyzing -> implementing -> testing -> submitting -> completed
                    \                           /
                     \-------> failed <--------/
```

CLI commands:
```bash
rtmx activity                              # Show recent activity
rtmx activity --agent claude:abc123        # Filter by agent
rtmx activity --req REQ-AGENT-001          # Filter by requirement
rtmx activity --action implementing        # Filter by action
rtmx activity --since 1h                   # Last hour
rtmx activity --follow                     # Stream live activity
```

## Dependencies
- REQ-AGENT-001: Agent registration protocol (agent must be registered)
- REQ-RT-002: WebSocket push updates (broadcast mechanism)
- REQ-COLLAB-006: Real-time cursors/presence (presence awareness)

## Blocks
- REQ-AGENT-006: Verification/acceptance workflow
