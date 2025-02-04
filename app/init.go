package app

import (
	sdkapi "sdk/api"

	"com.flarego.wired-coinslot/app/config"
)

func Init(api sdkapi.IPluginApi) {
	// Set default wired coinslots if not exists
	config.InitWiredCoinslots(api)

	// Setup routes
	SetRoutes(api)

	// Register the payment provider
	provider := NewPaymentProvider(api)
	api.Payments().NewPaymentProvider(provider)
}
