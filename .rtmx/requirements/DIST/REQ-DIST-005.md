# REQ-DIST-005: Mainstream Package Repository Submission

## Metadata
- **Category**: DIST
- **Subcategory**: Mainstream
- **Priority**: MEDIUM
- **Phase**: 19
- **Status**: MISSING
- **Effort**: 3 weeks
- **Dependencies**: REQ-GO-047 (v1.0.0 release gate), REQ-MIG-002 (Trunk migration)

## Requirement

RTMX shall be submitted to mainstream package repositories so that users can install via their platform's native package manager without custom taps, manual downloads, or third-party repositories.

## Rationale

Currently `brew install rtmx` does not work -- users must use `brew install rtmx-ai/tap/rtmx`. Similarly, Debian/RPM users must download packages from GitHub releases manually. Mainstream repository presence reduces adoption friction to a single command.

## Target Repositories

| Platform | Repository | Install command | Submission process |
|----------|-----------|-----------------|-------------------|
| macOS/Linux | homebrew-core | `brew install rtmx` | PR to Homebrew/homebrew-core with formula |
| Arch Linux | AUR | `yay -S rtmx` | Create PKGBUILD, submit to aur.archlinux.org |
| Fedora/RHEL | COPR | `sudo dnf install rtmx` | Create COPR project, upload spec |
| Ubuntu/Debian | PPA or apt repo | `sudo apt install rtmx` | Launchpad PPA or self-hosted apt repo |
| Windows | Chocolatey | `choco install rtmx` | Submit package to community.chocolatey.org |
| Linux | Snapcraft | `snap install rtmx` | Create snapcraft.yaml, publish to Snap Store |

## Acceptance Criteria

1. `brew install rtmx` installs from homebrew-core (no custom tap required)
2. At least one Linux-native repository submission (AUR, COPR, or PPA)
3. At least one Windows-native repository submission (Chocolatey or winget)
4. All submissions pass upstream review and are publicly available
5. CI validates that submitted package metadata stays current with releases

## Notes

- homebrew-core requires the project to be "notable" (sufficient GitHub stars, users)
- AUR is community-maintained and has the lowest submission barrier
- COPR and PPA can be self-hosted without upstream approval
- Chocolatey requires package moderation (~1 week review)
- Custom tap (rtmx-ai/homebrew-tap) and Scoop bucket remain as fallbacks
