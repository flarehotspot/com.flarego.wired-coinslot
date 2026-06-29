package src

import (
	"sync/atomic"
	"testing"
	"time"

	"com.flarego.wired-coinslot/src/gpio"
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

func TestDetectBoard(t *testing.T) {
	cases := map[string]struct {
		driver    string
		chipLabel string
	}{
		"orangepi-one":    {"gpiod", "1c20800"},
		"orangepi-pc":     {"gpiod", "1c20800"},
		"orangepi-zero-3": {"gpiod", "300b000"},
		"rpi-4":           {"rpi", ""},
		"unknown-device":  {"rpi", ""}, // unknown -> RPi.GPIO fallback
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
