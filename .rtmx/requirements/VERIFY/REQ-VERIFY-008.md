# REQ-VERIFY-008: Configurable Warn Threshold for Status Changes

## Metadata
- **Category**: VERIFY
- **Subcategory**: UX
- **Priority**: MEDIUM
- **Phase**: 28
- **Status**: MISSING
- **Dependencies**: REQ-VERIFY-001
- **Blocks**: 
- **External ID**: aegis-cli

## Requirement

`rtmx verify` shall accept a `--warn-threshold=N` flag to configure the
number of status changes that trigger a warning message, replacing the
current hard-coded threshold of 5.

## Rationale

The current verify command emits a warning when more than 5 status changes
occur in a single run:

```
WARNING: 6 status changes exceed warn threshold (5)
```

During normal batch development -- especially when bootstrapping a new
project or completing a sprint of work -- it is common to have many
requirements transition status simultaneously. The hard-coded threshold
of 5 fires frequently during legitimate development, creating noise that
trains users to ignore warnings.

The aegis-cli team (rtmx-ai/aegis-cli) reported this as a usability issue
during their initial onboarding, where dozens of requirements transition
from MISSING to COMPLETE as tests are linked for the first time.

## Design

### New Flag

`--warn-threshold=N` (integer, default 5) controls the threshold. Setting
N=0 disables the warning entirely.

### Configuration File Support

The threshold can also be set in `rtmx.yaml`:

```yaml
verify:
  warn_threshold: 20
```

The flag takes precedence over the configuration file value.

### Behavior

- `--warn-threshold=5` (default): current behavior preserved.
- `--warn-threshold=0`: warning suppressed entirely.
- `--warn-threshold=50`: warning only fires for very large batches.
- Negative values are treated as 0 (disabled).

## Acceptance Criteria

1. `--warn-threshold=N` flag accepted by `rtmx verify`.
2. Default value of 5 preserves current behavior.
3. Setting threshold to 0 suppresses the warning.
4. Setting threshold higher than the number of changes suppresses the warning.
5. Configuration file `verify.warn_threshold` is respected.
6. Flag value overrides configuration file value.
7. Help text documents the flag and its default.

## Files to Create/Modify

- `internal/cmd/verify.go` -- Add flag, replace hard-coded threshold
- `internal/cmd/verify_test.go` -- Tests for configurable threshold
- `internal/config/config.go` -- Add verify.warn_threshold field
- `.rtmx/database.csv` -- Add this requirement

## Effort Estimate

0.5 weeks
