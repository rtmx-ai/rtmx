# REQ-REL-006: Security Policy and Install Script

## Metadata
- **Category**: RELEASE
- **Subcategory**: Security
- **Priority**: HIGH
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-REL-001

## Requirement

Repository shall include SECURITY.md for vulnerability disclosure and a universal install script at `https://rtmx.ai/install.sh` that installs the correct binary for any platform.

## Design

### SECURITY.md

```markdown
# Security Policy

## Supported Versions
| Version | Supported |
|---------|-----------|
| 0.1.x   | ✅        |

## Reporting a Vulnerability
Email security@rtmx.ai with:
- Description
- Steps to reproduce
- Expected vs actual behavior

We will acknowledge within 48 hours and provide a fix within 7 days for critical issues.

## Verification
All binaries are GPG-signed. Verify with:
gpg --verify checksums.txt.sig checksums.txt
```

### Install Script

```bash
#!/bin/sh
# https://rtmx.ai/install.sh
set -e
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in x86_64) ARCH=amd64;; aarch64|arm64) ARCH=arm64;; esac
VERSION=$(curl -s https://api.github.com/repos/rtmx-ai/rtmx-go/releases/latest | grep tag_name | cut -d'"' -f4)
URL="https://github.com/rtmx-ai/rtmx-go/releases/download/${VERSION}/rtmx_${VERSION#v}_${OS}_${ARCH}.tar.gz"
curl -fsSL "$URL" | tar xz -C /usr/local/bin rtmx
echo "rtmx ${VERSION} installed to /usr/local/bin/rtmx"
```

Usage: `curl -fsSL https://rtmx.ai/install.sh | sh`

## Acceptance Criteria

1. SECURITY.md exists in repository root
2. Install script works on Linux (amd64, arm64) and macOS (amd64, arm64)
3. Install script verifies checksums
4. Install script provides clear error on unsupported platforms

## Files to Create

- `SECURITY.md` - Vulnerability disclosure policy
- `scripts/install.sh` - Universal install script
