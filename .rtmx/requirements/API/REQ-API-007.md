# REQ-API-007: Agent Activity and Claims Endpoint

## Metadata
- **Category**: API
- **Subcategory**: REST
- **Priority**: MEDIUM
- **Phase**: 25
- **Status**: MISSING
- **Dependencies**: REQ-API-001
- **Blocks**: REQ-TUI-007, REQ-DASH-008

## Requirement

The RTMX serve command shall expose `GET /api/agents` and
`GET /api/agents/claims` endpoints that return active agent claims,
heartbeat status, and claim history, providing visibility into multi-agent
coordination activity.

## Rationale

When multiple AI agents or human developers are working concurrently via
the RTMX orchestration layer, operators need visibility into who is working
on what. The TUI agent monitor and GUI dashboard both need this data to
show real-time coordination status.

## Design

### GET /api/agents/claims

```json
{
  "active_claims": [
    {
      "req_id": "REQ-MCP-007",
      "agent_id": "claude-001",
      "claimed_at": "2026-06-04T10:30:00Z",
      "last_heartbeat": "2026-06-04T10:35:00Z",
      "stale": false,
      "requirement_text": "RTMX MCP server shall log response byte..."
    }
  ],
  "summary": {
    "total_active": 1,
    "stale_count": 0,
    "agents": ["claude-001"]
  }
}
```

### Stale Detection

A claim is marked `stale: true` when `last_heartbeat` is older than the
configured stale timeout (default 15 minutes). Stale claims are candidates
for `ForceRelease` by operators.

### Implementation

Reads from the existing `.rtmx/claims/` directory managed by the
orchestration package. No new persistence layer needed.

## Acceptance Criteria

1. `GET /api/agents/claims` returns all active claims from `.rtmx/claims/`.
2. Each claim includes agent_id, timestamps, and staleness flag.
3. `summary.agents` is deduplicated list of active agent IDs.
4. Stale detection uses configurable timeout threshold.
5. Empty claims directory returns empty array (not error).
6. Response includes requirement_text for context without additional lookup.

## Files to Create/Modify

- `internal/cmd/serve_api.go` -- Add /api/agents/claims handler
- `internal/cmd/serve_api_test.go` -- Claims endpoint tests

## Effort Estimate

0.5 weeks

## Test Strategy

- Create test claim files, verify endpoint returns them
- Stale detection with manipulated timestamps
- Empty claims directory returns clean empty response
- Concurrent claim file creation during read does not crash
