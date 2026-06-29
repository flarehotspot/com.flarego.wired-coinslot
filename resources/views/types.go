package views

import (
	sdkapi "sdk/api"

	sdkutils "github.com/flarewifi/sdk-utils"
)

// CoinslotSettingsData is the flat view model for one coinslot's hardware
// configuration form (kept in the views package to avoid an import cycle with src).
type CoinslotSettingsData struct {
	ID                 string
	Name               string
	CoinPin            int
	RelayPin           int
	RelayActive        int
	Pull               string
	Edge               string
	DebounceMs         int
	WindowMs           int
	PaymentTimeoutSecs int
	BoardModel         string
	Denominations      []DenominationData
}

// DenominationData is one editable pulses→amount row in the dynamic
// denominations table.
type DenominationData struct {
	Index  int
	Pulses int
	Amount float64
}

// getCurrencySymbol returns the configured currency symbol, defaulting to the
// peso sign if the application config can't be read.
func getCurrencySymbol(api sdkapi.IPluginApi) string {
	appCfg, err := api.Config().Application().Get()
	if err != nil {
		return "₱"
	}
	return sdkutils.GetCurrencySymbol(appCfg.Currency)
}
