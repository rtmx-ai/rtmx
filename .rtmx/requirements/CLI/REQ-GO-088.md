# REQ-GO-088: Infer issue-store provider hint from git remotes

## Status: COMPLETE
## Priority: P0
## Phase: 22
## External ID: rtmx-ai/REQ-MONO-044a

## Description

Go CLI SHALL expose a testable API that maps git remote URLs to a provider
hint (`github`, `gitlab`, or empty) for the link picker — no network I/O.

## Acceptance Criteria

- [x] Table-driven tests for github.com, git@github.com, gitlab.com, unknown hosts, empty remotes
- [x] Used by `rtmx link` / init handoff (044b/c)

## Dependencies

- rtmx-ai/REQ-MONO-044a
