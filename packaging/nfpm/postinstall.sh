#!/bin/sh
# Optional: remind operators to enable the packaged broker (REQ-MONO-069j).
# Does not prompt for credentials; fleet admins run: rtmx broker install
set -e
echo "rtmx: verification broker binary installed at /usr/bin/verify"
echo "rtmx: enable with: sudo rtmx broker install"
