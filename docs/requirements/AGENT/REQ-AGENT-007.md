# REQ-AGENT-007: System shall match agent capabilities to requirements

## Status: MISSING
## Priority: MEDIUM
## Phase: 16
## Estimated Effort: 2.0 weeks

## Description
System shall match agent capabilities to requirements enabling intelligent work assignment. Capability matching maps requirement categories and subcategories to agent capabilities, scoring matches to optimize work distribution and ensure appropriate skill alignment.

## Acceptance Criteria
- [ ] Capability matching uses requirement category and subcategory fields
- [ ] Match score calculated as float between 0.0 (no match) and 1.0 (perfect match)
- [ ] Default capability mappings defined for standard categories (CLI, API, TESTING, etc.)
- [ ] Custom capability mappings can be configured in `.rtmx/config.yaml`
- [ ] Agent can register with multiple capabilities
- [ ] Capability wildcards supported (e.g., `*` matches all categories)
- [ ] Match scoring considers partial matches (e.g., "code" matches "code-python")
- [ ] `rtmx match` CLI command shows capability matches for agent/requirement
- [ ] Unmatched requirements reported with suggested capabilities
- [ ] Capability matching feeds into work partitioning algorithm

## Test Cases
- `tests/test_agent.py::test_capability_match_exact` - Exact match scores 1.0
- `tests/test_agent.py::test_capability_match_partial` - Partial match scores < 1.0
- `tests/test_agent.py::test_capability_match_none` - No match scores 0.0
- `tests/test_agent.py::test_capability_wildcard` - Wildcard matches all
- `tests/test_agent.py::test_capability_mapping` - Category to capability mapping
- `tests/test_agent.py::test_capability_custom_config` - Custom mappings from config
- `tests/test_agent.py::test_capability_suggestions` - Unmatched requirements show suggestions

## Technical Notes
Default category-to-capability mappings:
```yaml
capability_mappings:
  CLI: ["code", "python", "cli"]
  API: ["code", "api", "rest"]
  TESTING: ["test", "pytest", "qa"]
  DOCUMENTATION: ["docs", "writing", "markdown"]
  WEB_UI: ["frontend", "react", "typescript"]
  SECURITY: ["security", "audit", "review"]
  AGENT: ["code", "python", "agent", "distributed"]
  CRDT: ["code", "distributed", "sync"]
  REALTIME: ["websocket", "async", "networking"]
```

Match scoring algorithm:
```python
def calculate_match_score(requirement, agent_capabilities):
    category = requirement.category
    required_capabilities = capability_mappings.get(category, [])

    if not required_capabilities:
        return 0.5  # Unknown category, neutral score

    if "*" in agent_capabilities:
        return 1.0  # Wildcard matches all

    matches = set(required_capabilities) & set(agent_capabilities)
    return len(matches) / len(required_capabilities)
```

Match result format:
```json
{
  "req_id": "REQ-AGENT-001",
  "category": "AGENT",
  "agent_id": "claude:abc123",
  "agent_capabilities": ["code", "python", "test"],
  "required_capabilities": ["code", "python", "agent", "distributed"],
  "matched_capabilities": ["code", "python"],
  "score": 0.5,
  "recommendation": "Agent partially capable. Consider handoff for distributed aspects."
}
```

CLI commands:
```bash
rtmx match REQ-AGENT-001                    # Show matches for requirement
rtmx match --agent claude:abc123            # Show matches for agent
rtmx match --threshold 0.8                  # Only show high-quality matches
rtmx match --unmatched                      # Show requirements with no good match
rtmx capabilities                           # List capability mappings
rtmx capabilities --add "code-rust"         # Add capability to current agent
```

## Dependencies
- REQ-AGENT-001: Agent registration protocol (capabilities stored in registration)

## Blocks
- None (leaf requirement, feeds into REQ-AGENT-004)
