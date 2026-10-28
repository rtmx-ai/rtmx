# REQ-GO-093: rtmx connect github

## Status: COMPLETE
## Priority: P0
## Phase: 22
## External ID: rtmx-ai/REQ-MONO-038e

## Description

Go CLI SHALL support `rtmx connect github` with the same OAuth start flow as
Asana/Jira/Monday/GitLab against managed Sync `/connections/github/start`.
This is GitHub Issues vault OAuth — not forge org prove.

## Acceptance Criteria

- [x] `github` accepted by `rtmx connect`
- [x] Mock HTTP test covers start path and 503 env hint (`RTMX_GITHUB_ISSUES_CLIENT_ID`)
- [x] Unsupported-provider error lists github among supported providers

## Dependencies

- rtmx-ai/REQ-MONO-038e
