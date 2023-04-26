package payment

import (
	"context"
	"log"

	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/wired-coinslot/app/models"
)

func PaymentSetup(api plugin.IPluginApi, mdl *models.WiredCoinslotModel) {
	paymentApi := api.PaymentsApi()

	coinslots, err := mdl.All(context.Background())
	if err != nil {
		log.Println(err)
		return
	}

	for _, c := range coinslots {
		wiredCoinslot := NewPaymentMethod(api, c)
		err := paymentApi.AddPaymentMethod(wiredCoinslot)
		if err != nil {
			log.Printf("Unable to register payment method %+v", err)
		}
	}
}
