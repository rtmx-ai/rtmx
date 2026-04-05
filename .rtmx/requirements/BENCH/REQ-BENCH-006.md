# REQ-BENCH-006: Java Language Benchmark (google/gson)

## Metadata
- **Category**: BENCH
- **Subcategory**: Java
- **Priority**: HIGH
- **Phase**: 21
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-001, REQ-LANG-008
- **Blocks**: (none)

## Requirement

RTMX benchmark framework shall validate the Java scanner against Google's Gson library, confirming JUnit 5 annotation extraction and Maven test output parsing on a production Java project.

## Rationale

Gson has ~1000 tests using JUnit 5 with annotations, parameterized tests, and nested test classes -- patterns that stress the Java scanner's `@Req("...")` annotation detection across a multi-module Maven project.

## Exemplar

| Field | Value |
|-------|-------|
| Repository | [google/gson](https://github.com/google/gson) |
| Pinned ref | v2.11.0 (or latest stable at implementation time) |
| License | Apache-2.0 |
| Test count | ~1000 |
| Test framework | JUnit 5 |
| Build time | ~3 min |

## Design

### Marker Patch

Add `@Req("REQ-BENCH-JAVA-NNN")` annotations to a representative sample:
- `gson/src/test/java/com/google/gson/` -- core serialization (~10 tests)
- `gson/src/test/java/com/google/gson/stream/` -- streaming API (~5 tests)
- `gson/src/test/java/com/google/gson/functional/` -- functional tests (~10 tests)

Minimum 20 markers across at least 3 test packages.

Patch also adds the marker annotation class if not already present:
```java
@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.METHOD)
public @interface Req { String value(); }
```

### Benchmark Config

```yaml
language: java
exemplar:
  repo: google/gson
  ref: v2.11.0
  license: Apache-2.0
clone_depth: 1
setup_commands:
  - mvn dependency:resolve -q
marker_patch: patches/java/gson.patch
expected_markers: 20
scan_command: rtmx from-tests --format json .
verify_command: mvn test -q
timeout_minutes: 10
```

### Validation Checks

1. `rtmx from-tests` extracts >= 20 markers from patched source
2. Scanner detects `@Req` annotations in JUnit 5 test methods
3. Markers span >= 3 test packages
4. `mvn test` succeeds on patched source
5. `rtmx verify --command "mvn test"` parses output correctly

## Acceptance Criteria

1. `benchmarks/configs/java.yaml` exists with valid config
2. Patch applies cleanly to pinned ref
3. `make -C benchmarks run LANG=java` completes successfully
4. Scanner handles JUnit 5 annotations, parameterized tests, nested classes
5. Verify output maps all markers to correct status

## Effort Estimate

0.5 weeks

## Test Strategy

- `make -C benchmarks run LANG=java` in CI
- Baseline stored in `benchmarks/results/baselines/java.json`
