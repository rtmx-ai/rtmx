# REQ-LANG-030: Multi-file JUnit ingestion in from-pytest

## Metadata
- **Category**: LANG
- **Subcategory**: Python
- **Priority**: HIGH
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-LANG-004

## Requirement

`rtmx from-pytest` shall accept more than one `--junitxml` value — the flag is
repeatable and each value is glob-expanded — and shall ingest the test cases from
every matched file as a single result set. Duplicate paths are de-duplicated; a
literal path that matches nothing is preserved verbatim so a genuinely missing
file still surfaces as a read error.

## Rationale

A sharded CI matrix emits one JUnit report per leg (e.g. `output/junit-*.xml`).
Previously `from-pytest` accepted a single `--junitxml`, forcing every consumer to
pre-merge the per-leg reports themselves before calling rtmx (the Phoenix radar
project carried a hand-rolled `merge_junit` for exactly this). Ingesting many
files — or a glob — in one call removes that workaround and makes
`rtmx from-pytest --no-run --junitxml 'output/junit-*.xml'` the whole story.

## Acceptance Criteria

- [ ] `--junitxml` is repeatable and glob-expanded; cases from all matched files
      are combined.
- [ ] Duplicate matched paths are de-duplicated.
- [ ] A literal non-matching path is preserved (surfaces as a read error), keeping
      prior single-file behavior.
- [ ] Verified by `internal/cmd/from_pytest_test.go::TestExpandJUnitPaths` and
      `::TestFromPytestMultiFileJUnit`.
