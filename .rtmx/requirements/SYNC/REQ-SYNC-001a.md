# REQ-SYNC-001a: Durable and Idempotent Export Linkage

## Metadata
- **Category**: SYNC
- **Subcategory**: Persistence
- **Priority**: P0
- **Phase**: 29
- **Status**: COMPLETE
- **Dependencies**: REQ-GO-028
- **Blocks**: REQ-SYNC-001
- **Target Release**: v1.10.1

## Requirement

After an external adapter successfully creates an item, `rtmx sync --export`
shall persist the returned external key in the requirement's `external_id`
field before reporting the export as successful or creating another remote
item.

## Rationale

The export loop currently prints the returned key but does not assign it to the
requirement or save the database. A second export therefore treats the same
requirement as new and creates a duplicate Jira ticket or issue.

## Persistence Boundary

For each successful `CreateItem` call:

1. Reject an empty returned external key as an adapter error.
2. Assign the key to the in-memory requirement.
3. Atomically save the database to its configured path.
4. Only after the save succeeds, report the export as created.

Checkpointing each key separately is required. A single save after the loop
would leave all previously created items unlinked if the process were
interrupted before the final write.

If the remote item is created but the local save fails, export shall stop. The
error shall include both the requirement ID and returned remote key so an
operator can recover the linkage without creating another ticket.

## Acceptance Criteria

1. A successful create writes the returned key to `external_id`.
2. Reloading the CSV preserves the key.
3. Re-running export updates the existing remote key and does not create a
   second item.
4. Multiple successful creates are checkpointed independently.
5. A create error leaves `external_id` blank and allows other requirements to
   follow existing error handling.
6. An empty returned key is an error and is not counted as created.
7. A save failure after remote creation is an error containing the requirement
   ID and remote key; no later remote item is created.
8. Dry-run mode neither calls the adapter nor modifies the database.

## Files to Create/Modify

- `internal/cmd/sync.go`
- `internal/cmd/sync_test.go`

## Test Strategy

- **Test Module**: `internal/cmd/sync_test.go`
- **Test Functions**:
  - `TestRunExportActualCreatePersistsExternalID`
  - `TestRunExportRepeatIsIdempotent`
  - `TestRunExportCheckpointsBeforeLaterCreateFailure`
  - `TestRunExportEmptyExternalID`
  - `TestRunExportSaveFailureStopsWithRemoteID`
- **Validation Method**: Integration Test
