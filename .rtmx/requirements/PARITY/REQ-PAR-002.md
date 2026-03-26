# REQ-PAR-002: Fail-Under Threshold for Status

## Metadata
- **Category**: PARITY
- **Subcategory**: CI
- **Priority**: P0
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-010
- **Blocks**: REQ-GO-047

## Requirement

`rtmx status --fail-under N` shall exit with code 1 if completion percentage is below the threshold N, enabling CI/CD quality gates.

## Design

```bash
rtmx status --fail-under 80    # Exit 1 if completion < 80%
rtmx status --fail-under 100   # Exit 1 unless 100% complete
```

## Acceptance Criteria

1. `--fail-under N` exits 0 if completion >= N%
2. `--fail-under N` exits 1 if completion < N%
3. Threshold applies to overall completion, not per-phase
4. Output still renders normally before exit
5. Works with `--json` (includes threshold result in JSON)

## Files to Modify

- `internal/cmd/status.go` - Add `--fail-under` flag and exit logic
- `internal/cmd/status_test.go` - Threshold tests
