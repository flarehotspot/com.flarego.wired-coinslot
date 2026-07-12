package gpio

import "sort"

// Board describes how a given single-board computer's GPIO header is reached.
// Two driver families exist:
//
//   - "rpi"/"opi": the sysfs-based Python libraries (RPi.GPIO / OPi.GPIO). Both
//     use BOARD (physical header pin) numbering, so the coin/relay pins are
//     passed straight through and only the library + board module differ. Kept
//     only as a last-resort fallback for unrecognized boards — every board in
//     the registry now uses "gpiod", because the sysfs interface these Python
//     libraries drive is deprecated/removed on modern kernels (Linux 6.1+),
//     where it no longer fires edge events and renumbers the pins.
//   - "gpiod": the Linux GPIO character device, driven in-process by the pure-Go
//     go-gpiocdev library (no Python, no pip). Lines are addressed by
//     (chip-label, line-offset); the offset is derived from the header entry
//     per the board's Scheme (Allwinner port name for sunxi, BCM number for the
//     Raspberry Pi) because these kernels leave individual line names unset.
type Board struct {
	Driver    string         // "rpi" | "opi" | "gpiod"
	Scheme    string         // "sunxi" | "bcm": how a Header entry -> line offset; gpiod only
	OpiBoard  string         // OPi.GPIO board module name (orangepi.<name>); opi only
	ChipLabel string         // pinctrl label fragment to match a chip's label (substring); gpiod only
	Header    map[int]string // physical header pin -> line name (Allwinner port "PH5" for sunxi, "GPIO17" for bcm); gpiod only
}

// boardRegistry is keyed by device_model (IMachineApi.DeviceModel, decrypted from
// core/product.json — see manager.go's StartManager), which equals the device
// folder name under go/builder/imagebuilder/devices/. The image variant
// (device_config) is irrelevant to GPIO and intentionally ignored.
//
// Adding a future char-device board is a single line: give it Driver "gpiod"
// and the pinctrl Label of the chip its header GPIOs live on (read with
// `gpioinfo` / `gpiodetect` on the target).
var boardRegistry = map[string]Board{
	// Allwinner H3. Uses the gpiod char-device driver (not OPi.GPIO): OPi.GPIO is
	// a PyPI package that the on-device preinstall can only fetch with internet,
	// so an offline coin-vendo box fails to install it. gpiod is compiled into
	// the plugin — no pip, no network. The H3 header GPIOs (banks PA/PC/PD/PG)
	// live on the "pio" controller at register 1c20800. One and PC share the
	// same H3 40-pin header map.
	"orangepi-one": {Driver: "gpiod", Scheme: "sunxi", ChipLabel: "1c20800", Header: orangePiH3Header},
	"orangepi-pc":  {Driver: "gpiod", Scheme: "sunxi", ChipLabel: "1c20800", Header: orangePiH3Header},
	// Allwinner H618. OPi.GPIO has no zero3 board module and relies on the
	// removed sysfs interface, so the Zero 3 uses the gpiod char-device driver.
	// Its 26-pin header GPIOs live on the "pio" controller (banks PC/PH) at
	// register 300b000. Header maps physical pin -> port name so the admin UI
	// can stay in familiar BOARD (physical pin) numbers.
	"orangepi-zero-3": {Driver: "gpiod", Scheme: "sunxi", ChipLabel: "300b000", Header: orangePiZero3Header},
	// Broadcom BCM2711 (Raspberry Pi 4). Uses the gpiod char-device driver, NOT
	// RPi.GPIO: RPi.GPIO drives the sysfs GPIO interface, which is deprecated and
	// effectively removed on the modern kernel this board's OpenWRT image ships
	// (bcm27xx/bcm2711, Linux 6.x) — edge events stop firing and pins get
	// renumbered, so coin pulses are never seen. The 40-pin header GPIOs are on
	// gpiochip0, labeled "pinctrl-bcm2711" (58 lines), where the line offset is
	// exactly the BCM GPIO number (Scheme "bcm"). Header maps physical pin ->
	// "GPIO<bcm>" so the admin UI stays in familiar BOARD (physical pin) numbers.
	"rpi-4": {Driver: "gpiod", Scheme: "bcm", ChipLabel: "pinctrl-bcm2711", Header: rpiBcm2711Header},
}

// orangePiH3Header maps the Allwinner H3 40-pin header's physical pin numbers
// to their Allwinner port names, shared by the OrangePi One and PC (identical
// header). Derived from OPi.GPIO's orangepi.pc.BOARD map (global GPIO number ->
// port name, e.g. 12=PA12, 110=PD14, 68=PC4, 200=PG8), so it reproduces exactly
// the pin assignments the old OPi.GPIO driver used. Power/ground pins omitted.
var orangePiH3Header = map[int]string{
	3:  "PA12",
	5:  "PA11",
	7:  "PA6",
	8:  "PA13",
	10: "PA14",
	11: "PA1",
	12: "PD14",
	13: "PA0",
	15: "PA3",
	16: "PC4",
	18: "PC7",
	19: "PC0",
	21: "PC1",
	22: "PA2",
	23: "PC2",
	24: "PC3",
	26: "PA21",
	27: "PA19",
	28: "PA18",
	29: "PA7",
	31: "PA8",
	32: "PG8",
	33: "PA9",
	35: "PA10",
	36: "PG9",
	37: "PA20",
	38: "PG6",
	40: "PG7",
}

// orangePiZero3Header maps the OrangePi Zero 3 26-pin (DIP26) header's physical
// pin numbers to their Allwinner port names. Only GPIO-capable pins are listed;
// power (1,2,4,17) and ground (6,9,14,20,25) pins are omitted. This is the
// board's fixed PCB layout (not a kernel assumption), so hardcoding it is safe.
// Special-function silk: 3/5 = I2C3 SDA/SCL, 8/10 = UART5 TX/RX,
// 19/21/23/24 = SPI1 MOSI/MISO/CLK/CS. Verify against your actual harness.
var orangePiZero3Header = map[int]string{
	3:  "PH5",
	5:  "PH4",
	7:  "PC9",
	8:  "PH2",
	10: "PH3",
	11: "PC6",
	12: "PC11",
	13: "PC5",
	15: "PC8",
	16: "PC15",
	18: "PC14",
	19: "PH7",
	21: "PH8",
	22: "PC7",
	23: "PH6",
	24: "PH9",
	26: "PC10",
}

// rpiBcm2711Header maps the Raspberry Pi 40-pin (J8) header's physical pin
// numbers to their BCM GPIO line names on the "pinctrl-bcm2711" char-device
// chip (gpiochip0). Only GPIO-capable pins are listed; power (1,2,4,17) and
// ground (6,9,14,20,25,30,34,39) pins are omitted. This is the Pi's fixed PCB
// layout (not a kernel assumption), so hardcoding it is safe. The "GPIO<n>"
// name doubles as the line offset under Scheme "bcm": on bcm2711 the line offset
// equals the BCM GPIO number (bcmOffset just parses the trailing number).
var rpiBcm2711Header = map[int]string{
	3:  "GPIO2",
	5:  "GPIO3",
	7:  "GPIO4",
	8:  "GPIO14",
	10: "GPIO15",
	11: "GPIO17",
	12: "GPIO18",
	13: "GPIO27",
	15: "GPIO22",
	16: "GPIO23",
	18: "GPIO24",
	19: "GPIO10",
	21: "GPIO9",
	22: "GPIO25",
	23: "GPIO11",
	24: "GPIO8",
	26: "GPIO7",
	27: "GPIO0",
	28: "GPIO1",
	29: "GPIO5",
	31: "GPIO6",
	32: "GPIO12",
	33: "GPIO13",
	35: "GPIO19",
	36: "GPIO16",
	37: "GPIO26",
	38: "GPIO20",
	40: "GPIO21",
}

// DetectBoard resolves a device_model to its GPIO Board. Unknown models fall
// back to RPi.GPIO/BOARD; the admin board override (handled by the caller)
// always takes precedence over this guess.
func DetectBoard(deviceModel string) Board {
	if b, ok := boardRegistry[deviceModel]; ok {
		return b
	}
	return Board{Driver: "rpi"}
}

// BoardModels returns the supported board model names (sorted) for the admin
// override dropdown. These are the device_model keys the registry knows how to
// map to a GPIO driver.
func BoardModels() []string {
	models := make([]string, 0, len(boardRegistry))
	for model := range boardRegistry {
		models = append(models, model)
	}
	sort.Strings(models)
	return models
}
