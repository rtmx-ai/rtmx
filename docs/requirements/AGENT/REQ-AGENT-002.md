# REQ-AGENT-002: System shall support work claiming with distributed locks

## Status: MISSING
## Priority: HIGH
## Phase: 16
## Estimated Effort: 2.5 weeks

## Description
System shall support work claiming with distributed locks enabling multiple agents to safely claim and work on requirements without conflicts. Lock files prevent multiple agents from working on the same requirement simultaneously, ensuring work coordination in multi-agent environments.

## Acceptance Criteria
- [ ] Claim locks stored as `.rtmx/claims/{req_id}.lock` files
- [ ] Lock files are CRDT-backed for distributed consistency
- [ ] Locks have TTL of 30 minutes with automatic expiration
- [ ] Agent can hold maximum 3 concurrent claims
- [ ] Claim includes agent_id, claimed_at timestamp, and TTL
- [ ] Agent can release claim explicitly before TTL expiration
- [ ] Claim refresh extends TTL without releasing lock
- [ ] Attempting to claim already-locked requirement returns error with lock holder info
- [ ] Expired locks are automatically cleaned up
- [ ] Orphaned locks (agent deregistered) are released automatically
- [ ] `rtmx claim` CLI command claims requirement for current agent
- [ ] `rtmx unclaim` CLI command releases claim

## Test Cases
- `tests/test_agent.py::test_claim_requirement` - Agent can claim available requirement
- `tests/test_agent.py::test_claim_already_claimed` - Claiming locked requirement returns error
- `tests/test_agent.py::test_claim_max_concurrent` - Fourth claim rejected when at limit
- `tests/test_agent.py::test_claim_ttl_expiration` - Lock expires after TTL
- `tests/test_agent.py::test_claim_refresh` - Refreshing claim extends TTL
- `tests/test_agent.py::test_unclaim` - Releasing claim removes lock file
- `tests/test_agent.py::test_claim_orphan_cleanup` - Orphaned locks cleaned on agent deregister

## Technical Notes
Lock file format (`.rtmx/claims/REQ-AGENT-001.lock`):
```json
{
  "req_id": "REQ-AGENT-001",
  "agent_id": "claude:abc123",
  "claimed_at": "2024-01-15T10:30:00Z",
  "ttl_seconds": 1800,
  "expires_at": "2024-01-15T11:00:00Z",
  "refresh_count": 0
}
```

The CRDT backing enables lock state to converge across distributed systems. Lock acquisition uses last-writer-wins with vector clocks for conflict resolution.

CLI commands:
```bash
rtmx claim REQ-AGENT-001              # Claim requirement
rtmx claim REQ-AGENT-001 --ttl 3600   # Claim with custom TTL (1 hour)
rtmx claim --refresh REQ-AGENT-001    # Refresh existing claim
rtmx unclaim REQ-AGENT-001            # Release claim
rtmx claims                           # List current claims for agent
```

## Dependencies
- REQ-AGENT-001: Agent registration protocol (must be registered to claim)
- REQ-COLLAB-003: User/agent collaboration (claim ownership transfer)
- REQ-COLLAB-004: Conflict resolution (claim conflict handling)

## Blocks
- REQ-AGENT-004: Parallel work partitioning
- REQ-AGENT-005: Work handoff protocol
- REQ-AGENT-006: Verification/acceptance workflow
