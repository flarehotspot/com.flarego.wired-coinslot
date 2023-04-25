package payment

import (
	"log"

	"github.com/flarehotspot/sdk/api/plugin"
)

func PaymentSetup(api plugin.IPluginApi) {
	paymentApi := api.PaymentsApi()
	wiredCoinslot := NewPaymentMethod(api)
  err := paymentApi.AddPaymentMethod(wiredCoinslot)
  if err != nil {
    log.Printf("Unable to register payment method %+v", err)
  }
}
