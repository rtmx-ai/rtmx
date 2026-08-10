# REQ-SYNC-001b: Durable Imported State and Linkage

## Metadata
- **Category**: SYNC
- **Subcategory**: Persistence
- **Priority**: HIGH
- **Phase**: 29
- **Status**: MISSING
- **Dependencies**: REQ-GO-028
- **Blocks**: REQ-SYNC-001

## Requirement

`rtmx sync --import` shall persist accepted remote status changes and discovered
requirement-to-item linkages before reporting them as updated.

## Rationale

Import currently detects status differences and external items that reference a
known requirement, but it only prints the proposed mutation. Without assigning
the new status or `external_id` and saving the database, every later import
rediscovers the same change.

## Acceptance Criteria

1. A linked external item's mapped status is assigned to the local requirement
   and survives database reload.
2. An external item carrying a known RTMX requirement ID sets that
   requirement's `external_id` and survives database reload.
3. Unchanged linked items do not rewrite the database.
4. A persistence failure is returned as a sync error and is not reported as a
   successful update.
5. Dry-run mode reports changes without mutating memory or disk.
6. Database load failures are surfaced rather than silently treated as an empty
   database.

## Files to Create/Modify

- `internal/cmd/sync.go`
- `internal/cmd/sync_test.go`

## Test Strategy

- **Test Module**: `internal/cmd/sync_test.go`
- **Planned Tests**:
  - imported status survives reload;
  - discovered linkage survives reload;
  - dry run remains write-free;
  - save and load failures are surfaced.
- **Validation Method**: Integration Test
