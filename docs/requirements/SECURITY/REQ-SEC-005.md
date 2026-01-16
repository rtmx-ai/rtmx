# REQ-SEC-005: Immutable Audit Log

## Requirement
All sync operations shall be logged immutably.

## Phase
10 (Collaboration) - Enterprise Security

## Rationale
Audit logs are essential for compliance (SOC 2, ISO 27001), incident investigation, and accountability. An immutable log provides tamper-evident records of all changes to requirements.

## Acceptance Criteria
- [ ] All CRDT operations logged with timestamp
- [ ] User identity recorded for each operation
- [ ] Operation type recorded (create, update, delete)
- [ ] Affected requirement IDs recorded
- [ ] Log entries cannot be modified or deleted
- [ ] Log integrity verifiable via hash chain
- [ ] Logs exportable in standard format (JSON, CSV)
- [ ] Log retention configurable (default 7 years)

## Log Entry Schema
```json
{
  "id": "uuid",
  "timestamp": "2025-01-04T22:45:00Z",
  "user_id": "user-123",
  "user_email": "alice@example.com",
  "operation": "update",
  "project_id": "proj-456",
  "requirement_ids": ["REQ-001", "REQ-002"],
  "ip_address": "192.168.1.1",
  "user_agent": "rtmx-cli/0.1.0",
  "prev_hash": "sha256:abc123...",
  "hash": "sha256:def456..."
}
```

## Technical Notes
- Use append-only storage (PostgreSQL with row-level security, or dedicated audit service)
- Hash chain: each entry includes hash of previous entry
- Consider write-once storage (S3 Object Lock, Azure Immutable Blob)
- Separate audit log from operational database

## Crypto-Agility Integration

The audit log integrity mechanism must integrate with the crypto-agility architecture (REQ-SEC-002):

### Hash Chain Algorithm Selection
The hash chain algorithm is selected based on the active crypto profile:

| Profile | Hash Algorithm | Signature Algorithm |
|---------|---------------|---------------------|
| `classic` | SHA-256 | Ed25519 (optional) |
| `pqc-hybrid` | SHA-256 | Ed25519 + ML-DSA-65 |
| `fips` | SHA-256 (FIPS validated) | ECDSA P-384 |
| `pqc-only` | SHA-256 | ML-DSA-65 |

### Signed Audit Entries
When `crypto.sign_audit_log: true`:
- Each audit entry includes a cryptographic signature
- Signature algorithm follows the active crypto profile
- ML-DSA signatures provide quantum-resistant audit integrity
- Enables third-party verification of log authenticity

### Entry Schema Extension
```json
{
  "hash": "sha256:def456...",
  "signature": "base64-encoded-signature",
  "signature_algorithm": "ml-dsa-65",  // or ed25519, ecdsa-p384
  "signer_key_id": "audit-signing-key-001"
}
```

### Algorithm Migration
When crypto profile changes:
- New entries use new algorithm
- Old entries remain verifiable with original algorithm
- Migration log entry records algorithm change event

## Test Cases
1. Create operation is logged
2. Update operation is logged
3. Delete operation is logged
4. Log entries cannot be modified via API
5. Hash chain validates correctly
6. Export produces valid JSON/CSV

## Dependencies
- REQ-COLLAB-005 (basic audit logging)

## Effort
1.5 weeks
