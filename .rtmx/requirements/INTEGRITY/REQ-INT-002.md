# REQ-INT-002: CRDT Proof-of-Verification

## Metadata
- **Category**: INTEGRITY
- **Subcategory**: Decentralized
- **Priority**: HIGH
- **Phase**: 17
- **Status**: MISSING
- **Dependencies**: REQ-INT-001, REQ-GO-042

## Requirement

RTMX Sync shall require cryptographic proof-of-verification for CRDT operations that change requirement status, with invalid operations rejected during merge.

## Rationale

In a decentralized/federated system without central authority, enforcement shifts from "prevent writes" to "reject invalid operations during sync." This preserves the closed-loop verification guarantee across trust boundaries.

## Design

### Proof Payload Structure

```
StatusChangeOp {
  req_id: string
  old_status: Status
  new_status: Status
  timestamp: ISO8601

  proof: {
    test_output_hash: string    // sha256 of test results
    test_timestamp: ISO8601     // when tests were run
    verifier_id: string         // DID or public key fingerprint
    signature: string           // proof signed by verifier
  }
}
```

### CRDT Merge Behavior

1. Operations without proof payload → rejected
2. Operations with invalid signature → rejected
3. Operations from untrusted verifiers → rejected (per trust policy)
4. Valid operations → merged into local state

### Trust Policy Options

| Policy | Description | Use Case |
|--------|-------------|----------|
| `self` | Accept proofs signed by local key only | Single developer |
| `team` | Accept proofs from configured team keys | Small team |
| `delegated` | Accept proofs from org-designated verifiers | Enterprise |
| `web-of-trust` | Accept if N-of-M trusted keys attest | Federation |

## Acceptance Criteria

1. Status change operations include proof payload
2. CRDT merge rejects operations with missing/invalid proofs
3. Trust policy is configurable per repository
4. Proofs are portable (can be verified by any node)
5. Offline-generated proofs are valid when synced later

## Test Strategy

- Unit tests: proof generation and validation
- Integration tests: CRDT merge with valid/invalid proofs
- Adversarial tests: attempt to forge proofs

## Open Questions

1. Key management: how are verifier keys distributed and rotated?
2. Revocation: how do we handle compromised verifier keys?
3. Proof size: impact on sync bandwidth for large databases?
4. Clock skew: how do we handle timestamp validation across nodes?

## References

- Certificate Transparency (RFC 6962)
- Git signed commits
- DID (Decentralized Identifiers) specification
