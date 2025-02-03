package app

import (
	sdkapi "sdk/api"
)

func Init(api sdkapi.IPluginApi) {
	SetRoutes(api)
	// SetComponents(api, mdl)

	provider := NewPaymentProvider(api)
	api.Payments().NewPaymentProvider(provider)
}
