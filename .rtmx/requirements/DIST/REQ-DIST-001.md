# REQ-DIST-001: Scoop Package (Windows)

## Metadata
- **Category**: DIST
- **Subcategory**: Windows
- **Priority**: HIGH
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-043

## Requirement

RTMX shall be installable on Windows via Scoop package manager.

## Rationale

Scoop is the preferred package manager for developer tools on Windows, providing easy installation and updates without administrator privileges.

## Design

### Installation

```powershell
# Add RTMX bucket
scoop bucket add rtmx https://github.com/rtmx-ai/scoop-bucket

# Install
scoop install rtmx

# Update
scoop update rtmx
```

### Manifest

```json
{
    "version": "0.1.0",
    "description": "Requirements Traceability Matrix toolkit",
    "homepage": "https://rtmx.ai",
    "license": "Apache-2.0",
    "architecture": {
        "64bit": {
            "url": "https://github.com/rtmx-ai/rtmx/releases/download/v0.1.0/rtmx_0.1.0_windows_amd64.zip",
            "hash": "sha256:..."
        },
        "arm64": {
            "url": "https://github.com/rtmx-ai/rtmx/releases/download/v0.1.0/rtmx_0.1.0_windows_arm64.zip",
            "hash": "sha256:..."
        }
    },
    "bin": "rtmx.exe",
    "checkver": "github",
    "autoupdate": {
        "architecture": {
            "64bit": {
                "url": "https://github.com/rtmx-ai/rtmx/releases/download/v$version/rtmx_$version_windows_amd64.zip"
            },
            "arm64": {
                "url": "https://github.com/rtmx-ai/rtmx/releases/download/v$version/rtmx_$version_windows_arm64.zip"
            }
        }
    }
}
```

## Infrastructure Required

1. Create `rtmx-ai/scoop-bucket` repository
2. Configure `SCOOP_BUCKET_TOKEN` secret in rtmx-go
3. Enable Scoop publisher in .goreleaser.yaml

## Acceptance Criteria

1. `scoop bucket add rtmx ...` succeeds
2. `scoop install rtmx` installs working binary
3. `rtmx version` shows correct version
4. `scoop update rtmx` updates to new versions
5. Both amd64 and arm64 architectures supported

## Test Strategy

- Manual testing on Windows
- CI job to verify manifest syntax
- Automated installation test in GitHub Actions

## References

- Scoop bucket documentation
- GoReleaser Scoop publisher
- REQ-GO-043 GoReleaser configuration
