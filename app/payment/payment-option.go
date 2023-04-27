package payment

import (
	"fmt"
	"log"
	"net/http"

	mdlI "github.com/flarehotspot/sdk/api/models"
	"github.com/flarehotspot/sdk/api/payments"
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/wired-coinslot/app/models"
)

type PaymentOption struct {
	api      plugin.IPluginApi
	coinslot *models.WiredCoinslot
}

func (self *PaymentOption) Name() string {
	return self.coinslot.Name()
}

func (self *PaymentOption) PaymentHandler(w http.ResponseWriter, r *http.Request, purchase mdlI.IPurchase) {
	log.Printf("%+v", purchase)
	fmt.Fprintf(w, "Please insert coin")
}

func NewPaymentOpt(api plugin.IPluginApi, c *models.WiredCoinslot) payments.IPaymentOpt {
	return &PaymentOption{api, c}
}
