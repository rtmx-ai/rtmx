# REQ-ADAPT-011: Slack Slash Command Handler

## Metadata
- **Category**: ADAPT
- **Subcategory**: Slack
- **Priority**: MEDIUM
- **Phase**: 27
- **Status**: MISSING
- **Dependencies**: REQ-ADAPT-010
- **Blocks**: (none)

## Requirement

The RTMX serve command shall expose a `/api/slack/commands` endpoint that
handles Slack slash command webhooks, enabling team members to query RTMX
status, backlog, and requirement details directly from Slack via commands
like `/rtmx status` and `/rtmx req REQ-MCP-007`.

## Rationale

Slack slash commands let team members query project status without leaving
their communication context. This is particularly valuable for managers
and stakeholders who do not use the CLI. Slash commands provide a low-
friction entry point to RTMX data.

## Design

### Slash Commands

| Command | Response |
|---------|----------|
| `/rtmx status` | Completion summary with progress bar |
| `/rtmx backlog` | Top 5 incomplete requirements |
| `/rtmx req REQ-MCP-007` | Requirement detail card |
| `/rtmx health` | Health check summary |
| `/rtmx release v1.2.0` | Release scope summary |

### Endpoint

```
POST /api/slack/commands
Content-Type: application/x-www-form-urlencoded

token=...&command=/rtmx&text=status&user_id=U123&channel_id=C456
```

### Response Format

Slack Block Kit messages with:
- Status badges (colored emoji indicators)
- Collapsible sections for long lists
- Action buttons (e.g., "View in Dashboard" linking to `rtmx serve` URL)

### Security

- Slack request signature verification (HMAC-SHA256)
- Signing secret stored in `SLACK_SIGNING_SECRET` env var
- Timestamp validation (reject requests > 5 minutes old)

### Interactive Messages

Responses include buttons for common follow-up actions:
- "View Details" -> opens dashboard URL
- "Show Backlog" -> triggers backlog response
- Action handler at `/api/slack/interactions`

## Acceptance Criteria

1. `/rtmx status` returns formatted completion summary.
2. `/rtmx backlog` returns top 5 incomplete requirements.
3. `/rtmx req REQ-MCP-007` returns requirement detail card.
4. `/rtmx health` returns health check results.
5. `/rtmx release v1.2.0` returns release scope.
6. Request signature verification rejects unsigned requests.
7. Unknown subcommands return help text.
8. Interactive buttons trigger correct follow-up actions.
9. Response time < 3 seconds (Slack timeout).

## Files to Create/Modify

- `internal/cmd/serve_slack.go` -- Slack command and interaction handlers
- `internal/cmd/serve_slack_test.go` -- Handler tests with signature verification
- `internal/adapters/slack.go` -- Block Kit message formatting helpers

## Effort Estimate

1 week

## Test Strategy

- Command parsing: each subcommand produces correct response
- Signature verification: valid signature passes, invalid rejected
- Timestamp rejection: old requests (> 5 minutes) rejected
- Unknown command: returns help text
- Response format: verify Block Kit JSON structure
- Timeout: verify response generated within 3 seconds
