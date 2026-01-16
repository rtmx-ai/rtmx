# REQ-COLLAB-005: All CRDT operations shall be logged for audit

## Status: MISSING
## Priority: MEDIUM
## Phase: 10

## Description
All CRDT operations shall be logged for audit

## Acceptance Criteria
- [ ] Audit log exists and is append-only
- [ ] Each CRDT operation logged with timestamp, user, and operation type
- [ ] Log entries include affected requirement IDs
- [ ] Log integrity protected via hash chain
- [ ] When `crypto.sign_audit_log: true`, entries are cryptographically signed

## Test Cases
- `tests/test_collab.py::test_audit_log`
- `tests/test_collab.py::test_audit_log_hash_chain`
- `tests/test_collab.py::test_audit_log_signed`

## Security Integration

This requirement provides the basic audit logging that REQ-SEC-005 extends with immutable storage and compliance features.

### Hash Chain Integrity
Each audit entry includes hash of previous entry, creating tamper-evident log:
```json
{
  "id": "uuid",
  "timestamp": "2026-01-14T12:00:00Z",
  "user_id": "user-123",
  "operation": "update",
  "requirement_ids": ["REQ-001"],
  "prev_hash": "sha256:abc123...",
  "hash": "sha256:def456..."
}
```

### Cryptographic Signatures (Crypto-Agility)
When signatures enabled, algorithm follows active crypto profile (REQ-SEC-002):
- `classic`: Ed25519 signatures
- `pqc-hybrid`: Ed25519 + ML-DSA-65 dual signatures
- `fips`: ECDSA P-384 signatures
- `pqc-only`: ML-DSA-65 signatures

Signed entries provide:
- Non-repudiation of user actions
- Third-party verifiable audit trail
- Quantum-resistant integrity when PQC enabled

## Notes
Append-only log of who changed what and when. This is the foundation for compliance, debugging, and accountability in collaborative environments.

## Dependencies
- REQ-COLLAB-001: Sync server exists to capture operations
- REQ-SEC-002: Crypto-agility layer for signature algorithms
