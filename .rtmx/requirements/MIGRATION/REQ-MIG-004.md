# REQ-MIG-004: Seamless Local Schema Migration on CLI Upgrade

## Metadata
- **Category**: MIGRATION
- **Subcategory**: SchemaEvolution
- **Priority**: HIGH
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-GO-007

## Requirement

When a user upgrades the rtmx binary (e.g., `brew upgrade rtmx`) and the
CSV database schema has changed, rtmx shall automatically detect the
schema version mismatch and migrate the local `.rtmx/database.csv` to
the current schema without data loss. The migration shall be seamless:
no manual intervention required, no data lost, and a clear log of what
changed.

## Rationale

Today, rtmx handles missing columns via implicit migration: the CSV
parser returns empty strings for missing columns, and the next write
operation silently adds new columns to the file. This works for additive
changes but has several gaps:

1. No schema version tracking -- rtmx cannot distinguish "old schema"
   from "new schema with empty values."
2. No proactive detection -- the user is not told their schema is
   outdated until they happen to run a write command.
3. No migration log -- when columns appear or disappear, there is no
   record of what changed.
4. Column removal risk -- if a future version removes a column, the
   current design would silently drop data.
5. `rtmx migrate` only checks Python-to-Go compatibility, not schema
   evolution between Go versions.

## Design

### Schema Version Tracking

Add a `schema_version` field to `.rtmx/config.yaml`:

```yaml
rtmx:
  database: .rtmx/database.csv
  schema: core
  schema_version: 2   # tracks the database column layout version
```

Each RTMX release that changes the standard column set increments this
version. The version is checked on every database load.

### Migration Logic

On `database.Load()`:

1. Read the CSV header and compare to `standardColumns`.
2. Detect missing columns (present in code, absent in file).
3. Detect unknown columns (present in file, absent in code) -- these
   are either custom extensions or removed columns.
4. If missing columns exist:
   a. Add the columns with empty default values.
   b. Write the migrated database back.
   c. Print a migration summary: "Migrated database schema: added
      columns [external_id]. Run `rtmx health` to verify."
5. If unknown columns exist that match previously-standard columns
   (tracked in a removal registry): warn but preserve in `Extra`.
6. Update `schema_version` in config.

### CLI Surface

- `rtmx health` gains a "Schema version" check that reports whether
  the local database matches the current schema.
- `rtmx setup` runs migration as part of its idempotent setup.
- All commands that load the database trigger migration detection, but
  only commands that write (verify --update, release assign, etc.)
  actually perform the migration. Read-only commands print a warning.

### Backward Compatibility

- Databases without `schema_version` are treated as version 1 (the
  original 20-column core schema).
- Migration is always additive -- new columns get empty defaults.
- Custom columns (not in any known version) are always preserved in
  `Extra`.
- A removed-column registry prevents accidental data loss: if a column
  was standard in version N but removed in version N+1, it is kept in
  `Extra` with a deprecation notice.

## Acceptance Criteria

1. `rtmx status` on a database missing the `external_id` column prints
   a migration warning and adds the column on next write
2. `rtmx health` reports schema version mismatch
3. `rtmx setup` migrates the database schema to current version
4. Migration preserves all existing data including custom columns
5. Migration log shows which columns were added
6. No data loss when columns are removed from the standard set
7. Schema version persisted in config after migration
8. Round-trip test: load old-format CSV, save, reload -- all data intact

## Files to Create/Modify

- `internal/database/csv.go` -- Migration detection and execution
- `internal/database/csv_test.go` -- Migration tests
- `internal/database/migration.go` -- Migration registry (new)
- `internal/config/config.go` -- schema_version field
- `internal/cmd/health.go` -- Schema version health check
- `internal/cmd/setup.go` -- Trigger migration during setup

## Effort Estimate

2.0 weeks
