package routes

import (
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/sdk/api/http/router"
	"github.com/flarehotspot/com.flarego.wired-coinslot/app/controllers"
	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	"github.com/flarehotspot/com.flarego.wired-coinslot/app/payment"
	"github.com/flarehotspot/com.flarego.wired-coinslot/app/routes/names"
)

func SetRoutes(api plugin.IPluginApi, mdl *models.WiredCoinslotModel) {
	coinctrl := controllers.NewCoinslotsCtrl(api, mdl)
	adminRtr := api.HttpApi().Router().AdminRouter()
	adminRtr.Get("/index", coinctrl.IndexPage).Name(names.RouteCoinslotsIndex)

	paymentApi := api.PaymentsApi()
	provider := payment.NewPaymentProvider(api, mdl)
	paymentApi.NewPaymentProvider(provider)
	deviceMw := api.HttpApi().Middlewares().Device()

	plugRtr := api.HttpApi().Router().PluginRouter()
	plugRtr.Group("/payment", func(subrouter router.IRouter) {
		subrouter.Use(deviceMw)
		subrouter.Get("/received", provider.PaymentReceived).Name("payment:received")
		subrouter.Get("/wallet", provider.UseWalletBal).Name("use:wallet")
		subrouter.Get("/done", provider.Done).Name("payment:done")
		subrouter.Get("/cancel", provider.Cancel).Name("payment:cancel")
	})
}
