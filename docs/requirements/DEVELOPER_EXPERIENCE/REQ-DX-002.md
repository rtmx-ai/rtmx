# REQ-DX-002: Auto-Migration to .rtmx/

## Status: NOT_STARTED
## Priority: HIGH
## Phase: 4
## Effort: 1.0 weeks

## Description

Automatically migrate existing projects from the legacy layout to `.rtmx/` on first run, with user confirmation and backup.

## Acceptance Criteria

- [ ] Detect old layout (rtmx.yaml in root, docs/rtm_database.csv)
- [ ] Prompt user for confirmation before migrating
- [ ] Create backup of old files before migration
- [ ] Move rtmx.yaml → .rtmx/config.yaml
- [ ] Move docs/rtm_database.csv → .rtmx/database.csv
- [ ] Move docs/requirements/ → .rtmx/requirements/
- [ ] Update any internal path references in config
- [ ] Display migration summary with changes made
- [ ] `--no-migrate` flag to suppress auto-migration
- [ ] Idempotent - running again doesn't break anything

## Migration Flow

```
$ rtmx status

╔══════════════════════════════════════════════════════════════╗
║  RTMX Layout Migration Available                              ║
╠══════════════════════════════════════════════════════════════╣
║  Detected legacy layout. The following files will be moved:   ║
║                                                               ║
║    rtmx.yaml           → .rtmx/config.yaml                   ║
║    docs/rtm_database.csv → .rtmx/database.csv                ║
║    docs/requirements/   → .rtmx/requirements/                 ║
║                                                               ║
║  Backups will be created with .rtmx-backup-TIMESTAMP suffix  ║
╚══════════════════════════════════════════════════════════════╝

Proceed with migration? [y/N]: y

✓ Backed up rtmx.yaml to rtmx.yaml.rtmx-backup-20260102
✓ Created .rtmx/ directory
✓ Moved rtmx.yaml → .rtmx/config.yaml
✓ Moved docs/rtm_database.csv → .rtmx/database.csv
✓ Moved docs/requirements/ → .rtmx/requirements/
✓ Created .rtmx/.gitignore
✓ Migration complete!

$ rtmx status
[normal output]
```

## Suppressing Migration

```bash
# Suppress for single command
rtmx status --no-migrate

# Suppress permanently via environment
export RTMX_NO_MIGRATE=1
rtmx status
```

## Files to Modify

- `src/rtmx/cli/main.py` - Pre-command migration check
- NEW `src/rtmx/migration.py` - Migration logic

## Dependencies

- REQ-DX-001 (requires new structure to exist)

## Notes

Migration should be safe and reversible. Backups ensure no data loss.
