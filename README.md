# Wired Coinslot

Turn spare change into WiFi sales. Wired Coinslot lets your machine accept real, physical coins as payment for WiFi access — no separate payment terminal, card reader, or online wallet required. Just wire a coin acceptor to your machine and start selling connectivity for cash on the spot. It's a simple, low-cost way to serve customers who just want to pay with the coins in their pocket.

## Features

- **Accept cash the simplest way possible** — customers drop coins in, WiFi access follows automatically, no extra payment hardware to buy or manage.
- **Set your own coin values** — decide how much each coin is worth so pricing matches your local currency and rates.
- **Live payment progress on screen** — customers see their running total update in real time as each coin is counted, so there's never any doubt a coin was accepted.
- **Automatic checkout** — once enough has been paid, access is granted right away with no extra taps or confirmations needed.
- **Idle timeout protection** — if a customer walks away mid-payment, the session automatically finishes or cancels after a short wait, keeping the coinslot free for the next customer.
- **One customer at a time** — each coinslot handles a single payment at a time, so there's no mix-up over whose coins paid for what.
- **Works alongside your other payment methods** — sits right next to vouchers and other ways customers can already pay, giving them one more easy option.

Built for machines running on Orange Pi or Raspberry Pi hardware with a coin acceptor and relay wired in.

## Supported Devices

| Board | GPIO Driver | Dependencies |
|-------|-------------|--------------|
| Orange Pi One (Allwinner H3) | `gpiod` (compiled in) | None |
| Orange Pi PC (Allwinner H3) | `gpiod` (compiled in) | None |
| Orange Pi Zero 3 (Allwinner H618) | `gpiod` (compiled in) | None |
| Raspberry Pi 4 | Python `RPi.GPIO` | Installed automatically during plugin setup |
