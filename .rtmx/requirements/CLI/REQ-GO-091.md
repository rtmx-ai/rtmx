# REQ-GO-091: rtmx connect monday

## Status: COMPLETE
## Priority: P0
## Phase: 22
## External ID: rtmx-ai/REQ-MONO-038c

## Description

Go CLI SHALL support `rtmx connect monday` with the same OAuth start flow as
Asana/Jira against managed Sync `/connections/monday/start`.

## Acceptance Criteria

- [x] `monday` accepted by `rtmx connect`
- [x] Mock HTTP test covers start path and 503 env hint (`RTMX_MONDAY_CLIENT_ID`)
- [x] Unsupported-provider error lists monday among supported providers

## Dependencies

- rtmx-ai/REQ-MONO-038c
