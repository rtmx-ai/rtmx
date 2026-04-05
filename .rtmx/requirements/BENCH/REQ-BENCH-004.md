# REQ-BENCH-004: Rust Language Benchmark (rtmx-ai/aegis-cli)

## Metadata
- **Category**: BENCH
- **Subcategory**: Rust
- **Priority**: HIGH
- **Phase**: 21
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-001, REQ-LANG-005
- **Blocks**: (none)

## Requirement

RTMX benchmark framework shall validate the Rust scanner against `aegis-cli`, confirming marker extraction and cargo test output parsing on a production Rust project.

## Rationale

Aegis is an internal RTMX project, giving full control over marker placement. Rust uses three marker styles (`#[req("...")]`, `// @req REQ-ID`, `rtmx::req()`), and this benchmark validates all three in a real workspace with multiple crates.

## Exemplar

| Field | Value |
|-------|-------|
| Repository | [rtmx-ai/aegis-cli](https://github.com/rtmx-ai/aegis-cli) |
| Pinned ref | Latest stable tag at implementation time |
| License | Internal / Apache-2.0 |
| Test count | ~437 |
| Test framework | `cargo test` (stdlib) |
| Build time | ~3 min |

## Design

### Marker Patch

Since aegis-cli is an internal project, markers may already exist or can be added upstream. Patch adds markers to:
- Unit tests across at least 3 crates
- Integration tests in `tests/`
- All three marker styles represented

Minimum 30 markers.

### Benchmark Config

```yaml
language: rust
exemplar:
  repo: rtmx-ai/aegis-cli
  ref: v0.5.0
  license: Apache-2.0
clone_depth: 1
marker_patch: patches/rust/aegis-cli.patch
expected_markers: 30
scan_command: rtmx from-tests --format json .
verify_command: cargo test --workspace
timeout_minutes: 15
```

### Validation Checks

1. `rtmx from-tests` extracts >= 30 markers from patched source
2. All three Rust marker styles detected
3. Markers span >= 3 crates in the workspace
4. `cargo test --workspace` succeeds on patched source
5. `rtmx verify --command "cargo test --workspace"` parses output correctly

## Acceptance Criteria

1. `benchmarks/configs/rust.yaml` exists with valid config
2. Patch applies cleanly to pinned ref
3. `make -C benchmarks run LANG=rust` completes successfully
4. All three marker styles represented in extracted markers
5. Verify output maps all markers to correct status

## Effort Estimate

0.5 weeks

## Test Strategy

- `make -C benchmarks run LANG=rust` in CI
- Baseline stored in `benchmarks/results/baselines/rust.json`
