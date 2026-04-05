# REQ-BENCH-001: Benchmark Framework and Orchestration

## Metadata
- **Category**: BENCH
- **Subcategory**: Framework
- **Priority**: HIGH
- **Phase**: 21
- **Status**: MISSING
- **Dependencies**: REQ-LANG-007
- **Blocks**: REQ-BENCH-002, REQ-BENCH-003, REQ-BENCH-004, REQ-BENCH-005, REQ-BENCH-006, REQ-BENCH-007

## Requirement

RTMX shall provide a benchmark framework within the monorepo that validates language scanners against real-world open source projects, with nightly CI regression tracking.

## Rationale

The 23 language scanners in `from_tests_langs.go` are tested against synthetic fixtures. Benchmarks against real codebases prove that scanners work on idiomatic, production-quality test suites -- catching edge cases that synthetic tests miss (nested markers, unusual formatting, framework-specific patterns).

## Design

### Directory Structure

```
benchmarks/
  README.md
  Makefile                    # orchestrate all benchmarks
  scripts/
    run-benchmark.sh          # clone, patch, scan, verify one exemplar
    report.sh                 # compare results to baseline, flag regressions
  configs/
    go.yaml                   # exemplar repo, pinned commit, expected markers
    python.yaml
    rust.yaml
    javascript.yaml
    java.yaml
    csharp.yaml
  patches/
    go/cli-cli.patch          # add rtmx markers to exemplar tests
    python/requests.patch
    ...
  results/
    latest.json               # most recent benchmark results
    baselines/                # known-good baseline per language
  .github/workflows/
    benchmark.yml             # nightly scheduled workflow
```

### Benchmark Config Schema (per language)

```yaml
language: go
exemplar:
  repo: cli/cli
  ref: v2.60.0                # pinned tag or commit SHA
  license: MIT
clone_depth: 1
marker_patch: patches/go/cli-cli.patch
expected_markers: 25          # minimum marker count after patching
scan_command: rtmx from-tests --format json .
verify_command: go test -json ./...
timeout_minutes: 10
```

### Execution Model

For each configured language:

1. Clone exemplar at pinned ref (shallow clone)
2. Apply marker patch (add `rtmx:req` markers to selected tests)
3. Run `rtmx from-tests --format json .` -- verify marker extraction
4. Run the language's test command -- verify test output parsing
5. Run `rtmx verify --results` -- verify closed-loop status update
6. Compare marker count and pass rate to baseline -- flag regressions

### Nightly CI Workflow

```yaml
name: Benchmarks
on:
  schedule:
    - cron: '0 4 * * *'       # 04:00 UTC daily
  workflow_dispatch: {}

jobs:
  benchmark:
    strategy:
      matrix:
        language: [go, python, rust, javascript, java, csharp]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: make -C benchmarks run LANG=${{ matrix.language }}
      - run: make -C benchmarks compare LANG=${{ matrix.language }}
      - uses: actions/upload-artifact@v4
        with:
          name: benchmark-${{ matrix.language }}
          path: benchmarks/results/
```

### Regression Detection

A benchmark regresses when:
- Marker count drops below expected (scanner broke)
- Previously passing tests fail to parse (output parser broke)
- Verify produces different status than baseline (status logic broke)

Regressions post a GitHub issue automatically.

## Acceptance Criteria

1. `benchmarks/` directory exists with Makefile and at least one language config
2. `make -C benchmarks run LANG=go` clones, patches, scans, and verifies successfully
3. `make -C benchmarks compare LANG=go` compares results to baseline and exits 0 on match
4. `.github/workflows/benchmark.yml` runs on schedule and matrix of languages
5. Regression in marker count or verify results causes non-zero exit and GitHub issue creation

## Files to Create/Modify

- `benchmarks/Makefile`
- `benchmarks/scripts/run-benchmark.sh`
- `benchmarks/scripts/report.sh`
- `benchmarks/configs/*.yaml`
- `.github/workflows/benchmark.yml`

## Effort Estimate

3 weeks (framework + first language integration)

## Test Strategy

- Integration test: `test/benchmark_framework_test.go` validates config parsing and execution model
- E2E test: `make -C benchmarks run LANG=go` succeeds in CI
