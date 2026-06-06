# REQ-API-002: Requirement Detail Endpoint with Dependencies

## Metadata
- **Category**: API
- **Subcategory**: REST
- **Priority**: P0
- **Phase**: 25
- **Status**: MISSING
- **Dependencies**: REQ-API-001
- **Blocks**: REQ-TUI-003, REQ-DASH-003

## Requirement

The RTMX serve command shall expose a `GET /api/requirements/:id` endpoint
that returns full detail for a single requirement, including resolved
upstream and downstream dependency chains with their statuses.

## Rationale

Drill-down from a requirements list into a single requirement is the most
common navigation pattern in any project management tool. The detail view
needs dependency context to answer "what blocks this?" and "what does this
unblock?" without additional round trips.

## Design

### Response Schema

```json
{
  "requirement": {
    "req_id": "REQ-MCP-007",
    "category": "MCP",
    "subcategory": "Observability",
    "requirement_text": "RTMX MCP server shall log response byte and token counts...",
    "status": "MISSING",
    "priority": "P0",
    "phase": 27,
    "effort_weeks": 0.5,
    "assignee": "",
    "sprint": "",
    "dependencies": ["REQ-MCP-003", "REQ-MCP-006"],
    "blocks": ["REQ-MCP-008", "REQ-MCP-009"],
    "started_date": "",
    "completed_date": "",
    "notes": "",
    "target_value": "Log line emitted on every tools/call response",
    "test_module": "internal/adapters/mcp/server_test.go",
    "test_function": "TestMCPResponseSizeLogging",
    "validation_method": "Integration Test",
    "requirement_file": ".rtmx/requirements/MCP/REQ-MCP-007.md",
    "external_id": ""
  },
  "dependency_detail": {
    "upstream": [
      {"req_id": "REQ-MCP-003", "status": "COMPLETE", "requirement_text": "..."},
      {"req_id": "REQ-MCP-006", "status": "COMPLETE", "requirement_text": "..."}
    ],
    "downstream": [
      {"req_id": "REQ-MCP-008", "status": "MISSING", "requirement_text": "..."},
      {"req_id": "REQ-MCP-009", "status": "MISSING", "requirement_text": "..."}
    ],
    "transitive_upstream_count": 2,
    "transitive_downstream_count": 2,
    "all_upstream_complete": true
  }
}
```

### Implementation

```go
mux.HandleFunc("/api/requirements/", func(w http.ResponseWriter, r *http.Request) {
    reqID := strings.TrimPrefix(r.URL.Path, "/api/requirements/")
    req := db.Get(reqID)
    if req == nil {
        writeError(w, 404, "requirement not found: "+reqID)
        return
    }
    detail := buildDependencyDetail(db, req)
    writeJSON(w, RequirementDetail{Requirement: req, DependencyDetail: detail})
})
```

## Acceptance Criteria

1. `GET /api/requirements/REQ-MCP-007` returns full requirement with dependency detail.
2. Upstream dependencies list all direct `dependencies` with their current status.
3. Downstream dependencies list all direct `blocks` with their current status.
4. `transitive_upstream_count` and `transitive_downstream_count` reflect full closure.
5. `all_upstream_complete` is true only when every transitive upstream req is COMPLETE.
6. Unknown req_id returns 404 with descriptive message.
7. Response time < 10ms for single requirement lookup.

## Files to Create/Modify

- `internal/cmd/serve_api.go` -- Add /api/requirements/:id handler, buildDependencyDetail
- `internal/cmd/serve_api_test.go` -- Detail endpoint tests

## Effort Estimate

0.5 weeks

## Test Strategy

- Valid req_id returns full detail with correct dependency resolution
- Unknown req_id returns 404
- Requirement with no dependencies returns empty arrays
- Requirement with deep transitive chains returns correct counts
- Concurrent reads safe under -race
