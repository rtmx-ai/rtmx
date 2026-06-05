# REQ-ADAPT-007: GitLab REST API Client

## Metadata
- **Category**: ADAPT
- **Subcategory**: GitLab
- **Priority**: HIGH
- **Phase**: 27
- **Status**: MISSING
- **Dependencies**: (none)
- **Blocks**: REQ-ADAPT-008, REQ-ADAPT-009

## Requirement

RTMX shall implement a GitLab adapter that connects to the GitLab REST
API (v4) using Personal Access Tokens, implementing the `ServiceAdapter`
interface for issue CRUD operations, label management, and milestone
enumeration. The adapter shall support both GitLab.com (SaaS) and
self-hosted GitLab instances.

## Rationale

GitLab is the primary alternative to GitHub for teams that need self-hosted
source control, particularly in defense and government environments where
air-gapped deployments are common. The RTMX air-gapped deployment topology
(Zarf) often coexists with self-hosted GitLab. Supporting GitLab issues
enables requirements traceability for these teams without requiring GitHub.

## Design

### Authentication

```yaml
rtmx:
  adapters:
    gitlab:
      server_url: "https://gitlab.example.com"  # or https://gitlab.com
      project_id: "group/project"  # or numeric ID
      token_env: "GITLAB_TOKEN"
      sync_mode: "bidirectional"
```

### API Client

```go
type GitLabAdapter struct {
    client    HTTPClient
    serverURL string
    projectID string
    token     string
}

func NewGitLabAdapter(cfg *config.GitLabConfig, opts ...AdapterOption) (*GitLabAdapter, error)
```

### API Endpoints Used

| Endpoint | Purpose |
|----------|---------|
| `GET /projects/:id/issues` | List project issues |
| `GET /projects/:id/issues/:iid` | Get issue detail |
| `POST /projects/:id/issues` | Create issue |
| `PUT /projects/:id/issues/:iid` | Update issue |
| `GET /projects/:id/labels` | List labels |
| `GET /projects/:id/milestones` | List milestones |

### Security

- Server URL validated (HTTPS required, private IP blocking for SaaS mode)
- Private IP blocking disabled for self-hosted mode (`allow_private_ip: true`)
- Token never logged or included in error messages

### Status Mapping

| GitLab State | RTMX Status |
|-------------|-------------|
| opened | MISSING |
| closed | COMPLETE |

Label-based refinement:
- Label `in-progress` + opened -> PARTIAL
- Label `blocked` + opened -> MISSING (with note)

## Acceptance Criteria

1. `NewGitLabAdapter` returns configured adapter for both SaaS and self-hosted.
2. `IsConfigured()` returns false when token env var is empty.
3. `TestConnection()` verifies API access with a lightweight project query.
4. `FetchItems()` retrieves issues with pagination handling.
5. `GetItem()` retrieves a single issue by IID.
6. `CreateItem()` creates a GitLab issue with requirement details.
7. `UpdateItem()` updates issue state and labels.
8. Server URL validation enforces HTTPS.
9. Private IP blocking active for SaaS, disabled for self-hosted.
10. All HTTP calls go through injected HTTPClient.

## Files to Create/Modify

- `internal/adapters/gitlab.go` -- GitLab adapter implementation
- `internal/adapters/gitlab_test.go` -- Adapter tests with mock HTTP
- `internal/config/config.go` -- Add GitLabConfig struct

## Effort Estimate

1 week

## Test Strategy

- Mock HTTP server: simulate GitLab API responses
- Self-hosted URL: verify private IP allowed when configured
- SaaS URL: verify private IP blocked
- Pagination: verify all pages fetched for large issue lists
- Status mapping: table-driven tests for state + label combinations
- Auth failure: verify clean error on 401/403
