# REQ-DIST-002: APT/DEB Repository (Linux)

## Metadata
- **Category**: DIST
- **Subcategory**: Linux
- **Priority**: HIGH
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-043

## Requirement

RTMX shall be installable on Debian/Ubuntu Linux via APT package manager.

## Rationale

APT is the standard package manager for Debian-based distributions (Ubuntu, Debian, Linux Mint, Pop!_OS), representing a large share of Linux developer workstations and CI environments.

## Design

### Installation

```bash
# Add RTMX repository
curl -fsSL https://rtmx.ai/gpg.key | sudo gpg --dearmor -o /usr/share/keyrings/rtmx-archive-keyring.gpg
echo "deb [signed-by=/usr/share/keyrings/rtmx-archive-keyring.gpg] https://apt.rtmx.ai stable main" | sudo tee /etc/apt/sources.list.d/rtmx.list

# Install
sudo apt update
sudo apt install rtmx

# Update
sudo apt update && sudo apt upgrade rtmx
```

### Alternative: Install Script

```bash
# One-liner for quick install
curl -fsSL https://rtmx.ai/install.sh | bash

# Script handles:
# 1. Detect platform (deb vs rpm vs binary)
# 2. Add appropriate repository
# 3. Install package
```

## Infrastructure Required

1. APT repository hosting (Cloudflare R2, S3, or packagecloud.io)
2. GPG key for package signing
3. CI workflow to publish .deb on release
4. apt.rtmx.ai subdomain or CDN path

### Repository Structure

```
apt.rtmx.ai/
├── dists/
│   └── stable/
│       └── main/
│           ├── binary-amd64/
│           │   ├── Packages
│           │   ├── Packages.gz
│           │   └── Release
│           └── binary-arm64/
│               └── ...
├── pool/
│   └── main/
│       └── r/
│           └── rtmx/
│               ├── rtmx_0.1.0_amd64.deb
│               └── rtmx_0.1.0_arm64.deb
└── gpg.key
```

## Acceptance Criteria

1. APT repository accessible at apt.rtmx.ai (or equivalent)
2. `apt install rtmx` installs working binary
3. `rtmx version` shows correct version
4. `apt upgrade rtmx` updates to new versions
5. Both amd64 and arm64 architectures supported
6. Package signed with GPG key

## Test Strategy

- CI job to verify .deb package integrity
- Docker-based installation test (ubuntu:latest)
- Manual testing on Ubuntu LTS

## References

- Debian repository format
- GoReleaser nfpm (already produces .deb)
- packagecloud.io for hosted APT repos
