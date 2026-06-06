# REQ-ADAPT-012: Outbound Webhook Adapter

## Metadata
- **Category**: ADAPT
- **Subcategory**: Webhook
- **Priority**: MEDIUM
- **Phase**: 27
- **Status**: MISSING
- **Dependencies**: REQ-ADAPT-010
- **Blocks**: (none)

## Requirement

RTMX shall implement a generic outbound webhook adapter that sends HTTP
POST requests to configurable endpoints when requirements change status,
enabling integration with any external system that can receive webhooks
(CI/CD pipelines, monitoring systems, custom dashboards, PagerDuty, etc.).

## Rationale

Not every team uses Asana, Monday, GitLab, or Slack. A generic webhook
adapter provides an escape hatch for integrating with any HTTP-capable
system. Webhooks are the universal integration primitive -- every
monitoring, CI/CD, and workflow tool supports inbound webhooks.

## Design

### Configuration

```yaml
rtmx:
  adapters:
    webhooks:
      - name: "ci-trigger"
        url: "https://ci.example.com/api/trigger"
        events: ["status_change"]
        headers:
          Authorization: "Bearer ${CI_TOKEN}"
        method: "POST"
        timeout_seconds: 10
        retry_count: 3

      - name: "pagerduty"
        url: "https://events.pagerduty.com/v2/enqueue"
        events: ["health_degradation"]
        headers:
          Content-Type: "application/json"
        method: "POST"
        timeout_seconds: 5
        retry_count: 2

      - name: "custom-dashboard"
        url: "https://internal.example.com/rtmx/webhook"
        events: ["status_change", "release_gate", "agent_stale"]
        method: "POST"
        retry_count: 1
```

### Payload Schema

```json
{
  "event": "status_change",
  "timestamp": "2026-06-04T10:30:00Z",
  "data": {
    "req_id": "REQ-MCP-007",
    "old_status": "MISSING",
    "new_status": "PARTIAL",
    "changed_by": "rhino11",
    "requirement_text": "RTMX MCP server shall log response byte..."
  },
  "source": {
    "project": "rtmx",
    "version": "v1.2.0",
    "database": ".rtmx/database.csv"
  }
}
```

### Event Types

| Event | Trigger |
|-------|---------|
| `status_change` | Requirement status updated |
| `release_gate` | Release gate checked (pass or fail) |
| `health_degradation` | Health check transitions from PASS to WARN/FAIL |
| `agent_stale` | Agent claim becomes stale |
| `requirement_created` | New requirement added to database |

### Delivery Guarantees

- At-least-once delivery with configurable retry count
- Exponential backoff between retries (1s, 2s, 4s)
- Dead letter log for failed deliveries (written to `.rtmx/webhook-failures.log`)
- HMAC-SHA256 signature in `X-RTMX-Signature` header for payload verification

### Security

- URLs validated (HTTPS required unless `allow_insecure: true`)
- Environment variable expansion in headers (`${CI_TOKEN}`)
- No secret values logged in failure messages

## Acceptance Criteria

1. Webhook fires on configured event types.
2. Payload matches documented JSON schema.
3. HMAC-SHA256 signature included in request header.
4. Retry with exponential backoff on HTTP 5xx responses.
5. Failed deliveries logged to dead letter file.
6. Environment variables expanded in header values.
7. HTTPS required by default (configurable override).
8. Multiple webhooks can fire for the same event.
9. Webhook delivery does not block the triggering operation.
10. `rtmx webhook test <name>` sends a test payload to verify configuration.

## Files to Create/Modify

- `internal/adapters/webhook.go` -- Webhook adapter implementation
- `internal/adapters/webhook_test.go` -- Delivery and retry tests
- `internal/config/config.go` -- Add WebhookConfig struct
- `internal/cmd/webhook.go` -- `rtmx webhook test` command

## Effort Estimate

0.5 weeks

## Test Strategy

- Mock HTTP server: verify payload format and headers
- Retry: simulate 500 response, verify backoff and retry
- Dead letter: verify failed delivery logged
- HMAC: verify signature computation matches expected
- Multiple webhooks: verify all fire for same event
- Async delivery: verify non-blocking behavior
- Env var expansion: verify header values resolved
