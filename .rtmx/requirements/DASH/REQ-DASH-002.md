# REQ-DASH-002: Requirements List with Server-Side Filtering

## Metadata
- **Category**: DASH
- **Subcategory**: View
- **Priority**: P0
- **Phase**: 28
- **Status**: MISSING
- **Dependencies**: REQ-DASH-001, REQ-API-001
- **Blocks**: REQ-DASH-003

## Requirement

The web dashboard shall provide a requirements list view with a data table
supporting server-side filtering, sorting, and pagination. The table shall
fetch data from the `GET /api/requirements` endpoint and update dynamically
as users interact with filter controls.

## Rationale

The requirements list is the primary navigation surface in the GUI. It must
handle databases of 200+ requirements with responsive filtering and sorting.
Server-side operations (via REQ-API-001) keep the client lightweight and
ensure consistent behavior between the API and the UI.

## Design

### Layout

```
+-- Requirements --------------------------------------------------------+
| [Filter: category v] [Status v] [Priority v] [Search...        ] [Clear]|
|                                                                         |
| REQ ID       | Status   | Priority | Category | Effort | Assignee | Desc|
| REQ-MCP-007  | MISSING  | P0       | MCP      | 0.5w   |          | Res.|
| REQ-MCP-008  | MISSING  | P0       | MCP      | 1.5w   |          | Too.|
| REQ-CLI-001  | COMPLETE | P0       | CLI      | 1.0w   | rhino11  | Sta.|
| ...                                                                     |
|                                                                         |
| Page 1 of 5  [<< < 1 2 3 4 5 > >>]          Showing 1-50 of 241       |
+-------------------------------------------------------------------------+
```

### Implementation

Uses htmx to fetch table rows from the server. Filter changes trigger
`hx-get="/api/requirements?category=MCP&status=MISSING"` which returns
an HTML table body fragment (accept header negotiation: HTML for htmx
requests, JSON for API clients).

### Filter Controls

- Category dropdown: populated from distinct categories in database
- Status dropdown: COMPLETE, PARTIAL, MISSING, NOT_STARTED
- Priority dropdown: P0, HIGH, MEDIUM, LOW
- Free-text search: debounced 300ms, triggers server-side search
- Clear button: resets all filters

### Sorting

Column headers are clickable. Clicking toggles sort direction and
re-fetches with `?sort=priority&order=desc`.

### Pagination

Bottom pagination control with page numbers, prev/next, and
items-per-page selector (25, 50, 100).

## Acceptance Criteria

1. Requirements table loads with all requirements on first visit.
2. Each filter control correctly narrows displayed requirements.
3. Filters can be combined (category + status + search).
4. Column header click sorts by that column.
5. Pagination controls navigate between pages.
6. Page size selector changes items per page.
7. URL hash updates with filter/sort/page state (bookmarkable).
8. Table updates without full page reload (htmx swap).
9. Empty results show "No requirements match filters" message.
10. Loading state shown during fetch.

## Files to Create/Modify

- `dashboard/requirements.html` -- Requirements list template
- `dashboard/components/table.html` -- Reusable table component
- `dashboard/components/filters.html` -- Filter controls
- `internal/cmd/serve_api.go` -- Content negotiation (HTML vs JSON)

## Effort Estimate

1 week

## Test Strategy

- Filter combinations produce correct server requests
- Pagination state management (page resets on filter change)
- Sort indicator reflects active sort column and direction
- Empty state message displayed when no results
- Content negotiation: htmx requests get HTML, curl gets JSON
