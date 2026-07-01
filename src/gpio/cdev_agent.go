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
// subprocess, no interpreter, and no pip-installed library. The relay is a held
// output line; the coin pin is read by a high-frequency polling goroutine. It is
// the driver for sunxi boards (e.g. OrangePi One/Zero 3) whose modern kernels
// expose GPIO only via /dev/gpiochipN, not the removed sysfs.
//
// Coin detection POLLS rather than using edge interrupts. On Allwinner sunxi
// the GPIO *value* is reliably readable via the char device, but cdev *edge
// events* do not actually fire on many pins/kernels (the IRQ shows as armed yet
// never delivers) — verified on an OrangePi One (H3, kernel 5.15) where every
// coin pulse was visible by polling the line value but produced zero edge
// events. Coin-acceptor pulses are tens of milliseconds wide, so a ~1ms poll
// with software debounce catches every one reliably and portably.
//
// Lines are addressed by (chip-label, line-offset). The chip is resolved by its
// pinctrl label substring rather than its /dev name, because /dev/gpiochipN
// ordering is not stable across kernels. The offset is computed from the
// Allwinner port name via the sunxi formula because sunxi kernels leave
// individual line names unset, so name-based lookup is unavailable.
type CdevAgent struct {
	cfg    Config
	logger Logger
	pulses chan struct{}

	mu        sync.Mutex
	started   bool
	coin      *gpiocdev.Line
	relay     *gpiocdev.Line
	relayOpen bool

	done     chan struct{}
	stopOnce sync.Once
}

// coinPollInterval is how often the coin line value is sampled. Coin-acceptor
// pulses are tens of ms wide, so 1ms (1kHz) catches every edge with margin at
// negligible CPU cost (one ioctl read per tick).
const coinPollInterval = time.Millisecond

func NewCdevAgent(cfg Config, logger Logger) *CdevAgent {
	return &CdevAgent{
		cfg:    cfg,
		logger: logger,
		pulses: make(chan struct{}, pulseBufferSize),
		done:   make(chan struct{}),
	}
}

// Compile-time guarantee that the char-device agent satisfies the interface.
var _ CoinAgent = (*CdevAgent)(nil)

// Pulses delivers one value per coin-acceptor pulse detected on the coin line.
func (a *CdevAgent) Pulses() <-chan struct{} { return a.pulses }

// Start requests the coin (polled input) and relay (held output) lines and
// launches the coin poller. Hardware-unavailable conditions — no matching chip,
// a busy line, missing permissions — are logged once and leave the agent in a
// disabled state rather than failing: the coinslot runtime (and the dev
// mock-pulse path) keep working.
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

// Stop signals the coin poller to exit and releases both lines. go-gpiocdev
// reverts a released line to its default (input) state, so the relay is
// de-energized on shutdown — coins rejected.
func (a *CdevAgent) Stop() {
	a.stopOnce.Do(func() { close(a.done) })
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

	// Coin line: biased input, read by the poller (NOT edge interrupts — see the
	// type doc for why cdev edge events are unreliable on sunxi).
	coin, err := gpiocdev.RequestLine(chipName, coinOffset, gpiocdev.AsInput, pullOption(a.cfg.Pull))
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
	go a.pollCoin(coin)

	a.logf(fmt.Sprintf("gpiod agent ready on %s (coin pin %d=%s polled, relay pin %d=%s)",
		chipName, a.cfg.CoinPin, coinPort, a.cfg.RelayPin, relayPort))
	return nil
}

// pollCoin samples the coin line at coinPollInterval and emits one pulse per
// configured edge (default falling), with a software debounce so contact bounce
// within DebounceMs of an accepted pulse is ignored. It exits when the agent is
// stopped (a.done closed) or the line read fails (line closed on release).
func (a *CdevAgent) pollCoin(line *gpiocdev.Line) {
	debounce := time.Duration(a.cfg.DebounceMs) * time.Millisecond
	ticker := time.NewTicker(coinPollInterval)
	defer ticker.Stop()

	prev, err := line.Value()
	if err != nil {
		a.logf("coin poll: initial read failed: " + err.Error())
		return
	}
	var lastPulse time.Time

	for {
		select {
		case <-a.done:
			return
		case <-ticker.C:
		}

		cur, err := line.Value()
		if err != nil {
			return // line released/closed
		}
		if cur == prev {
			continue
		}
		if isPulseEdge(a.cfg.Edge, prev, cur) {
			now := time.Now()
			if debounce <= 0 || now.Sub(lastPulse) >= debounce {
				lastPulse = now
				select {
				case a.pulses <- struct{}{}:
				default: // drop if the consumer is momentarily behind
				}
			}
		}
		prev = cur
	}
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

// isPulseEdge reports whether a prev->cur level change counts as one pulse for
// the configured edge: "rising" (low->high), "both" (any change), or the
// default "falling" (high->low, the common open-collector coin-acceptor pulse).
func isPulseEdge(edge string, prev, cur int) bool {
	switch edge {
	case "rising":
		return prev == 0 && cur == 1
	case "both":
		return true
	default: // falling
		return prev == 1 && cur == 0
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

// pullOption maps the configured bias to a go-gpiocdev request option,
// defaulting to pull-up (matches open-collector acceptor outputs).
func pullOption(pull string) gpiocdev.LineReqOption {
	if pull == "down" {
		return gpiocdev.WithPullDown
	}
	return gpiocdev.WithPullUp
}
