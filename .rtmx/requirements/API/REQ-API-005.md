# REQ-API-005: Backlog Endpoint with Views

## Metadata
- **Category**: API
- **Subcategory**: REST
- **Priority**: HIGH
- **Phase**: 25
- **Status**: MISSING
- **Dependencies**: REQ-API-001
- **Blocks**: REQ-TUI-005, REQ-DASH-005

## Requirement

The RTMX serve command shall expose a `GET /api/backlog` endpoint that
returns the prioritized backlog with the same view modes as `rtmx backlog`
(all, critical, quick-wins, blockers), providing structured JSON suitable
for rendering Kanban boards and prioritized lists.

## Rationale

The backlog is the primary work-planning view. The CLI already supports
multiple views (critical, quick-wins, blockers, list) but the current
serve command exposes none of this. Both TUI and GUI dashboards need
server-side backlog computation to avoid duplicating priority/blocking
analysis in client code.

## Design

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| view | string | all | View mode: `all`, `critical`, `quick-wins`, `blockers` |
| category | string | (all) | Filter by category |
| version | string | (all) | Filter by target version |
| limit | int | 50 | Max items to return |

### Response Schema

```json
{
  "view": "all",
  "sections": [
    {
      "name": "Critical Path",
      "items": [
        {
          "req_id": "REQ-MCP-007",
          "requirement_text": "...",
          "priority": "P0",
          "status": "MISSING",
          "effort_weeks": 0.5,
          "blocked": false,
          "blocks_count": 2,
          "transitive_blocks_count": 2,
          "category": "MCP",
          "assignee": "",
          "sprint": ""
        }
      ]
    },
    {
      "name": "Quick Wins",
      "items": [...]
    }
  ],
  "summary": {
    "total_incomplete": 3,
    "total_effort_weeks": 2.5,
    "unblocked_count": 1,
    "blocked_count": 2
  }
}
```

### View Modes

- **all**: Critical path items, then quick wins, then remaining (matches CLI)
- **critical**: P0/HIGH priority + blocking multiple items
- **quick-wins**: Low effort (<1 week), high priority, unblocked
- **blockers**: Requirements blocking others, sorted by transitive block count

## Acceptance Criteria

1. `GET /api/backlog` returns sectioned backlog matching CLI `rtmx backlog` output.
2. Each view mode returns correctly filtered and sorted items.
3. `blocks_count` and `transitive_blocks_count` are accurate.
4. `blocked` flag reflects whether all upstream dependencies are complete.
5. `summary` totals are consistent with the items returned.
6. `?category=MCP` filters to MCP items only.
7. `?limit=10` caps the total items returned across all sections.

## Files to Create/Modify

- `internal/cmd/serve_api.go` -- Add /api/backlog handler
- `internal/cmd/serve_api_test.go` -- Backlog endpoint tests

## Effort Estimate

0.5 weeks

## Test Strategy

- Each view mode returns expected items for a known test database
- Blocking analysis matches CLI `rtmx backlog --view blockers` output
- Quick-wins criteria correctly applied (effort < 1 week, unblocked)
- Empty backlog (all complete) returns empty sections
- Filter combinations work correctly
