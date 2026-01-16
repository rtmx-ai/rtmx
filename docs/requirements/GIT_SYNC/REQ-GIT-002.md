# REQ-GIT-002: Pre-commit hook for requirement validation

## Status: NOT STARTED
## Priority: HIGH
## Phase: 15
## Effort: 1.0 weeks

## Description

Implement a pre-commit hook that validates staged RTM CSV files before allowing commits. The hook shall perform schema validation, duplicate ID detection, and dependency cycle detection to prevent invalid requirements from entering the repository.

## Acceptance Criteria

- [ ] Hook validates only staged RTM CSV files (not working tree)
- [ ] Schema validation ensures all required columns present
- [ ] Schema validation ensures column values match expected types/formats
- [ ] Duplicate `req_id` detection across all staged RTM files
- [ ] Dependency cycle detection via graph analysis
- [ ] Orphaned dependency detection (refs to non-existent requirements)
- [ ] Hook outputs specific line numbers and fields with errors
- [ ] Exit code 0 when validation passes, non-zero on failure
- [ ] Hook is fast (<1 second for typical RTM size)
- [ ] `rtmx install --hooks --validate` installs the validation hook

## Test Cases

- `tests/test_hooks.py::test_precommit_valid_rtm` - Valid RTM passes
- `tests/test_hooks.py::test_precommit_missing_column` - Missing column fails
- `tests/test_hooks.py::test_precommit_invalid_status` - Invalid status value fails
- `tests/test_hooks.py::test_precommit_duplicate_id` - Duplicate req_id fails
- `tests/test_hooks.py::test_precommit_cycle_detection` - Dependency cycle fails
- `tests/test_hooks.py::test_precommit_orphaned_dependency` - Missing dependency warns
- `tests/test_hooks.py::test_precommit_only_staged` - Ignores unstaged changes

## Technical Notes

### Hook Implementation

The pre-commit hook shall be implemented as a shell script that invokes `rtmx validate-staged`:

```bash
#!/bin/sh
# RTMX pre-commit validation hook
# Installed by: rtmx install --hooks --validate

# Get list of staged RTM CSV files
STAGED_RTM=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.csv$')

if [ -n "$STAGED_RTM" ]; then
    echo "Validating staged RTM files..."
    rtmx validate-staged $STAGED_RTM
    exit $?
fi
```

### Staged File Validation

To validate staged content (not working tree):

```bash
# Extract staged content to temp file
git show :path/to/file.csv > /tmp/staged_rtm.csv
rtmx validate /tmp/staged_rtm.csv
```

### Validation Rules

1. **Schema Validation**
   - Required columns: req_id, category, requirement_text, status
   - Valid status values: MISSING, PARTIAL, COMPLETE
   - Valid priority values: LOW, MEDIUM, HIGH, CRITICAL
   - Phase must be numeric

2. **Referential Integrity**
   - All `depends_on` references must exist in RTM
   - All `blocks` references must exist in RTM
   - Warn on references to requirements with MISSING status

3. **Graph Validation**
   - No cycles in dependency graph
   - No self-references in dependencies

## Files to Create/Modify

- `src/rtmx/cli/validate.py` - Add `validate-staged` command
- `src/rtmx/cli/install.py` - Add `--validate` option to hooks
- `src/rtmx/validation.py` - Add staged file validation logic
- `tests/test_hooks.py` - Hook validation tests

## Dependencies

- REQ-DX-005: Git hook integration (basic hook infrastructure)

## Blocks

- REQ-GIT-003: Post-merge hook (uses same validation infrastructure)
