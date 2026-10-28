# REQ-GO-086: Bootstrap flags for all adapter sources

## Status: COMPLETE
## Priority: P0
## Phase: 21
## External ID: rtmx-ai/REQ-MONO-040a

## Description

`rtmx bootstrap` SHALL accept `--from-github`, `--from-jira`, `--from-asana`,
`--from-monday`, `--from-gitlab`, and `--from-tests`. Help and empty-source
errors SHALL list all of them.

## Acceptance Criteria

- [x] Flags registered on bootstrap command
- [x] `--help` lists each source
- [x] Unit test asserts help or flag presence

## Dependencies

- REQ-GO-027
- rtmx-ai/REQ-MONO-040a
