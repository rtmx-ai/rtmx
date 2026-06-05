# REQ-ADAPT-004: Monday.com GraphQL API Client

## Metadata
- **Category**: ADAPT
- **Subcategory**: Monday
- **Priority**: HIGH
- **Phase**: 27
- **Status**: MISSING
- **Dependencies**: (none)
- **Blocks**: REQ-ADAPT-005, REQ-ADAPT-006

## Requirement

RTMX shall implement a Monday.com adapter that connects to the Monday.com
GraphQL API (v2) using API token authentication, implementing the
`ServiceAdapter` interface for board item CRUD operations, column value
reading/writing, and group enumeration.

## Rationale

Monday.com is a leading work management platform used by engineering and
product teams. Teams using Monday for sprint planning and task tracking
need bidirectional sync with RTMX to maintain requirements traceability
alongside their existing workflow. Monday's GraphQL API provides efficient
querying of board data.

## Design

### Authentication

```yaml
rtmx:
  adapters:
    monday:
      board_id: "1234567890"
      token_env: "MONDAY_TOKEN"
      sync_mode: "bidirectional"
```

### GraphQL Client

```go
type MondayAdapter struct {
    client  HTTPClient
    boardID string
    token   string
    apiURL  string  // default: "https://api.monday.com/v2"
}

func NewMondayAdapter(cfg *config.MondayConfig, opts ...AdapterOption) (*MondayAdapter, error)
```

### Key Queries

```graphql
# Fetch board items
query {
  boards(ids: [$boardID]) {
    items_page(limit: 100) {
      items {
        id
        name
        column_values { id text value }
        group { id title }
      }
    }
  }
}

# Update item
mutation {
  change_multiple_column_values(
    item_id: $itemID,
    board_id: $boardID,
    column_values: $values
  ) { id }
}
```

### Column Type Handling

Monday.com uses typed columns (status, person, number, text, date).
The adapter maps each column type to its RTMX equivalent:

| Monday Column Type | RTMX Field | Notes |
|-------------------|------------|-------|
| status | status | Via label-to-status mapping |
| person | assignee | User name extraction |
| numbers | effort_weeks | Direct numeric |
| text | notes | Direct text |
| date | started_date / completed_date | ISO format conversion |

### Rate Limiting

Monday API: 10,000,000 complexity points/minute. Each query has a
complexity cost. The adapter tracks cumulative cost and pauses when
approaching the limit.

## Acceptance Criteria

1. `NewMondayAdapter` returns configured adapter when token is available.
2. `IsConfigured()` returns false when token env var is empty.
3. `TestConnection()` executes a lightweight board query.
4. `FetchItems()` retrieves items from the configured board.
5. `GetItem()` retrieves a single item by Monday ID.
6. `CreateItem()` creates a board item with requirement details.
7. `UpdateItem()` updates column values for an existing item.
8. Column type mapping handles status, person, number, text, and date.
9. Complexity tracking prevents rate limit violations.
10. All HTTP calls go through injected HTTPClient.

## Files to Create/Modify

- `internal/adapters/monday.go` -- Monday adapter implementation
- `internal/adapters/monday_test.go` -- Adapter tests with mock HTTP
- `internal/config/config.go` -- Add MondayConfig struct

## Effort Estimate

1 week

## Test Strategy

- Mock HTTP server: simulate GraphQL responses
- Column type mapping: table-driven tests for each column type
- Complexity tracking: verify pause at threshold
- Auth failure: verify clean error on 401
- GraphQL error responses: verify error extraction from response body
