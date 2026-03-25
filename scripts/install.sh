#!/bin/sh
# RTMX CLI Installer
# Usage: curl -fsSL https://rtmx.ai/install.sh | sh
#
# Environment variables:
#   RTMX_VERSION  - Specific version to install (default: latest)
#   RTMX_INSTALL  - Installation directory (default: /usr/local/bin)

set -e

REPO="rtmx-ai/rtmx-go"
INSTALL_DIR="${RTMX_INSTALL:-/usr/local/bin}"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux)  OS="linux" ;;
  darwin) OS="darwin" ;;
  *)      echo "Error: Unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64)   ARCH="amd64" ;;
  aarch64|arm64)   ARCH="arm64" ;;
  *)               echo "Error: Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Determine version
if [ -z "$RTMX_VERSION" ]; then
  RTMX_VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
  if [ -z "$RTMX_VERSION" ]; then
    echo "Error: Could not determine latest version" >&2
    exit 1
  fi
fi

VERSION_NUM="${RTMX_VERSION#v}"
ARCHIVE="rtmx-go_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${RTMX_VERSION}/${ARCHIVE}"
CHECKSUM_URL="https://github.com/${REPO}/releases/download/${RTMX_VERSION}/checksums.txt"

echo "Installing rtmx ${RTMX_VERSION} for ${OS}/${ARCH}..."

# Create temp directory
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

# Download binary and checksums
echo "Downloading ${URL}..."
curl -fsSL "$URL" -o "${TMPDIR}/${ARCHIVE}"
curl -fsSL "$CHECKSUM_URL" -o "${TMPDIR}/checksums.txt"

# Verify checksum
echo "Verifying checksum..."
cd "$TMPDIR"
EXPECTED=$(grep "$ARCHIVE" checksums.txt | awk '{print $1}')
ACTUAL=$(sha256sum "$ARCHIVE" | awk '{print $1}')
if [ "$EXPECTED" != "$ACTUAL" ]; then
  echo "Error: Checksum mismatch!" >&2
  echo "  Expected: $EXPECTED" >&2
  echo "  Actual:   $ACTUAL" >&2
  exit 1
fi
echo "Checksum verified."

# Extract and install
tar xzf "$ARCHIVE"
if [ -w "$INSTALL_DIR" ]; then
  mv rtmx "$INSTALL_DIR/rtmx"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv rtmx "$INSTALL_DIR/rtmx"
fi

echo ""
echo "rtmx ${RTMX_VERSION} installed to ${INSTALL_DIR}/rtmx"
echo ""
echo "Verify installation:"
echo "  rtmx version"
echo ""
echo "Get started:"
echo "  rtmx init        # Initialize new project"
echo "  rtmx setup       # Full project setup"
echo "  rtmx status      # Check RTM status"

# Detect existing Python rtmx installation and print migration notice
PYTHON_RTMX=""
if command -v pip >/dev/null 2>&1 && pip show rtmx >/dev/null 2>&1; then
  PYTHON_RTMX="pip"
elif command -v pip3 >/dev/null 2>&1 && pip3 show rtmx >/dev/null 2>&1; then
  PYTHON_RTMX="pip3"
elif command -v pipx >/dev/null 2>&1 && pipx list 2>/dev/null | grep -q rtmx; then
  PYTHON_RTMX="pipx"
fi

if [ -n "$PYTHON_RTMX" ]; then
  echo ""
  echo "=========================================="
  echo "  Python rtmx CLI detected (via ${PYTHON_RTMX})"
  echo "=========================================="
  echo ""
  echo "The Python rtmx CLI is deprecated and will reach end-of-life on 2026-09-25."
  echo "You have just installed the Go replacement. To complete the migration:"
  echo ""
  echo "  1. Verify the Go CLI works: rtmx status"
  echo "  2. Remove the Python CLI:   ${PYTHON_RTMX} uninstall rtmx"
  echo ""
  echo "Migration guide: https://github.com/rtmx-ai/rtmx-go#migrating-from-python-cli"
  echo ""
fi
