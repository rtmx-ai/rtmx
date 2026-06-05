# REQ-ADAPT-002: Asana Bidirectional Sync

## Metadata
- **Category**: ADAPT
- **Subcategory**: Asana
- **Priority**: HIGH
- **Phase**: 27
- **Status**: MISSING
- **Dependencies**: REQ-ADAPT-001
- **Blocks**: REQ-ADAPT-003

## Requirement

The Asana adapter shall support bidirectional synchronization between
Asana tasks and RTMX requirements, mapping task completion state to
RTMX status, custom fields to requirement metadata, and task assignee
to RTMX assignee.

## Rationale

One-way sync forces teams to choose between tools. Bidirectional sync
lets developers update status in whichever tool they prefer, with changes
flowing to the other within the sync interval. This is essential for
adoption in teams where not everyone uses the CLI.

## Design

### Status Mapping

| Asana Status | RTMX Status |
|-------------|-------------|
| Incomplete (no section) | NOT_STARTED |
| Incomplete (in active section) | MISSING |
| Incomplete (in progress section) | PARTIAL |
| Completed | COMPLETE |

Section-to-status mapping is configurable:

```yaml
rtmx:
  adapters:
    asana:
      status_mapping:
        sections:
          "To Do": "NOT_STARTED"
          "In Progress": "PARTIAL"
          "Review": "PARTIAL"
          "Done": "COMPLETE"
```

### Requirement ID Linking

Each synced task includes the RTMX requirement ID in a custom field
or in the task description as a `[REQ-XXX-NNN]` marker. The adapter
uses this to correlate tasks with requirements.

### Sync Algorithm

```
for each requirement with external_id matching an Asana GID:
    fetch Asana task
    if Asana task updated more recently:
        update RTMX status from Asana
    elif RTMX requirement updated more recently:
        update Asana task from RTMX
    else:
        no-op (in sync)
```

Conflict resolution: last-write-wins based on `updated_at` timestamps.

### Field Mapping

| Asana Field | RTMX Field |
|------------|------------|
| name | requirement_text (first line) |
| completed | status (COMPLETE if true) |
| assignee.name | assignee |
| custom_field:priority | priority |
| custom_field:effort | effort_weeks |
| section | status (via mapping) |
| notes | notes |

### Sync Command

```bash
rtmx sync --adapter asana              # sync all linked requirements
rtmx sync --adapter asana --dry-run    # preview changes
rtmx sync --adapter asana --pull-only  # one-way pull from Asana
```

## Acceptance Criteria

1. `rtmx sync --adapter asana` syncs status bidirectionally.
2. Status mapping respects configured section-to-status rules.
3. Requirement ID in task description is preserved across syncs.
4. `--dry-run` shows planned changes without executing them.
5. `--pull-only` only updates RTMX from Asana (no Asana writes).
6. Conflict resolution uses last-write-wins timestamp comparison.
7. New requirements without Asana tasks can be pushed via CreateItem.
8. Sync report shows counts: updated, created, skipped, conflicted.

## Files to Create/Modify

- `internal/adapters/asana.go` -- Sync logic in adapter
- `internal/adapters/asana_test.go` -- Sync scenario tests
- `internal/cmd/sync.go` -- --adapter flag routing

## Effort Estimate

1 week

## Test Strategy

- Sync scenarios: Asana newer, RTMX newer, in-sync, both changed
- Status mapping: each section maps to correct RTMX status
- Dry run: verify no HTTP mutations in dry-run mode
- Pull-only: verify no Asana writes
- Conflict: verify last-write-wins resolution
- New requirement push: verify CreateItem called
