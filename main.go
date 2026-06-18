//go:build !mono

package main

import (
	sdkapi "sdk/api"

	"com.flarego.wired-coinslot/src"
)

func main() {}

// Init must return error: the core plugin loader only invokes Init when it
// type-asserts to func(sdkapi.IPluginApi) error (see core plugin-init.go). A
// no-return Init is silently skipped, so the plugin never registers anything.
func Init(api sdkapi.IPluginApi) error {
	// Ensure a default coinslot config exists.
	src.InitWiredCoinslots(api)

	// Portal + admin routes and navigation.
	src.SetRoutes(api)
	src.SetAdminRoutes(api)
	src.SetAdminNavs(api)

	// Start the GPIO agents, pulse counters and payment session manager.
	src.StartManager(api)

	// Register the payment provider so coinslots appear as a payment option.
	provider := src.NewPaymentProvider(api)
	api.Payments().NewPaymentProvider(provider)

	return nil
}
