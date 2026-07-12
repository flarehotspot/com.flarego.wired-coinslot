package gpio

import "testing"

// TestBcmOffset covers the Raspberry Pi BCM line-name -> char-device offset
// conversion. On pinctrl-bcm2711 the offset equals the BCM GPIO number.
func TestBcmOffset(t *testing.T) {
	ok := map[string]int{
		"GPIO0":  0,
		"GPIO2":  2,  // physical pin 3 (default coin)
		"GPIO3":  3,  // physical pin 5 (default relay)
		"GPIO17": 17,
		"GPIO27": 27,
		"gpio21": 21, // case-insensitive
		"GPIO57": 57, // last line on the 58-line chip
	}
	for name, want := range ok {
		got, err := bcmOffset(name)
		if err != nil {
			t.Errorf("bcmOffset(%q) unexpected error: %v", name, err)
			continue
		}
		if got != want {
			t.Errorf("bcmOffset(%q) = %d, want %d", name, got, want)
		}
	}

	for _, bad := range []string{"", "GPIO", "GPIO58", "GPIO-1", "PH5", "GPIOxx"} {
		if _, err := bcmOffset(bad); err == nil {
			t.Errorf("bcmOffset(%q) expected error, got nil", bad)
		}
	}
}

// TestRpiHeaderResolves guards the rpi-4 registry entry: every physical header
// pin must resolve through the "bcm" scheme to the expected BCM GPIO number, so
// a typo in rpiBcm2711Header (which addresses real relay/coin lines) is caught
// at build time rather than on-device.
func TestRpiHeaderResolves(t *testing.T) {
	// Reference physical-pin -> BCM number for the Pi 40-pin (J8) header.
	wantBCM := map[int]int{
		3: 2, 5: 3, 7: 4, 8: 14, 10: 15, 11: 17, 12: 18, 13: 27, 15: 22,
		16: 23, 18: 24, 19: 10, 21: 9, 22: 25, 23: 11, 24: 8, 26: 7, 27: 0,
		28: 1, 29: 5, 31: 6, 32: 12, 33: 13, 35: 19, 36: 16, 37: 26, 38: 20, 40: 21,
	}

	board := DetectBoard("rpi-4")
	if board.Driver != "gpiod" || board.Scheme != "bcm" {
		t.Fatalf("rpi-4 board = {Driver:%q Scheme:%q}, want {gpiod bcm}", board.Driver, board.Scheme)
	}
	if len(board.Header) != len(wantBCM) {
		t.Errorf("rpi-4 header has %d pins, want %d", len(board.Header), len(wantBCM))
	}

	a := &CdevAgent{cfg: Config{Scheme: board.Scheme, Header: board.Header}}
	for pin, bcm := range wantBCM {
		_, offset, err := a.resolvePin(pin)
		if err != nil {
			t.Errorf("resolvePin(pin %d) error: %v", pin, err)
			continue
		}
		if offset != bcm {
			t.Errorf("resolvePin(pin %d) offset = %d, want BCM %d", pin, offset, bcm)
		}
	}
}

// TestSunxiOffsetUnchanged is a regression guard that the shared scheme switch
// did not disturb the existing Allwinner path.
func TestSunxiOffsetUnchanged(t *testing.T) {
	cases := map[string]int{
		"PA12": 12,  // OrangePi H3 physical pin 3
		"PC5":  69,  // validated in sunxiOffset doc
		"PH5":  229, // ('H'-'A')*32 + 5
	}
	for port, want := range cases {
		got, err := sunxiOffset(port)
		if err != nil {
			t.Errorf("sunxiOffset(%q) error: %v", port, err)
			continue
		}
		if got != want {
			t.Errorf("sunxiOffset(%q) = %d, want %d", port, got, want)
		}
	}
}
