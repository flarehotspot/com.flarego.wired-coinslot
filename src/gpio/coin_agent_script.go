package gpio

// coinAgentScript is the Python GPIO agent source, kept as a Go string literal
// rather than a //go:embed of a .py file on purpose: the plugin/release build
// pipeline stages only Go sources, so a separate coin_agent.py would not be
// copied into plugins/installed/ and the embed would compile to nothing (or
// fail). Embedding the source directly in a .go file guarantees it is always
// compiled into the plugin binary and materialized at runtime by the Agent.
//
// Keep this in sync with the protocol documented in Agent (stdin: relay
// open|close|ping|shutdown; stdout: JSON events ready|pulse|pong|error).
var coinAgentScript = []byte(`#!/usr/bin/env python3
"""
Wired coinslot GPIO agent.

One long-lived process per coinslot, supervised by the Go plugin. All parameters
come from a single JSON config passed as argv[1] -- there is no string templating.

Protocol (line-oriented, both directions):

  stdin  (commands, one per line):
      relay open      -> energize the relay (accept coins)
      relay close     -> de-energize the relay (reject coins)
      ping            -> health check
      shutdown        -> clean exit

  stdout (events, one JSON object per line, always flushed):
      {"event": "ready"}
      {"event": "pulse"}
      {"event": "pong"}
      {"event": "error", "msg": "..."}

Lifetime is tied to the parent: SIGTERM or stdin EOF (parent died) triggers
GPIO.cleanup() and exit, so no GPIO line is ever left exported/driven.

Config JSON:
  {
    "library":      "rpi" | "opi",   # which sysfs GPIO library to use
    "board":        "one",            # OPi.GPIO board module name (opi only)
    "coin_pin":     3,                # physical header pin (BOARD numbering)
    "relay_pin":    5,                # physical header pin (BOARD numbering)
    "pull":         "up" | "down",    # input bias for the coin pin
    "edge":         "falling" | "rising" | "both",
    "debounce_ms":  30,               # hardware debounce for the coin pin
    "relay_active": 1                 # output value that energizes the relay
  }
"""

import importlib
import json
import signal
import sys
import threading


def emit(obj):
    """Write a single JSON event line to stdout and flush immediately."""
    try:
        sys.stdout.write(json.dumps(obj) + "\n")
        sys.stdout.flush()
    except (BrokenPipeError, ValueError):
        # Parent went away mid-write; nothing useful we can do.
        pass


def die(gpio, code=0, msg=None):
    if msg is not None:
        emit({"event": "error", "msg": msg})
    if gpio is not None:
        try:
            gpio.cleanup()
        except Exception:
            pass
    sys.exit(code)


def load_config():
    if len(sys.argv) < 2:
        return None, "missing config argument"
    try:
        return json.loads(sys.argv[1]), None
    except ValueError as e:
        return None, "invalid config json: %s" % e


def setup_gpio(cfg):
    """Import the requested library, configure pins, return the GPIO module."""
    library = cfg.get("library", "rpi")

    if library == "opi":
        import OPi.GPIO as GPIO  # noqa: N814
        board_name = cfg.get("board", "")
        if not board_name:
            raise ValueError("opi library requires a 'board' name")
        board_mod = importlib.import_module("orangepi." + board_name)
        GPIO.setmode(board_mod.BOARD)
    else:
        import RPi.GPIO as GPIO  # noqa: N814
        GPIO.setmode(GPIO.BOARD)

    GPIO.setwarnings(False)

    pull = GPIO.PUD_UP if cfg.get("pull", "up") == "up" else GPIO.PUD_DOWN
    edge_name = cfg.get("edge", "falling")
    if edge_name == "rising":
        edge = GPIO.RISING
    elif edge_name == "both":
        edge = GPIO.BOTH
    else:
        edge = GPIO.FALLING

    coin_pin = int(cfg["coin_pin"])
    relay_pin = int(cfg["relay_pin"])
    relay_active = int(cfg.get("relay_active", 1))
    debounce_ms = int(cfg.get("debounce_ms", 30))

    # Coin pin: input with bias + hardware-debounced edge interrupt.
    GPIO.setup(coin_pin, GPIO.IN, pull_up_down=pull)
    GPIO.add_event_detect(
        coin_pin, edge, callback=lambda _pin: emit({"event": "pulse"}),
        bouncetime=debounce_ms,
    )

    # Relay pin: output, initially de-energized (coins rejected) until told otherwise.
    GPIO.setup(relay_pin, GPIO.OUT, initial=(1 - relay_active))

    return GPIO


def main():
    cfg, err = load_config()
    if err is not None:
        emit({"event": "error", "msg": err})
        sys.exit(2)

    try:
        gpio = setup_gpio(cfg)
    except Exception as e:  # import errors, bad board, busy pin, etc.
        emit({"event": "error", "msg": "gpio setup failed: %s" % e})
        sys.exit(2)

    relay_pin = int(cfg["relay_pin"])
    relay_active = int(cfg.get("relay_active", 1))

    # SIGTERM -> clean shutdown. add_event_detect runs callbacks on a helper
    # thread, so the main thread is free to block on stdin.
    signal.signal(signal.SIGTERM, lambda *_: die(gpio, 0))

    emit({"event": "ready"})

    # Command loop. readline() returns "" on EOF (parent closed our stdin),
    # which is our signal to shut down.
    for raw in sys.stdin:
        cmd = raw.strip().lower()
        if cmd == "relay open":
            gpio.output(relay_pin, relay_active)
        elif cmd == "relay close":
            gpio.output(relay_pin, 1 - relay_active)
        elif cmd == "ping":
            emit({"event": "pong"})
        elif cmd == "shutdown":
            break
        elif cmd == "":
            continue
        else:
            emit({"event": "error", "msg": "unknown command: %s" % cmd})

    die(gpio, 0)


if __name__ == "__main__":
    # Keep a reference so the interpreter doesn't GC the callback thread early.
    threading.current_thread().name = "coin-agent-main"
    main()
`)
