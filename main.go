package main

import (
	sdkapi "sdk/api"

	"com.flarego.wired-coinslot/src"
)

func main() {}

func Init(api sdkapi.IPluginApi) {
	// Set default wired coinslots if not exists
	src.InitWiredCoinslots(api)

	// // Setup routes
	src.SetRoutes(api)

	// // Register the payment provider
	provider := src.NewPaymentProvider(api)
	api.Payments().NewPaymentProvider(provider)
}
