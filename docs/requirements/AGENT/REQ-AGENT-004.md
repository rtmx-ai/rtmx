# REQ-AGENT-004: System shall partition work for parallel agent execution

## Status: MISSING
## Priority: MEDIUM
## Phase: 16
## Estimated Effort: 2.5 weeks

## Description
System shall partition work for parallel agent execution enabling multiple agents to work on independent requirements simultaneously. Work partitioning respects dependency ordering, matches agent capabilities to requirement types, and balances effort across available agents.

## Acceptance Criteria
- [ ] Work partitioning uses DAG-based algorithm respecting requirement dependencies
- [ ] Requirements with unsatisfied dependencies are not assigned
- [ ] Agent capabilities are matched to requirement categories/subcategories
- [ ] Effort is balanced across agents based on estimated hours
- [ ] `rtmx partition` CLI command generates work assignments
- [ ] Partition output includes assignment rationale
- [ ] Maximum concurrent work per agent respects claim limits
- [ ] Partitioning considers agent current workload
- [ ] Partitioning can be filtered by phase or priority
- [ ] Dry-run mode shows proposed assignments without claiming

## Test Cases
- `tests/test_agent.py::test_partition_dag_order` - Dependencies respected in assignments
- `tests/test_agent.py::test_partition_capability_match` - Capabilities matched to categories
- `tests/test_agent.py::test_partition_effort_balance` - Work balanced across agents
- `tests/test_agent.py::test_partition_blocked_deps` - Blocked requirements excluded
- `tests/test_agent.py::test_partition_agent_workload` - Current claims considered
- `tests/test_agent.py::test_partition_dry_run` - Dry-run shows assignments only
- `tests/test_agent.py::test_partition_by_phase` - Phase filter works correctly

## Technical Notes
Partitioning algorithm:
1. Build requirement dependency DAG
2. Topological sort to identify available requirements (no blocked deps)
3. Score each (agent, requirement) pair based on capability match
4. Assign highest-scoring pairs while respecting:
   - Agent claim limits (max 3)
   - Current agent workload
   - Effort balance target
5. Output assignment plan with rationale

Assignment output format:
```json
{
  "partition_id": "part_abc123",
  "created_at": "2024-01-15T10:30:00Z",
  "assignments": [
    {
      "agent_id": "claude:abc123",
      "req_id": "REQ-AGENT-001",
      "capability_score": 0.95,
      "estimated_hours": 16,
      "rationale": "Agent has 'code' capability matching AGENT category"
    }
  ],
  "unassigned": [
    {
      "req_id": "REQ-AGENT-006",
      "reason": "Blocked by REQ-AGENT-002 (in progress)"
    }
  ]
}
```

CLI commands:
```bash
rtmx partition                           # Generate assignments for registered agents
rtmx partition --phase 16                # Partition specific phase
rtmx partition --dry-run                 # Show assignments without claiming
rtmx partition --agent claude:abc123     # Partition for specific agent
rtmx partition --max-per-agent 2         # Override concurrent limit
```

## Dependencies
- REQ-AGENT-001: Agent registration protocol (agents must be registered)
- REQ-AGENT-002: Work claiming with distributed locks (claims enable work)

## Blocks
- None (leaf requirement in AGENT phase)
