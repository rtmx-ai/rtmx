# REQ-SEC-011: ML-KEM Key Encapsulation

## Requirement
RTMX shall support ML-KEM (FIPS 203) for quantum-resistant key encapsulation.

## Phase
13 (Security/Compliance) - Post-Quantum Cryptography

## Rationale
ML-KEM (Module-Lattice Key Encapsulation Mechanism), standardized as FIPS 203, provides quantum-resistant key encapsulation. As quantum computers advance, traditional key exchange mechanisms (RSA, ECDH) will become vulnerable. ML-KEM is one of the primary algorithms selected by NIST for post-quantum cryptography and is required for CNSA 2.0 compliance starting January 1, 2027.

## Acceptance Criteria
- [ ] ML-KEM-768 supported (default for most use cases)
- [ ] ML-KEM-1024 supported (CNSA 2.0 requirement)
- [ ] Key generation API available
- [ ] Encapsulation API available
- [ ] Decapsulation API available
- [ ] Compatible with liboqs-python implementation
- [ ] Compatible with OpenSSL 3.5+ implementation (when available)
- [ ] Algorithm selection configurable via rtmx.yaml

## Technical Notes
- Use `liboqs-python` (Open Quantum Safe project) as primary implementation
- Fall back to OpenSSL 3.5+ when available and configured
- ML-KEM-768: 2400 byte public key, 1088 byte ciphertext, NIST Level 3 security
- ML-KEM-1024: 3168 byte public key, 1568 byte ciphertext, NIST Level 5 security
- Key generation is computationally intensive; cache and reuse where appropriate
- Store key material securely in user's keychain/credential store

## Security Levels

| Variant | Security Level | Public Key | Ciphertext | Shared Secret |
|---------|---------------|------------|------------|---------------|
| ML-KEM-768 | NIST Level 3 | 2400 bytes | 1088 bytes | 32 bytes |
| ML-KEM-1024 | NIST Level 5 | 3168 bytes | 1568 bytes | 32 bytes |

## Configuration
```yaml
crypto:
  pqc:
    enabled: true
    kem:
      algorithm: ml-kem-768  # or ml-kem-1024
      provider: liboqs       # or openssl (when available)
```

## Test Cases
1. ML-KEM-768 key generation produces valid key pair
2. ML-KEM-1024 key generation produces valid key pair
3. Encapsulation produces valid ciphertext
4. Decapsulation recovers shared secret
5. Invalid ciphertext is rejected
6. Algorithm selection respects configuration
7. Interoperability with reference implementation (liboqs)

## Dependencies
- REQ-SEC-010 (cryptographic agility framework)

## Effort
2.0 weeks
