# REQ-DIST-009: APT Repository Hosting and Publication

## Metadata
- **Category**: DIST
- **Subcategory**: APT
- **Priority**: HIGH
- **Phase**: 26
- **Status**: MISSING
- **Effort**: 1 week
- **Dependencies**: REQ-DIST-007 (apt-repo.sh script exists), REQ-REL-007 (v1.0.0 tagged)
- **Blocks**: REQ-LAUNCH-002

## Requirement

RTMX shall host a GPG-signed APT repository so that Debian/Ubuntu users can
install via `apt install rtmx` after adding the repository. The repository shall
be published to a stable URL and updated automatically on each release.

## Rationale

REQ-DIST-007 created the apt-repo.sh script for generating repository metadata.
This requirement covers hosting the repository at a public URL, publishing the
GPG public key, and automating updates on release.

## Repository Design

### Structure
```
apt.rtmx.ai/
  dists/
    stable/
      Release
      Release.gpg
      InRelease
      main/
        binary-amd64/Packages
        binary-amd64/Packages.gz
        binary-arm64/Packages
        binary-arm64/Packages.gz
  pool/
    main/
      r/
        rtmx/
          rtmx_1.0.0_amd64.deb
          rtmx_1.0.0_arm64.deb
```

### Hosting Options
1. **GitHub Pages** on a dedicated repo (rtmx-ai/apt) -- simplest
2. **Cloudflare R2** with custom domain (apt.rtmx.ai) -- fastest
3. **GitHub Releases** with apt-repo metadata alongside -- no extra infra

### User Install Flow
```bash
# Add GPG key
curl -fsSL https://apt.rtmx.ai/gpg.key | sudo gpg --dearmor -o /etc/apt/keyrings/rtmx.gpg

# Add repository
echo "deb [signed-by=/etc/apt/keyrings/rtmx.gpg] https://apt.rtmx.ai stable main" \
  | sudo tee /etc/apt/sources.list.d/rtmx.list

# Install
sudo apt update && sudo apt install rtmx
```

## Acceptance Criteria

1. APT repository hosted at a stable public URL
2. GPG public key published and downloadable
3. Repository signed with the org GPG key (93FA984F65C38A73)
4. `apt install rtmx` works on a clean Debian/Ubuntu machine
5. Release workflow automatically publishes new .deb to the repository
6. Install instructions documented in README

## Verification

Test validates repository structure, GPG signing, and metadata generation.
End-to-end install verified on clean container.

## Files to Create/Modify

- `.github/workflows/release.yml` -- add apt repo publish step
- `scripts/apt-repo.sh` -- already exists, may need updates
- `README.md` -- add apt install instructions

## Notes

- GoReleaser already produces .deb packages in the release artifacts
- The apt-repo.sh script generates Packages/Release/InRelease files
- Consider GitHub Pages for simplicity (free, no infra to manage)
- arm64 .deb support is important for Linux ARM servers
