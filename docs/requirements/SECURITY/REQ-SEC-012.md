# REQ-SEC-012: ML-DSA Digital Signatures

## Requirement
RTMX shall support ML-DSA (FIPS 204) for quantum-resistant digital signatures.

## Phase
13 (Security/Compliance) - Post-Quantum Cryptography

## Rationale
ML-DSA (Module-Lattice Digital Signature Algorithm), standardized as FIPS 204, provides quantum-resistant digital signatures. Traditional signature algorithms (RSA, ECDSA, Ed25519) will become vulnerable to quantum attacks. ML-DSA is one of the primary algorithms selected by NIST for post-quantum cryptography and is required for CNSA 2.0 compliance.

## Acceptance Criteria
- [ ] ML-DSA-65 supported (default for most use cases)
- [ ] ML-DSA-87 supported (CNSA 2.0 requirement)
- [ ] Key generation API available
- [ ] Sign API available
- [ ] Verify API available
- [ ] Compatible with liboqs-python implementation
- [ ] Compatible with OpenSSL 3.5+ implementation (when available)
- [ ] Audit log signatures use ML-DSA when PQC mode enabled
- [ ] Algorithm selection configurable via rtmx.yaml

## Technical Notes
- Use `liboqs-python` (Open Quantum Safe project) as primary implementation
- Fall back to OpenSSL 3.5+ when available and configured
- ML-DSA-65: 1952 byte public key, 3309 byte signature, NIST Level 3 security
- ML-DSA-87: 2592 byte public key, 4627 byte signature, NIST Level 5 security
- Signatures are larger than traditional algorithms; consider storage implications
- Signing is faster than RSA; verification is comparable to ECDSA

## Security Levels

| Variant | Security Level | Public Key | Private Key | Signature |
|---------|---------------|------------|-------------|-----------|
| ML-DSA-65 | NIST Level 3 | 1952 bytes | 4032 bytes | 3309 bytes |
| ML-DSA-87 | NIST Level 5 | 2592 bytes | 4896 bytes | 4627 bytes |

## Configuration
```yaml
crypto:
  pqc:
    enabled: true
    signature:
      algorithm: ml-dsa-65  # or ml-dsa-87
      provider: liboqs      # or openssl (when available)
    audit_signatures: true  # Sign audit logs with ML-DSA
```

## Test Cases
1. ML-DSA-65 key generation produces valid key pair
2. ML-DSA-87 key generation produces valid key pair
3. Sign operation produces valid signature
4. Verify operation validates authentic signatures
5. Verify operation rejects tampered signatures
6. Verify operation rejects wrong public key
7. Audit log entries signed when audit_signatures enabled
8. Algorithm selection respects configuration
9. Interoperability with reference implementation (liboqs)

## Dependencies
- REQ-SEC-010 (cryptographic agility framework)

## Effort
2.0 weeks
