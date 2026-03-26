# REQ-SEC-011: Configurable Verify Throughput Thresholds

## Metadata
- **Category**: SECURITY
- **Subcategory**: Governance
- **Priority**: HIGH
- **Phase**: 20
- **Status**: MISSING
- **Effort**: 1.5 weeks
- **Dependencies**: REQ-SEC-010

## Requirement

`rtmx verify --update` shall enforce configurable thresholds on the number of requirements that can change status in a single run. Exceeding the warn threshold produces a warning; exceeding the fail threshold causes the command to exit non-zero without writing changes.

## Rationale

Teams have different risk tolerances for batch requirement updates. A solo developer may accept 5 changes per commit. A safety-critical system may allow only 1. A migration sprint may temporarily need 50. The threshold should be configurable at the repo level via `rtmx.yaml`, auditable in git history, and enforceable in CI.

This also enables future organizational policy enforcement via rtmx-sync, where a central authority can set maximum thresholds for downstream repos (e.g., IL4/IL5 systems capped at N=1).

## Design

### Configuration

```yaml
rtmx:
  verify:
    auto_update: true
    thresholds:
      warn: 5       # Warn when status changes exceed this count
      fail: 15      # Fail (exit 1, no write) when changes exceed this count
    audit_log: true  # Log status changes as structured output for CI artifacts
```

### Behavior

| Changes | Action |
|---------|--------|
| 0 | "No status changes needed" |
| 1 to warn | Auto-commit silently |
| warn+1 to fail | Auto-commit with WARNING annotation in commit message |
| fail+1 and above | Exit 1, do not write, print error with guidance |

### CLI Output

Normal (within warn threshold):
```
Verification Results: 2 status changes -> updated
```

Warning band:
```
Verification Results: 8 status changes
  WARNING: Exceeds warn threshold (5). Review changes carefully.
  Changes written. Commit message will note batch update.
```

Failure band:
```
Verification Results: 20 status changes
  ERROR: Exceeds fail threshold (15). Changes NOT written.
  Run `rtmx verify --update --force` to override, or adjust
  rtmx.verify.thresholds.fail in your config.
```

### Override

`rtmx verify --update --force` bypasses the fail threshold for one invocation. This is for planned batch operations. The `--force` flag is never used by the CI auto-commit job, so it cannot be exploited remotely.

### CI Integration

The verify-requirements CI job:
1. Runs `rtmx verify --update --verbose` (respects thresholds)
2. If the command exits 0, commits the changes
3. If it exits 1 (threshold exceeded), the job fails visibly
4. Before committing, validates `git diff --name-only` shows only `.rtmx/database.csv`

### Defaults

- `warn: 5` -- generous enough for normal development, flags batch work
- `fail: 15` -- catches anomalies and attacks, allows deliberate batches
- `auto_update: true` -- opt-out, not opt-in
- `audit_log: true` -- structured log of all status changes

## Acceptance Criteria

1. Thresholds configurable in `rtmx.yaml` under `verify.thresholds.warn` and `verify.thresholds.fail`
2. Changes within warn threshold commit silently
3. Changes between warn and fail thresholds commit with WARNING
4. Changes exceeding fail threshold cause exit 1 with no database write
5. `--force` flag overrides the fail threshold for one invocation
6. Sensible defaults (warn=5, fail=15) when no config is set
7. CI auto-commit validates only database.csv was modified before committing

## Test Strategy

- **Test Module**: `internal/cmd/verify_test.go`
- **Test Function**: `TestVerifyThresholds`
- **Validation Method**: Integration Test

### Test Cases

1. Changes within warn threshold -- silent commit, exit 0
2. Changes exceeding warn but within fail -- commit with warning, exit 0
3. Changes exceeding fail threshold -- no write, exit 1
4. --force flag overrides fail threshold -- writes, exit 0
5. Custom thresholds from config respected
6. Default thresholds used when config absent
7. CI file validation -- only database.csv modified
