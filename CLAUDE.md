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

### One-payer-at-a-time (atomic claim + ownership re-checks)

Only one client may pay through a coinslot at a time, enforced in three layers — don't weaken any of them:

- **Atomic claim:** `WiredCoinslot.TryUseBy` uses `sync.Map.LoadOrStore` so check-and-claim is one step (the
  old `CanBeUsedBy()`+`UseBy()` pair was a check-then-act race where two clients could both see "free").
  Back out only with `ReleaseIfOwner` (`CompareAndDelete`) — never `DoneUsing`/`Delete` after a failed
  start, or you can delete another device's claim.
- **`Begin` re-validates** the claim (`UsedCoinslots.Load == clientID`) and refuses to overwrite a session
  owned by a different client; returns `bool`.
- **`Subscribe(coinslotID, clientID)` re-validates** that the subscriber owns the session before handing out
  the SSE channel (and thus the relay). `CoinEventsHandler` passes the authenticated client device's ID.

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
| `orangepi-one`, `orangepi-pc` (H3) | `gpiod` (scheme `sunxi`) | `cdev_agent.go` (pure-Go, register `1c20800`) |
| `orangepi-zero-3` (H618) | `gpiod` (scheme `sunxi`) | `cdev_agent.go` (register `300b000`) |
| `rpi-4` (bcm2711) | `gpiod` (scheme `bcm`) | `cdev_agent.go` (chip label `pinctrl-bcm2711`) |
| unknown fallback only | `rpi`/`opi` | `agent.go` (Python `coin_agent_script.go`) |

### gpiod driver (`cdev_agent.go`) — the driver for EVERY registered board

Pure-Go via `github.com/warthog618/go-gpiocdev` (no CGO, no Python, no pip, no network at install). Chosen
because the Python libs `OPi.GPIO`/`RPi.GPIO` drive the **sysfs GPIO interface, which is deprecated/removed
on modern kernels** (Linux 6.1+) — there it stops firing edge events and renumbers pins — and `OPi.GPIO`'s
pip-install additionally **fails on an offline coin-vendo box**. The `rpi-4` was the last board on the Python
path; it moved to `gpiod` because its OpenWRT image (`bcm27xx/bcm2711`) now ships one of those kernels.

- **Physical-pin addressing is preserved for every board.** Config/UI use BOARD pin numbers
  (`CoinPin`/`RelayPin`, default 3/5). The driver converts internally: physical pin → line name
  (`Board.Header` map) → line offset, via the board's **`Scheme`**:
  - `sunxi` — Allwinner port name (e.g. `PH5`) → `sunxiOffset` (`offset = (bank-'A')*32 + pin`). H3 map
    derived from OPi.GPIO's `orangepi.pc.BOARD` so pins match the old driver.
  - `bcm` — Raspberry Pi BCM name (e.g. `GPIO17`) → `bcmOffset`; on `pinctrl-bcm2711` the line **offset ==
    BCM GPIO number**, so it just parses the trailing number. `rpiBcm2711Header` is the fixed J8 layout.
  The `Header` map is the board's fixed PCB layout (safe to hardcode).
- **Chip is resolved by pinctrl-label substring** (`resolveChip` → `strings.Contains`): `1c20800`/`300b000`
  for the Orange Pis, `pinctrl-bcm2711` for the Pi 4. Matched NOT by `/dev/gpiochipN` (unstable numbering)
  and NOT by kernel line name (these SoCs leave line names unset).

### 🚨 CRITICAL: gpiod coin detection POLLS, never edge-interrupts

On Allwinner sunxi (verified H3 / kernel 5.15, OrangePi One), the GPIO **value** is reliably readable via
the char device, but cdev **edge events DO NOT FIRE** on many pins — debugfs shows the IRQ armed
(`gpio-N ... in hi IRQ`) yet `go-gpiocdev`'s `WithEventHandler` callback never runs. Field proof: 3×5-peso
coins = 15 clean `hi→lo` transitions visible by **polling**, but **zero** edge events.

So `pollCoin()` samples `line.Value()` every `coinPollInterval` (1ms) and detects the configured edge
(`isPulseEdge`, default falling) with a software debounce (`DebounceMs`). The relay path is unaffected
(direct `SetValue`, no interrupts). **Do NOT "optimize" this back to `WithEventHandler`/`WithFallingEdge`
for sunxi — it silently detects nothing.** Coin pulses are tens of ms wide, so 1ms polling is ample.

On `bcm2711` (Pi 4) cdev edge events DO work, but the driver still polls: it's the single shared code path,
and 1ms polling is more than fast enough for coin pulses. No need to special-case the Pi.

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
  dev/sandbox skip). It is now effectively a **no-op for every recognized board**: all of them (Orange Pis
  **and** `rpi-4`) use the compiled-in gpiod driver, so nothing is installed. It's kept as the seam for any
  future board that needs on-device setup. (Historical note: `rpi-4` used to pip-install RPi.GPIO here —
  removed because RPi.GPIO drives the now-gone sysfs interface.) Bump `version` in plugin.json so install
  hooks re-run on redeploy.
- The Python fallback agent (`coin_agent_script.go`) is now reached **only** by the unknown-board `rpi`/`opi`
  fallback, never by a registered board. It lives as a Go string (NOT `//go:embed` — the build stages only
  `.go`) and logs once then stops on exit code 2 (persistent setup failure), not a retry loop.

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
