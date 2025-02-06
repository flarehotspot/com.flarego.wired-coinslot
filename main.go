//go:build !mono

package main

import (
	sdkapi "sdk/api"
)

func main() {}

func Init(api sdkapi.IPluginApi) {
	// Set default wired coinslots if not exists
	InitWiredCoinslots(api)

	// Setup routes
	SetRoutes(api)

	// Register the payment provider
	provider := NewPaymentProvider(api)
	api.Payments().NewPaymentProvider(provider)
}
