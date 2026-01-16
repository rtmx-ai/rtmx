# REQ-AGENT-005: System shall support work handoff between agents

## Status: MISSING
## Priority: MEDIUM
## Phase: 16
## Estimated Effort: 2.0 weeks

## Description
System shall support work handoff between agents enabling transfer of claim ownership when an agent cannot complete work. Handoffs preserve context, enable graceful capacity management, and maintain audit trail of work transitions.

## Acceptance Criteria
- [ ] Handoff transfers claim ownership from source agent to target agent
- [ ] Handoff includes context message describing current state and next steps
- [ ] Handoff reason is recorded (capacity, capability, timeout, error)
- [ ] Audit log captures handoff with source, target, reason, and context
- [ ] Target agent must be registered and have capacity for claim
- [ ] Target agent must have capability matching requirement category
- [ ] Source agent claim is released atomically with target acquisition
- [ ] Handoff fails gracefully if target agent unavailable
- [ ] `rtmx handoff` CLI command initiates handoff
- [ ] Handoff notification sent to both agents via activity broadcast

## Test Cases
- `tests/test_agent.py::test_handoff_transfer` - Claim transfers to target agent
- `tests/test_agent.py::test_handoff_context` - Context message preserved
- `tests/test_agent.py::test_handoff_audit` - Handoff logged in audit trail
- `tests/test_agent.py::test_handoff_atomic` - Source released only on target success
- `tests/test_agent.py::test_handoff_capacity_check` - Target must have capacity
- `tests/test_agent.py::test_handoff_capability_check` - Target must have capability
- `tests/test_agent.py::test_handoff_notification` - Both agents notified

## Technical Notes
Handoff reasons enum:
- `capacity`: Agent at claim limit, handing off to free up capacity
- `capability`: Requirement needs capability agent doesn't have
- `timeout`: Agent unable to complete before TTL expiration
- `error`: Agent encountered unrecoverable error

Handoff record format:
```json
{
  "handoff_id": "ho_abc123",
  "req_id": "REQ-AGENT-001",
  "source_agent_id": "claude:abc123",
  "target_agent_id": "cursor:xyz789",
  "reason": "capability",
  "context": "Completed analysis phase. Implementation requires frontend skills. Files modified: src/rtmx/agent.py. Next step: implement UI components.",
  "timestamp": "2024-01-15T10:45:00Z"
}
```

Handoff protocol:
1. Source agent requests handoff with target and context
2. System validates target capacity and capability
3. System atomically releases source claim and acquires target claim
4. Audit log updated with handoff record
5. Activity broadcast sent to both agents
6. Target agent receives handoff notification with context

CLI commands:
```bash
rtmx handoff REQ-AGENT-001 --to cursor:xyz789 --reason capability \
  --context "Completed analysis, needs frontend implementation"
rtmx handoff REQ-AGENT-001 --reason timeout  # Handoff to any capable agent
rtmx handoffs                                 # List recent handoffs
rtmx handoffs --agent claude:abc123           # Handoffs involving agent
```

## Dependencies
- REQ-AGENT-002: Work claiming with distributed locks (handoff transfers claims)
- REQ-COLLAB-005: Audit logging (handoff audit trail)

## Blocks
- None (leaf requirement in AGENT phase)
