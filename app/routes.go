package app

import (
	"github.com/flarehotspot/com.flarego.wired-coinslot/app/handlers"
	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	plugin "github.com/flarehotspot/sdk/api/plugin"
)

func SetRoutes(api plugin.PluginApi, mdl *models.WiredCoinslotModel) {
	rtr := api.Http().HttpRouter().PluginRouter()
	paymentReceivedHandler := handlers.PaymentReceivedHandler(api, mdl)
	donePaymentHandler := handlers.DonePayingHandler(api)

	rtr.Post("/payment-received/{optname}/{amount}", paymentReceivedHandler).Name("payment:received")
	rtr.Post("/done-paying", donePaymentHandler).Name("payment:done")
}
