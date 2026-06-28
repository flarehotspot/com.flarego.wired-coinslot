package src

import (
	sdkapi "sdk/api"
)

func SetRoutes(api sdkapi.IPluginApi) {
	rtr := api.Http().Router().HttpRouter(nil)

	rtr.Group("/payments", func(subrouter sdkapi.IHttpRouterInstance) {
		subrouter.Get("/insert-coin/{id}", InsertCoinHandler(api)).Name("payments.insert_coin")
		subrouter.Get("/coin-events/{id}", CoinEventsHandler(api)).Name("payments.coin_events")
		subrouter.Get("/done", DonePayingHandler(api)).Name("payments.done")

		// Synthetic-coin endpoint, registered only in dev builds (no-op in prod).
		RegisterMockRoutes(api, subrouter)
	})
}
