#!/usr/bin/env bash
# Generate an APT repository structure from .deb packages.
#
# Usage:
#   scripts/apt-repo.sh <deb-dir> <output-dir> <gpg-key-id>
#
# Example:
#   scripts/apt-repo.sh dist/ apt-repo/ 93FA984F65C38A73
#
# This script:
#   1. Creates the apt repository directory structure
#   2. Copies .deb files into pool/
#   3. Generates Packages and Release files
#   4. GPG-signs the Release file (Release.gpg and InRelease)
#
# Prerequisites:
#   - dpkg-scanpackages (from dpkg-dev)
#   - apt-ftparchive (from apt-utils) -- optional, falls back to manual
#   - gpg with the signing key available

set -euo pipefail

DEB_DIR="${1:?Usage: $0 <deb-dir> <output-dir> <gpg-key-id>}"
OUTPUT_DIR="${2:?Usage: $0 <deb-dir> <output-dir> <gpg-key-id>}"
GPG_KEY="${3:?Usage: $0 <deb-dir> <output-dir> <gpg-key-id>}"

CODENAME="stable"
COMPONENT="main"

echo "==> Building APT repository"
echo "    Source:  ${DEB_DIR}"
echo "    Output:  ${OUTPUT_DIR}"
echo "    Key:     ${GPG_KEY}"

# Create directory structure
mkdir -p "${OUTPUT_DIR}/pool/${COMPONENT}/r/rtmx"
mkdir -p "${OUTPUT_DIR}/dists/${CODENAME}/${COMPONENT}/binary-amd64"
mkdir -p "${OUTPUT_DIR}/dists/${CODENAME}/${COMPONENT}/binary-arm64"

# Copy .deb files to pool
echo "==> Copying .deb packages to pool"
cp "${DEB_DIR}"/*.deb "${OUTPUT_DIR}/pool/${COMPONENT}/r/rtmx/" 2>/dev/null || {
    echo "ERROR: No .deb files found in ${DEB_DIR}"
    exit 1
}

# Generate Packages files
echo "==> Generating Packages index"
cd "${OUTPUT_DIR}"

for arch in amd64 arm64; do
    dpkg-scanpackages --arch "${arch}" "pool/${COMPONENT}" > \
        "dists/${CODENAME}/${COMPONENT}/binary-${arch}/Packages"
    gzip -9c "dists/${CODENAME}/${COMPONENT}/binary-${arch}/Packages" > \
        "dists/${CODENAME}/${COMPONENT}/binary-${arch}/Packages.gz"
    echo "    ${arch}: $(grep -c '^Package:' "dists/${CODENAME}/${COMPONENT}/binary-${arch}/Packages" 2>/dev/null || echo 0) package(s)"
done

# Generate Release file
echo "==> Generating Release file"
cat > "dists/${CODENAME}/Release" <<EOF
Origin: RTMX
Label: RTMX
Suite: ${CODENAME}
Codename: ${CODENAME}
Architectures: amd64 arm64
Components: ${COMPONENT}
Description: RTMX CLI APT Repository
Date: $(date -u '+%a, %d %b %Y %H:%M:%S UTC')
EOF

# Add checksums to Release
if command -v apt-ftparchive &>/dev/null; then
    apt-ftparchive release "dists/${CODENAME}" >> "dists/${CODENAME}/Release"
else
    # Manual checksum generation
    echo "MD5Sum:" >> "dists/${CODENAME}/Release"
    for f in $(find "dists/${CODENAME}/${COMPONENT}" -type f); do
        rel="${f#dists/${CODENAME}/}"
        size=$(wc -c < "$f" | tr -d ' ')
        md5=$(md5sum "$f" | cut -d' ' -f1)
        printf " %s %s %s\n" "$md5" "$size" "$rel" >> "dists/${CODENAME}/Release"
    done
    echo "SHA256:" >> "dists/${CODENAME}/Release"
    for f in $(find "dists/${CODENAME}/${COMPONENT}" -type f); do
        rel="${f#dists/${CODENAME}/}"
        size=$(wc -c < "$f" | tr -d ' ')
        sha=$(sha256sum "$f" | cut -d' ' -f1)
        printf " %s %s %s\n" "$sha" "$size" "$rel" >> "dists/${CODENAME}/Release"
    done
fi

# GPG sign
echo "==> Signing Release file"
gpg --batch --yes --local-user "${GPG_KEY}" \
    --detach-sign --armor --output "dists/${CODENAME}/Release.gpg" \
    "dists/${CODENAME}/Release"

gpg --batch --yes --local-user "${GPG_KEY}" \
    --clearsign --output "dists/${CODENAME}/InRelease" \
    "dists/${CODENAME}/Release"

echo ""
echo "==> APT repository generated at ${OUTPUT_DIR}"
echo ""
echo "User setup:"
echo "  curl -fsSL https://apt.rtmx.ai/gpg.key | sudo gpg --dearmor -o /usr/share/keyrings/rtmx.gpg"
echo "  echo 'deb [signed-by=/usr/share/keyrings/rtmx.gpg] https://apt.rtmx.ai stable main' | \\"
echo "    sudo tee /etc/apt/sources.list.d/rtmx.list"
echo "  sudo apt update && sudo apt install rtmx"
