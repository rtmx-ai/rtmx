# REQ-DIST-007: APT Repository for Debian/Ubuntu

## Metadata
- **Category**: DIST
- **Subcategory**: APT
- **Priority**: HIGH
- **Phase**: 25
- **Status**: MISSING
- **Effort**: 1 week
- **Dependencies**: REQ-REL-007 (v1.0.0 tagged)
- **Blocks**: REQ-LAUNCH-002

## Requirement

RTMX shall be installable via `apt install rtmx` on Debian and Ubuntu systems
through a hosted APT repository. The repository shall be GPG-signed and updated
automatically on each release via the GitHub Actions release workflow.

## Rationale

GoReleaser already builds `.deb` packages. Currently users must download them
manually from GitHub releases. An APT repository turns this into a one-time
setup and `apt install rtmx` / `apt upgrade rtmx` thereafter.

## Design

### Repository Hosting

Host the APT repository on GitHub Pages at `apt.rtmx.ai` (or a subdirectory
of the main site). Alternatively, use a dedicated GitHub repository
`rtmx-ai/apt` with GitHub Pages enabled.

### Repository Structure

```
apt/
  dists/
    stable/
      main/
        binary-amd64/
          Packages
          Packages.gz
        binary-arm64/
          Packages
          Packages.gz
      Release
      Release.gpg
      InRelease
  pool/
    main/
      r/
        rtmx/
          rtmx_1.0.0_amd64.deb
          rtmx_1.0.0_arm64.deb
```

### User Setup

```bash
# Add GPG key
curl -fsSL https://apt.rtmx.ai/gpg.key | sudo gpg --dearmor -o /usr/share/keyrings/rtmx.gpg

# Add repository
echo "deb [signed-by=/usr/share/keyrings/rtmx.gpg] https://apt.rtmx.ai stable main" | \
  sudo tee /etc/apt/sources.list.d/rtmx.list

# Install
sudo apt update && sudo apt install rtmx
```

### Release Automation

The GitHub Actions release workflow shall:
1. Copy `.deb` artifacts from GoReleaser output
2. Generate `Packages` and `Release` files using `dpkg-scanpackages` and `apt-ftparchive`
3. GPG-sign the Release file
4. Push to the apt repository (GitHub Pages or S3)

## Acceptance Criteria

1. APT repository is accessible at a public URL
2. Repository is GPG-signed with the RTMX release key
3. `apt update` succeeds after adding the repository
4. `apt install rtmx` installs the correct version
5. Both amd64 and arm64 architectures are available
6. Release workflow automatically publishes new `.deb` packages
7. Setup instructions documented in README and rtmx.ai

## Verification Test

Test validates that the repository generation script exists and the
GoReleaser nfpms section produces `.deb` packages. Actual APT repository
hosting is verified by inspection.

## Files to Create

- `scripts/apt-repo.sh` -- Script to generate APT repository structure from .deb files
- `.github/workflows/release.yml` -- Add apt repo publish step (modify existing)

## Notes

- Launchpad PPA is an alternative but requires Ubuntu-specific packaging
- Self-hosted apt repo gives us control over all Debian-family distros
- The same approach can be extended to RPM repos (yum/dnf) later
