package gpio

import "sort"

// Board describes how a given single-board computer's GPIO header is reached
// through the sysfs-based Python libraries (RPi.GPIO / OPi.GPIO). Both use BOARD
// (physical header pin) numbering, so the coin/relay pins are passed straight
// through and only the library + board module differ per device.
type Board struct {
	Library  string // "rpi" or "opi"
	OpiBoard string // OPi.GPIO board module name (orangepi.<name>); empty for rpi
}

// boardRegistry is keyed by os_release.json device_model, which equals the
// device folder name under go/builder/imagebuilder/devices/. The image variant
// (device_config) is irrelevant to GPIO and intentionally ignored.
var boardRegistry = map[string]Board{
	"orangepi-one":    {Library: "opi", OpiBoard: "one"},   // Allwinner H3
	"orangepi-pc":     {Library: "opi", OpiBoard: "pc"},    // Allwinner H3
	"orangepi-zero-3": {Library: "opi", OpiBoard: "zero3"}, // Allwinner H618
	"rpi-4":           {Library: "rpi"},
}

// DetectBoard resolves a device_model to its GPIO Board. Unknown models fall
// back to RPi.GPIO/BOARD; the admin board override (handled by the caller)
// always takes precedence over this guess.
func DetectBoard(deviceModel string) Board {
	if b, ok := boardRegistry[deviceModel]; ok {
		return b
	}
	return Board{Library: "rpi"}
}

// BoardModels returns the supported board model names (sorted) for the admin
// override dropdown. These are the os_release device_model keys the registry
// knows how to map to a GPIO library + board module.
func BoardModels() []string {
	models := make([]string, 0, len(boardRegistry))
	for model := range boardRegistry {
		models = append(models, model)
	}
	sort.Strings(models)
	return models
}
