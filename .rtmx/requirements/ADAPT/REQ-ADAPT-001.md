# REQ-ADAPT-001: Asana REST API Client

## Metadata
- **Category**: ADAPT
- **Subcategory**: Asana
- **Priority**: HIGH
- **Phase**: 27
- **Status**: MISSING
- **Dependencies**: (none)
- **Blocks**: REQ-ADAPT-002, REQ-ADAPT-003

## Requirement

RTMX shall implement an Asana adapter that connects to the Asana REST API
(v1) using Personal Access Tokens (PAT) or OAuth2 Service Account
credentials, implementing the `ServiceAdapter` interface for task CRUD
operations, project listing, and section enumeration.

## Rationale

Asana is one of the most widely adopted project management tools in
enterprise environments. Teams using Asana for day-to-day work tracking
need bidirectional sync with RTMX so that requirements traceability does
not require abandoning their existing PM tool. The Asana REST API is
well-documented and stable.

## Design

### Authentication

```yaml
# rtmx.yaml
rtmx:
  adapters:
    asana:
      workspace_id: "1234567890"
      project_id: "9876543210"
      token_env: "ASANA_TOKEN"  # PAT or OAuth2 token
      sync_mode: "bidirectional"  # or "pull-only", "push-only"
```

Token is read from the environment variable specified in `token_env`
(default: `ASANA_TOKEN`). This follows the same pattern as the GitHub
adapter's `GITHUB_TOKEN`.

### API Client

```go
type AsanaAdapter struct {
    client      HTTPClient
    workspaceID string
    projectID   string
    token       string
}

func NewAsanaAdapter(cfg *config.AsanaConfig, opts ...AdapterOption) (*AsanaAdapter, error)
```

### API Endpoints Used

| Endpoint | Purpose |
|----------|---------|
| `GET /projects/{id}/tasks` | List tasks in project |
| `GET /tasks/{id}` | Get task detail |
| `POST /tasks` | Create task |
| `PUT /tasks/{id}` | Update task |
| `GET /projects/{id}/sections` | List sections for category mapping |
| `GET /workspaces/{id}/users` | List users for assignee mapping |

### Rate Limiting

Asana enforces 150 requests/minute per PAT. The adapter implements a
token-bucket rate limiter that sleeps before exceeding the limit, with
exponential backoff on 429 responses.

### Error Handling

- 401: Invalid or expired token -> return `ErrNotConfigured`
- 403: Insufficient permissions -> return descriptive error
- 404: Task/project not found -> return nil item (not error)
- 429: Rate limited -> backoff and retry (max 3 retries)

## Acceptance Criteria

1. `NewAsanaAdapter` returns configured adapter when token is available.
2. `IsConfigured()` returns false when token env var is empty.
3. `TestConnection()` verifies API access with a lightweight call.
4. `FetchItems()` retrieves tasks from the configured project.
5. `GetItem()` retrieves a single task by GID.
6. `CreateItem()` creates an Asana task with requirement details.
7. `UpdateItem()` updates an existing task's status and assignee.
8. Rate limiter prevents 429 errors under sustained load.
9. All HTTP calls go through injected HTTPClient (testable).
10. No direct os.Getenv calls (token via config/option injection).

## Files to Create/Modify

- `internal/adapters/asana.go` -- Asana adapter implementation
- `internal/adapters/asana_test.go` -- Adapter tests with mock HTTP
- `internal/config/config.go` -- Add AsanaConfig struct

## Effort Estimate

1 week

## Test Strategy

- Mock HTTP server: simulate all Asana API responses
- Table-driven tests for status mapping
- Rate limiter: verify backoff on 429 response
- Auth failure: verify clean error on 401
- Token injection: verify no direct env access
- Integration test (manual): connect to Asana sandbox project
