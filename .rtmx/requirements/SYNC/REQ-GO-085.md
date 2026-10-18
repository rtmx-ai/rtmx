# REQ-GO-085: Managed sync adoption heartbeat

## Metadata
- **Status**: MISSING
- **Priority**: HIGH
- **Phase**: 15
- **External ID**: rtmx-ai/REQ-MONO-033g

## Requirement

After a successful managed Sync room session for a git working tree, the CLI
SHALL POST an adoption heartbeat to Sync including canonical `external_ref`
(e.g. `github.com/owner/repo` from `git remote`), `org_slug`, and `room`.

## Acceptance Criteria

- [ ] Heartbeat sent only for managed Sync (not GitHub Issues/Jira adapters alone)
- [ ] Failure to heartbeat does not fail the sync UX (best-effort, logged)
- [ ] Unit test with mock HTTP client
- [ ] Dogfood path works for rtmx-ai/rtmx

## Files

- `internal/cmd/sync_room.go` (or sync client helper)
- `internal/sync/` adoption client
