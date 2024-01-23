package routes

import (
	// "github.com/flarehotspot/com.flarego.wired-coinslot/app/controllers"
	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	// "github.com/flarehotspot/com.flarego.wired-coinslot/app/payment"
	// "github.com/flarehotspot/com.flarego.wired-coinslot/app/routes/names"
	// "github.com/flarehotspot/core/sdk/api/http/router"
	plugin "github.com/flarehotspot/core/sdk/api/plugin"
)

func SetRoutes(api plugin.IPluginApi, mdl *models.WiredCoinslotModel) {
	// coinctrl := controllers.NewCoinslotsCtrl(api, mdl)
	// adminRtr := api.HttpApi().HttpRouter().AdminRouter()
	// adminRtr.Get("/index", coinctrl.IndexPage).Name(names.RouteCoinslotsIndex)

	// paymentApi := api.PaymentsApi()
	// provider := payment.NewPaymentProvider(api, mdl)
	// paymentApi.NewPaymentProvider(provider)
	// deviceMw := api.HttpApi().Middlewares().Device()

	// plugRtr := api.HttpApi().HttpRouter().PluginRouter()
	// plugRtr.Group("/payment", func(subrouter router.IHttpRouter) {
	// 	subrouter.Use(deviceMw)
	// 	subrouter.Get("/received", provider.PaymentReceived).Name("payment:received")
	// 	subrouter.Get("/wallet", provider.UseWalletBal).Name("use:wallet")
	// 	subrouter.Get("/done", provider.Done).Name("payment:done")
	// 	subrouter.Get("/cancel", provider.Cancel).Name("payment:cancel")
	// })
}
