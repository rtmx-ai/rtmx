# REQ-LANG-031: JUnit-driven marker discovery in from-pytest --no-run

## Metadata
- **Category**: LANG
- **Subcategory**: Python
- **Priority**: HIGH
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-LANG-030

## Requirement

When converting existing JUnit XML (`from-pytest --no-run`), `rtmx` shall discover
requirement markers not only from the given/default test paths but also from the
source files referenced by the JUnit test cases. Each case's source file is
resolved from its `file` attribute, or — when absent — from its dotted class name
(dots become slashes plus `.py`, a trailing ClassName segment is dropped, and
hyphenated directory components are preserved). A missing implicit default test
path (`tests`) is tolerated when markers can be discovered from the JUnit cases.

## Rationale

`from-pytest --no-run` scanned only the given test paths (default `tests`) for
markers, so a caller converting a CI matrix's JUnit — without enumerating every
directory — silently dropped every test outside `tests/`. In a package layout
(e.g. `packages/<name>/tests/`) that meant those requirements never joined a
result and never promoted. The Phoenix radar project worked around this with a
bespoke AST "self-join" over the JUnit class names; discovering markers from the
JUnit cases directly moves that capability into rtmx and makes the common
`from-pytest --no-run --junitxml <files>` invocation work for any project layout,
including hyphenated (non-importable) package directories.

## Acceptance Criteria

- [ ] With `--no-run` and no path argument, a marker under a package directory is
      discovered from the JUnit class name and joined.
- [ ] A hyphenated package directory (e.g. `packages/signal-processing/tests`) is
      resolved the same as any other (dots→slashes, hyphens preserved).
- [ ] A case's `file` attribute takes precedence over its class name.
- [ ] A missing implicit `tests` path does not error when JUnit-discovered markers
      exist; an explicitly-passed missing path still errors.
- [ ] Verified by `internal/cmd/from_pytest_test.go::TestClassNameToPath` and
      `::TestFromPytestNoRunDiscoversPackageMarkers`.
