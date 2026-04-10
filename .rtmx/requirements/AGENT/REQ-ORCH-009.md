# REQ-ORCH-009: Heartbeat and Staleness Reaping for Claims

## Metadata
- **Category**: ORCH
- **Subcategory**: Orchestration
- **Priority**: MEDIUM
- **Phase**: 20
- **Status**: MISSING
- **Dependencies**: REQ-ORCH-005
- **Blocks**: (none)

## Requirement
The claim protocol shall support periodic heartbeat updates and automatic detection/reaping of stale claims from dead or abandoned agents.

## Design
- `rtmx heartbeat [--agent NAME]` updates heartbeat timestamp for all agent claims
- Stale = heartbeat older than configured timeout (default 1 hour, via agent.claim_timeout in rtmx.yaml)
- `rtmx release --stale` reaps all stale claims
- Optionally: stale check on `rtmx next` so dead claims don't block new work

## Acceptance Criteria
1. `rtmx heartbeat` updates timestamps for the calling agent.
2. Claims older than timeout are detected as stale.
3. `rtmx release --stale` removes stale claims.
4. Timeout configurable via rtmx.yaml.
5. `rtmx next` skips stale claims when selecting work.

## Files to Create/Modify
- internal/orchestration/claims.go
- internal/orchestration/claims_test.go
- internal/config/config.go
- internal/cmd/claim.go
