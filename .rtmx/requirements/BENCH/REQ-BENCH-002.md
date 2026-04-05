# REQ-BENCH-002: Go Language Benchmark (cli/cli)

## Metadata
- **Category**: BENCH
- **Subcategory**: Go
- **Priority**: HIGH
- **Phase**: 21
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-001, REQ-LANG-003
- **Blocks**: (none)

## Requirement

RTMX benchmark framework shall validate the Go scanner against the GitHub CLI (`cli/cli`) codebase, confirming marker extraction and test output parsing on a production Go project.

## Rationale

Go is the implementation language for RTMX. The GitHub CLI has ~800 tests using stdlib `testing`, table-driven patterns, and subtests -- the exact patterns the Go scanner must handle. This is the reference benchmark.

## Exemplar

| Field | Value |
|-------|-------|
| Repository | [cli/cli](https://github.com/cli/cli) |
| Pinned ref | v2.60.0 (or latest stable at implementation time) |
| License | MIT |
| Test count | ~800 |
| Test framework | Go stdlib `testing` |
| Build time | ~2 min |

## Design

### Marker Patch

Add `rtmx.Req(t, "REQ-BENCH-GO-NNN")` calls to a representative sample of tests across packages:
- `pkg/cmd/auth/` -- authentication commands (~10 tests)
- `pkg/cmd/issue/` -- issue management (~10 tests)
- `pkg/cmd/pr/` -- pull request operations (~10 tests)
- `internal/config/` -- configuration logic (~5 tests)

Minimum 25 markers across at least 4 packages.

### Benchmark Config

```yaml
language: go
exemplar:
  repo: cli/cli
  ref: v2.60.0
  license: MIT
clone_depth: 1
marker_patch: patches/go/cli-cli.patch
expected_markers: 25
scan_command: rtmx from-tests --format json .
verify_command: go test -json ./...
timeout_minutes: 10
```

### Validation Checks

1. `rtmx from-tests` extracts >= 25 markers from patched source
2. Markers span >= 4 distinct packages
3. `go test -json` runs successfully on patched source
4. `rtmx verify --results` maps test results to requirement IDs
5. All patched tests pass (no marker insertion side effects)

## Acceptance Criteria

1. `benchmarks/configs/go.yaml` exists with valid config
2. `benchmarks/patches/go/cli-cli.patch` applies cleanly to pinned ref
3. `make -C benchmarks run LANG=go` completes successfully
4. Extracted marker count matches or exceeds baseline
5. Verify output maps all markers to COMPLETE status

## Effort Estimate

0.5 weeks

## Test Strategy

- `make -C benchmarks run LANG=go` in CI
- Baseline stored in `benchmarks/results/baselines/go.json`
