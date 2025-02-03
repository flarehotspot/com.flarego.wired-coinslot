package app

import (
	sdkapi "sdk/api"

	"com.flarego.wired-coinslot/app/handlers"
)

func SetRoutes(api sdkapi.IPluginApi) {
	rtr := api.Http().HttpRouter().PluginRouter()
	insertCoinHandler := handlers.InsertCoinHandler(api)
	paymentReceivedHandler := handlers.PaymentReceivedHandler(api)
	donePaymentHandler := handlers.DonePayingHandler(api)

	rtr.Group("/payments", func(subrouter sdkapi.IHttpRouterInstance) {
		subrouter.Get("/insert-coin/{id}", insertCoinHandler).Name("payment:insert_coin")
		subrouter.Post("/received/{id}/{amount}", paymentReceivedHandler).Name("payment:received")
		subrouter.Post("/done", donePaymentHandler).Name("payment:done")
	})

}
