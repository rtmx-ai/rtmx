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

CONFIG="${1:?Usage: run-benchmark.sh <config.yaml> <results_dir>}"
RESULTS_DIR="${2:?Usage: run-benchmark.sh <config.yaml> <results_dir>}"

# Parse config fields using grep/sed (no yq dependency)
get_field() {
    grep "^${1}:" "$CONFIG" | head -1 | sed "s/^${1}:[[:space:]]*//"
}

get_nested_field() {
    local parent="$1"
    local child="$2"
    awk "/^${parent}:/,/^[^ ]/" "$CONFIG" | grep "^  ${child}:" | head -1 | sed "s/^  ${child}:[[:space:]]*//"
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

WORKDIR="workdir/${LANGUAGE}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BENCH_DIR="$(dirname "$SCRIPT_DIR")"

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
SETUP_LINES=$(awk '/^setup_commands:/,/^[^ ]/' "$CONFIG" | grep '^ *- ' | sed 's/^ *- //')
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
if SCAN_OUTPUT=$(cd "$WORKDIR" && eval "$SCAN_COMMAND" 2>&1); then
    MARKER_COUNT=$(echo "$SCAN_OUTPUT" | grep -c "REQ-" || true)
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
