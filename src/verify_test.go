package src

import (
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
	pc := NewPulseCounter(40*time.Millisecond, DefaultDenominations(), input, func(amount float64) {
		got <- amount
	})
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
}

func TestDetectBoard(t *testing.T) {
	cases := map[string]gpio.Board{
		"orangepi-one":    {Library: "opi", OpiBoard: "one"},
		"orangepi-zero-3": {Library: "opi", OpiBoard: "zero3"},
		"rpi-4":           {Library: "rpi"},
		"unknown-device":  {Library: "rpi"},
	}
	for model, want := range cases {
		if got := gpio.DetectBoard(model); got != want {
			t.Errorf("DetectBoard(%q) = %+v, want %+v", model, got, want)
		}
	}
}
