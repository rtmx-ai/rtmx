# REQ-DASH-008: WebSocket Live Updates

## Metadata
- **Category**: DASH
- **Subcategory**: Realtime
- **Priority**: MEDIUM
- **Phase**: 28
- **Status**: MISSING
- **Dependencies**: REQ-DASH-001, REQ-API-007
- **Blocks**: (none)

## Requirement

The web dashboard shall support WebSocket-based live updates that push
database changes and agent claim events to connected browsers in real
time, using the existing CRDT sync infrastructure when `--sync-url` is
configured, or a local file-watch fallback.

## Rationale

When multiple users or agents are working simultaneously, a dashboard
that requires manual refresh shows stale data and creates confusion.
Real-time updates via WebSocket make the dashboard a live command center
for multi-agent coordination. The `--sync-url` flag already exists on
`rtmx serve` but is not wired to any real-time mechanism.

## Design

### WebSocket Endpoint

```
ws://localhost:8080/ws
```

### Message Types

```json
{"type": "requirements_changed", "data": {"changed": ["REQ-MCP-007"], "timestamp": "..."}}
{"type": "claim_event", "data": {"action": "claim", "req_id": "REQ-MCP-007", "agent_id": "claude-001"}}
{"type": "health_update", "data": {"status": "HEALTHY", "completion_pct": 98.8}}
```

### Data Sources

1. **With `--sync-url`**: Connect to the CRDT sync server and forward
   document change events to WebSocket clients.
2. **Without `--sync-url`**: Use fsnotify to watch the database CSV file
   and claims directory, broadcasting changes to connected browsers.

### Client Integration

```javascript
// Alpine.js component
const ws = new WebSocket(`ws://${location.host}/ws`);
ws.onmessage = (e) => {
    const msg = JSON.parse(e.data);
    if (msg.type === 'requirements_changed') {
        htmx.trigger('#requirements-table', 'refresh');
    }
};
```

### Connection Management

- Auto-reconnect with exponential backoff (1s, 2s, 4s, max 30s)
- Connection status indicator in dashboard header
- Graceful degradation: dashboard works without WebSocket (manual refresh)

## Acceptance Criteria

1. WebSocket endpoint at `/ws` accepts connections.
2. Database file changes broadcast `requirements_changed` messages.
3. Claim file changes broadcast `claim_event` messages.
4. Connected browsers refresh affected views automatically.
5. Auto-reconnect works after server restart.
6. Connection status indicator shows connected/disconnected state.
7. Dashboard degrades gracefully when WebSocket is unavailable.
8. With `--sync-url`, changes from remote sync are forwarded.

## Files to Create/Modify

- `internal/cmd/serve_ws.go` -- WebSocket handler and hub
- `internal/cmd/serve_ws_test.go` -- WebSocket tests
- `dashboard/js/ws.js` -- Client-side WebSocket with reconnect

## Effort Estimate

1 week

## Test Strategy

- WebSocket connection: verify upgrade and message receipt
- File change broadcast: modify database, verify message sent
- Claim event: create claim file, verify event broadcast
- Reconnect: close connection, verify reconnect with backoff
- Multiple clients: verify broadcast reaches all connected browsers
- Graceful degradation: verify dashboard works without WebSocket
