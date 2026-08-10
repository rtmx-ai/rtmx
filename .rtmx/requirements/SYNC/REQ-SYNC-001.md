# REQ-SYNC-001: Durable Local State for External-Service Sync

## Metadata
- **Category**: SYNC
- **Subcategory**: Persistence
- **Priority**: P0
- **Phase**: 29
- **Status**: MISSING
- **Dependencies**: REQ-SYNC-001a|REQ-SYNC-001b|REQ-SYNC-001c
- **Blocks**: (none)

## Requirement

External-service sync operations shall durably persist every local RTMX state
change before reporting that change as successful. After a successful command
returns, restarting RTMX and reloading the CSV shall preserve all external
linkages and locally accepted remote state.

## Rationale

Sync currently reports successful Jira and issue-tracker operations without
persisting the corresponding local mutations. In export mode, a created remote
key is printed and then forgotten, so the next export creates a duplicate. The
same persistence boundary must ultimately cover import and bidirectional modes.

## Decomposition

This requirement was decomposed once by sync direction because each direction
has a distinct mutation boundary and can be delivered and verified
independently:

1. **REQ-SYNC-001a** — exported-item linkage durability and idempotency.
2. **REQ-SYNC-001b** — imported status and linkage durability.
3. **REQ-SYNC-001c** — bidirectional conflict-resolution durability.

REQ-SYNC-001 is satisfied only when all three child requirements are COMPLETE.
The urgent patch release implements REQ-SYNC-001a; the remaining children stay
visible in the backlog.

## Acceptance Criteria

1. Every successful sync mutation survives database reload.
2. A command never reports a local mutation as successful before it is durable.
3. Dry-run mode performs no local or remote writes.
4. Persistence failures produce recoverable diagnostics and halt additional
   side effects where continuing could create unlinked remote records.
5. Export, import, and bidirectional persistence are independently traceable to
   child requirements and tests.

## Files to Create/Modify

- `internal/cmd/sync.go`
- `internal/cmd/sync_test.go`
- `.rtmx/database.csv`
- `.rtmx/config.yaml`

## Test Strategy

The child requirements own executable integration tests for each sync
direction. Parent completion is derived from all children becoming COMPLETE.
