# REQ-SYNC-001c: Durable Bidirectional Conflict Resolution

## Metadata
- **Category**: SYNC
- **Subcategory**: Persistence
- **Priority**: HIGH
- **Phase**: 29
- **Status**: MISSING
- **Dependencies**: REQ-GO-028
- **Blocks**: REQ-SYNC-001

## Requirement

`rtmx sync --bidirectional` shall persist every local mutation selected by
conflict resolution and shall establish durable linkages for matched or newly
created items.

## Rationale

Bidirectional sync currently reports remote-wins status changes without
assigning them locally, while unlinked external and local records are presented
only as candidates. Durable, deterministic state transitions are required
before bidirectional mode can converge across repeated runs.

## Acceptance Criteria

1. `--prefer-remote` assigns the mapped remote status locally and the change
   survives database reload.
2. `--prefer-local` reports success only when the remote update succeeds.
3. Items matched through an embedded RTMX requirement ID gain a durable
   `external_id` linkage.
4. Any implemented import or export candidate action checkpoints its resulting
   local state before continuing.
5. Adapter and persistence failures are represented in `SyncResult.Errors`.
6. Re-running bidirectional sync after success converges to skipped/no-change
   results rather than repeating mutations.
7. Dry-run mode performs no local or remote writes.

## Files to Create/Modify

- `internal/cmd/sync.go`
- `internal/cmd/sync_test.go`

## Test Strategy

- **Test Module**: `internal/cmd/sync_test.go`
- **Planned Tests**:
  - remote-wins status survives reload;
  - failed local-wins adapter update is reported;
  - requirement-ID matching persists linkage;
  - second run converges;
  - dry run remains write-free.
- **Validation Method**: Integration Test
