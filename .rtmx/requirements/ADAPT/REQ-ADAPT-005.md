# REQ-ADAPT-005: Monday.com Bidirectional Sync

## Metadata
- **Category**: ADAPT
- **Subcategory**: Monday
- **Priority**: HIGH
- **Phase**: 27
- **Status**: MISSING
- **Dependencies**: REQ-ADAPT-004
- **Blocks**: REQ-ADAPT-006

## Requirement

The Monday.com adapter shall support bidirectional synchronization between
Monday board items and RTMX requirements, mapping status column labels to
RTMX statuses, person columns to assignees, and group membership to
categories.

## Rationale

Bidirectional sync is the minimum bar for adoption in teams already
invested in Monday.com. One-way sync forces manual reconciliation.
Monday's status column is label-based (customizable per board), requiring
configurable mapping to RTMX's fixed status values.

## Design

### Status Mapping

Monday status columns use custom labels. Mapping is configured per board:

```yaml
rtmx:
  adapters:
    monday:
      status_mapping:
        column_id: "status"  # Monday column ID
        labels:
          "Working on it": "PARTIAL"
          "Stuck": "MISSING"
          "Done": "COMPLETE"
          "": "NOT_STARTED"  # default/empty label
```

### Sync Algorithm

Same last-write-wins algorithm as Asana adapter (REQ-ADAPT-002),
using Monday's `updated_at` field for conflict resolution.

### Requirement ID Linking

The adapter stores the RTMX requirement ID in a dedicated text column
on the Monday board (configurable via `req_id_column`). This provides
deterministic linking without parsing item names or descriptions.

```yaml
rtmx:
  adapters:
    monday:
      req_id_column: "text0"  # Monday column ID for req_id
```

### Sync Command

```bash
rtmx sync --adapter monday              # sync all linked items
rtmx sync --adapter monday --dry-run    # preview changes
rtmx sync --adapter monday --push-only  # one-way push to Monday
```

## Acceptance Criteria

1. `rtmx sync --adapter monday` syncs status bidirectionally.
2. Status label mapping respects configured column/label rules.
3. Requirement ID stored in dedicated Monday column.
4. `--dry-run` shows planned changes without mutations.
5. `--push-only` only updates Monday from RTMX (no RTMX writes).
6. Conflict resolution uses last-write-wins timestamp comparison.
7. New requirements pushed to Monday as new items.
8. Sync report shows counts: updated, created, skipped, conflicted.

## Files to Create/Modify

- `internal/adapters/monday.go` -- Sync logic
- `internal/adapters/monday_test.go` -- Sync scenario tests

## Effort Estimate

1 week

## Test Strategy

- Sync scenarios: Monday newer, RTMX newer, in-sync, both changed
- Status label mapping: each label maps to correct RTMX status
- Req ID column: verify stored and retrieved correctly
- Dry run: verify no GraphQL mutations
- Push-only: verify no RTMX database writes
