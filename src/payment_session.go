package src

import (
	"context"
	"sync"
	"time"

	sdkapi "sdk/api"
)

// subscriberGrace is how long a session is kept alive after its last SSE
// subscriber disconnects, so a brief EventSource reconnect doesn't tear it down.
const subscriberGrace = 15 * time.Second

// RelayController is the slice of Manager that the session manager drives.
type RelayController interface {
	OpenRelay(coinslotID string)
	CloseRelay(coinslotID string)
	// CoinslotName returns the coinslot's configured display name, or "" if unknown.
	CoinslotName(coinslotID string) string
}

// CoinEvent is the SSE payload pushed to the insert-coin page on every coin.
//
// Counting marks the leading-edge "a coin's pulses are arriving" signal that
// precedes the resolved amount, so the page can show an immediate cue. Any event
// with Counting=true or LastCoin>0 (and the initial snapshot) carries the live
// TimeoutSecs, which the page uses to (re)start its idle countdown — the
// countdown therefore resets on every real pulse/coin.
type CoinEvent struct {
	Total       float64 `json:"total"`
	LastCoin    float64 `json:"last_coin"`
	Price       float64 `json:"price"`
	Sufficient  bool    `json:"sufficient"`
	Counting    bool    `json:"counting"`
	TimeoutSecs int     `json:"timeout_secs"`
}

// PaymentSessionManager tracks the at-most-one active paying client per coinslot
// and gates the relay accordingly: the relay is only open while a client is
// actively on the insert-coin page (has an SSE subscription).
type PaymentSessionManager struct {
	api    sdkapi.IPluginApi
	relays RelayController

	mu   sync.Mutex
	byID map[string]*paymentSession
}

type paymentSession struct {
	coinslotID  string
	clientID    int64
	purchase    sdkapi.IPurchaseRequest
	price       float64
	fixedPrice  bool
	timeoutSecs int

	mu         sync.Mutex
	total      float64
	subs       map[chan CoinEvent]struct{}
	graceTimer *time.Timer
}

func NewPaymentSessionManager(api sdkapi.IPluginApi, relays RelayController) *PaymentSessionManager {
	return &PaymentSessionManager{
		api:    api,
		relays: relays,
		byID:   map[string]*paymentSession{},
	}
}

// Begin starts (or resumes) a paying session for a client on a coinslot and
// opens the relay so coins are accepted. timeoutSecs is the page's idle
// countdown, broadcast to the client so it can auto-finalize when it elapses.
//
// It re-validates the claim independently of the caller: only the device that
// holds the coinslot claim (the atomic UsedCoinslots entry) may start/resume a
// session, and an existing session for a different client is never clobbered. It
// returns false if the caller does not own the claim or another client already
// holds the session — defense in depth behind the atomic claim in TryUseBy.
func (m *PaymentSessionManager) Begin(coinslotID string, clientID int64, purchase sdkapi.IPurchaseRequest, timeoutSecs int) bool {
	// Re-check ownership against the atomic claim map.
	if v, ok := UsedCoinslots.Load(coinslotID); !ok || v.(int64) != clientID {
		return false
	}

	m.mu.Lock()
	if s, ok := m.byID[coinslotID]; ok {
		if s.clientID != clientID {
			// A session already belongs to a different client — refuse rather
			// than overwrite it (should be unreachable given the atomic claim).
			m.mu.Unlock()
			return false
		}
		s.purchase = purchase
		s.timeoutSecs = timeoutSecs
		m.mu.Unlock()
		s.cancelGrace()
		m.relays.OpenRelay(coinslotID)
		return true
	}
	m.byID[coinslotID] = &paymentSession{
		coinslotID:  coinslotID,
		clientID:    clientID,
		purchase:    purchase,
		price:       purchase.Price(),
		fixedPrice:  purchase.IsFixedPrice(),
		timeoutSecs: timeoutSecs,
		subs:        map[chan CoinEvent]struct{}{},
	}
	m.mu.Unlock()
	m.relays.OpenRelay(coinslotID)
	return true
}

// Credit records an accepted coin against the active session's purchase and
// broadcasts the new running total to subscribers. Coins arriving with no active
// session are ignored (the relay should be closed in that state anyway).
func (m *PaymentSessionManager) Credit(coinslotID string, amount float64) {
	s := m.get(coinslotID)
	if s == nil {
		return
	}

	paymentMethod := m.relays.CoinslotName(coinslotID)
	if paymentMethod == "" {
		paymentMethod = "Coinslot"
	}

	if err := s.purchase.CreatePayment(context.Background(), sdkapi.CreatePaymentParams{
		Amount:        amount,
		PaymentMethod: paymentMethod,
	}); err != nil {
		_ = m.api.Logger().Error("[wired-coinslot] failed to record coin payment: " + err.Error())
		return
	}

	s.mu.Lock()
	s.total += amount
	ev := s.snapshotLocked(amount)
	s.broadcastLocked(ev)
	s.mu.Unlock()
}

// Counting signals that a coin's pulse train has started arriving but its amount
// isn't resolved yet. It broadcasts a Counting event (no money recorded) so the
// page can show an immediate "counting" cue and reset its idle countdown on the
// very first pulse, well before the resolved coin value follows.
func (m *PaymentSessionManager) Counting(coinslotID string) {
	s := m.get(coinslotID)
	if s == nil {
		return
	}
	s.mu.Lock()
	ev := s.snapshotLocked(0)
	ev.Counting = true
	s.broadcastLocked(ev)
	s.mu.Unlock()
}

// Subscribe attaches an SSE listener for clientID. The returned unsubscribe func
// must be called when the connection ends; when the last subscriber leaves, the
// relay is closed and a grace timer is armed to end the session. It re-validates
// ownership: only the client that owns the session may subscribe (and thus hold
// the relay open), so a racing/stale request for someone else's session is
// refused with ok=false.
func (m *PaymentSessionManager) Subscribe(coinslotID string, clientID int64) (<-chan CoinEvent, func(), bool) {
	s := m.get(coinslotID)
	if s == nil || s.clientID != clientID {
		return nil, nil, false
	}

	ch := make(chan CoinEvent, 8)
	s.mu.Lock()
	s.subs[ch] = struct{}{}
	s.cancelGraceLocked()
	s.mu.Unlock()

	m.relays.OpenRelay(coinslotID)

	unsub := func() {
		s.mu.Lock()
		delete(s.subs, ch)
		empty := len(s.subs) == 0
		if empty {
			s.startGraceLocked(func() { m.End(coinslotID) })
		}
		s.mu.Unlock()
		if empty {
			m.relays.CloseRelay(coinslotID)
		}
	}
	return ch, unsub, true
}

// Snapshot returns the current payment state for a coinslot's session.
func (m *PaymentSessionManager) Snapshot(coinslotID string) (CoinEvent, bool) {
	s := m.get(coinslotID)
	if s == nil {
		return CoinEvent{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.snapshotLocked(0), true
}

// End tears down a session: closes the relay, releases the coinslot, and closes
// any open SSE channels. Idempotent.
func (m *PaymentSessionManager) End(coinslotID string) {
	m.mu.Lock()
	s := m.byID[coinslotID]
	delete(m.byID, coinslotID)
	m.mu.Unlock()
	if s == nil {
		return
	}

	s.mu.Lock()
	s.cancelGraceLocked()
	for ch := range s.subs {
		close(ch)
		delete(s.subs, ch)
	}
	s.mu.Unlock()

	m.relays.CloseRelay(coinslotID)
	UsedCoinslots.Delete(coinslotID)
}

// =============================================================================
// HELPER FUNCTIONS (internal)
// =============================================================================

func (m *PaymentSessionManager) get(coinslotID string) *paymentSession {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.byID[coinslotID]
}

func (s *paymentSession) snapshotLocked(lastCoin float64) CoinEvent {
	return CoinEvent{
		Total:       s.total,
		LastCoin:    lastCoin,
		Price:       s.price,
		Sufficient:  s.sufficientLocked(),
		TimeoutSecs: s.timeoutSecs,
	}
}

func (s *paymentSession) sufficientLocked() bool {
	if s.fixedPrice {
		return s.total >= s.price
	}
	return s.total > 0
}

func (s *paymentSession) broadcastLocked(ev CoinEvent) {
	for ch := range s.subs {
		select {
		case ch <- ev:
		default: // drop if a subscriber is momentarily behind
		}
	}
}

func (s *paymentSession) startGraceLocked(fn func()) {
	if s.graceTimer != nil {
		s.graceTimer.Stop()
	}
	s.graceTimer = time.AfterFunc(subscriberGrace, fn)
}

func (s *paymentSession) cancelGraceLocked() {
	if s.graceTimer != nil {
		s.graceTimer.Stop()
		s.graceTimer = nil
	}
}

func (s *paymentSession) cancelGrace() {
	s.mu.Lock()
	s.cancelGraceLocked()
	s.mu.Unlock()
}
