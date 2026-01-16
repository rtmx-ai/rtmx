# REQ-SEC-009: FIPS-Validated Crypto Provider

## Requirement
Cryptographic operations shall use a FIPS 140-3 validated provider when FIPS mode enabled.

## Phase
13 (Security/Compliance)

## Rationale
FIPS 140-3 compliance requires not just using approved algorithms, but using implementations that have been independently validated by NIST-accredited labs. This requirement ensures RTMX integrates with validated cryptographic modules and provides proper runtime verification.

## Acceptance Criteria
- [ ] Support OpenSSL 3.x FIPS provider (Certificate #4282 or newer)
- [ ] Support AWS-LC FIPS as alternative (Certificate #4698 or newer)
- [ ] Provider selection via configuration (`fips_provider: openssl` or `fips_provider: aws-lc`)
- [ ] Validation certificate number logged on startup in FIPS mode
- [ ] Self-tests run on initialization per FIPS requirements
- [ ] Provider module integrity verification before use
- [ ] Graceful fallback with clear error if no validated provider available
- [ ] Documentation of provider installation and configuration

## Supported FIPS Providers

| Provider | Certificate | Validation Level | Platforms |
|----------|-------------|------------------|-----------|
| OpenSSL 3.0 FIPS Provider | #4282 | Level 1 | Linux, Windows, macOS |
| AWS-LC FIPS | #4698 | Level 1 | Linux (AL2, Ubuntu), Windows |
| BoringCrypto (Go) | #4407 | Level 1 | Linux only |

## Technical Notes
- OpenSSL FIPS provider requires separate `fips.so` module
- Provider configuration via `openssl.cnf` or runtime API
- AWS-LC requires linking against FIPS-validated build
- Self-test execution adds ~50ms startup overhead
- Provider module hash verification prevents tampering
- Certificate numbers should be verified against NIST CMVP database

## Provider Detection Logic

```python
# Pseudocode for provider initialization
def init_fips_provider(config):
    provider = config.security.fips_provider

    if provider == "openssl":
        # Check OpenSSL FIPS provider availability
        # Verify module integrity
        # Run self-tests
        # Log certificate number
    elif provider == "aws-lc":
        # Check AWS-LC FIPS availability
        # Verify FIPS mode active
        # Log certificate number
    else:
        raise FIPSConfigError(f"Unknown provider: {provider}")
```

## Configuration Example

```yaml
# rtmx.yaml
security:
  fips_mode: true
  fips_provider: openssl
  fips_provider_path: /usr/lib64/ossl-modules/fips.so  # Optional, auto-detected
```

## Installation Requirements

### OpenSSL FIPS Provider (Linux)
```bash
# Ubuntu/Debian
apt install openssl libssl3

# RHEL/Fedora
dnf install openssl openssl-fips-provider

# Verify FIPS provider
openssl list -providers | grep fips
```

### AWS-LC FIPS
```bash
# Install AWS-LC FIPS from AWS repositories
# or build from source with FIPS flag
cmake -DFIPS=1 ../aws-lc
```

## Test Cases
1. Verify OpenSSL FIPS provider detected and loaded
2. Verify AWS-LC FIPS provider detected and loaded
3. Verify self-tests execute on startup
4. Verify certificate number logged correctly
5. Verify error when invalid provider specified
6. Verify error when provider not installed
7. Verify module integrity check fails on tampered provider
8. Verify crypto operations use validated provider

## Dependencies
- REQ-SEC-008 (FIPS mode infrastructure)

## Effort
3.0 weeks
