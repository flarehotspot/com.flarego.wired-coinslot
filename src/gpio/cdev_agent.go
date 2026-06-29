package gpio

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	gpiocdev "github.com/warthog618/go-gpiocdev"
)

// CdevAgent drives one coinslot through the Linux GPIO character device using
// the pure-Go go-gpiocdev library. Unlike the Python *Agent it needs no
// subprocess, no interpreter, and no pip-installed library: the kernel delivers
// coin-pulse edges straight to an in-process callback, and the relay is a held
// output line. It is the driver for sunxi boards (e.g. OrangePi Zero 3) whose
// modern kernels expose GPIO only via /dev/gpiochipN, not the removed sysfs.
//
// Lines are addressed by (chip-label, line-offset). The chip is resolved by its
// pinctrl label rather than its /dev name, because /dev/gpiochipN ordering is
// not stable across kernels. The offset is computed from the Allwinner port
// name (e.g. "PH5") because sunxi kernels leave individual line names unset, so
// name-based lookup is unavailable.
type CdevAgent struct {
	cfg    Config
	logger Logger
	pulses chan struct{}

	mu        sync.Mutex
	started   bool
	coin      *gpiocdev.Line
	relay     *gpiocdev.Line
	relayOpen bool
}

func NewCdevAgent(cfg Config, logger Logger) *CdevAgent {
	return &CdevAgent{
		cfg:    cfg,
		logger: logger,
		pulses: make(chan struct{}, pulseBufferSize),
	}
}

// Compile-time guarantee that the char-device agent satisfies the interface.
var _ CoinAgent = (*CdevAgent)(nil)

// Pulses delivers one value per coin-acceptor pulse detected on the coin line.
func (a *CdevAgent) Pulses() <-chan struct{} { return a.pulses }

// Start requests the coin (edge-detected input) and relay (held output) lines.
// Hardware-unavailable conditions — no matching chip, a busy line, missing
// permissions — are logged once and leave the agent in a disabled state rather
// than failing: the coinslot runtime (and the dev mock-pulse path) keep working.
func (a *CdevAgent) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.started {
		return nil
	}
	a.started = true

	if err := a.setup(); err != nil {
		a.logf("gpiod agent disabled: " + err.Error())
		a.releaseLocked()
	}
	return nil
}

// Stop releases both lines. go-gpiocdev reverts a released line to its default
// (input) state, so the relay is de-energized on shutdown — coins rejected.
func (a *CdevAgent) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.releaseLocked()
	a.started = false
}

// OpenRelay energizes the relay so the acceptor takes coins.
func (a *CdevAgent) OpenRelay() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.relayOpen = true
	a.applyRelayLocked()
}

// CloseRelay de-energizes the relay so the acceptor rejects coins.
func (a *CdevAgent) CloseRelay() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.relayOpen = false
	a.applyRelayLocked()
}

// =============================================================================
// HELPER FUNCTIONS (internal)
// =============================================================================

// setup resolves the chip + offsets and requests both lines. Caller holds a.mu.
func (a *CdevAgent) setup() error {
	chipName, err := resolveChip(a.cfg.ChipLabel)
	if err != nil {
		return err
	}

	coinPort, coinOffset, err := a.resolvePin(a.cfg.CoinPin)
	if err != nil {
		return fmt.Errorf("coin pin: %w", err)
	}
	relayPort, relayOffset, err := a.resolvePin(a.cfg.RelayPin)
	if err != nil {
		return fmt.Errorf("relay pin: %w", err)
	}

	// Coin line: biased input with a hardware-debounced edge interrupt. Every
	// delivered event is one pulse (only the requested edge fires, except in
	// "both" mode where counting every transition is intentional).
	coinOpts := []gpiocdev.LineReqOption{
		gpiocdev.WithEventHandler(a.onEvent),
		edgeOption(a.cfg.Edge),
		pullOption(a.cfg.Pull),
	}
	if a.cfg.DebounceMs > 0 {
		coinOpts = append(coinOpts, gpiocdev.WithDebounce(time.Duration(a.cfg.DebounceMs)*time.Millisecond))
	}
	coin, err := gpiocdev.RequestLine(chipName, coinOffset, coinOpts...)
	if err != nil {
		return fmt.Errorf("request coin pin %d / %s (offset %d): %w", a.cfg.CoinPin, coinPort, coinOffset, err)
	}
	a.coin = coin

	// Relay line: output, initially de-energized (coins rejected) until a
	// client starts paying. relayOpen is honored on (re)start.
	relay, err := gpiocdev.RequestLine(chipName, relayOffset, gpiocdev.AsOutput(a.idleValue()))
	if err != nil {
		return fmt.Errorf("request relay pin %d / %s (offset %d): %w", a.cfg.RelayPin, relayPort, relayOffset, err)
	}
	a.relay = relay

	a.applyRelayLocked()
	a.logf(fmt.Sprintf("gpiod agent ready on %s (coin pin %d=%s, relay pin %d=%s)",
		chipName, a.cfg.CoinPin, coinPort, a.cfg.RelayPin, relayPort))
	return nil
}

// resolvePin maps a physical header pin to its Allwinner port name (via the
// board's Header map) and then to a char-device line offset.
func (a *CdevAgent) resolvePin(pin int) (port string, offset int, err error) {
	port, ok := a.cfg.Header[pin]
	if !ok {
		return "", 0, fmt.Errorf("physical pin %d is not a GPIO pin on this board", pin)
	}
	offset, err = sunxiOffset(port)
	if err != nil {
		return port, 0, err
	}
	return port, offset, nil
}

func (a *CdevAgent) onEvent(_ gpiocdev.LineEvent) {
	select {
	case a.pulses <- struct{}{}:
	default: // drop if the consumer is momentarily behind
	}
}

// applyRelayLocked drives the relay to match relayOpen. Caller holds a.mu.
func (a *CdevAgent) applyRelayLocked() {
	if a.relay == nil {
		return
	}
	v := a.idleValue()
	if a.relayOpen {
		v = a.cfg.RelayActive
	}
	if err := a.relay.SetValue(v); err != nil {
		a.logf("failed to set relay value: " + err.Error())
	}
}

// releaseLocked closes any held lines. Caller holds a.mu.
func (a *CdevAgent) releaseLocked() {
	if a.coin != nil {
		_ = a.coin.Close()
		a.coin = nil
	}
	if a.relay != nil {
		_ = a.relay.Close()
		a.relay = nil
	}
}

// idleValue is the output level that de-energizes the relay (coins rejected).
func (a *CdevAgent) idleValue() int { return 1 - a.cfg.RelayActive }

func (a *CdevAgent) logf(msg string) {
	if a.logger != nil {
		_ = a.logger.Error("[wired-coinslot] " + msg)
	}
}

// resolveChip finds the /dev/gpiochipN whose label contains want (a pinctrl
// register-address fragment such as "1c20800"). Matching by label rather than
// the /dev name survives unstable chip numbering across kernels; matching a
// substring rather than the exact label tolerates suffix differences (e.g. a
// node labeled "1c20800.pinctrl" vs "1c20800.pio") since the register address
// is fixed by the SoC.
func resolveChip(want string) (string, error) {
	if want == "" {
		return "", fmt.Errorf("no gpiochip label configured for this board")
	}
	for _, name := range gpiocdev.Chips() {
		c, err := gpiocdev.NewChip(name)
		if err != nil {
			continue
		}
		label := c.Label
		_ = c.Close()
		if strings.Contains(label, want) {
			return name, nil
		}
	}
	return "", fmt.Errorf("no gpiochip found with label containing %q", want)
}

// sunxiOffset converts an Allwinner port name (e.g. "PH5") into the line offset
// within its character-device chip. For the "pio" controller the layout matches
// the classic global numbering: offset = (bank-'A')*32 + pin (validated:
// PC5 -> 69). PL-bank pins live on a separate r_pio chip and aren't broken out
// on the supported boards' headers, so they're rejected with a clear error.
func sunxiOffset(port string) (int, error) {
	p := strings.ToUpper(strings.TrimSpace(port))
	if len(p) < 3 || p[0] != 'P' {
		return 0, fmt.Errorf("invalid port name %q (expected like PH5)", port)
	}
	bank := p[1]
	if bank < 'A' || bank > 'L' {
		return 0, fmt.Errorf("invalid GPIO bank in %q", port)
	}
	if bank == 'L' {
		return 0, fmt.Errorf("PL-bank pin %q is on the r_pio chip and not supported", port)
	}
	num, err := strconv.Atoi(p[2:])
	if err != nil || num < 0 || num > 31 {
		return 0, fmt.Errorf("invalid pin number in %q", port)
	}
	return int(bank-'A')*32 + num, nil
}

// edgeOption maps the configured edge to a go-gpiocdev request option,
// defaulting to falling (the common open-collector coin-acceptor pulse).
func edgeOption(edge string) gpiocdev.LineReqOption {
	switch edge {
	case "rising":
		return gpiocdev.WithRisingEdge
	case "both":
		return gpiocdev.WithBothEdges
	default:
		return gpiocdev.WithFallingEdge
	}
}

// pullOption maps the configured bias to a go-gpiocdev request option,
// defaulting to pull-up (matches open-collector acceptor outputs).
func pullOption(pull string) gpiocdev.LineReqOption {
	if pull == "down" {
		return gpiocdev.WithPullDown
	}
	return gpiocdev.WithPullUp
}
