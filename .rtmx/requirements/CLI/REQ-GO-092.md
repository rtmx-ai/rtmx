# REQ-GO-092: rtmx connect gitlab

## Status: COMPLETE
## Priority: P0
## Phase: 22
## External ID: rtmx-ai/REQ-MONO-038d

## Description

Go CLI SHALL support `rtmx connect gitlab` with the same OAuth start flow as
Asana/Jira/Monday against managed Sync `/connections/gitlab/start`.

## Acceptance Criteria

- [x] `gitlab` accepted by `rtmx connect`
- [x] Mock HTTP test covers start path and 503 env hint (`RTMX_GITLAB_CLIENT_ID`)
- [x] Unsupported-provider error lists gitlab among supported providers

## Dependencies

- rtmx-ai/REQ-MONO-038d
