#!/bin/sh
# Pre-install hook for com.flarego.wired-coinslot.
#
# Runs BEFORE the plugin's Init, and only once the device has internet. It is now
# effectively a no-op for every recognized board: ALL of them (the Orange Pis and
# the Raspberry Pi 4) drive GPIO in-process via the compiled-in Go go-gpiocdev
# char-device library — no Python, no pip, no network dependency at install time.
# The rpi-4 formerly pip-installed RPi.GPIO, but RPi.GPIO drives the sysfs GPIO
# interface, which is deprecated/removed on the modern kernel its OpenWRT image
# ships (bcm27xx/bcm2711, Linux 6.x) — so it no longer detects coin pulses. The
# board now uses gpiod like the others (see src/gpio/board.go).
#
# The script is kept (rather than deleted) as the board-aware seam for any future
# board that needs on-device setup. Debug CLIs come from plugin.json
# "system_packages" (gpiod-tools). An unrecognized/empty device_model installs
# and compiles nothing.

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
# any setup. Every recognized board now uses the compiled-in gpiod char-device
# driver, so none of them install anything here. An unrecognized/empty
# device_model installs and compiles NOTHING — the plugin simply has no coinslot
# hardware support on that board.
case "$device_model" in
  orangepi-zero-3 | orangepi-one | orangepi-pc | rpi-4)
    # gpiod (char-device) boards: GPIO is compiled into the plugin via the Go
    # go-gpiocdev library, so there is NO Python GPIO package to install. The
    # rpi-4 used to pip-install RPi.GPIO, but that drives the sysfs interface
    # gone on modern kernels; it now uses gpiod like the Orange Pis.
    log "gpiod board ($device_model) — GPIO is built into the plugin; no Python GPIO library needed"
    ;;
  *)
    # No matching board in the registry — do not install or compile anything.
    log "device_model '${device_model:-unknown}' is not a recognized board — skipping GPIO setup"
    exit 0
    ;;
esac

log "done"
exit 0
