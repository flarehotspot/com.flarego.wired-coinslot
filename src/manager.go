package src

import (
	"sync"
	"time"

	sdkapi "sdk/api"

	"com.flarego.wired-coinslot/src/gpio"
	sdkutils "github.com/flarewifi/sdk-utils"
)

const osReleaseFile = "/etc/os_release.json"

// coinslotRuntime bundles the live hardware objects for one coinslot.
type coinslotRuntime struct {
	coinslot *WiredCoinslot
	agent    gpio.CoinAgent
	counter  *PulseCounter
}

// Manager owns the per-coinslot GPIO agents + pulse counters and the payment
// session manager. It is the single entry point the HTTP handlers use to open
// or close relays. There is one Manager per plugin process.
type Manager struct {
	api         sdkapi.IPluginApi
	deviceModel string

	mu       sync.RWMutex
	runtimes map[string]*coinslotRuntime

	sessions *PaymentSessionManager
}

var (
	managerMu sync.RWMutex
	manager   *Manager
)

// StartManager builds the runtime for every configured coinslot and stores the
// process-wide singleton. Safe to call once from Init.
func StartManager(api sdkapi.IPluginApi) *Manager {
	m := &Manager{
		api:      api,
		runtimes: map[string]*coinslotRuntime{},
	}
	m.sessions = NewPaymentSessionManager(api, m)

	if release, err := sdkutils.ReadOsRelease(osReleaseFile); err == nil {
		m.deviceModel = release.DeviceModel
	} else {
		_ = api.Logger().Error("[wired-coinslot] unable to read os_release for board detection: " + err.Error())
	}

	managerMu.Lock()
	manager = m
	managerMu.Unlock()

	m.startAll()
	return m
}

// GetManager returns the running Manager, or nil if StartManager hasn't run.
func GetManager() *Manager {
	managerMu.RLock()
	defer managerMu.RUnlock()
	return manager
}

// Sessions exposes the payment session manager to the HTTP handlers.
func (m *Manager) Sessions() *PaymentSessionManager { return m.sessions }

// DeviceModel returns the detected os_release device_model (may be empty).
func (m *Manager) DeviceModel() string { return m.deviceModel }

// OpenRelay energizes the coinslot's relay so coins are accepted.
func (m *Manager) OpenRelay(coinslotID string) {
	if rt := m.runtime(coinslotID); rt != nil {
		rt.agent.OpenRelay()
	}
}

// CloseRelay de-energizes the coinslot's relay so coins are rejected.
func (m *Manager) CloseRelay(coinslotID string) {
	if rt := m.runtime(coinslotID); rt != nil {
		rt.agent.CloseRelay()
	}
}

// InjectPulse feeds a synthetic pulse into a coinslot's counter. Used by the
// dev/mock route to exercise the full pipeline without real hardware.
func (m *Manager) InjectPulse(coinslotID string) bool {
	rt := m.runtime(coinslotID)
	if rt == nil || rt.counter == nil {
		return false
	}
	return rt.counter.Inject()
}

// Reload rebuilds a single coinslot's runtime after its config changed.
func (m *Manager) Reload(coinslotID string) {
	m.stopCoinslot(coinslotID)
	c, err := LoadWiredCoinslot(m.api, coinslotID)
	if err != nil {
		_ = m.api.Logger().Error("[wired-coinslot] reload failed: " + err.Error())
		return
	}
	m.startCoinslot(c)
}

// =============================================================================
// HELPER FUNCTIONS (internal)
// =============================================================================

func (m *Manager) startAll() {
	coinslots, err := GetAllWiredCoinslots(m.api)
	if err != nil {
		_ = m.api.Logger().Error("[wired-coinslot] unable to load coinslots: " + err.Error())
		return
	}
	for _, c := range coinslots {
		if c == nil {
			continue
		}
		m.startCoinslot(c)
	}
}

func (m *Manager) startCoinslot(c *WiredCoinslot) {
	board := gpio.DetectBoard(m.modelFor(c))

	cfg := gpio.Config{
		Driver:      board.Driver,
		Board:       board.OpiBoard,
		CoinPin:     c.CoinPin,
		RelayPin:    c.RelayPin,
		Pull:        c.Pull,
		Edge:        c.Edge,
		DebounceMs:  c.DebounceMs,
		RelayActive: c.RelayActive,
	}
	if board.Driver == "gpiod" {
		// Char-device driver: resolve the same physical pins to line offsets
		// internally via the board's header map + chip (matched by label).
		cfg.ChipLabel = board.ChipLabel
		cfg.Header = board.Header
	} else {
		// Python rpi/opi drivers use BOARD pin numbers directly; Library is what
		// coin_agent.py imports ("rpi" or "opi").
		cfg.Library = board.Driver
	}

	agent := gpio.NewCoinAgent(cfg, m.api.Logger())
	if err := agent.Start(); err != nil {
		_ = m.api.Logger().Error("[wired-coinslot] failed to start gpio agent for " + c.Name + ": " + err.Error())
		return
	}

	coinslotID := c.ID
	counter := NewPulseCounter(
		time.Duration(c.WindowMs)*time.Millisecond,
		c.Denominations,
		agent.Pulses(),
		func(amount float64) { m.sessions.Credit(coinslotID, amount) },
		func() { m.sessions.Counting(coinslotID) },
	)
	go counter.Run()

	// Relay starts closed: coins are only accepted while a client is paying.
	agent.CloseRelay()

	m.mu.Lock()
	m.runtimes[coinslotID] = &coinslotRuntime{coinslot: c, agent: agent, counter: counter}
	m.mu.Unlock()
}

func (m *Manager) stopCoinslot(coinslotID string) {
	m.mu.Lock()
	rt := m.runtimes[coinslotID]
	delete(m.runtimes, coinslotID)
	m.mu.Unlock()
	if rt == nil {
		return
	}
	rt.agent.Stop()
	rt.counter.Stop()
}

func (m *Manager) runtime(coinslotID string) *coinslotRuntime {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.runtimes[coinslotID]
}

func (m *Manager) modelFor(c *WiredCoinslot) string {
	if c.BoardModel != "" {
		return c.BoardModel
	}
	return m.deviceModel
}
