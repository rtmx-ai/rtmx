# REQ-GO-046: Migrate Command for Python Users

## Metadata
- **Category**: DIST
- **Subcategory**: Migration
- **Priority**: MEDIUM
- **Phase**: 14
- **Status**: MISSING
- **Effort**: 1.5 weeks
- **Dependencies**: REQ-GO-045 (Deprecation notice)

## Requirement

Go CLI shall include a `rtmx migrate` command that detects an existing Python rtmx installation, validates config/database compatibility, and performs any necessary conversions to ensure seamless transition.

## Rationale

Users transitioning from `pip install rtmx` to the Go binary need confidence that their existing project configuration, RTM database, and requirement specs will work identically. The migrate command automates this validation.

## Design

### Command: `rtmx migrate`

```
rtmx migrate [--check] [--fix] [--from PATH]
```

**Flags:**
- `--check` (default): Validate compatibility without making changes
- `--fix`: Auto-fix any compatibility issues found
- `--from PATH`: Path to existing Python rtmx project (default: current directory)

### Migration Checks

1. **Config compatibility**: Verify `.rtmx/config.yaml` or `rtmx.yaml` is readable by Go CLI
2. **Database compatibility**: Verify CSV schema matches (all 21 columns present)
3. **Database path**: Check both `.rtmx/database.csv` and `docs/rtm_database.csv` (Python default)
4. **Requirement specs**: Verify `.rtmx/requirements/` directory structure
5. **Python artifacts**: Detect `conftest.py` with rtmx markers, `pyproject.toml` with rtmx config
6. **Git hooks**: Check if Python-era hooks need updating to call Go binary

### Fix Actions (with --fix)

1. Move `docs/rtm_database.csv` to `.rtmx/database.csv` if needed
2. Update git hooks to use Go binary path
3. Create `.rtmx/config.yaml` from `pyproject.toml` rtmx section if needed
4. Report all changes made

### Output

```
Migration Check Results
  [PASS] Config file: .rtmx/config.yaml
  [PASS] Database: .rtmx/database.csv (108 requirements)
  [PASS] Schema: All 21 columns present
  [PASS] Requirements directory: .rtmx/requirements/ (42 specs)
  [WARN] Legacy database path: docs/rtm_database.csv (use --fix to move)
  [WARN] Git hooks reference Python rtmx (use --fix to update)

2 warnings found. Run `rtmx migrate --fix` to resolve.
```

## Acceptance Criteria

1. `rtmx migrate --check` validates config, database, and schema compatibility
2. `rtmx migrate --fix` auto-fixes legacy database paths and git hooks
3. Reports pass/warn/fail for each check
4. Detects both `.rtmx/database.csv` and `docs/rtm_database.csv` locations
5. Validates all 21 CSV columns are present
6. Non-destructive: --fix creates backups before modifying files

## Test Strategy

- **Test Module**: `internal/cmd/migrate_test.go`
- **Test Function**: `TestMigrateCommand`
- **Validation Method**: Integration Test
