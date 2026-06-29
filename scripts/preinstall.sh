#!/bin/sh
# Pre-install hook for com.flarego.wired-coinslot.
#
# Runs BEFORE the plugin's Init, and only once the device has internet (so opkg
# and pip can fetch). BOARD-AWARE — acts ONLY on boards that have a matching
# entry in src/gpio/board.go's registry; an unrecognized/empty device_model
# installs and compiles nothing (see the `*)` branch):
#   - gpiod boards (orangepi-one/pc/zero-3) -> NOTHING to install. Their GPIO is
#     driven in-process by the compiled-in Go go-gpiocdev library: no Python, no
#     pip, no network dependency at install time (the failure mode that made
#     pip-installed OPi.GPIO unreliable on offline coin-vendo boxes). Debug CLIs
#     come from plugin.json "system_packages" (gpiod-tools).
#   - rpi-4 -> RPi.GPIO (a C extension, compiles from sdist).
#
# RPi.GPIO is PyPI-only (not in the opkg feed), so it is pip-installed here. Its
# prerequisites — python3-pip, python3-setuptools, gcc, python3-dev — are
# opkg-installed ON DEMAND in the rpi-4 branch only. They are intentionally NOT
# in system_packages, so gpiod boards (the common case) never install Python.

set -u

log() { echo "[wired-coinslot/preinstall] $*"; }

# Environment guard: the core passes GO_ENV (development|sandbox|staging|
# production). Staging runs on real hardware (a deployed device), so it is
# treated like production and runs device setup; only development/sandbox skip.
case "${GO_ENV:-production}" in
  production | staging) ;; # real device — run setup
  *)
    log "GO_ENV=${GO_ENV:-unset} (not production/staging) — skipping GPIO library install"
    exit 0
    ;;
esac

# ensure_pip opkg-installs python3-pip + python3-setuptools (idempotent) and
# resolves a pip entrypoint into $PIP. Called only on the rpi/opi branches —
# gpiod boards never need Python, so these packages are kept out of
# system_packages and installed here on demand instead.
ensure_pip() {
  if command -v opkg >/dev/null 2>&1; then
    log "ensuring python3-pip + python3-setuptools"
    opkg install python3-pip python3-setuptools || log "WARN: could not install python3-pip/setuptools"
  fi
  if command -v pip3 >/dev/null 2>&1; then
    PIP="pip3"
  elif command -v pip >/dev/null 2>&1; then
    PIP="pip"
  else
    PIP="python3 -m pip"
  fi
}

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

# Only recognized boards (those with a src/gpio/board.go registry entry) trigger
# any setup. An unrecognized/empty device_model installs and compiles NOTHING —
# the plugin simply has no coinslot hardware support on that board.
case "$device_model" in
  orangepi-zero-3 | orangepi-one | orangepi-pc)
    # gpiod (char-device) boards: GPIO is compiled into the plugin via the Go
    # go-gpiocdev library, so there is NO Python GPIO package to install.
    log "gpiod board ($device_model) — GPIO is built into the plugin; no Python GPIO library needed"
    ;;
  rpi-4)
    # The only Python-driver board. RPi.GPIO is a C extension with no musl/ARM
    # wheels -> compiles from sdist, which needs python3-dev (Python.h) + a C
    # compiler, installed on demand here.
    ensure_pip
    if command -v opkg >/dev/null 2>&1; then
      log "ensuring build deps (gcc, python3-dev)"
      opkg install gcc python3-dev || log "WARN: could not install build deps; RPi.GPIO build may fail"
    fi
    log "installing RPi.GPIO via $PIP"
    if ! $PIP install --no-cache-dir "RPi.GPIO"; then
      log "ERROR: failed to install RPi.GPIO (needs gcc + python3-dev)"
      exit 1
    fi
    ;;
  *)
    # No matching board in the registry — do not install or compile anything.
    log "device_model '${device_model:-unknown}' is not a recognized board — skipping GPIO setup"
    exit 0
    ;;
esac

log "done"
exit 0
