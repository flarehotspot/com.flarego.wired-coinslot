package app

import (
	"github.com/flarehotspot/com.flarego.wired-coinslot/app/handlers"
	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	sdkhttp "github.com/flarehotspot/sdk/api/http"
	plugin "github.com/flarehotspot/sdk/api/plugin"
)

func SetRoutes(api plugin.PluginApi, mdl *models.WiredCoinslotModel) {
	rtr := api.Http().HttpRouter().PluginRouter()
	insertCoinHandler := handlers.InsertCoinHandler(api, mdl)
	paymentReceivedHandler := handlers.PaymentReceivedHandler(api, mdl)
	donePaymentHandler := handlers.DonePayingHandler(api, mdl)

	rtr.Group("/payments", func(subrouter sdkhttp.HttpRouterInstance) {
        subrouter.Get("/insert-coin/{id}", insertCoinHandler).Name("payment.insert_coin")
		subrouter.Post("/received/{id}/{amount}", paymentReceivedHandler).Name("payment:received")
		subrouter.Post("/done", donePaymentHandler).Name("payment:done")
	})

}
