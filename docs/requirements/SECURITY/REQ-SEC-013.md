# REQ-SEC-013: Hybrid Classical+PQC Mode

## Requirement
RTMX shall support hybrid encryption combining classical and post-quantum algorithms.

## Phase
13 (Security/Compliance) - Post-Quantum Cryptography

## Rationale
During the transition period to post-quantum cryptography, hybrid mode provides defense-in-depth by combining proven classical algorithms with new PQC algorithms. If either algorithm is later found to be weak, the other provides protection. This approach specifically protects against "harvest now, decrypt later" attacks where adversaries collect encrypted data today to decrypt with future quantum computers.

## Acceptance Criteria
- [ ] X25519 + ML-KEM-768 hybrid key exchange supported
- [ ] Ed25519 + ML-DSA-65 hybrid signatures supported
- [ ] Hybrid mode is the default during transition period
- [ ] Both algorithms must succeed for operation to succeed
- [ ] Combined shared secret derived from both KEM outputs
- [ ] Combined signature includes both signature schemes
- [ ] Fallback to classical-only mode configurable
- [ ] Upgrade path to PQC-only mode when standards mature

## Technical Notes
- Key Exchange: Combine X25519 and ML-KEM shared secrets using HKDF
  - `shared_secret = HKDF(x25519_secret || mlkem_secret)`
- Signatures: Concatenate Ed25519 and ML-DSA signatures
  - Both signatures must verify for combined signature to be valid
- NIST recommends hybrid mode during transition (2024-2030)
- Hybrid mode increases key/signature sizes but maintains classical security floor
- Consider TLS 1.3 hybrid key exchange patterns (draft-ietf-tls-hybrid-design)

## Security Properties
- **Classical Protection**: X25519/Ed25519 proven over 10+ years
- **Quantum Protection**: ML-KEM/ML-DSA provide NIST Level 3 security
- **Defense in Depth**: Compromise of one algorithm doesn't break encryption
- **Harvest Protection**: Data encrypted today remains secure against future quantum attacks

## Key/Signature Sizes

| Mode | Public Key | Ciphertext/Signature |
|------|------------|---------------------|
| X25519 only | 32 bytes | 32 bytes |
| ML-KEM-768 only | 2400 bytes | 1088 bytes |
| X25519 + ML-KEM-768 | 2432 bytes | 1120 bytes |
| Ed25519 only | 32 bytes | 64 bytes |
| ML-DSA-65 only | 1952 bytes | 3309 bytes |
| Ed25519 + ML-DSA-65 | 1984 bytes | 3373 bytes |

## Configuration
```yaml
crypto:
  pqc:
    enabled: true
    mode: hybrid  # hybrid (default), classical, pqc-only
    kem:
      classical: x25519
      quantum: ml-kem-768
    signature:
      classical: ed25519
      quantum: ml-dsa-65
```

## Test Cases
1. Hybrid key exchange produces combined shared secret
2. Hybrid signature produces combined signature
3. Hybrid verification requires both signatures valid
4. X25519 failure causes hybrid key exchange to fail
5. ML-KEM failure causes hybrid key exchange to fail
6. Ed25519 failure causes hybrid signature verification to fail
7. ML-DSA failure causes hybrid signature verification to fail
8. Classical-only fallback works when configured
9. PQC-only mode works when configured

## Dependencies
- REQ-SEC-011 (ML-KEM key encapsulation)
- REQ-SEC-012 (ML-DSA digital signatures)

## Effort
1.5 weeks
