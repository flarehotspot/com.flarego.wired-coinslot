package src

import (
	sdkapi "sdk/api"
)

func SetAdminRoutes(api sdkapi.IPluginApi) {
	adminR := api.Http().Router().AdminRouter()

	adminR.Group("/wired-coinslots", func(subrouter sdkapi.IHttpRouterInstance) {
		subrouter.
			Get("/", ListCoinslotsHandler(api)).
			Name("admin:wired-coinslots:index")

		subrouter.
			Post("/settings/{id}", SaveCoinslotSettingsHandler(api)).
			Name("admin:wired-coinslots:settings-save")

		// Per-coinslot dynamic denomination management.
		subrouter.
			Post("/denominations/{id}/add", AddDenominationHandler(api)).
			Name("admin:wired-coinslots:denomination-add")

		subrouter.
			Post("/denominations/{id}/update/{index}", UpdateDenominationHandler(api)).
			Name("admin:wired-coinslots:denomination-update")

		subrouter.
			Post("/denominations/{id}/delete/{index}", DeleteDenominationHandler(api)).
			Name("admin:wired-coinslots:denomination-delete")
	})
}
