package payment

import "github.com/flarehotspot/sdk/api/plugin"

func PaymentSetup(api plugin.IPluginApi) {
	paymentApi := api.PaymentsApi()
	wiredCoinslot := NewPaymentMethod(api.Utils())
	paymentApi.NewPaymentMethod(wiredCoinslot)
}
