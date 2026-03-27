# REQ-SEC-013: Evidence-Based Status Display

## Metadata
- **Category**: SECURITY
- **Subcategory**: Integrity
- **Priority**: CRITICAL
- **Phase**: 20
- **Status**: MISSING
- **Effort**: 1.5 weeks

## Requirement

`rtmx status` and `rtmx backlog` shall warn when displayed status has not been verified against current test evidence. `rtmx status --verify` shall run tests before displaying status, ensuring the displayed RTM always reflects reality.

## Rationale

The foundational principle of RTMX is that requirement status is derived from evidence (passing tests), never asserted by humans or agents. The original Phoenix implementation enforced this by running tests before showing status (`make rtm` ran `make test` first, then displayed status from fresh coverage data).

The current Go CLI reads `database.csv` and trusts whatever status is written there. If a developer manually edits the status column, or if tests have regressed since the last `verify --update`, the displayed status is a lie. This undermines the entire closed-loop verification model.

## Design

### Staleness Detection

Track when verify last ran by storing a timestamp:

```yaml
# .rtmx/verify.meta (auto-generated, not hand-edited)
last_verified: "2026-03-26T15:00:00Z"
last_verify_commit: "abc123"
```

When `rtmx status` or `rtmx backlog` runs:
1. Check if `.rtmx/verify.meta` exists
2. Compare `last_verify_commit` to current `HEAD`
3. If HEAD has advanced since last verify, display a warning:

```
WARNING: Status not verified since commit abc123 (3 commits behind HEAD).
Run `rtmx verify --update` or `rtmx status --verify` to refresh.
```

### Verified Status Mode

```
rtmx status --verify          # Run tests, update status, then display
rtmx status                   # Display with staleness warning if needed
rtmx status --no-warn         # Suppress staleness warning
```

`rtmx status --verify` is equivalent to:
```
rtmx verify --update && rtmx status
```

### Makefile Integration

```makefile
rtm:
	@rtmx reconcile --execute 2>/dev/null || true
	@rtmx verify --update
	@rtmx status

rtm-v:
	@rtmx reconcile --execute 2>/dev/null || true
	@rtmx verify --update
	@rtmx status -v
```

This mirrors the Phoenix Makefile pattern where `make rtm` always shows verified status.

### verify.meta Update

`rtmx verify --update` writes `.rtmx/verify.meta` after successful verification:
- `last_verified`: current timestamp
- `last_verify_commit`: current git HEAD SHA

### Configuration

```yaml
rtmx:
  verify:
    warn_stale: true       # Warn when status is unverified (default: true)
    auto_verify: false     # Auto-run verify before status (default: false)
```

`auto_verify: true` makes every `rtmx status` call equivalent to `rtmx status --verify`. This is the Phoenix behavior. Default is false for performance (tests can be slow).

## Acceptance Criteria

1. `rtmx verify --update` writes `.rtmx/verify.meta` with timestamp and commit SHA
2. `rtmx status` warns when HEAD has advanced past last verified commit
3. `rtmx status --verify` runs tests then displays verified status
4. `rtmx status --no-warn` suppresses the staleness warning
5. `verify.warn_stale` config controls warning behavior
6. `verify.auto_verify` config enables automatic verification before display
7. Staleness warning includes commit distance (e.g., "3 commits behind HEAD")

## Test Strategy

- **Test Module**: `internal/cmd/status_test.go`
- **Test Function**: `TestStatusStalenessWarning`
- **Validation Method**: Integration Test
