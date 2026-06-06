# REQ-ADAPT-010: Slack Notification Adapter

## Metadata
- **Category**: ADAPT
- **Subcategory**: Slack
- **Priority**: HIGH
- **Phase**: 27
- **Status**: MISSING
- **Dependencies**: (none)
- **Blocks**: REQ-ADAPT-011, REQ-ADAPT-012

## Requirement

RTMX shall implement a Slack adapter that connects to the Slack Web API
using Bot Token authentication, supporting outbound notifications for
requirement status changes, release gate events, and health alerts with
configurable channel routing.

## Rationale

Slack is the de facto communication platform for engineering teams.
Automatic notifications when requirements change status, releases gate,
or health degrades keeps the team informed without requiring them to
actively check the dashboard. Channel routing lets teams direct different
notification types to appropriate channels (e.g., releases to #releases,
health alerts to #ops).

## Design

### Authentication

```yaml
rtmx:
  adapters:
    slack:
      token_env: "SLACK_BOT_TOKEN"
      channels:
        status_changes: "#rtmx-updates"
        release_gates: "#releases"
        health_alerts: "#ops-alerts"
        agent_activity: "#rtmx-agents"
      notify_on:
        - status_change
        - release_gate
        - health_degradation
        - agent_stale
```

### Notification Types

| Event | Channel | Message Format |
|-------|---------|---------------|
| Status change | status_changes | `[REQ-MCP-007] Status: MISSING -> PARTIAL (by rhino11)` |
| Release gate pass | release_gates | `Release v1.2.0 gate: PASS (50/50 complete)` |
| Release gate fail | release_gates | `Release v1.2.0 gate: FAIL (3 incomplete: REQ-MCP-007, ...)` |
| Health degradation | health_alerts | `Health check WARN: 3 requirements stale > 30 days` |
| Agent stale | agent_activity | `Agent claude-001 stale: REQ-MCP-007 claimed 30m ago, no heartbeat` |

### Message Formatting

Uses Slack Block Kit for rich formatting:

```go
type SlackAdapter struct {
    client   HTTPClient
    token    string
    channels map[string]string
    notifyOn []string
}

func (s *SlackAdapter) NotifyStatusChange(req *database.Requirement, oldStatus, newStatus database.Status) error
func (s *SlackAdapter) NotifyReleaseGate(version string, pass bool, failures []string) error
func (s *SlackAdapter) NotifyHealthAlert(check string, status string, message string) error
```

### Rate Limiting

Slack API: 1 message/second per channel. The adapter queues messages and
sends at most 1/second per channel, batching rapid status changes into
a single summary message when multiple changes occur within 5 seconds.

## Acceptance Criteria

1. `NotifyStatusChange` posts formatted message to configured channel.
2. `NotifyReleaseGate` posts gate pass/fail to release channel.
3. `NotifyHealthAlert` posts health degradation to ops channel.
4. Channel routing respects per-event-type configuration.
5. `notify_on` controls which event types trigger notifications.
6. Rate limiter batches rapid changes (< 5s apart).
7. Invalid channel name returns descriptive error.
8. Missing token returns `IsConfigured() == false`.
9. All HTTP calls go through injected HTTPClient.

## Files to Create/Modify

- `internal/adapters/slack.go` -- Slack adapter implementation
- `internal/adapters/slack_test.go` -- Adapter tests with mock HTTP
- `internal/config/config.go` -- Add SlackConfig struct

## Effort Estimate

0.5 weeks

## Test Strategy

- Mock Slack API: verify message payload format
- Channel routing: each event type routes to correct channel
- Rate limiter: rapid events batched into single message
- Auth failure: verify clean error on invalid token
- Notify filter: disabled event types do not send messages
