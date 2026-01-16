# REQ-SEC-008: FIPS 140-3 Mode

## Requirement
RTMX shall support FIPS 140-3 compliant operation mode.

## Phase
13 (Security/Compliance)

## Rationale
Federal agencies and defense contractors are required to use FIPS 140-3 validated cryptography for sensitive data. RTMX must support a FIPS mode that ensures all cryptographic operations use approved algorithms from validated providers, enabling adoption in regulated government environments.

## Acceptance Criteria
- [ ] FIPS mode can be enabled via configuration (`fips_mode: true` in rtmx.yaml)
- [ ] Only FIPS-approved algorithms used when enabled (AES, SHA-2/SHA-3, RSA, ECDSA)
- [ ] Runtime validation of FIPS provider availability on startup
- [ ] Clear error messages when FIPS requirements not met
- [ ] Documentation of FIPS deployment requirements
- [ ] Non-FIPS algorithms rejected with actionable error when FIPS enabled
- [ ] FIPS mode status displayed in `rtmx config show`

## FIPS 140-3 Approved Algorithms

| Category | Approved | Not Approved |
|----------|----------|--------------|
| Symmetric Encryption | AES-128/192/256 | ChaCha20, Blowfish, 3DES |
| Hash Functions | SHA-256, SHA-384, SHA-512, SHA-3 | MD5, SHA-1 |
| Key Exchange | RSA (2048+), ECDH (P-256/384/521) | X25519, Curve25519 |
| Digital Signatures | RSA-PSS, ECDSA | Ed25519, Ed448 |
| Key Derivation | HKDF, PBKDF2 | Argon2, scrypt, bcrypt |
| Random Number | SP 800-90A DRBG | System urandom (unless FIPS-backed) |

## Technical Notes
- Use OpenSSL 3.x FIPS provider as primary option
- AWS-LC FIPS as alternative for AWS environments
- Configure via environment variable `RTMX_FIPS_MODE=1` or config file
- Validate provider availability before any crypto operations
- Provider detection at import time with lazy error on first use
- Log warning if FIPS mode requested but provider unavailable

## Configuration Example

```yaml
# rtmx.yaml
security:
  fips_mode: true
  fips_provider: openssl  # or aws-lc
```

## Test Cases
1. Verify FIPS mode disables non-approved algorithms
2. Verify clear error when FIPS provider unavailable
3. Verify AES-256-GCM works in FIPS mode
4. Verify ChaCha20-Poly1305 rejected in FIPS mode
5. Verify X25519 key exchange rejected in FIPS mode
6. Verify config show displays FIPS status
7. Verify environment variable overrides config file

## Dependencies
- REQ-SEC-002 (E2E encryption exists)

## Effort
2.0 weeks
