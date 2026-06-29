# CLAUDE.md — com.flarego.wired-coinslot

Reads coin-acceptor pulses and drives the coin relay over GPIO to grant WiFi sessions.
This plugin is a **separate, gitignored repo** (the parent flarewifi repo won't show its history).

> Spell the product **"Flarewifi"** (capital F only) in all user-facing text/translations.
> Use `api.Translate(type, text, pairs...)` for ALL user-facing strings (`<% .key %>` delimiters, not `{{ }}`).

## What it does (payment flow)

1. A coin acceptor pulses an input GPIO; the relay (output GPIO) gates whether the acceptor takes coins.
2. `gpio` agent emits one value per pulse on a channel → `PulseCounter` aggregates pulses within an idle
   window (`WindowMs`, default 400ms) → greedy-resolves the count to a denomination amount.
3. `PaymentSessionManager.Credit` → `purchase.CreatePayment` (incremental, per coin) → SSE broadcast to the
   insert-coin page. "Done Payment" → `purchase.Execute`. **The plugin does NOT call CreateSession** —
   wifi-hotspot grants the session on purchase execute.
4. **Relay gating = concurrency guard:** the relay is OPEN only while a client holds the coin-events SSE
   subscription (`GET /payments/coin-events/{id}` → `OpenRelay`; disconnect + grace → `CloseRelay`). It
   boots CLOSED. Most acceptors are **inhibited while the relay is closed**, so coins only pulse during an
   active payment session — keep this in mind when testing ("no pulses" is often "relay was closed").

### "Counting payment" cue + idle countdown (insert-coin page UX)

- **Counting cue:** a coin's pulse burst takes `WindowMs` of silence to resolve into an amount, so there's a
  visible gap between "coin dropped" and "Received N". `PulseCounter` fires a second callback `onCounting`
  on the **leading edge** (count 0→1) of each burst → `PaymentSessionManager.Counting` broadcasts a
  `CoinEvent{Counting:true}` (no money recorded) so the page shows "Counting payment…" immediately, then the
  resolved `Credit` event (`counting:false, last_coin>0`) overwrites it with the real value.
- **Idle countdown / auto-finalize:** `WiredCoinslot.PaymentTimeoutSecs` (default 30, adjustable in
  Settings → Payment) is broadcast in every `CoinEvent.TimeoutSecs`. The page runs the countdown and
  **resets it on every pulse/coin** (each `Counting` and `last_coin>0` event). On expiry the page navigates:
  total>0 → `payments.done` (executes the accumulated amount), total==0 → `payments.cancel`
  (`CancelPayingHandler` → `Execute{Success:false}` → wifi-hotspot calls `purchase.Cancel`). The clock is
  authoritative on the **server** (only it sees real pulses) but the *navigation* is client-side because
  only an HTTP handler can call `RedirectToCallback` (it mints a core-internal JWT). This is safe because
  **relay-open ⟺ SSE-subscribed ⟺ page JS running**, so the countdown can't be bypassed by killing JS.

## GPIO architecture — TWO drivers behind one interface

`gpio.CoinAgent` (in `src/gpio/agent.go`) is the interface the `Manager` drives: `Pulses() / Start() /
Stop() / OpenRelay() / CloseRelay()`. The factory `gpio.NewCoinAgent(cfg, logger)` picks the impl by
`cfg.Driver`, resolved from `src/gpio/board.go`'s registry (keyed on os_release `device_model`):

| Board | Driver | Impl |
|-------|--------|------|
| `orangepi-one`, `orangepi-pc` (H3) | `gpiod` | `cdev_agent.go` (pure-Go, register `1c20800`) |
| `orangepi-zero-3` (H618) | `gpiod` | `cdev_agent.go` (register `300b000`) |
| `rpi-4`, unknown fallback | `rpi`/`opi` | `agent.go` (Python `coin_agent_script.go`) |

### gpiod driver (`cdev_agent.go`) — the default for all Orange Pi boards

Pure-Go via `github.com/warthog618/go-gpiocdev` (no CGO, no Python, no pip, no network at install). Chosen
because `OPi.GPIO` is a PyPI package whose pip-install **fails on an offline coin-vendo box**, and sysfs
(which RPi/OPi.GPIO use) is removed on modern kernels.

- **Physical-pin addressing is preserved.** Config/UI use BOARD pin numbers (`CoinPin`/`RelayPin`, default
  3/5). The driver converts internally: physical pin → Allwinner port name (`Board.Header` map) → line
  offset (`sunxiOffset`, `offset = (bank-'A')*32 + pin`). The `Header` map is the board's fixed PCB layout
  (safe to hardcode); the H3 map is derived from OPi.GPIO's `orangepi.pc.BOARD` so pins match the old driver.
- **Chip is resolved by pinctrl-label substring** (`resolveChip` → `strings.Contains(label, "1c20800")`),
  NOT by `/dev/gpiochipN` (unstable numbering) and NOT by line name (sunxi leaves line names unset).

### 🚨 CRITICAL: gpiod coin detection POLLS, never edge-interrupts

On Allwinner sunxi (verified H3 / kernel 5.15, OrangePi One), the GPIO **value** is reliably readable via
the char device, but cdev **edge events DO NOT FIRE** on many pins — debugfs shows the IRQ armed
(`gpio-N ... in hi IRQ`) yet `go-gpiocdev`'s `WithEventHandler` callback never runs. Field proof: 3×5-peso
coins = 15 clean `hi→lo` transitions visible by **polling**, but **zero** edge events.

So `pollCoin()` samples `line.Value()` every `coinPollInterval` (1ms) and detects the configured edge
(`isPulseEdge`, default falling) with a software debounce (`DebounceMs`). The relay path is unaffected
(direct `SetValue`, no interrupts). **Do NOT "optimize" this back to `WithEventHandler`/`WithFallingEdge`
for sunxi — it silently detects nothing.** Coin pulses are tens of ms wide, so 1ms polling is ample.

## Config (`src/wired-coinslot.go`)

`WiredCoinslot` JSON in plugin config under `wired_coinslots/<id>`. Fields: `CoinPin`/`RelayPin` (physical
BOARD pins), `RelayActive` (output level that energizes; 1=active-high), `Pull` (up/down), `Edge`
(falling/rising/both), `DebounceMs`, `WindowMs`, `BoardModel` (override; empty = auto-detect from
os_release), `Denominations` (`[]{Pulses, Amount}`). `ApplyDefaults` seeds "Main Vendo" and fills unset
fields. Admin UI = `resources/views/settings.templ` (physical-pin inputs for every board; Board Model
dropdown from `gpio.BoardModels()`).

## Build / install

- `system_packages` (plugin.json): `gpiod-tools` ONLY (the OpenWRT package name — NOT
  `libgpiod-tools`/`-utils`; v1.6.4, gives `gpiodetect`/`gpiomon`/`gpioget` for debugging).
- `scripts/preinstall.sh` is board-aware and `GO_ENV`-gated (**`staging` is treated like `production`**;
  dev/sandbox skip). It acts ONLY on boards with a `board.go` registry entry — an unknown/empty
  `device_model` installs and compiles **nothing**. gpiod boards install nothing (GPIO is compiled in);
  `rpi-4` pip-installs RPi.GPIO, with `python3-pip`/`python3-setuptools`/`gcc`/`python3-dev`
  opkg-installed **on demand in that branch** (deliberately kept out of `system_packages` so gpiod boards
  never install Python). Bump `version` in plugin.json so install hooks re-run on redeploy.
- The Python fallback agent lives as a Go string in `coin_agent_script.go` (NOT `//go:embed` — the build
  stages only `.go`). It logs once and stops on exit code 2 (persistent setup failure), not a retry loop.

## Dev vs prod

- **Mock coin is dev-only** via paired `src/mock_dev.go` / `src/mock_prod.go` build-tag files (single source
  of dev/prod truth): dev mounts `POST /payments/mock-coin/{id}` and `MockCoinURL` returns a URL; prod
  no-ops both. Don't add a `views`-package build-tag flag (views can't import src — cycle); pass the URL down.
- The dev container's `device_model` is `arm64-generic` → `rpi` fallback → Python → `No module named 'RPi'`
  (harmless, logs once). To exercise gpiod in dev, set Board Model = `orangepi-zero-3` (it'll log
  `no gpiochip found with label "300b000"` since the container has no such chip).

## Testing on real hardware

- Tools: `gpiodetect` (confirm chip label contains the register fragment), `gpioget`/`gpiomon` (but the
  running app holds the lines — `gpiomon`/`gpioget` get EBUSY; use `/sys/kernel/debug/gpio` to read live
  line state non-disruptively: `mount -t debugfs none /sys/kernel/debug` then `grep " gpio-N "`).
- The app runs as `flare server` (cdev consumer label `gpiocdev-<pid>`), started by `/etc/rc.d/S99flarehotspot`
  (there is no `/etc/init.d/flare*`). `curl` is absent on the device.
- To see coin pulses you must have the relay OPEN (start a payment session via the portal) so the acceptor
  is enabled.

## Don'ts

- Don't modify core files (`core/`, `sdk/`) — this is plugin code.
- Don't use edge interrupts for sunxi coin detection (see CRITICAL above).
- Don't add foreign keys to `sessions.id` from plugin tables — use `session_uuid` (cloud-sync may lack local IDs).
- ES5 JS only; wrap templ URLs with `templ.SafeURL()`; `int64` for IDs; handle ALL errors with rollback.
- `func Init(api sdkapi.IPluginApi) error` MUST return `error` or the core loader silently never calls it.
