//go:build !dev

package src

import (
	sdkapi "sdk/api"
)

// RegisterMockRoutes is a no-op in production builds: there is no synthetic-coin
// endpoint, so payments can only come from real coin pulses read off the GPIO.
func RegisterMockRoutes(api sdkapi.IPluginApi, subrouter sdkapi.IHttpRouterInstance) {}

// MockCoinURL is empty in production: with no synthetic-coin route, the insert-coin
// view renders no mock button.
func MockCoinURL(api sdkapi.IPluginApi, coinslotID string) string { return "" }
