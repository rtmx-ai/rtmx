# Security Policy

## Supported Versions

| Version | Supported          |
|---------|--------------------|
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

## Reporting a Vulnerability

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, email **security@rtmx.ai** with:

- Description of the vulnerability
- Steps to reproduce
- Expected vs actual behavior
- Impact assessment (if known)

We will acknowledge receipt within **48 hours** and provide a fix timeline within **7 business days** for critical issues.

## Verification

All release binaries are GPG-signed. Verify your download:

```bash
# Import RTMX public key
curl -fsSL https://rtmx.ai/gpg.key | gpg --import

# Verify checksums signature
gpg --verify checksums.txt.sig checksums.txt

# Verify binary checksum
sha256sum -c <(grep linux_amd64 checksums.txt)
```

## Security Practices

- All dependencies are audited with `govulncheck` in CI
- CodeQL analysis runs on every push
- Binaries are statically compiled (CGO_ENABLED=0) with no external runtime dependencies
- Release artifacts include SBOM (Software Bill of Materials) in SPDX format

## Contact

- Security issues: security@rtmx.ai
- General support: dev@rtmx.ai
- Company: ioTACTICAL LLC
