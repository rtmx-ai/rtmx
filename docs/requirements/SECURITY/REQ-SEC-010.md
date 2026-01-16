# REQ-SEC-010: Crypto Agility Architecture

## Requirement
RTMX shall implement a crypto-agile architecture enabling algorithm migration without code changes.

## Phase
13 (Security/Compliance)

## Priority
HIGH

## Rationale
Cryptographic agility is essential for long-term security. As quantum computing advances threaten classical cryptography, organizations must be able to migrate to post-quantum cryptography (PQC) algorithms without rewriting application code. Additionally, different deployment environments may require different algorithms (e.g., FIPS-140 compliance, export restrictions, performance constraints).

A crypto-agile architecture decouples encryption logic from specific algorithms, allowing:
- Transparent migration from classical to hybrid to post-quantum algorithms
- Compliance with evolving standards (NIST PQC, FIPS-140-3)
- Decryption of legacy data encrypted with older algorithms
- Environment-specific algorithm selection via configuration

## Acceptance Criteria
- [ ] CryptoProvider protocol/interface defined with clear abstraction
- [ ] Algorithm selection via configuration, not code
- [ ] Support for multiple concurrent providers (classical, hybrid, PQC)
- [ ] Provider can be changed without data migration
- [ ] Encryption metadata includes algorithm identifier
- [ ] Graceful fallback when preferred algorithm unavailable
- [ ] Key derivation abstracted through provider interface
- [ ] Algorithm negotiation for multi-party scenarios

## CryptoProvider Protocol
```python
from typing import Protocol, runtime_checkable

@runtime_checkable
class CryptoProvider(Protocol):
    """Abstract cryptographic provider interface."""

    @property
    def algorithm_id(self) -> str:
        """Unique identifier for this algorithm (stored with ciphertext)."""
        ...

    @property
    def name(self) -> str:
        """Human-readable name for logging and configuration."""
        ...

    def encrypt(self, plaintext: bytes, key: bytes) -> bytes:
        """Encrypt data, returns ciphertext with embedded algorithm ID."""
        ...

    def decrypt(self, ciphertext: bytes, key: bytes) -> bytes:
        """Decrypt data, validates algorithm ID matches."""
        ...

    def derive_key(self, password: str, salt: bytes) -> bytes:
        """Derive encryption key from password."""
        ...

    def generate_key(self) -> bytes:
        """Generate a new random encryption key."""
        ...

    def is_available(self) -> bool:
        """Check if required cryptographic libraries are available."""
        ...
```

## Provider Implementations

| Provider | Algorithm ID | Description | Use Case |
|----------|--------------|-------------|----------|
| `ClassicalProvider` | `aes-256-gcm` | AES-256-GCM with Argon2id KDF | Default, widest compatibility |
| `HybridProvider` | `x25519-kyber768` | X25519 + Kyber768 hybrid | Transition to PQC |
| `PQCProvider` | `ml-kem-1024` | ML-KEM (NIST standardized) | Future-proof encryption |
| `FIPSProvider` | `aes-256-gcm-fips` | FIPS 140-3 validated | Government/regulated use |

## Configuration
```yaml
crypto:
  # Default provider for new encryption operations
  default_provider: classical

  # Fallback order if preferred provider unavailable
  fallback_order:
    - hybrid
    - classical

  # Provider-specific configuration
  providers:
    classical:
      cipher: aes-256-gcm
      kdf: argon2id
      kdf_params:
        memory_cost: 65536
        time_cost: 3
        parallelism: 4

    hybrid:
      classical_cipher: aes-256-gcm
      kem: x25519-kyber768
      kdf: hkdf-sha384

    pqc:
      kem: ml-kem-1024
      cipher: aes-256-gcm
      kdf: hkdf-sha384

    fips:
      cipher: aes-256-gcm
      kdf: pbkdf2-sha256
      # FIPS mode uses system OpenSSL in FIPS mode
      fips_mode: true

  # Algorithm negotiation for sync scenarios
  negotiation:
    # Minimum acceptable security level
    minimum_algorithm: classical
    # Preferred algorithms (in order)
    preferred: [pqc, hybrid, classical]
```

## Ciphertext Envelope Format
```json
{
  "version": 1,
  "algorithm_id": "aes-256-gcm",
  "kdf_id": "argon2id",
  "salt": "base64-encoded-salt",
  "nonce": "base64-encoded-nonce",
  "ciphertext": "base64-encoded-ciphertext",
  "created_at": "2025-01-04T22:45:00Z"
}
```

## Technical Notes
- Define abstract `CryptoProvider` protocol using Python's `typing.Protocol`
- Implementations: `ClassicalProvider`, `HybridProvider`, `PQCProvider`, `FIPSProvider`
- Store algorithm ID with ciphertext for future decryption
- Enable "encrypt with new, decrypt with any" pattern for seamless migration
- Use `cryptography` library for classical algorithms
- Use `pqcrypto` or `liboqs-python` for post-quantum algorithms
- Provider registry enables runtime discovery of available algorithms
- Consider using envelope encryption for large payloads (encrypt data key, not data)

## Migration Strategy

1. **Phase 1**: Deploy crypto-agile architecture with classical-only provider
2. **Phase 2**: Add hybrid provider, continue encrypting with classical
3. **Phase 3**: Switch default to hybrid for new encryption
4. **Phase 4**: Add PQC provider when standards finalize
5. **Phase 5**: Switch default to PQC, maintain hybrid/classical for decryption

All phases maintain backward compatibility - old data remains decryptable.

## Test Cases
1. Classical provider encrypts and decrypts correctly
2. Hybrid provider encrypts and decrypts correctly
3. PQC provider encrypts and decrypts correctly
4. Cross-provider decryption works (encrypt with old, decrypt with new)
5. Algorithm ID is correctly embedded in ciphertext
6. Unknown algorithm ID raises clear error
7. Fallback activates when preferred provider unavailable
8. Configuration changes algorithm without code changes
9. Provider availability detection works correctly
10. Key derivation produces consistent results across providers

## Security Considerations
- Never log or expose plaintext or keys
- Use constant-time comparison for authentication tags
- Validate algorithm ID before attempting decryption
- Reject unknown or deprecated algorithms with clear errors
- Implement key rotation without service interruption

## Dependencies
- REQ-SEC-002 (E2E encryption - provides the encryption context)

## Effort
2.5 weeks
