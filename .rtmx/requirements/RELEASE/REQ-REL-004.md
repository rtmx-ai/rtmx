# REQ-REL-004: Homebrew and Scoop Package Publishing

## Metadata
- **Category**: RELEASE
- **Subcategory**: Distribution
- **Priority**: HIGH
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-044, REQ-DIST-001
- **Blocks**: REQ-GO-047

## Requirement

Release workflow shall publish to Homebrew tap and Scoop bucket automatically on tag push, enabling `brew install` and `scoop install` to work.

## Rationale

README documents `brew install rtmx-ai/tap/rtmx` and `scoop install rtmx` but neither works because tokens aren't configured and publishing is commented out in `.goreleaser.yaml`. This is misleading documentation for a released product.

## Design

### Infrastructure Setup

1. Create `rtmx-ai/homebrew-tap` repository (if not exists)
2. Create `rtmx-ai/scoop-bucket` repository (if not exists)
3. Generate fine-grained PATs with repo write access for each
4. Configure as GitHub Actions secrets:
   - `HOMEBREW_TAP_TOKEN`
   - `SCOOP_BUCKET_TOKEN`

### GoReleaser Configuration (uncomment and verify)

```yaml
brews:
  - name: rtmx
    repository:
      owner: rtmx-ai
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
    directory: Formula
    homepage: "https://rtmx.ai"
    description: "Requirements Traceability Matrix toolkit"
    license: "Apache-2.0"
    test: |
      system "#{bin}/rtmx", "version"
    install: |
      bin.install "rtmx"

scoops:
  - repository:
      owner: rtmx-ai
      name: scoop-bucket
      token: "{{ .Env.SCOOP_BUCKET_TOKEN }}"
    homepage: "https://rtmx.ai"
    description: "Requirements Traceability Matrix toolkit"
    license: "Apache-2.0"
```

## Acceptance Criteria

1. `brew install rtmx-ai/tap/rtmx` installs working binary
2. `scoop bucket add rtmx https://github.com/rtmx-ai/scoop-bucket && scoop install rtmx` works
3. Formula/manifest auto-updated on each release tag
4. `rtmx version` shows correct version after install
5. README installation instructions are accurate

## Files to Modify

- `.goreleaser.yaml` - Uncomment brews/scoops sections
- `.github/workflows/release.yml` - Verify secrets are used
- `README.md` - Verify installation instructions match reality
