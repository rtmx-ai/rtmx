# REQ-AGENT-006: System shall require human verification for agent work

## Status: MISSING
## Priority: HIGH
## Phase: 16
## Estimated Effort: 3.0 weeks

## Description
System shall require human verification for agent work implementing a human-in-the-loop workflow. Agent-completed work enters PENDING_VERIFICATION status until a human reviewer accepts or rejects the changes, ensuring quality control and maintaining accountability.

## Acceptance Criteria
- [ ] New requirement status PENDING_VERIFICATION added to status enum
- [ ] Agent-completed work automatically transitions to PENDING_VERIFICATION
- [ ] `rtmx verify` CLI command lists requirements awaiting verification
- [ ] `rtmx verify --accept REQ-ID` accepts agent work, transitions to COMPLETE
- [ ] `rtmx verify --reject REQ-ID --reason "..."` rejects work with feedback
- [ ] Rejected work transitions back to IN_PROGRESS with rejection context
- [ ] Verification decision logged in audit trail with reviewer identity
- [ ] Activity broadcast sent on verification accept/reject
- [ ] Verification dashboard shows pending items with age and agent
- [ ] Verification can include review comments attached to requirement
- [ ] Auto-escalation notification when verification pending > 24 hours

## Test Cases
- `tests/test_agent.py::test_verify_status` - PENDING_VERIFICATION status works
- `tests/test_agent.py::test_verify_auto_transition` - Agent completion triggers status
- `tests/test_agent.py::test_verify_accept` - Accept transitions to COMPLETE
- `tests/test_agent.py::test_verify_reject` - Reject transitions to IN_PROGRESS
- `tests/test_agent.py::test_verify_audit` - Verification logged with reviewer
- `tests/test_agent.py::test_verify_rejection_context` - Rejection reason preserved
- `tests/test_agent.py::test_verify_escalation` - Escalation after 24 hours

## Technical Notes
Status transition diagram with PENDING_VERIFICATION:
```
MISSING -> IN_PROGRESS -> PENDING_VERIFICATION -> COMPLETE
                ^                    |
                |                    v
                +-------- (rejected) ---------+
```

Verification record format:
```json
{
  "verification_id": "ver_abc123",
  "req_id": "REQ-AGENT-001",
  "agent_id": "claude:abc123",
  "submitted_at": "2024-01-15T10:45:00Z",
  "reviewer_id": "user:ryan",
  "decision": "accepted",
  "decided_at": "2024-01-15T11:30:00Z",
  "comments": "Good implementation, tests comprehensive",
  "artifacts": [
    "src/rtmx/agent.py",
    "tests/test_agent.py"
  ]
}
```

Rejection context (attached to requirement for agent retry):
```json
{
  "rejection_id": "rej_xyz789",
  "req_id": "REQ-AGENT-001",
  "reason": "Tests missing edge case for empty input",
  "feedback": "Please add test for empty capabilities list",
  "rejected_at": "2024-01-15T11:30:00Z",
  "reviewer_id": "user:ryan"
}
```

CLI commands:
```bash
rtmx verify                                  # List pending verifications
rtmx verify --agent claude:abc123            # Filter by agent
rtmx verify --accept REQ-AGENT-001           # Accept work
rtmx verify --accept REQ-AGENT-001 --comment "LGTM"
rtmx verify --reject REQ-AGENT-001 --reason "Missing edge case test"
rtmx verify --stats                          # Verification statistics
```

## Dependencies
- REQ-AGENT-002: Work claiming with distributed locks (claim lifecycle)
- REQ-AGENT-003: Agent activity broadcast (verification notifications)
- REQ-COLLAB-005: Audit logging (verification audit trail)

## Blocks
- None (leaf requirement in AGENT phase)
