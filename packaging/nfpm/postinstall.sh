#!/bin/sh
# Enable the packaged verification broker after deb/rpm install (REQ-MONO-069j).
# Runs as root in the package manager context. If enable fails, first
# `rtmx verify` elevates once as a fallback.
set -e
echo "rtmx: verification broker binary installed at /usr/bin/verify"
if command -v rtmx >/dev/null 2>&1; then
  if rtmx broker status >/dev/null 2>&1; then
    echo "rtmx: refreshing broker via upgrade"
    if ! rtmx broker upgrade; then
      echo "rtmx: broker upgrade deferred; first rtmx verify will elevate once" >&2
    fi
  else
    echo "rtmx: enabling broker via install"
    if ! rtmx broker install; then
      echo "rtmx: broker enable deferred; first rtmx verify will elevate once" >&2
    fi
  fi
else
  echo "rtmx: rtmx not on PATH during postinstall; enable with: sudo rtmx broker install" >&2
fi
