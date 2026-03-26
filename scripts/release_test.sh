#!/usr/bin/env bash
# Release test script for REQ-GO-043
# Tests GoReleaser snapshot builds

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

echo "=== GoReleaser Release Test ==="
echo ""

# Check goreleaser is available
if ! command -v goreleaser &> /dev/null; then
    echo "ERROR: goreleaser not found in PATH"
    echo "Install: go install github.com/goreleaser/goreleaser/v2@latest"
    exit 1
fi

echo "1. Checking GoReleaser version..."
goreleaser --version

echo ""
echo "2. Validating configuration..."
goreleaser check

echo ""
echo "3. Running snapshot build..."
goreleaser build --snapshot --clean

echo ""
echo "4. Verifying build artifacts..."
EXPECTED_PLATFORMS=(
    "linux_amd64"
    "linux_arm64"
    "darwin_amd64"
    "darwin_arm64"
    "windows_amd64"
    "windows_arm64"
)

ARTIFACTS_FOUND=0
for platform in "${EXPECTED_PLATFORMS[@]}"; do
    # Find matching directory (version suffix varies like _v1 or _v8.0)
    MATCH=""
    for dir in dist/rtmx_${platform}*; do
        if [[ -d "$dir" ]]; then
            MATCH="$dir"
            break
        fi
    done

    if [[ -n "$MATCH" ]]; then
        if [[ "$platform" == windows_* ]]; then
            BINARY="$MATCH/rtmx.exe"
        else
            BINARY="$MATCH/rtmx"
        fi

        if [[ -f "$BINARY" ]]; then
            SIZE=$(stat -c%s "$BINARY" 2>/dev/null || stat -f%z "$BINARY" 2>/dev/null)
            SIZE_MB=$((SIZE / 1024 / 1024))
            echo "  ✓ ${platform}: $(basename "$MATCH") (${SIZE_MB}MB)"
            ARTIFACTS_FOUND=$((ARTIFACTS_FOUND + 1))
        else
            echo "  ✗ ${platform}: binary not found at $BINARY"
        fi
    else
        echo "  ✗ ${platform}: directory not found"
    fi
done

echo ""
echo "5. Testing native binary..."
# Find native binary
NATIVE_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
NATIVE_ARCH=$(uname -m)
case "$NATIVE_ARCH" in
    x86_64) NATIVE_ARCH="amd64" ;;
    aarch64) NATIVE_ARCH="arm64" ;;
esac

NATIVE_DIR=$(find dist -maxdepth 1 -type d -name "rtmx_${NATIVE_OS}_${NATIVE_ARCH}*" 2>/dev/null | head -1)
if [[ -n "$NATIVE_DIR" && -x "$NATIVE_DIR/rtmx" ]]; then
    echo "  Running: $NATIVE_DIR/rtmx version"
    "$NATIVE_DIR/rtmx" version

    # Verify it's statically linked (Linux only)
    if [[ "$NATIVE_OS" == "linux" ]]; then
        echo ""
        echo "6. Verifying static linking..."
        LDD_OUTPUT=$(ldd "$NATIVE_DIR/rtmx" 2>&1 || true)
        if echo "$LDD_OUTPUT" | grep -q "not a dynamic executable"; then
            echo "  ✓ Binary is statically linked"
        elif echo "$LDD_OUTPUT" | grep -q "statically linked"; then
            echo "  ✓ Binary is statically linked"
        else
            echo "  ✗ Binary has dynamic dependencies:"
            echo "$LDD_OUTPUT"
            exit 1
        fi
    fi
else
    echo "  ⚠ Native binary not found, skipping execution test"
fi

echo ""
echo "=== Summary ==="
echo "Artifacts built: $ARTIFACTS_FOUND / ${#EXPECTED_PLATFORMS[@]}"

if [[ $ARTIFACTS_FOUND -eq ${#EXPECTED_PLATFORMS[@]} ]]; then
    echo "✓ All platform builds successful"
    exit 0
else
    echo "✗ Some builds missing"
    exit 1
fi
