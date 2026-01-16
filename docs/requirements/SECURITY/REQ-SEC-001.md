# REQ-SEC-001: TLS 1.3 Transport Security

## Requirement
All sync connections shall use TLS 1.3.

## Phase
10 (Collaboration) - Sync Server Security

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

## Crypto-Agility Integration

This requirement establishes baseline transport security. Future crypto-agility extensions include:

- **FIPS Mode**: When `crypto.fips_mode: true` in config, enforce FIPS 140-3 validated TLS implementations (e.g., AWS-LC FIPS, BoringCrypto)
- **Hybrid TLS**: As post-quantum cryptography matures, support hybrid key exchange combining X25519 with ML-KEM-768 (NIST PQC standard). This provides quantum-resistant key exchange while maintaining classical security as fallback
- **Algorithm Configuration**: Transport cipher suites should be configurable via `crypto.tls.cipher_suites` to allow organizations to enforce specific algorithm requirements

The crypto-agility architecture (REQ-CRYPTO-*) will extend this requirement with:
- Runtime algorithm selection based on `CryptoProvider` configuration
- Automatic fallback when PQC algorithms unavailable
- Compliance logging for algorithm usage auditing

## Test Cases
1. Verify TLS 1.3 connection succeeds
2. Verify TLS 1.2 connection is rejected
3. Verify invalid certificate is rejected
4. Verify HSTS header presence

## Dependencies
- REQ-COLLAB-001 (sync server exists)

## Effort
0.5 weeks
