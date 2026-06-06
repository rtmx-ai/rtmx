# REQ-API-006: Release Scope and Gate Endpoint

## Metadata
- **Category**: API
- **Subcategory**: REST
- **Priority**: HIGH
- **Phase**: 25
- **Status**: MISSING
- **Dependencies**: REQ-API-001
- **Blocks**: REQ-DASH-006

## Requirement

The RTMX serve command shall expose `GET /api/releases` and
`GET /api/releases/:version` endpoints that return release scope, gate
status, and version-scoped requirement summaries, matching the data
available via `rtmx release scope` and `rtmx release gate`.

## Rationale

Release planning is a core project management activity. The GUI dashboard
needs release scope visibility to support sprint planning views and release
gate checks without shelling out to CLI commands.

## Design

### GET /api/releases

```json
{
  "versions": [
    {
      "version": "v1.0.0",
      "total": 150,
      "complete": 150,
      "partial": 0,
      "missing": 0,
      "completion_pct": 100.0,
      "gate_status": "PASS"
    },
    {
      "version": "v1.2.0",
      "total": 50,
      "complete": 47,
      "partial": 0,
      "missing": 3,
      "completion_pct": 94.0,
      "gate_status": "FAIL"
    },
    {
      "version": "",
      "label": "unversioned",
      "total": 41,
      "complete": 41,
      "partial": 0,
      "missing": 0,
      "completion_pct": 100.0
    }
  ]
}
```

### GET /api/releases/:version

```json
{
  "version": "v1.2.0",
  "gate_status": "FAIL",
  "gate_failures": [
    "3 requirements not COMPLETE: REQ-MCP-007, REQ-MCP-008, REQ-MCP-009"
  ],
  "requirements": [...],
  "summary": {
    "total": 50,
    "complete": 47,
    "completion_pct": 94.0,
    "total_effort_remaining": 2.5
  }
}
```

## Acceptance Criteria

1. `GET /api/releases` lists all versions with completion summaries.
2. Unversioned requirements grouped under empty-string version with label "unversioned".
3. `gate_status` is "PASS" when all requirements in version are COMPLETE.
4. `GET /api/releases/v1.2.0` returns version-scoped detail with gate failures.
5. Unknown version returns 404.
6. Data matches `rtmx release scope` and `rtmx release gate` CLI output.

## Files to Create/Modify

- `internal/cmd/serve_api.go` -- Add /api/releases handlers
- `internal/cmd/serve_api_test.go` -- Release endpoint tests

## Effort Estimate

0.5 weeks

## Test Strategy

- Multiple versions in test database produce correct summaries
- Gate pass/fail logic matches CLI behavior
- Unknown version returns 404
- Empty database returns empty versions array
