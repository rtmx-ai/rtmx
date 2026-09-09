# REQ-DATA-001b: AC-level verify spike (ATDD evidence)

## Metadata
- **Category**: DATA
- **Subcategory**: DOCUMENT_MODEL
- **Priority**: HIGH
- **Phase**: 34
- **Status**: MISSING
- **Dependencies**: REQ-DATA-001a
- **Blocks**: REQ-DATA-001e|REQ-DATA-001f
- **External ID**:
- **ADR**: docs/adr/0007-requirement-document-model.md

## Requirement

On a fixture project only, implement enough reader and verify logic to compute
per-AC pass/fail from markers or results JSON, report per-AC gaps, and
aggregate to requirement status under a documented ATDD completeness policy
(example: COMPLETE iff all required ACs pass; PARTIAL if any required AC
passes; MISSING if none). Req-level-only bindings remain supported as a
compat path.

## Rationale

The primary product outcome is AC-to-test ATDD. Schema alone does not prove
agents can see which ACs lack evidence. This spike is the decisive fitness
test for the document model.

## Acceptance Criteria

1. Fixture with one requirement, three ACs, and three tests (one binding per AC); verify (spike path) prints an AC matrix with pass/fail/gap per `ac_id`.
2. Failing or unbound required AC yields PARTIAL (or policy-documented equivalent), never silent COMPLETE.
3. Completeness policy is written beside the spike (policy name + aggregation rules).
4. Req-level-only bindings still produce req-level status (v1 compat path).
5. Spike code may live under `testdata/` / experimental package; it is not wired as the default CLI verify path.

## Test Strategy

Table-driven tests on the fixture project: all green → COMPLETE; one AC fail →
PARTIAL; one AC unbound → gap reported; skip/non-evidence behavior documented
relative to VERIFY completeness rules where applicable.

## Out of Scope

Default CLI flip; monorepo RTM migration; CRDT sync of AC status (covered in 001e sketch only).
