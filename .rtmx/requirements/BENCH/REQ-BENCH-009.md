# REQ-BENCH-009: Java/Defense Benchmark (TAK-Product-Center/Server)

## Metadata
- **Category**: BENCH
- **Subcategory**: Java
- **Priority**: HIGH
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-001, REQ-LANG-008
- **Blocks**: (none)

## Requirement

RTMX benchmark framework shall validate the Java scanner against TAK Server, a defense-grade situational awareness platform, confirming marker extraction and test output parsing on a large-scale government Java project.

## Rationale

Google Gson (REQ-BENCH-006) validates the scanner on a small, clean library. TAK Server is the opposite end of the spectrum: a multi-module Gradle project with 77 test files, 345 @Test annotations, military-grade code (DO-178C adjacent), and real-world complexity (Ignite caches, CoT message parsing, KML services, federation). This validates that the Java scanner handles:
- Multi-module Gradle builds (vs Maven in gson)
- JUnit 4 @Test annotations with expected exceptions
- Mixed test naming conventions
- Large codebases with deep package hierarchies

TAK Server is also a strategic exemplar: it demonstrates rtmx's value in exactly the defense/government domain where requirements traceability is mandated.

## Exemplar

| Field | Value |
|-------|-------|
| Repository | [TAK-Product-Center/Server](https://github.com/TAK-Product-Center/Server) |
| Pinned ref | 5.7-RELEASE-14 |
| License | BSD-like (TAK Server license) |
| Test count | ~345 @Test annotations |
| Test framework | JUnit 4 |
| Build time | ~5 min |

## Design

### Marker Patch

Add `// rtmx:req REQ-BENCH-TAK-NNN` comments to tests across 3+ modules:
- `takserver-core/takserver-war/test/` -- utility and service tests (~8 tests)
- `takserver-core/src/test/` -- message conversion and cache tests (~7 tests)
- `takserver-fig-core/rol/src/test/` -- ROL builder and mission tests (~5 tests)

Minimum 20 markers across at least 3 modules.

### Benchmark Config

```yaml
language: java
exemplar:
  repo: TAK-Product-Center/Server
  ref: 5.7-RELEASE-14
  license: BSD
clone_depth: 1
marker_patch: patches/java/tak-server.patch
expected_markers: 20
scan_command: rtmx from-tests --format json .
verify_command: gradle test
timeout_minutes: 15
```

### Validation Checks

1. `rtmx from-tests` extracts >= 20 markers from patched source
2. Markers span >= 3 Gradle modules
3. Scanner handles JUnit 4 `@Test(expected=...)` pattern
4. Verify output maps markers to correct status

## Acceptance Criteria

1. `benchmarks/configs/tak-server.yaml` exists with valid config
2. Patch applies cleanly to pinned ref
3. `make -C benchmarks run LANG=tak-server` completes successfully
4. Extracted marker count matches or exceeds baseline

## Effort Estimate

0.5 weeks

## Test Strategy

- `make -C benchmarks run LANG=tak-server` in CI
- Baseline stored in `benchmarks/results/baselines/tak-server.json`
