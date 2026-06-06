# REQ-API-001: Requirements List Endpoint with Filter/Sort/Paginate

## Metadata
- **Category**: API
- **Subcategory**: REST
- **Priority**: P0
- **Phase**: 25
- **Status**: MISSING
- **Dependencies**: (none)
- **Blocks**: REQ-API-002, REQ-API-003, REQ-API-004, REQ-API-005, REQ-API-006, REQ-API-007, REQ-DASH-001, REQ-DASH-002, REQ-DASH-007, REQ-TUI-002

## Requirement

The RTMX serve command shall expose a `GET /api/requirements` endpoint that
returns a paginated, filterable, sortable list of requirements as JSON. This
endpoint replaces the current `/api/status` summary-only endpoint as the
primary data source for both the TUI and GUI dashboards.

## Rationale

The current `/api/status` endpoint returns only aggregate counts (complete,
partial, missing, total). Both the TUI and GUI dashboards need access to
individual requirement records with filtering to support table views, search,
and drill-down navigation. A well-designed REST endpoint avoids duplicating
query logic across clients.

## Design

### Query Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| category | string | (all) | Filter by category (e.g., `MCP`, `CLI`) |
| status | string | (all) | Filter by status (`COMPLETE`, `PARTIAL`, `MISSING`) |
| priority | string | (all) | Filter by priority (`P0`, `HIGH`, `MEDIUM`, `LOW`) |
| version | string | (all) | Filter by target version/sprint |
| assignee | string | (all) | Filter by assignee |
| search | string | (all) | Full-text search across req_id, requirement_text, notes |
| sort | string | req_id | Sort field: `req_id`, `category`, `priority`, `status`, `effort_weeks` |
| order | string | asc | Sort order: `asc` or `desc` |
| page | int | 1 | Page number (1-based) |
| per_page | int | 50 | Items per page (max 200) |

### Response Schema

```json
{
  "requirements": [
    {
      "req_id": "REQ-CLI-001",
      "category": "CLI",
      "subcategory": "Foundation",
      "requirement_text": "...",
      "status": "COMPLETE",
      "priority": "P0",
      "phase": 1,
      "effort_weeks": 1.0,
      "assignee": "rhino11",
      "sprint": "v1.0.0",
      "dependencies": ["REQ-GO-001"],
      "blocks": ["REQ-CLI-002"],
      "started_date": "2026-02-11",
      "completed_date": "2026-02-11"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 50,
    "total": 241,
    "total_pages": 5
  },
  "filters_applied": {
    "category": "CLI",
    "status": null
  }
}
```

### Implementation

```go
// internal/cmd/serve.go -- extend NewDashboardMux

mux.HandleFunc("/api/requirements", func(w http.ResponseWriter, r *http.Request) {
    opts := parseFilterOpts(r.URL.Query())
    reqs := db.Filter(opts)
    page := paginateRequirements(reqs, opts.Page, opts.PerPage)
    writeJSON(w, page)
})
```

The endpoint reuses `database.FilterOptions` extended with pagination and
sort fields, keeping query logic in the database package where it belongs.

### Input Validation

- `per_page` capped at 200 to prevent memory abuse
- `sort` validated against allowed field names (reject unknown fields)
- `page` must be >= 1
- `search` sanitized (no regex injection; plain substring match)

## Acceptance Criteria

1. `GET /api/requirements` returns all requirements as paginated JSON.
2. Each filter parameter correctly narrows the result set.
3. `search` matches against req_id, requirement_text, and notes fields.
4. `sort` and `order` correctly order results.
5. Pagination metadata includes total count and total_pages.
6. Invalid parameters return 400 with descriptive error message.
7. Response time < 50ms for 500-requirement database.
8. Concurrent requests are safe (read lock on database).

## Files to Create/Modify

- `internal/cmd/serve.go` -- Add /api/requirements handler
- `internal/cmd/serve_api.go` -- Extract API helpers (parseFilterOpts, paginateRequirements, writeJSON)
- `internal/cmd/serve_api_test.go` -- Table-driven tests for all filter/sort/pagination combinations
- `internal/database/database.go` -- Extend FilterOptions with Page, PerPage, Sort, Order fields

## Effort Estimate

0.5 weeks

## Test Strategy

- Table-driven tests: each filter parameter in isolation and combined
- Pagination edge cases: empty results, last page partial, page beyond range
- Sort correctness: verify ordering for each sortable field
- Concurrent request test with -race flag
- Golden file test for response schema stability
