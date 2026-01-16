# REQ-SEC-014: CNSA 2.0 Compliance Profile

## Requirement
RTMX shall provide a CNSA 2.0 compliance profile for National Security Systems.

## Phase
13 (Security/Compliance) - Post-Quantum Cryptography

## Rationale
The NSA's Commercial National Security Algorithm Suite 2.0 (CNSA 2.0) defines mandatory cryptographic algorithms for protecting National Security Systems (NSS). Starting January 1, 2027, all DoD contracts and NSS deployments must use CNSA 2.0 compliant algorithms. RTMX must provide a strict compliance mode to serve defense contractors and government customers.

## Acceptance Criteria
- [ ] AES-256 for symmetric encryption
- [ ] ML-KEM-1024 for key establishment (NIST Level 5)
- [ ] ML-DSA-87 for digital signatures (NIST Level 5)
- [ ] SHA-384 minimum for hashing (SHA-512 preferred)
- [ ] LMS/XMSS for firmware signing (if applicable)
- [ ] Compliance mode enforces all CNSA 2.0 algorithms
- [ ] Non-compliant algorithm usage blocked in strict mode
- [ ] Compliance validation command available
- [ ] Compliance report generation for audits

## Technical Notes
- CNSA 2.0 Timeline:
  - 2025: Prefer CNSA 2.0 algorithms
  - 2027: Required for new systems
  - 2030: Required for all systems
  - 2033: Legacy algorithm support ends
- Strict mode rejects any non-CNSA 2.0 algorithm selection
- Advisory mode warns but allows non-compliant algorithms
- Consider FIPS 140-3 validated implementations for government use
- Document which algorithm providers are FIPS validated

## CNSA 2.0 Algorithm Suite

| Function | Required Algorithm | RTMX Support |
|----------|-------------------|--------------|
| Symmetric Encryption | AES-256 | REQ-SEC-002 |
| Key Establishment | ML-KEM-1024 | REQ-SEC-011 |
| Digital Signatures | ML-DSA-87 | REQ-SEC-012 |
| Hashing | SHA-384/SHA-512 | Built-in |
| Firmware Signing | LMS/XMSS | Future |

## Configuration
```yaml
crypto:
  compliance:
    profile: cnsa-2.0  # or fips-140-3, commercial
    mode: strict       # strict, advisory, disabled

  # CNSA 2.0 strict mode overrides these settings:
  pqc:
    enabled: true
    mode: pqc-only     # No hybrid mode for CNSA 2.0
    kem:
      algorithm: ml-kem-1024  # Required for CNSA 2.0
    signature:
      algorithm: ml-dsa-87    # Required for CNSA 2.0

  symmetric:
    algorithm: aes-256-gcm  # Required for CNSA 2.0

  hash:
    algorithm: sha-512      # SHA-384 minimum for CNSA 2.0
```

## Compliance Validation
```bash
# Check compliance status
rtmx security compliance-check --profile cnsa-2.0

# Generate compliance report
rtmx security compliance-report --format json > compliance.json

# Enable strict mode
rtmx config set crypto.compliance.mode strict
```

## Test Cases
1. CNSA 2.0 profile enforces ML-KEM-1024
2. CNSA 2.0 profile enforces ML-DSA-87
3. CNSA 2.0 profile enforces AES-256
4. CNSA 2.0 profile enforces SHA-384 minimum
5. Strict mode rejects ML-KEM-768 selection
6. Strict mode rejects ML-DSA-65 selection
7. Strict mode rejects AES-128 selection
8. Strict mode rejects SHA-256 selection
9. Advisory mode warns on non-compliant selection
10. Compliance check command reports violations
11. Compliance report includes all algorithm usage

## Dependencies
- REQ-SEC-011 (ML-KEM key encapsulation)
- REQ-SEC-012 (ML-DSA digital signatures)
- REQ-SEC-008 (symmetric encryption - AES-256)

## Notes
- Required for DoD contracts starting January 1, 2027
- Defense contractors should enable advisory mode now to prepare
- Consider partnering with FIPS-validated crypto providers

## Effort
2.0 weeks
