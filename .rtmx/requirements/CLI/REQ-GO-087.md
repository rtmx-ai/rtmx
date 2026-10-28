# REQ-GO-087: Bootstrap import implementation for all PM adapters

## Status: COMPLETE
## Priority: P0
## Phase: 21
## External ID: rtmx-ai/REQ-MONO-040b

## Description

Implement bootstrap import via each ServiceAdapter (`github`, `jira`, `asana`,
`monday`, `gitlab`): remove “coming in future release” stubs; map FetchItems
to BootstrapRequirement; honor `--dry-run`, `--merge`, `--prefix`. Clear errors
when adapter disabled or misconfigured.

## Acceptance Criteria

- [x] No future-release stubs for the five adapter sources
- [x] Table-driven tests with mocks per service
- [x] Disabled-adapter failure path tested
- [x] Dry-run does not write CSV

## Dependencies

- REQ-GO-086
- rtmx-ai/REQ-MONO-040b
