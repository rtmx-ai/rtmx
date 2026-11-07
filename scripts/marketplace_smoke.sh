#!/usr/bin/env bash
# Post-release marketplace smoke (REQ-MONO-062r).
# Verifies first-party Claude marketplace manifest lists rtmx and that
# plugin / pack versions align. Does not drive Claude/Cursor GUIs.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MARKETPLACE="${ROOT}/.claude-plugin/marketplace.json"
PACK_VERSION_FILE="${ROOT}/plugins/rtmx/VERSION"
CLAUDE_PLUGIN="${ROOT}/plugins/claude-rtmx/.claude-plugin/plugin.json"
CURSOR_PLUGIN="${ROOT}/plugins/cursor-rtmx/.cursor-plugin/plugin.json"

die() { echo "marketplace_smoke FAIL: $*" >&2; exit 1; }

[[ -f "$MARKETPLACE" ]] || die "missing $MARKETPLACE"
[[ -f "$PACK_VERSION_FILE" ]] || die "missing $PACK_VERSION_FILE"

PACK_VER="$(tr -d '[:space:]' <"$PACK_VERSION_FILE")"
[[ -n "$PACK_VER" ]] || die "empty pack VERSION"

if command -v python3 >/dev/null 2>&1; then
  python3 - "$MARKETPLACE" "$PACK_VER" "$CLAUDE_PLUGIN" "$CURSOR_PLUGIN" <<'PY'
import json, sys
market_path, pack_ver, claude_path, cursor_path = sys.argv[1:5]
m = json.load(open(market_path))
plugins = m.get("plugins") or []
rtmx = next((p for p in plugins if p.get("name") == "rtmx"), None)
if not rtmx:
    raise SystemExit("marketplace.json has no plugin name=rtmx")
mv = str(rtmx.get("version") or "")
if mv != pack_ver:
    raise SystemExit(f"marketplace plugin version {mv!r} != pack VERSION {pack_ver!r}")
for label, path in (("claude", claude_path), ("cursor", cursor_path)):
    p = json.load(open(path))
    if p.get("name") != "rtmx":
        raise SystemExit(f"{label} plugin.json name != rtmx")
    if str(p.get("version") or "") != pack_ver:
        raise SystemExit(f"{label} plugin version {p.get('version')!r} != pack {pack_ver!r}")
print(f"OK marketplace lists rtmx@{pack_ver}; Claude/Cursor plugin.json match")
PY
else
  grep -q '"name": "rtmx"' "$MARKETPLACE" || die "marketplace missing rtmx name"
  grep -q "\"version\": \"$PACK_VER\"" "$MARKETPLACE" || die "marketplace version mismatch $PACK_VER"
  echo "OK (grep fallback) marketplace lists rtmx version $PACK_VER"
fi

# Optional: remote manifest when RTMX_MARKETPLACE_URL is set (post-tag check).
if [[ -n "${RTMX_MARKETPLACE_URL:-}" ]]; then
  tmp="$(mktemp)"
  curl -fsSL "$RTMX_MARKETPLACE_URL" -o "$tmp" || die "curl $RTMX_MARKETPLACE_URL (see docs/AGENT_PLUGIN_RELEASE.md)"
  grep -q '"name": "rtmx"' "$tmp" || die "remote marketplace missing rtmx (see docs/AGENT_PLUGIN_RELEASE.md)"
  echo "OK remote marketplace reachable: $RTMX_MARKETPLACE_URL"
  rm -f "$tmp"
fi

# Optional public listing checks (REQ-MONO-063j). Unset = offline-friendly CI.
# Defaults used only when the env var is set to a non-empty value; set to the
# canonical URLs in post-release jobs, e.g.:
#   RTMX_DIRECTORY_URL=https://cursor.directory/plugins/rtmx
#   RTMX_REGISTRY_URL='https://registry.modelcontextprotocol.io/v0.1/servers?search=io.github.rtmx-ai/rtmx'
if [[ -n "${RTMX_DIRECTORY_URL:-}" ]]; then
  code="$(curl -fsSL -o /dev/null -w '%{http_code}' "$RTMX_DIRECTORY_URL" || true)"
  [[ "$code" == "200" ]] || die "directory listing HTTP $code for $RTMX_DIRECTORY_URL (see docs/AGENT_PLUGIN_RELEASE.md / COMMUNITY_LISTINGS.md)"
  echo "OK directory listing: $RTMX_DIRECTORY_URL ($code)"
fi

if [[ -n "${RTMX_REGISTRY_URL:-}" ]]; then
  tmp="$(mktemp)"
  curl -fsSL "$RTMX_REGISTRY_URL" -o "$tmp" || die "registry curl failed: $RTMX_REGISTRY_URL (see docs/MCP_REGISTRY.md / AGENT_PLUGIN_RELEASE.md)"
  # Accept either our server name or a non-empty JSON body (registry preview lag).
  if ! grep -q 'io.github.rtmx-ai/rtmx\|"servers"' "$tmp"; then
    die "registry response missing RTMX server hint (see docs/MCP_REGISTRY.md)"
  fi
  echo "OK MCP registry query: $RTMX_REGISTRY_URL"
  rm -f "$tmp"
fi

echo "marketplace_smoke PASS"
echo "On failure, see docs/AGENT_PLUGIN_RELEASE.md (REQ-MONO-062o / 063j checklist)."
