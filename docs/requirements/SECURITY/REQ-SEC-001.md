# REQ-SEC-001: TLS 1.3 Transport Security

## Requirement
All sync connections shall use TLS 1.3.

## Phase
9 (CRDT) - MVP Security

## Rationale
TLS 1.3 provides the strongest widely-supported transport encryption, eliminating vulnerabilities present in earlier versions (POODLE, BEAST, etc.). This is table-stakes for any sync service handling sensitive requirements data.

## Acceptance Criteria
- [ ] WebSocket connections require TLS 1.3 minimum
- [ ] HTTP API endpoints require TLS 1.3 minimum
- [ ] Server rejects connections attempting TLS 1.2 or lower
- [ ] Certificate validation is enforced (no self-signed in production)
- [ ] HSTS headers are set with minimum 1-year max-age

## Technical Notes
- Use `ssl.SSLContext` with `ssl.PROTOCOL_TLS_SERVER` and explicit minimum version
- Configure nginx/reverse proxy to enforce TLS 1.3
- Consider certificate pinning for mobile/desktop clients

## Test Cases
1. Verify TLS 1.3 connection succeeds
2. Verify TLS 1.2 connection is rejected
3. Verify invalid certificate is rejected
4. Verify HSTS header presence

## Dependencies
- REQ-COLLAB-001 (sync server exists)

## Effort
0.5 weeks
