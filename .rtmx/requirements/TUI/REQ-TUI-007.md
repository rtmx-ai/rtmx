# REQ-TUI-007: Agent Activity Monitor

## Metadata
- **Category**: TUI
- **Subcategory**: View
- **Priority**: MEDIUM
- **Phase**: 26
- **Status**: MISSING
- **Dependencies**: REQ-TUI-001, REQ-API-007
- **Blocks**: (none)

## Requirement

The TUI shall provide an agent activity monitor view that displays active
claims, agent heartbeat status, and a live feed of claim/release events,
providing real-time visibility into multi-agent coordination.

## Rationale

When multiple AI agents are working concurrently on requirements (via
`rtmx next --agent-id`), operators need visibility into who is working on
what and whether any agents have gone stale. This view turns the
orchestration layer from a hidden mechanism into an observable system.

## Design

### Layout

```
+-- Agent Activity Monitor -----------------------------------------------+
| Active Agents: 2     Active Claims: 3     Stale: 0                      |
|                                                                         |
| AGENT          CLAIM          SINCE         HEARTBEAT    STATUS          |
| claude-001     REQ-MCP-007    10m ago       2m ago       Active          |
| claude-001     REQ-MCP-008    8m ago        2m ago       Active          |
| cursor-002     REQ-TUI-001    3m ago        1m ago       Active          |
|                                                                         |
| EVENT LOG                                                               |
| 10:30:15  claude-001  CLAIM     REQ-MCP-007                             |
| 10:32:20  claude-001  CLAIM     REQ-MCP-008                             |
| 10:37:01  cursor-002  CLAIM     REQ-TUI-001                             |
| 10:38:45  claude-001  HEARTBEAT REQ-MCP-007                             |
+-------------------------------------------------------------------------+
```

### Data Source

Reads from `.rtmx/claims/` directory using the orchestration package.
File-watch (from REQ-TUI-006) detects new/removed claim files. The event
log is constructed from file modification timestamps at startup, then
updated in real-time from watch events.

### Stale Detection

Claims with heartbeat older than 15 minutes (configurable) are highlighted
in red with "STALE" status. A count of stale claims appears in the summary
header.

### Actions

| Key | Action |
|-----|--------|
| f | Force release a stale claim (prompts for confirmation) |
| Enter | Open detail pane for the claimed requirement |
| r | Refresh claims from disk |

## Acceptance Criteria

1. Agent summary shows correct counts for active agents, claims, and stale.
2. Claims table lists all active claims with agent, timing, and status.
3. Stale claims are visually highlighted and counted separately.
4. Event log shows recent claim/release activity.
5. Live refresh updates the view as claims change on disk.
6. `f` on a stale claim force-releases it after confirmation.
7. Enter opens the requirement detail for the claimed requirement.

## Files to Create/Modify

- `internal/tui/views/agents.go` -- Agent monitor view model
- `internal/tui/views/agents_test.go` -- Monitor tests

## Effort Estimate

0.5 weeks

## Test Strategy

- Create test claim files, verify table renders correct data
- Stale detection: set old heartbeat timestamp, verify STALE flag
- Force release: verify claim file removed after confirmation
- Event log construction from file timestamps
- Live refresh: add claim file, verify view updates
