# REQ-BENCH-007: C#/.NET Language Benchmark (jbogard/MediatR)

## Metadata
- **Category**: BENCH
- **Subcategory**: CSharp
- **Priority**: HIGH
- **Phase**: 21
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-001, REQ-LANG-009
- **Blocks**: (none)

## Requirement

RTMX benchmark framework shall validate the C#/.NET scanner against MediatR, confirming xUnit attribute extraction and `dotnet test` output parsing on a production .NET project.

## Rationale

MediatR uses xUnit with `[Fact]`, `[Theory]`, and custom attributes -- the exact patterns the C# scanner must handle. It is a well-structured, moderate-size project that exercises the scanner without excessive build times.

## Exemplar

| Field | Value |
|-------|-------|
| Repository | [jbogard/MediatR](https://github.com/jbogard/MediatR) |
| Pinned ref | v12.4.0 (or latest stable at implementation time) |
| License | Apache-2.0 |
| Test count | ~200 |
| Test framework | xUnit |
| Build time | ~2 min |

## Design

### Marker Patch

Add `[Req("REQ-BENCH-CS-NNN")]` attributes to a representative sample:
- `test/MediatR.Tests/` -- core mediator tests (~10 tests)
- Pipeline behavior tests (~5 tests)
- Notification handler tests (~5 tests)

Minimum 15 markers across at least 2 test projects/namespaces.

Patch also adds the `ReqAttribute` class:
```csharp
[AttributeUsage(AttributeTargets.Method)]
public class ReqAttribute : Attribute
{
    public string Id { get; }
    public ReqAttribute(string id) => Id = id;
}
```

### Benchmark Config

```yaml
language: csharp
exemplar:
  repo: jbogard/MediatR
  ref: v12.4.0
  license: Apache-2.0
clone_depth: 1
setup_commands:
  - dotnet restore
marker_patch: patches/csharp/mediatr.patch
expected_markers: 15
scan_command: rtmx from-tests --format json .
verify_command: dotnet test --no-build
timeout_minutes: 10
```

### Validation Checks

1. `rtmx from-tests` extracts >= 15 markers from patched source
2. Scanner detects `[Req("...")]` attributes in xUnit test methods
3. Markers span >= 2 test namespaces
4. `dotnet test` succeeds on patched source
5. `rtmx verify --command "dotnet test"` parses output correctly

## Acceptance Criteria

1. `benchmarks/configs/csharp.yaml` exists with valid config
2. Patch applies cleanly to pinned ref
3. `make -C benchmarks run LANG=csharp` completes successfully
4. Scanner handles xUnit `[Fact]`, `[Theory]`, and custom attributes
5. Verify output maps all markers to correct status

## Effort Estimate

0.5 weeks

## Test Strategy

- `make -C benchmarks run LANG=csharp` in CI
- Baseline stored in `benchmarks/results/baselines/csharp.json`
