#!/usr/bin/env bash
# report.sh -- Compare benchmark results to baseline and flag regressions.
#
# Usage: report.sh <current.json> <baseline.json>
#
# Exits 0 if no regressions, 1 if regressions detected.

set -euo pipefail

# REQ-BENCH-010: Diagnostic-on-exit.
trap 'echo "ERROR: ${BASH_SOURCE[0]}:${LINENO}: command \"${BASH_COMMAND}\" exited with status $?" >&2' ERR

CURRENT="${1:?Usage: report.sh <current.json> <baseline.json>}"
BASELINE="${2:?Usage: report.sh <current.json> <baseline.json>}"

if [ ! -f "$CURRENT" ]; then
    echo "ERROR: Current results not found: $CURRENT"
    exit 1
fi

if [ ! -f "$BASELINE" ]; then
    echo "WARNING: No baseline found at $BASELINE -- skipping comparison"
    echo "To create a baseline, copy the current results:"
    echo "  cp $CURRENT $BASELINE"
    exit 0
fi

# Requires jq
if ! command -v jq &>/dev/null; then
    echo "ERROR: jq is required for result comparison"
    exit 1
fi

LANG=$(jq -r '.language' "$CURRENT")
echo "Comparing ${LANG} benchmark results..."

REGRESSIONS=0

# Check marker count
BASELINE_MARKERS=$(jq '.marker_count' "$BASELINE")
CURRENT_MARKERS=$(jq '.markers_found' "$CURRENT")
if [ "$CURRENT_MARKERS" -lt "$BASELINE_MARKERS" ]; then
    echo "  REGRESSION: marker count dropped from ${BASELINE_MARKERS} to ${CURRENT_MARKERS}"
    REGRESSIONS=$((REGRESSIONS + 1))
else
    echo "  OK: markers ${CURRENT_MARKERS} >= baseline ${BASELINE_MARKERS}"
fi

# Check test failures
BASELINE_FAILURES=$(jq '.tests_failed' "$BASELINE")
CURRENT_FAILURES=$(jq '.tests_failed' "$CURRENT")
if [ "$CURRENT_FAILURES" -gt "$BASELINE_FAILURES" ]; then
    echo "  REGRESSION: test failures increased from ${BASELINE_FAILURES} to ${CURRENT_FAILURES}"
    REGRESSIONS=$((REGRESSIONS + 1))
else
    echo "  OK: test failures ${CURRENT_FAILURES} <= baseline ${BASELINE_FAILURES}"
fi

# Check verify status
BASELINE_STATUS=$(jq -r '.verify_status' "$BASELINE")
CURRENT_STATUS=$(jq -r '.verify_status' "$CURRENT")
if [ "$BASELINE_STATUS" = "pass" ] && [ "$CURRENT_STATUS" = "fail" ]; then
    echo "  REGRESSION: verify status changed from ${BASELINE_STATUS} to ${CURRENT_STATUS}"
    REGRESSIONS=$((REGRESSIONS + 1))
else
    echo "  OK: verify status ${CURRENT_STATUS}"
fi

echo ""
if [ "$REGRESSIONS" -gt 0 ]; then
    echo "FAILED: ${REGRESSIONS} regression(s) detected for ${LANG}"
    exit 1
else
    echo "PASSED: No regressions for ${LANG}"
    exit 0
fi
