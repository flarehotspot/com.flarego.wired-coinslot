package payment

import (
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/wired-coinslot/app/models"
)

func PaymentSetup(api plugin.IPluginApi, mdl *models.WiredCoinslotModel) {
	paymentApi := api.PaymentsApi()
	provider := NewPaymentProvider(api, mdl)
	paymentApi.NewPaymentProvider(provider)

  router := api.HttpApi().Router()
	router.PluginRouter().Get("/payment/received/{name}/{amount}", provider.PaymentReceived).Name("payment:received")
}
