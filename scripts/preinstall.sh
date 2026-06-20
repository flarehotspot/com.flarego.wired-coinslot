#!/bin/sh
# Pre-install hook for com.flarego.wired-coinslot.
#
# This runs BEFORE the plugin's Init (which starts the GPIO agent that imports
# the library installed here) and — like all install scripts — only once the
# device has internet, so pip can reach PyPI. That ordering is the reason this is
# a PREinstall, not a postinstall: Init depends on the library being present.
#
# RPi.GPIO and OPi.GPIO are NOT in the OpenWRT opkg feed (they are PyPI
# packages), so we pip-install them here. plugin.json "system_packages" provide
# the prerequisites: python3-pip (which pulls full python3 -> the ssl module via
# python3-openssl + ca-certs, so pip can reach PyPI over HTTPS) and
# python3-setuptools (the build backend). OpenWRT 23.05 ships no PEP-668
# EXTERNALLY-MANAGED marker, so a system-wide `pip install` works without
# --break-system-packages and without a venv (the target python is built
# --without-ensurepip, so venvs can't bootstrap pip anyway).
#
# This is BOARD-AWARE: it reads /etc/os_release.json device_model and installs
# only the library the coin agent will actually import, mirroring the Go-side
# board detection (src/gpio/board.go DetectBoard): "orangepi-*" -> OPi.GPIO
# (pure-Python, no compiler); anything else (rpi-*, or unknown -> the rpi
# fallback) -> RPi.GPIO (a C extension that compiles from source, so it needs
# gcc + python3-dev, installed on demand only on that branch).
#
# This runs on every (re)install/update; pip is idempotent ("already satisfied").

set -u

log() { echo "[wired-coinslot/preinstall] $*"; }

# Environment guard: the core passes GO_ENV (development|sandbox|staging|
# production). Only run device setup in production; clean no-op in development.
if [ "${GO_ENV:-production}" != "production" ]; then
  log "GO_ENV=${GO_ENV:-unset} (not production) — skipping GPIO library install"
  exit 0
fi

# Resolve a pip entrypoint.
if command -v pip3 >/dev/null 2>&1; then
  PIP="pip3"
elif command -v pip >/dev/null 2>&1; then
  PIP="pip"
else
  PIP="python3 -m pip"
fi

# Detect the board from os_release (same signal the Go side uses). Prefer
# jsonfilter (OpenWRT base), fall back to a grep/sed parse.
OS_RELEASE="/etc/os_release.json"
device_model=""
if [ -r "$OS_RELEASE" ]; then
  if command -v jsonfilter >/dev/null 2>&1; then
    device_model=$(jsonfilter -i "$OS_RELEASE" -e '@.device_model' 2>/dev/null)
  fi
  if [ -z "$device_model" ]; then
    device_model=$(grep -o '"device_model"[[:space:]]*:[[:space:]]*"[^"]*"' "$OS_RELEASE" 2>/dev/null \
      | sed -e 's/.*:[[:space:]]*"//' -e 's/".*//')
  fi
fi
log "device_model=${device_model:-unknown}"

case "$device_model" in
  orangepi-*)
    # Orange Pi: OPi.GPIO is pure-Python, no compiler needed.
    log "installing OPi.GPIO via $PIP (Orange Pi)"
    if ! $PIP install --no-cache-dir "OPi.GPIO"; then
      log "ERROR: failed to install OPi.GPIO"
      exit 1
    fi
    ;;
  *)
    # rpi-* and unknown both resolve to RPi.GPIO at runtime (board.go fallback).
    # RPi.GPIO is a C extension with no musl/ARM wheels -> compiles from sdist,
    # which needs python3-dev (Python.h) + a C compiler. Install those on demand
    # so the Orange Pi path above never pays for them.
    log "installing RPi.GPIO via $PIP (board: ${device_model:-unknown})"
    if command -v opkg >/dev/null 2>&1; then
      log "ensuring build deps (gcc, python3-dev)"
      opkg install gcc python3-dev || log "WARN: could not install build deps; RPi.GPIO build may fail"
    fi
    if ! $PIP install --no-cache-dir "RPi.GPIO"; then
      log "ERROR: failed to install RPi.GPIO (needs gcc + python3-dev)"
      exit 1
    fi
    ;;
esac

log "done"
exit 0
