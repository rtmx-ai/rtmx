# REQ-HYGIENE-001: Native RTM Hygiene Command

## Description
`rtmx hygiene` shall report requirement actionability and traceability hygiene findings so projects do not need custom local RTMX hygiene scripts.

## Target
Reports effort bounds, generic owner, missing test mapping, missing external ID, generic acceptance criteria, and dependency cycles.

## Acceptance Criteria
1. `rtmx hygiene` prints a text report with total requirements, finding count, summary counts, and representative findings.
2. `rtmx hygiene --json` emits machine-readable JSON containing total, findings, summary, and cycles.
3. `rtmx hygiene --strict` returns exit code 1 when findings are present.
4. `rtmx hygeine` is accepted as a compatibility alias for the common misspelling.

## Validation
- **Test**: `internal/cmd/hygiene_test.go::TestHygieneChecksDetectFindings`
- **Method**: Unit Test

## Dependencies
- `REQ-GO-012`
- `REQ-GO-013`
