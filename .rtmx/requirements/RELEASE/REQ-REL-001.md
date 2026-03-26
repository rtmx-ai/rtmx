# REQ-REL-001: GPG Binary Signing

## Metadata
- **Category**: RELEASE
- **Subcategory**: Security
- **Priority**: P0
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-043
- **Blocks**: REQ-GO-047

## Requirement

All release binaries and checksums shall be GPG-signed, enabling users to verify binary authenticity.

## Rationale

Without binary signing, users cannot distinguish official releases from tampered artifacts. This is a hard requirement for enterprise adoption and regulated industries (defense, healthcare, finance).

## Design

### GoReleaser Configuration

```yaml
signs:
  - artifacts: checksum
    args:
      - "--batch"
      - "--local-user"
      - "{{ .Env.GPG_FINGERPRINT }}"
      - "--output"
      - "${signature}"
      - "--detach-sign"
      - "${artifact}"
```

### Infrastructure

1. Generate GPG key for RTMX releases: `dev@rtmx.ai`
2. Publish public key at `https://rtmx.ai/gpg.key`
3. Store private key as GitHub Actions secret: `GPG_PRIVATE_KEY`
4. Store fingerprint as secret: `GPG_FINGERPRINT`

### Release Artifacts (after)

```
checksums.txt              # SHA256 checksums
checksums.txt.sig          # GPG detached signature
rtmx_0.2.0_linux_amd64.tar.gz
rtmx_0.2.0_linux_amd64.tar.gz.sig
...
```

### User Verification

```bash
# Import RTMX public key
curl -fsSL https://rtmx.ai/gpg.key | gpg --import

# Verify checksums
gpg --verify checksums.txt.sig checksums.txt

# Verify specific binary
sha256sum -c <(grep linux_amd64 checksums.txt)
```

## Acceptance Criteria

1. All release archives have `.sig` GPG detached signature files
2. `checksums.txt` has a `.sig` signature
3. GPG public key published at known URL
4. Verification instructions in release notes
5. CI release workflow passes with signing enabled

## Files to Modify

- `.goreleaser.yaml` - Add `signs:` section
- `.github/workflows/release.yml` - Add GPG key import step
- `README.md` - Add verification instructions
