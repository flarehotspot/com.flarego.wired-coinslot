# Wired Coinslot

Wired Coinslot turns a machine's GPIO header into a coin-operated payment method for Flarewifi. It reads pulses from a coin acceptor wired directly to the board and drives a relay that gates the acceptor, letting customers pay for WiFi access by dropping coins instead of using a card, e-wallet, or voucher code.

## Features

- **Direct GPIO coin acceptance** — counts pulses from a coin acceptor on one GPIO pin and switches a relay on another to enable/disable the acceptor, with no external controller board required.
- **Configurable pin wiring** — coin input pin, relay output pin, relay active level (HIGH/LOW), pull-up/pull-down bias, pulse edge (falling/rising/both), debounce time, and the idle window used to close out a coin's pulse burst are all adjustable from the admin settings page.
- **Per-coinslot denominations** — map a coin's pulse count to its monetary value (e.g. 1 pulse = 1 peso, 5 pulses = 5 pesos), with denominations added, edited, or removed per coinslot.
- **Multiple coinslots** — configure more than one coin acceptor on the same machine, each with its own name, wiring, and denomination table; a customer-facing payment method label ("alias") can be set per coinslot for the sales inventory and transaction records.
- **Automatic board detection** — resolves the correct GPIO driver from the machine's device model, with a manual board override in settings if auto-detection needs correcting.
- **Live payment progress over SSE** — the customer-facing insert-coin page updates in real time as coins are inserted, showing a "counting payment" indicator while a coin's pulses are still being resolved and the running total as each coin is credited.
- **Idle countdown with auto-finalize** — an adjustable countdown (default 30 seconds) resets on every coin inserted; on expiry the page automatically completes the purchase with whatever amount was paid, or cancels if nothing was inserted.
- **One payer at a time** — each coinslot can only be used by one client device at a time, so a coinslot already in use is not offered to another customer until the current payment finishes or times out.
- **Registers as a payment option** — appears alongside a machine's other payment methods (e.g. vouchers, online payment) wherever customers choose how to pay.

## How It Works

1. The relay stays closed (coin acceptor disabled) until a customer starts a payment session, at which point the relay opens and the acceptor begins accepting coins.
2. Each inserted coin produces a burst of pulses on the coin pin; the plugin counts the burst and matches the pulse count to a configured denomination to determine the coin's value.
3. Each resolved coin is credited to the payment session and pushed to the insert-coin page via Server-Sent Events, so the balance updates live without the page needing to poll.
4. When the customer finishes (or the idle countdown expires), the accumulated amount is submitted to complete the purchase; the relay closes again once the session ends.

## Requirements

- **GPIO-capable single-board hardware** — this plugin targets Orange Pi and Raspberry Pi boards with an accessible GPIO header. It does not apply to generic OpenWRT-only routers without GPIO pins.
- A physical **coin acceptor** wired to a GPIO input pin and a **relay** wired to a GPIO output pin (default header pins 3 and 5 respectively; both are reconfigurable).
- The **`gpiod-tools`** system package, used for on-device GPIO diagnostics.
- On Raspberry Pi boards, the plugin installs its Python GPIO dependency automatically during setup; supported Orange Pi boards use a built-in GPIO driver with no additional runtime dependency.

## Configuration

From the plugin's admin settings page, configure for each coinslot: the coinslot name and optional payment-method alias, coin/relay pin numbers, relay active level, board model (auto-detect or manual override), payment timeout, and the coin denomination table. Advanced signal settings (pull bias, pulse edge, debounce, pulse window) are available for fine-tuning acceptors that need non-default timing.
