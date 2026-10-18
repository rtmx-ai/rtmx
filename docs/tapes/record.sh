#!/usr/bin/env bash
# Record VHS demos inside Docker so GIFs never leak host paths.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
IMAGE="${RTMX_VHS_IMAGE:-rtmx-vhs-demos:local}"
OUT_DIR="$ROOT/docs/assets"
SITE_DEMO="${RTMX_SITE_DEMO:-}"

cd "$ROOT"
echo "Building Linux bin/rtmx for the recording image ..."
# VHS image is Linux; host macOS binaries will Exec-format-error.
mkdir -p bin
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/rtmx-linux-amd64 ./cmd/rtmx
cp bin/rtmx-linux-amd64 bin/rtmx
echo "Building recording image $IMAGE ..."
docker build -f docs/tapes/Dockerfile -t "$IMAGE" .

echo "Recording tapes ..."
docker run --rm \
  -v "$OUT_DIR:/workspace/docs/assets" \
  "$IMAGE"

echo "GIFs written to $OUT_DIR"
ls -la "$OUT_DIR"/rtmx-*.gif

if [[ -n "$SITE_DEMO" ]]; then
  echo "Concatenating product demo MP4 → $SITE_DEMO"
  mkdir -p "$(dirname "$SITE_DEMO")"
  # GIF → MP4 segments, then concat (no host PWD in frames)
  tmp="$(mktemp -d)"
  ffmpeg -y -i "$OUT_DIR/rtmx-workflow.gif" -movflags faststart -pix_fmt yuv420p -vf "scale=trunc(iw/2)*2:trunc(ih/2)*2" "$tmp/a.mp4"
  ffmpeg -y -i "$OUT_DIR/rtmx-mcp-setup.gif" -movflags faststart -pix_fmt yuv420p -vf "scale=trunc(iw/2)*2:trunc(ih/2)*2" "$tmp/b.mp4"
  printf "file '%s'\nfile '%s'\n" "$tmp/a.mp4" "$tmp/b.mp4" > "$tmp/list.txt"
  ffmpeg -y -f concat -safe 0 -i "$tmp/list.txt" -c copy "$SITE_DEMO"
  rm -rf "$tmp"
  ls -la "$SITE_DEMO"
fi
