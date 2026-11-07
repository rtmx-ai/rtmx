#!/usr/bin/env bash
# Build GitHub Release notes from CHANGELOG.md + verify/install header.
# Usage: scripts/github-release-notes.sh [VERSION]
# VERSION defaults to GITHUB_REF_NAME with a leading v stripped (e.g. v2.0.0 → 2.0.0).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VERSION="${1:-${GITHUB_REF_NAME:-}}"
VERSION="${VERSION#v}"
if [[ -z "$VERSION" ]]; then
  echo "usage: $0 VERSION" >&2
  exit 1
fi

CHANGELOG="$ROOT/CHANGELOG.md"
if [[ ! -f "$CHANGELOG" ]]; then
  echo "missing $CHANGELOG" >&2
  exit 1
fi

# Extract "## [X.Y.Z] - date" section through the next ## heading (exclusive).
section="$(awk -v ver="$VERSION" '
  $0 ~ "^## \\[" ver "\\]" {grab=1; print; next}
  grab && $0 ~ /^## / {exit}
  grab {print}
' "$CHANGELOG")"

if [[ -z "$section" ]]; then
  echo "CHANGELOG.md has no section for [$VERSION]" >&2
  exit 1
fi

cat <<EOF
## Verify Your Download

\`\`\`bash
# Import RTMX public key
curl -fsSL https://rtmx.ai/gpg.key | gpg --import

# Verify checksums
gpg --verify checksums.txt.sig checksums.txt
sha256sum -c <(grep linux_amd64 checksums.txt)
\`\`\`

## Install

\`\`\`bash
# Quick install (Linux/macOS)
curl -fsSL https://rtmx.ai/install.sh | sh

# Homebrew
brew install rtmx-ai/tap/rtmx

# Scoop (Windows)
scoop bucket add rtmx https://github.com/rtmx-ai/scoop-bucket
scoop install rtmx
\`\`\`

$section

Full history: [CHANGELOG.md](https://github.com/rtmx-ai/rtmx/blob/main/CHANGELOG.md).
EOF
