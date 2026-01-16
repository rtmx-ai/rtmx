# REQ-COLLAB-001: CRDT sync server shall use y-websocket protocol

## Status: MISSING
## Priority: HIGH
## Phase: 10

## Description
CRDT sync server shall use y-websocket protocol

## Acceptance Criteria
- [ ] Server accepts WebSocket connections on configurable port
- [ ] y-websocket protocol compatible with pycrdt/Yjs clients
- [ ] Server broadcasts CRDT updates to all connected clients
- [ ] Connection uses TLS 1.3 transport (REQ-SEC-001)
- [ ] CRDT payloads encrypted end-to-end before transmission (REQ-SEC-002)

## Test Cases
- `tests/test_collab.py::test_sync_server`
- `tests/test_collab.py::test_sync_server_tls`
- `tests/test_collab.py::test_sync_payload_encrypted`

## Security Integration

The sync server is the foundation for collaborative features and must integrate with the security layer:

### Transport Security (REQ-SEC-001)
- All WebSocket connections require TLS 1.3 minimum
- When `crypto.fips_mode: true`, use FIPS-validated TLS implementation
- Hybrid TLS with ML-KEM-768 supported when `crypto.profile: pqc-hybrid`

### Payload Encryption (REQ-SEC-002)
- CRDT updates are encrypted client-side before transmission
- Server relays only ciphertext; cannot read requirement content
- Encryption algorithm follows active crypto profile (classic, pqc-hybrid, fips, pqc-only)

### Authentication (REQ-SEC-003, REQ-COLLAB-007)
- WebSocket connections authenticated via OAuth/JWT tokens
- Token validation on connection establishment
- Connection rejected if token expired or invalid

## Notes
WebSocket server broadcasting CRDT updates to clients. This is the central hub for real-time collaboration, enabling multiple users and agents to work on the same RTM database simultaneously.

## Dependencies
- REQ-CRDT-001: CRDT layer exists
- REQ-WEB-007: WebSocket infrastructure
