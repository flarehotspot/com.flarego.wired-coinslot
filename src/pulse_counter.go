package src

import (
	"sort"
	"time"
)

// Denomination maps a coin's pulse train length to its monetary value. A ₱5
// coin typically emits 5 pulses, etc. (acceptor-dependent, configurable).
type Denomination struct {
	Pulses int     `json:"pulses"`
	Amount float64 `json:"amount"`
}

// PulseCounter aggregates raw coin-acceptor pulses into coin amounts. A coin
// emits a burst of pulses; after an idle window with no further pulses, the
// accumulated count is resolved into an amount and reported via onCoin.
type PulseCounter struct {
	window     time.Duration
	denoms     []Denomination
	input      <-chan struct{}
	inject     chan struct{}
	onCoin     func(amount float64)
	onCounting func()
	stop       chan struct{}
}

// NewPulseCounter wires a pulse source to two callbacks: onCounting fires once on
// the leading edge of a coin's pulse burst (the moment counting begins, before
// the amount is known) so the UI can show an immediate "counting" cue; onCoin
// fires after the idle window resolves the burst into a monetary amount. Either
// callback may be nil.
func NewPulseCounter(window time.Duration, denoms []Denomination, input <-chan struct{}, onCoin func(amount float64), onCounting func()) *PulseCounter {
	return &PulseCounter{
		window:     window,
		denoms:     denoms,
		input:      input,
		inject:     make(chan struct{}, pulseBufferSize),
		onCoin:     onCoin,
		onCounting: onCounting,
		stop:       make(chan struct{}),
	}
}

const pulseBufferSize = 64

// Inject simulates a single coin pulse (used by the dev/mock route). Returns
// false if the buffer is momentarily full.
func (p *PulseCounter) Inject() bool {
	select {
	case p.inject <- struct{}{}:
		return true
	default:
		return false
	}
}

// Run consumes pulses until Stop is called. It is meant to run in its own
// goroutine. A single timer drives the idle-window resolution; there is no
// per-pulse goroutine churn.
func (p *PulseCounter) Run() {
	var count int
	timer := time.NewTimer(p.window)
	if !timer.Stop() {
		<-timer.C
	}

	for {
		select {
		case <-p.stop:
			timer.Stop()
			return
		case <-p.input:
			p.onPulse(&count, timer)
		case <-p.inject:
			p.onPulse(&count, timer)
		case <-timer.C:
			if count > 0 {
				amount := resolveAmount(count, p.denoms)
				count = 0
				if amount > 0 && p.onCoin != nil {
					p.onCoin(amount)
				}
			}
		}
	}
}

func (p *PulseCounter) Stop() {
	close(p.stop)
}

// =============================================================================
// HELPER FUNCTIONS (internal)
// =============================================================================

// onPulse accounts for one pulse and (re)arms the idle window. On the leading
// edge of a new burst (count 0 -> 1) it signals onCounting so the UI can show a
// "counting" cue before the amount is resolved.
func (p *PulseCounter) onPulse(count *int, timer *time.Timer) {
	if *count == 0 && p.onCounting != nil {
		p.onCounting()
	}
	*count++
	resetTimer(timer, p.window)
}

// resetTimer safely drains and resets the idle-window timer.
func resetTimer(timer *time.Timer, d time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(d)
}

// resolveAmount maps a pulse count to a peso amount using a greedy,
// largest-denomination-first decomposition (handles fast multi-coin bursts that
// land in the same window).
func resolveAmount(pulses int, denoms []Denomination) float64 {
	sorted := make([]Denomination, len(denoms))
	copy(sorted, denoms)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Pulses > sorted[j].Pulses })

	remaining := pulses
	var amount float64
	for _, d := range sorted {
		if d.Pulses <= 0 {
			continue
		}
		n := remaining / d.Pulses
		amount += float64(n) * d.Amount
		remaining -= n * d.Pulses
	}
	return amount
}
