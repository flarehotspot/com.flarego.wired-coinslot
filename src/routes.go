package src

import (
	sdkapi "sdk/api"
)

func SetRoutes(api sdkapi.IPluginApi) {
	rtr := api.Http().Router().PluginRouter()
	insertCoinHandler := InsertCoinHandler(api)
	paymentReceivedHandler := PaymentReceivedHandler(api)
	donePaymentHandler := DonePayingHandler(api)

	rtr.Group("/payments", func(subrouter sdkapi.IHttpRouterInstance) {
		subrouter.Get("/insert-coin/{id}", insertCoinHandler).Name("payments.insert_coin")
		subrouter.Post("/received/{id}/{amount}", paymentReceivedHandler).Name("payments.received")
		subrouter.Get("/done", donePaymentHandler).Name("payments.done")
	})

}
