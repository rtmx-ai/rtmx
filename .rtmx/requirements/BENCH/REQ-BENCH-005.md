# REQ-BENCH-005: JavaScript/TypeScript Language Benchmark (sindresorhus/got)

## Metadata
- **Category**: BENCH
- **Subcategory**: JavaScript
- **Priority**: HIGH
- **Phase**: 21
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-001, REQ-LANG-006
- **Blocks**: (none)

## Requirement

RTMX benchmark framework shall validate the JavaScript/TypeScript scanner against the `got` HTTP client, confirming marker extraction and test output parsing on a production JS/TS project.

## Rationale

`got` uses AVA as its test runner with TypeScript source and ESM modules -- a modern JS/TS stack that exercises the scanner's handling of `// rtmx:req` comments, `describe.rtmx()`, and `req()` helper patterns in `.ts` test files.

## Exemplar

| Field | Value |
|-------|-------|
| Repository | [sindresorhus/got](https://github.com/sindresorhus/got) |
| Pinned ref | v14.4.0 (or latest stable at implementation time) |
| License | MIT |
| Test count | ~400 |
| Test framework | AVA |
| Build time | ~1 min |

## Design

### Marker Patch

Add `// rtmx:req REQ-BENCH-JS-NNN` comments to a representative sample:
- `test/http.ts` -- HTTP request tests (~8 tests)
- `test/retry.ts` -- retry logic (~5 tests)
- `test/hooks.ts` -- lifecycle hooks (~5 tests)
- `test/timeout.ts` -- timeout handling (~5 tests)

Minimum 20 markers across at least 3 test files.

### Benchmark Config

```yaml
language: javascript
exemplar:
  repo: sindresorhus/got
  ref: v14.4.0
  license: MIT
clone_depth: 1
setup_commands:
  - npm ci
marker_patch: patches/javascript/got.patch
expected_markers: 20
scan_command: rtmx from-tests --format json .
verify_command: npm test
timeout_minutes: 5
```

### Validation Checks

1. `rtmx from-tests` extracts >= 20 markers from patched source
2. Scanner correctly identifies `.ts` test files
3. Markers span >= 3 test files
4. `npm test` succeeds on patched source
5. `rtmx verify --command "npm test"` parses output correctly

## Acceptance Criteria

1. `benchmarks/configs/javascript.yaml` exists with valid config
2. Patch applies cleanly to pinned ref
3. `make -C benchmarks run LANG=javascript` completes successfully
4. Scanner handles TypeScript + ESM test files correctly
5. Verify output maps all markers to correct status

## Effort Estimate

0.5 weeks

## Test Strategy

- `make -C benchmarks run LANG=javascript` in CI
- Baseline stored in `benchmarks/results/baselines/javascript.json`
