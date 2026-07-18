package src

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"com.flarego.wired-coinslot/src/gpio"
	sdkutils "github.com/flarewifi/sdk-utils/v2"
)

func TestResolveAmount(t *testing.T) {
	denoms := DefaultDenominations()
	cases := map[int]float64{
		1:  1,
		5:  5,
		10: 10,
		6:  6,  // 5 + 1
		11: 11, // 10 + 1
		16: 16, // 10 + 5 + 1
	}
	for pulses, want := range cases {
		if got := resolveAmount(pulses, denoms); got != want {
			t.Errorf("resolveAmount(%d) = %v, want %v", pulses, got, want)
		}
	}
}

func TestSortDenominations(t *testing.T) {
	denoms := []Denomination{{Pulses: 10, Amount: 10}, {Pulses: 1, Amount: 1}, {Pulses: 5, Amount: 5}}
	sortDenominations(denoms)
	want := []int{1, 5, 10}
	for i, d := range denoms {
		if d.Pulses != want[i] {
			t.Errorf("sortDenominations[%d].Pulses = %d, want %d", i, d.Pulses, want[i])
		}
	}
}

func TestPulseCounterAggregation(t *testing.T) {
	input := make(chan struct{}, 16)
	got := make(chan float64, 4)
	var countingCalls int32
	pc := NewPulseCounter(40*time.Millisecond, DefaultDenominations(), input,
		func(amount float64) { got <- amount },
		func() { atomic.AddInt32(&countingCalls, 1) },
	)
	go pc.Run()
	defer pc.Stop()

	// Burst of 5 pulses = one ₱5 coin.
	for i := 0; i < 5; i++ {
		input <- struct{}{}
	}

	select {
	case amount := <-got:
		if amount != 5 {
			t.Errorf("expected amount 5, got %v", amount)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for coin amount")
	}

	// onCounting must fire exactly once for the burst — on the leading edge (the
	// first pulse), not per pulse — so the page shows the "counting" cue and
	// resets its idle countdown the moment a coin starts arriving.
	if c := atomic.LoadInt32(&countingCalls); c != 1 {
		t.Errorf("expected onCounting to fire once on the leading edge, got %d", c)
	}
}

func TestTryUseByOwnership(t *testing.T) {
	c := &WiredCoinslot{ID: "own-slot-" + sdkutils.RandomStr(6)}
	defer c.DoneUsing()

	if !c.TryUseBy(1) {
		t.Fatal("first claim should succeed")
	}
	if !c.TryUseBy(1) {
		t.Fatal("same-device re-claim should be idempotent (true)")
	}
	if c.TryUseBy(2) {
		t.Fatal("a different device must not steal an existing claim")
	}
	if !c.IsUsedBy(1) || c.IsUsedBy(2) {
		t.Fatal("device 1 must own the claim, device 2 must not")
	}

	// Releasing as the wrong owner must be a no-op (CompareAndDelete).
	c.ReleaseIfOwner(2)
	if !c.IsUsedBy(1) {
		t.Fatal("release by a non-owner must not free the claim")
	}
	// Releasing as the owner frees it for the next device.
	c.ReleaseIfOwner(1)
	if c.IsUsedBy(1) {
		t.Fatal("owner release should free the claim")
	}
	if !c.TryUseBy(2) {
		t.Fatal("after release, another device can claim")
	}
}

// TestTryUseByConcurrentSingleWinner is the regression guard for the original
// check-then-act race: many devices race to claim the same coinslot at once and
// exactly one must win (LoadOrStore is atomic, so there is no window where two
// callers both observe "free").
func TestTryUseByConcurrentSingleWinner(t *testing.T) {
	c := &WiredCoinslot{ID: "race-slot-" + sdkutils.RandomStr(6)}
	defer c.DoneUsing()

	const n = 64
	var wins int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			<-start // line everyone up so the claims truly race
			if c.TryUseBy(id) {
				atomic.AddInt32(&wins, 1)
			}
		}(int64(i + 1))
	}
	close(start)
	wg.Wait()

	if wins != 1 {
		t.Errorf("expected exactly one winner, got %d", wins)
	}
}

func TestDetectBoard(t *testing.T) {
	cases := map[string]struct {
		driver    string
		chipLabel string
	}{
		"orangepi-one":    {"gpiod", "1c20800"},
		"orangepi-pc":     {"gpiod", "1c20800"},
		"orangepi-zero-3": {"gpiod", "300b000"},
		"rpi-4":           {"gpiod", "pinctrl-bcm2711"}, // bcm2711 char device, not RPi.GPIO
		"unknown-device":  {"rpi", ""},                  // unknown -> RPi.GPIO fallback
	}
	for model, want := range cases {
		got := gpio.DetectBoard(model)
		// Board has a map field, so it isn't ==-comparable; check fields directly.
		if got.Driver != want.driver || got.ChipLabel != want.chipLabel {
			t.Errorf("DetectBoard(%q) = {Driver:%q ChipLabel:%q}, want {Driver:%q ChipLabel:%q}",
				model, got.Driver, got.ChipLabel, want.driver, want.chipLabel)
		}
		// gpiod boards must expose a header that covers the default coin/relay pins
		// so physical-pin -> line-offset resolution can succeed.
		if want.driver == "gpiod" {
			if got.Header[DefaultCoinPin] == "" || got.Header[DefaultRelayPin] == "" {
				t.Errorf("DetectBoard(%q) header missing default coin/relay pins: %+v", model, got.Header)
			}
		}
	}
}
