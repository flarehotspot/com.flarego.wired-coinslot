package app

import (
	"github.com/flarehotspot/com.flarego.wired-coinslot/app/handlers"
	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	plugin "github.com/flarehotspot/core/sdk/api/plugin"
)

func SetRoutes(api plugin.PluginApi, mdl *models.WiredCoinslotModel) {
	rtr := api.Http().HttpRouter().PluginRouter()
	deviceMw := api.Http().Middlewares().Device()
	paymentReceivedHandler := handlers.PaymentReceivedHandler(api, mdl)
	donePaymentHandler := handlers.DonePayingHandler(api)

	rtr.Post("/payment-received/{optname}/{amount}", paymentReceivedHandler, deviceMw).Name("payment:received")
	rtr.Post("/done-paying", donePaymentHandler, deviceMw).Name("payment:done")
}
