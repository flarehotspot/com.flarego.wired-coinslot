package src

import (
	"errors"
	"os"
	"path/filepath"
	sdkapi "sdk/api"
	"sync"

	sdkutils "github.com/flarewifi/sdk-utils"
	"github.com/goccy/go-json"
)

const (
	WiredCoinslotsPrefix string = "wired_coinslots"

	// Hardware defaults. Pin numbers are physical header pins (BOARD numbering)
	// for every driver: pin #3 = coin-acceptor pulse input, pin #5 = relay
	// output. The gpiod driver translates these to char-device line offsets
	// internally (see gpio.Board.Header), so the UI stays in physical pins.
	DefaultCoinPin     = 3
	DefaultRelayPin    = 5
	DefaultRelayActive = 1 // relay value that energizes the coil / accepts coins
	DefaultPull        = "up"
	DefaultEdge        = "falling"
	DefaultDebounceMs  = 30
	DefaultWindowMs    = 400

	// DefaultPaymentTimeoutSecs is the idle countdown shown on the insert-coin
	// page. It resets on every coin/pulse; on expiry the page auto-finalizes
	// (executes the accumulated payment) or cancels if nothing was inserted.
	DefaultPaymentTimeoutSecs = 30
)

// DefaultDenominations covers the common Philippine coin set where the acceptor
// emits one pulse per peso.
func DefaultDenominations() []Denomination {
	return []Denomination{
		{Pulses: 1, Amount: 1},
		{Pulses: 5, Amount: 5},
		{Pulses: 10, Amount: 10},
	}
}

var (
	UsedCoinslots sync.Map
)

func InitWiredCoinslots(api sdkapi.IPluginApi) {
	entries, err := api.Config().Plugin().List(WiredCoinslotsPrefix)
	// A missing config dir is the expected "fresh install" case; any other
	// error means we can't safely tell whether coinslots exist, so don't seed
	// (seeding on top of unreadable data could create a duplicate).
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		_ = api.Logger().Error("[wired-coinslot] failed to list coinslots during init: " + err.Error())
		return
	}
	// Seed the default "Main Vendo" whenever there are zero coinslots — covers
	// both a missing dir (ErrNotExist) and a present-but-empty dir, which some
	// config backends produce. Keying only on ErrNotExist (the previous
	// behavior) silently skipped seeding on present-but-empty dirs, leaving the
	// plugin with no coinslots, no payment option, and an empty settings page.
	if len(entries) > 0 {
		return
	}
	mainVendo := api.Translate("label", "Main Vendo")
	mainCoinslot := NewWiredCoinslot(api, mainVendo)
	if err := mainCoinslot.Save(); err != nil {
		_ = api.Logger().Error("[wired-coinslot] failed to seed default coinslot: " + err.Error())
	}
}

func NewWiredCoinslot(api sdkapi.IPluginApi, name string) *WiredCoinslot {
	c := &WiredCoinslot{
		api:  api,
		ID:   sdkutils.RandomStr(16),
		Name: name,
	}
	c.ApplyDefaults()
	return c
}

func GetAllWiredCoinslots(api sdkapi.IPluginApi) ([]*WiredCoinslot, error) {
	coinslotEntries, err := api.Config().Plugin().List(WiredCoinslotsPrefix)
	if err != nil {
		// No config dir yet means "no coinslots", not a failure — return an
		// empty list so callers (payment options, settings page) render cleanly.
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	// Append only successfully-parsed entries so a single corrupt config can
	// never leave a nil hole in the slice (which would panic nil-unsafe callers).
	coinslots := make([]*WiredCoinslot, 0, len(coinslotEntries))
	for _, entry := range coinslotEntries {
		b, err := api.Config().Plugin().Read(entry.Path)
		if err != nil {
			_ = api.Logger().Error("[wired-coinslot] failed to read coinslot config: " + err.Error())
			continue
		}

		var c WiredCoinslot
		if err := json.Unmarshal(b, &c); err != nil {
			_ = api.Logger().Error("[wired-coinslot] failed to parse coinslot config: " + err.Error())
			continue
		}

		c.api = api
		c.ApplyDefaults()
		coinslots = append(coinslots, &c)
	}

	return coinslots, nil
}

func FindUsedCoinslot(api sdkapi.IPluginApi, deviceID int64) (*WiredCoinslot, error) {
	var coinslotID string
	UsedCoinslots.Range(func(key, value any) bool {
		if value.(int64) == deviceID {
			coinslotID = key.(string)
			return false
		}
		return true
	})

	if coinslotID == "" {
		return nil, nil
	}

	return LoadWiredCoinslot(api, coinslotID)
}

func LoadWiredCoinslot(api sdkapi.IPluginApi, coinslotID string) (*WiredCoinslot, error) {
	var c WiredCoinslot
	b, err := api.Config().Plugin().Read(filepath.Join(WiredCoinslotsPrefix, coinslotID))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(b), &c); err != nil {
		return nil, err
	}

	c.api = api
	c.ApplyDefaults()
	return &c, nil
}

type WiredCoinslot struct {
	api  sdkapi.IPluginApi
	ID   string
	Name string

	// Alias is an optional, customer-facing display name (max 16 chars) shown as
	// the payment method in the sales inventory. Falls back to "Coinslot" when unset.
	Alias string

	// Hardware configuration (physical BOARD pin numbers). The gpiod driver
	// translates these to char-device line offsets internally, so the same
	// physical-pin addressing is used for every board.
	CoinPin     int    // coin-acceptor pulse input pin
	RelayPin    int    // relay output pin
	RelayActive int    // output value (0/1) that energizes the relay
	Pull        string // input bias for the coin pin: "up" | "down"
	Edge        string // pulse edge to count: "falling" | "rising" | "both"
	DebounceMs  int    // hardware debounce for the coin pin
	WindowMs    int    // idle window (ms) to finish counting a coin's pulses

	// PaymentTimeoutSecs is the insert-coin page's idle countdown (seconds). It
	// resets on each coin/pulse; on expiry the page auto-finalizes the payment
	// (or cancels if nothing was inserted).
	PaymentTimeoutSecs int

	// Board selection override. Empty falls back to auto-detection from
	// /etc/os_release.json device_model. When set to a known model, the board's
	// GPIO driver (rpi/opi/gpiod) and parameters are resolved from the registry.
	BoardModel string // override device_model key (e.g. "orangepi-zero-3")

	Denominations []Denomination
}

// ApplyDefaults fills unset hardware fields. A coinslot saved by the original
// scaffold has no hardware block at all, so when nothing is set we apply the
// full default set (this also avoids the active-low ambiguity of RelayActive==0,
// which is only honored once the config has been saved by the new code).
func (c *WiredCoinslot) ApplyDefaults() {
	legacy := c.CoinPin == 0 && c.RelayPin == 0 && len(c.Denominations) == 0
	if legacy {
		c.CoinPin = DefaultCoinPin
		c.RelayPin = DefaultRelayPin
		c.RelayActive = DefaultRelayActive
		c.Pull = DefaultPull
		c.Edge = DefaultEdge
		c.DebounceMs = DefaultDebounceMs
		c.WindowMs = DefaultWindowMs
		c.PaymentTimeoutSecs = DefaultPaymentTimeoutSecs
		c.Denominations = DefaultDenominations()
		return
	}
	if c.CoinPin == 0 {
		c.CoinPin = DefaultCoinPin
	}
	if c.RelayPin == 0 {
		c.RelayPin = DefaultRelayPin
	}
	if c.Pull == "" {
		c.Pull = DefaultPull
	}
	if c.Edge == "" {
		c.Edge = DefaultEdge
	}
	if c.DebounceMs == 0 {
		c.DebounceMs = DefaultDebounceMs
	}
	if c.WindowMs == 0 {
		c.WindowMs = DefaultWindowMs
	}
	if c.PaymentTimeoutSecs == 0 {
		c.PaymentTimeoutSecs = DefaultPaymentTimeoutSecs
	}
	if len(c.Denominations) == 0 {
		c.Denominations = DefaultDenominations()
	}
}

func (c *WiredCoinslot) ConfigPath() string {
	return filepath.Join(WiredCoinslotsPrefix, c.ID)
}

func (c *WiredCoinslot) GetID() string {
	return c.ID
}

func (c *WiredCoinslot) GetName() string {
	return c.Name
}

// PaymentMethod returns the coinslot's customer-facing alias for display in the
// sales inventory, falling back to a plain "Coinslot" label when no alias is set.
func (c *WiredCoinslot) PaymentMethod() string {
	if c.Alias != "" {
		return c.Alias
	}
	return "Coinslot"
}

// TryUseBy atomically claims the coinslot for deviceID and reports whether the
// caller now holds it. LoadOrStore makes the check-and-claim a single atomic
// step, closing the check-then-act race that a separate CanBeUsedBy()+UseBy()
// left open (two clients could both see "free" before either stored). It returns
// true when the slot was free (claimed now) or already held by the same device
// (idempotent re-entry, e.g. a page refresh), and false when a different device
// currently holds it.
func (c *WiredCoinslot) TryUseBy(deviceID int64) bool {
	actual, _ := UsedCoinslots.LoadOrStore(c.ID, deviceID)
	return actual.(int64) == deviceID
}

// IsUsedBy reports whether deviceID currently holds this coinslot's claim. It is
// the read-only ownership check used to re-validate a claim without mutating it.
func (c *WiredCoinslot) IsUsedBy(deviceID int64) bool {
	v, ok := UsedCoinslots.Load(c.ID)
	return ok && v.(int64) == deviceID
}

func (c *WiredCoinslot) DoneUsing() {
	UsedCoinslots.Delete(c.ID)
}

// ReleaseIfOwner deletes this coinslot's claim only if deviceID still holds it.
// It lets a caller safely back out of its own claim without risking deleting a
// claim that has meanwhile been taken by another device (CompareAndDelete is a
// no-op when the current value differs).
func (c *WiredCoinslot) ReleaseIfOwner(deviceID int64) {
	UsedCoinslots.CompareAndDelete(c.ID, deviceID)
}

func (c *WiredCoinslot) Save() error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return c.api.Config().Plugin().Write(c.ConfigPath(), b)
}
