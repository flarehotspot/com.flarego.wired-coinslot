package gpio

import "sort"

// Board describes how a given single-board computer's GPIO header is reached.
// Two driver families exist:
//
//   - "rpi"/"opi": the sysfs-based Python libraries (RPi.GPIO / OPi.GPIO). Both
//     use BOARD (physical header pin) numbering, so the coin/relay pins are
//     passed straight through and only the library + board module differ.
//   - "gpiod": the Linux GPIO character device, driven in-process by the pure-Go
//     go-gpiocdev library (no Python, no pip). Lines are addressed by
//     (chip-label, line-offset); the offset is derived from the Allwinner port
//     name (e.g. "PH5") because sunxi kernels leave line names unset.
type Board struct {
	Driver    string         // "rpi" | "opi" | "gpiod"
	OpiBoard  string         // OPi.GPIO board module name (orangepi.<name>); opi only
	ChipLabel string         // pinctrl label to match a /dev/gpiochipN against; gpiod only
	Header    map[int]string // physical header pin -> Allwinner port name; gpiod only
}

// boardRegistry is keyed by os_release.json device_model, which equals the
// device folder name under go/builder/imagebuilder/devices/. The image variant
// (device_config) is irrelevant to GPIO and intentionally ignored.
//
// Adding a future char-device board is a single line: give it Driver "gpiod"
// and the pinctrl Label of the chip its header GPIOs live on (read with
// `gpioinfo` / `gpiodetect` on the target).
var boardRegistry = map[string]Board{
	"orangepi-one": {Driver: "opi", OpiBoard: "one"}, // Allwinner H3
	"orangepi-pc":  {Driver: "opi", OpiBoard: "pc"},  // Allwinner H3
	// Allwinner H618. OPi.GPIO has no zero3 board module and relies on the
	// removed sysfs interface, so the Zero 3 uses the gpiod char-device driver.
	// Its 26-pin header GPIOs live on the "pio" controller (banks PC/PH), whose
	// pinctrl label is 300b000.pinctrl. Header maps physical pin -> port name so
	// the admin UI can stay in familiar BOARD (physical pin) numbers.
	"orangepi-zero-3": {Driver: "gpiod", ChipLabel: "300b000.pinctrl", Header: orangePiZero3Header},
	"rpi-4":           {Driver: "rpi"},
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
// override dropdown. These are the os_release device_model keys the registry
// knows how to map to a GPIO driver.
func BoardModels() []string {
	models := make([]string, 0, len(boardRegistry))
	for model := range boardRegistry {
		models = append(models, model)
	}
	sort.Strings(models)
	return models
}
