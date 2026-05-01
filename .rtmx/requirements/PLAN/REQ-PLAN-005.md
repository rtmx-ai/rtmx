# REQ-PLAN-005: Release Gate Command

## Metadata
- **Category**: PLAN
- **Subcategory**: Release
- **Priority**: P0
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: REQ-PLAN-003
- **Blocks**: REQ-PLAN-012

## Requirement

`rtmx release gate <version>` shall verify that all requirements assigned to
`<version>` are COMPLETE. Exit code 0 on pass, 1 on failure. Supports
`--verify` to run test verification first, and `--json` for CI consumption.

## Design

```bash
# Basic gate check
rtmx release gate v0.3.0

# Run verification first, then gate
rtmx release gate v0.3.0 --verify

# Machine-readable output for CI
rtmx release gate v0.3.0 --json
```

### Gate Logic

1. Load database
2. Filter requirements where sprint == version
3. If zero requirements match: exit 1, warn "no requirements assigned to version"
4. If all COMPLETE: exit 0, print pass report
5. If any MISSING or PARTIAL: exit 1, print failure report listing incomplete requirements

### CI Integration

```yaml
# In .github/workflows/release.yml
- name: Release gate
  run: rtmx release gate ${{ github.ref_name }} --json
```

## Acceptance Criteria

1. All COMPLETE -> exit 0
2. Any MISSING/PARTIAL -> exit 1 with report
3. No requirements assigned -> exit 1 with warning
4. `--verify` runs test verification before gate check
5. `--json` outputs machine-readable report
6. Report lists each incomplete requirement with status and priority

## Files to Create

- `internal/cmd/release.go`
- `internal/cmd/release_test.go`
