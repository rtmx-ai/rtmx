#!/usr/bin/env bash
# run-benchmark.sh -- Clone an exemplar project, apply rtmx markers, scan, and verify.
#
# Usage: run-benchmark.sh <config.yaml> <results_dir>
#
# Reads a benchmark config YAML and executes the full pipeline:
#   1. Clone exemplar at pinned ref (shallow)
#   2. Apply marker patch
#   3. Run rtmx from-tests to extract markers
#   4. Run verify command to execute tests
#   5. Write results JSON

set -euo pipefail

# REQ-BENCH-010: Diagnostic-on-exit. Print script:line:command:exit_code on
# any error so CI logs always contain an actionable diagnostic.
trap 'echo "ERROR: ${BASH_SOURCE[0]}:${LINENO}: command \"${BASH_COMMAND}\" exited with status $?" >&2' ERR

CONFIG="${1:?Usage: run-benchmark.sh <config.yaml> <results_dir>}"
RESULTS_DIR="${2:?Usage: run-benchmark.sh <config.yaml> <results_dir>}"

# Parse config fields using grep/sed (no yq dependency).
# Returns empty string (not error) when field is absent -- callers use
# validate_required to catch missing fields with actionable messages.
get_field() {
    grep "^${1}:" "$CONFIG" 2>/dev/null | head -1 | sed "s/^${1}:[[:space:]]*//" || true
}

get_nested_field() {
    local parent="$1"
    local child="$2"
    # The previous awk range /^parent:/,/^[^ ]/ terminated on the parent line
    # itself because ^[^ ] matches any line starting with a non-space character,
    # including the parent line. This caused all nested field extraction to
    # silently return empty strings, breaking every benchmark.
    # Fix: use an explicit state machine that skips the parent line, collects
    # indented children, and stops at the next top-level key.
    awk -v p="$parent" '
        $0 ~ "^"p":" { found=1; next }
        found && /^[^ ]/ { exit }
        found { print }
    ' "$CONFIG" | grep "^  ${child}:" 2>/dev/null | head -1 | sed "s/^  ${child}:[[:space:]]*//" || true
}

LANGUAGE=$(get_field "language")
REPO=$(get_nested_field "exemplar" "repo")
REF=$(get_nested_field "exemplar" "ref")
CLONE_DEPTH=$(get_field "clone_depth")
MARKER_PATCH=$(get_field "marker_patch")
EXPECTED_MARKERS=$(get_field "expected_markers")
SCAN_COMMAND=$(get_field "scan_command")
VERIFY_COMMAND=$(get_field "verify_command")
TIMEOUT_MINUTES=$(get_field "timeout_minutes")

CLONE_DEPTH="${CLONE_DEPTH:-1}"
TIMEOUT_MINUTES="${TIMEOUT_MINUTES:-10}"

# REQ-BENCH-013: Validate required fields before any execution.
# Missing fields cause exit 2 with an actionable message naming the config
# and the field, instead of cascading into confusing downstream failures.
validate_required() {
    local field_name="$1"
    local field_value="$2"
    if [ -z "$field_value" ]; then
        echo "ERROR: ${CONFIG} missing required field ${field_name}" >&2
        exit 2
    fi
}
validate_required "language" "$LANGUAGE"
validate_required "exemplar.repo" "$REPO"
validate_required "exemplar.ref" "$REF"
validate_required "expected_markers" "$EXPECTED_MARKERS"
validate_required "scan_command" "$SCAN_COMMAND"

WORKDIR="workdir/${LANGUAGE}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BENCH_DIR="$(dirname "$SCRIPT_DIR")"

# Resolve the rtmx binary. make build puts it at bin/rtmx relative to repo root.
RTMX_BIN="${BENCH_DIR}/../bin/rtmx"
if [ ! -x "$RTMX_BIN" ]; then
    echo "ERROR: rtmx binary not found at ${RTMX_BIN}" >&2
    echo "  Run 'make build' first." >&2
    exit 2
fi
RTMX_BIN="$(cd "$(dirname "$RTMX_BIN")" && pwd)/$(basename "$RTMX_BIN")"

# Replace bare 'rtmx' in scan command with the absolute binary path so it
# resolves correctly when run from within the cloned exemplar workdir.
SCAN_COMMAND="${SCAN_COMMAND/rtmx/${RTMX_BIN}}"

echo "Benchmark: ${LANGUAGE}"
echo "  Exemplar: ${REPO} @ ${REF}"
echo "  Expected markers: ${EXPECTED_MARKERS}"

# Step 1: Clone at pinned ref
rm -rf "$WORKDIR"
mkdir -p "$WORKDIR"
echo "  Cloning ${REPO}..."
git clone --depth "${CLONE_DEPTH}" --branch "${REF}" "https://github.com/${REPO}.git" "$WORKDIR" 2>/dev/null || \
git clone --depth "${CLONE_DEPTH}" "https://github.com/${REPO}.git" "$WORKDIR" 2>/dev/null && \
    git -C "$WORKDIR" checkout "${REF}" 2>/dev/null

# Step 2: Apply marker patch (if specified and exists)
if [ -n "${MARKER_PATCH}" ] && [ -f "${BENCH_DIR}/${MARKER_PATCH}" ]; then
    echo "  Applying marker patch..."
    git -C "$WORKDIR" apply "${BENCH_DIR}/${MARKER_PATCH}"
fi

# Step 3: Run setup commands (if any in config)
# Setup commands are parsed line by line from the setup_commands block
SETUP_LINES=$(awk '$0 ~ /^setup_commands:/ { found=1; next } found && /^[^ ]/ { exit } found { print }' "$CONFIG" | grep '^ *- ' | sed 's/^ *- //' || true)
if [ -n "$SETUP_LINES" ]; then
    echo "  Running setup commands..."
    while IFS= read -r cmd; do
        echo "    $ ${cmd}"
        (cd "$WORKDIR" && eval "$cmd")
    done <<< "$SETUP_LINES"
fi

# Step 4: Scan for markers
echo "  Scanning for markers..."
MARKER_COUNT=0
SCAN_EXIT=0
SCAN_OUTPUT=$(cd "$WORKDIR" && eval "$SCAN_COMMAND" 2>&1) || SCAN_EXIT=$?
if [ "$SCAN_EXIT" -eq 0 ]; then
    MARKER_COUNT=$(echo "$SCAN_OUTPUT" | grep -c "REQ-" || true)
else
    echo "  WARNING: scan command failed (exit ${SCAN_EXIT})" >&2
    echo "  Command: ${SCAN_COMMAND}" >&2
    echo "  Output: ${SCAN_OUTPUT:-(empty)}" >&2
fi

echo "  Markers found: ${MARKER_COUNT} (expected: ${EXPECTED_MARKERS})"

# Step 5: Run verify command
TESTS_PASSED=0
TESTS_FAILED=0
VERIFY_STATUS="skip"
if [ -n "${VERIFY_COMMAND}" ]; then
    echo "  Running verify command..."
    if timeout "${TIMEOUT_MINUTES}m" bash -c "cd '$WORKDIR' && $VERIFY_COMMAND" >/dev/null 2>&1; then
        VERIFY_STATUS="pass"
    else
        VERIFY_STATUS="fail"
    fi
fi

# Step 6: Write results
mkdir -p "$RESULTS_DIR"
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
cat > "${RESULTS_DIR}/${LANGUAGE}.json" <<RESULT_EOF
{
  "language": "${LANGUAGE}",
  "marker_count": ${EXPECTED_MARKERS},
  "markers_found": ${MARKER_COUNT},
  "tests_passed": ${TESTS_PASSED},
  "tests_failed": ${TESTS_FAILED},
  "verify_status": "${VERIFY_STATUS}",
  "timestamp": "${TIMESTAMP}"
}
RESULT_EOF

echo "  Results written to ${RESULTS_DIR}/${LANGUAGE}.json"
echo "  Done."
