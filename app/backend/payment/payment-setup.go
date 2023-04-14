package payment

import (
	"github.com/flarehotspot/sdk/api/payments"
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/wired-coinslot/app/backend/routes/names"
)

func PaymentSetup(api plugin.IPluginApi) {
	paymentApi := api.PaymentsApi()
	wiredCoinslot := NewPaymentMethod(api.Utils())
  adminNav := &payments.PaymentAdminNav{
    Label: "wired_coinslots",
    Translate: true,
    RouteName: names.RouteCoinslotsIndex,
  }
	paymentApi.AdminSettings(adminNav)
	paymentApi.NewPaymentsPlugin(wiredCoinslot)
}
