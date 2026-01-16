# REQ-AGENT-001: System shall provide agent registration protocol

## Status: MISSING
## Priority: HIGH
## Phase: 16
## Estimated Effort: 2.0 weeks

## Description
System shall provide an agent registration protocol enabling AI agents and automated workers to register themselves with the RTMX system. Registration establishes agent identity, declares capabilities, and enables participation in distributed work coordination.

## Acceptance Criteria
- [ ] Agent ID follows format `{agent_type}:{instance_id}` (e.g., `claude:abc123`, `cursor:xyz789`)
- [ ] Registration includes capabilities list declaring what work types the agent can perform
- [ ] MCP tool `rtmx_agent_register` is exposed for agent registration
- [ ] Registration is persisted in `.rtmx/agents/` directory as JSON
- [ ] Agent can update registration (re-register with updated capabilities)
- [ ] Agent can deregister (explicit cleanup)
- [ ] Registration includes heartbeat timestamp for liveness tracking
- [ ] Duplicate agent_id registration updates existing entry rather than creating duplicate
- [ ] Registration validates agent_type against allowed types (claude, cursor, copilot, custom)
- [ ] Registration fails gracefully with helpful error message on invalid input

## Test Cases
- `tests/test_agent.py::test_agent_register_valid` - Valid registration creates agent entry
- `tests/test_agent.py::test_agent_register_invalid_id` - Invalid ID format is rejected
- `tests/test_agent.py::test_agent_register_capabilities` - Capabilities list is stored correctly
- `tests/test_agent.py::test_agent_reregister` - Re-registration updates existing entry
- `tests/test_agent.py::test_agent_deregister` - Deregistration removes agent entry
- `tests/test_agent.py::test_agent_heartbeat` - Heartbeat updates timestamp

## Technical Notes
Agent registration format:
```json
{
  "agent_id": "claude:abc123",
  "agent_type": "claude",
  "instance_id": "abc123",
  "capabilities": ["code", "test", "documentation", "review"],
  "registered_at": "2024-01-15T10:30:00Z",
  "last_heartbeat": "2024-01-15T10:35:00Z",
  "metadata": {
    "version": "1.0.0",
    "model": "claude-opus-4-5-20251101"
  }
}
```

The `rtmx_agent_register` MCP tool signature:
```python
def rtmx_agent_register(
    agent_type: str,
    instance_id: str,
    capabilities: list[str],
    metadata: dict | None = None
) -> AgentRegistration
```

## Dependencies
- REQ-CRDT-001: CRDT layer for distributed state
- REQ-COLLAB-002: Presence tracking for awareness

## Blocks
- REQ-AGENT-002: Work claiming with distributed locks
- REQ-AGENT-003: Agent activity broadcast
- REQ-AGENT-004: Parallel work partitioning
- REQ-AGENT-007: Agent capability matching
