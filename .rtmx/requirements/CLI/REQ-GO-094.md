# REQ-GO-094: GitHub adapter vault credential

## Status: COMPLETE
## Priority: P0
## Phase: 23
## External ID: rtmx-ai/REQ-MONO-046c

## Description

GitHub adapter and `rtmx sync --service github` SHALL use a managed Sync
vault credential (`POST /connections/credential`) when `GITHUB_TOKEN` is unset.

## Acceptance Criteria

- [x] `RTMX_GITHUB_CONNECTION_TOKEN` accepted by NewGitHubAdapter
- [x] sync wires `githubEnvWithVault` fetch
- [x] Unit tests for adapter + credential HTTP client
