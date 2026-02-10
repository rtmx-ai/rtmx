# REQ-CLAUDE-003: Claude Cowork RTM Sharing

## Status: NOT_STARTED
## Priority: MEDIUM
## Phase: 19
## Effort: 3.0 weeks

## Description

RTMX shall integrate with Claude Cowork to enable real-time RTM synchronization across collaborative coding sessions. When multiple developers work together in a Cowork session, their requirement assignments, status updates, and context should be shared via RTMX's CRDT-based sync infrastructure.

## Rationale

Claude Cowork enables team collaboration, but without RTM integration:
- Developers may work on the same requirement unknowingly
- Status updates in one session don't propagate to others
- Context about "who's working on what" is lost

Integration provides:
- **Visibility** - See which requirements teammates are working on
- **Coordination** - Automatic conflict detection for requirement assignments
- **Traceability** - Audit trail of who worked on what, when

## Acceptance Criteria

- [ ] Cowork session detects RTMX-enabled project and initializes sync
- [ ] Requirement "claims" are broadcast to all session participants
- [ ] Status changes sync in real-time via CRDT merge
- [ ] Session sidebar shows "RTM Activity" panel with live updates
- [ ] Conflict resolution when two developers claim same requirement
- [ ] Session transcript includes RTM context for later reference
- [ ] Works with rtmx-sync server for persistence beyond session lifetime
- [ ] Graceful degradation when rtmx-sync is unavailable (local-only mode)
- [ ] Privacy controls: requirement text can be hidden while sharing status

## Technical Notes

### Integration Architecture

```
┌─────────────────┐     ┌─────────────────┐
│  Claude Cowork  │     │  Claude Cowork  │
│   Developer A   │     │   Developer B   │
└────────┬────────┘     └────────┬────────┘
         │                       │
         │    Cowork Protocol    │
         └───────────┬───────────┘
                     │
              ┌──────▼──────┐
              │ RTMX Plugin │
              │  (shared)   │
              └──────┬──────┘
                     │
              ┌──────▼──────┐
              │  rtmx-sync  │
              │   (CRDT)    │
              └─────────────┘
```

### Cowork Plugin API

The plugin hooks into Cowork's extension system:

```typescript
// Conceptual API - actual implementation depends on Cowork SDK
interface RTMXCoworkPlugin {
  onSessionStart(session: CoworkSession): void;
  onParticipantJoin(participant: Participant): void;
  onRequirementClaimed(req: RequirementClaim): void;
  onStatusChange(change: StatusChange): void;
  renderSidebarPanel(): ReactNode;
}
```

### CRDT Operations for Cowork

RTM operations that sync across sessions:
- `ClaimRequirement(req_id, user_id, timestamp)`
- `UnclaimRequirement(req_id, user_id, timestamp)`
- `UpdateStatus(req_id, old_status, new_status, timestamp)`
- `AddComment(req_id, user_id, comment, timestamp)`

LWW (Last-Writer-Wins) semantics with vector clocks for ordering.

### Privacy Levels

Configuration in rtmx.yaml:
```yaml
rtmx:
  cowork:
    share_level: status_only  # full | status_only | ids_only
    require_auth: true
```

## Gherkin Specification

```gherkin
@REQ-CLAUDE-003 @scope_system @technique_nominal
Feature: Claude Cowork RTM Sharing
  As a team using Claude Cowork
  I want our RTM status synchronized
  So that we avoid conflicts and see each other's progress

  Background:
    Given a Claude Cowork session with 2 developers
    And both are in an RTMX-enabled project
    And rtmx-sync is available

  Scenario: Requirement claim is broadcast
    Given Developer A has the session open
    And Developer B has the session open
    When Developer A claims REQ-AUTH-001
    Then Developer B sees "Alice claimed REQ-AUTH-001" in activity panel
    And the requirement shows as "in progress" for both

  Scenario: Conflict detection on double-claim
    Given REQ-AUTH-001 is unclaimed
    When Developer A claims REQ-AUTH-001
    And Developer B simultaneously claims REQ-AUTH-001
    Then the first claim wins (LWW with timestamp)
    And Developer B sees "REQ-AUTH-001 was claimed by Alice"
    And Developer B is offered alternative requirements

  Scenario: Status sync across sessions
    Given Developer A is working on REQ-AUTH-001
    When Developer A runs "rtmx verify --update"
    And REQ-AUTH-001 status changes to COMPLETE
    Then Developer B sees the status update in real-time
    And the shared RTM reflects the change

  Scenario: Offline graceful degradation
    Given rtmx-sync is unavailable
    When Developer A claims a requirement
    Then the claim is recorded locally
    And syncs when connection is restored
```

## Test Cases

1. `tests/test_cowork_integration.py::test_session_rtm_detection`
2. `tests/test_cowork_integration.py::test_claim_broadcast`
3. `tests/test_cowork_integration.py::test_status_sync`
4. `tests/test_cowork_integration.py::test_conflict_resolution`
5. `tests/test_cowork_integration.py::test_privacy_levels`
6. `tests/test_cowork_integration.py::test_offline_mode`

## Files to Create/Modify

- `src/rtmx/cowork/` (new directory) - Cowork integration
- `src/rtmx/cowork/plugin.py` - Plugin implementation
- `src/rtmx/cowork/sidebar.py` - Activity panel rendering
- `src/rtmx/sync/crdt.py` - Extend for Cowork operations
- `tests/test_cowork_integration.py` (new) - Integration tests

## Dependencies

- REQ-CLAUDE-001: Claude Code hooks (installation framework)
- REQ-COLLAB-001: Cross-repo dependency tracking (CRDT foundation)
- REQ-CRDT-001: CRDT implementation for RTM sync

## Blocks

- None (leaf requirement)
