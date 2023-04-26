package payment

import (
	"fmt"
	"log"
	"net/http"

	modelsI "github.com/flarehotspot/sdk/api/models"
	"github.com/flarehotspot/sdk/api/payments"
	"github.com/flarehotspot/sdk/api/plugin"
  "github.com/flarehotspot/wired-coinslot/app/models"
)

type PaymentMethod struct {
	api plugin.IPluginApi
  coinslot *models.WiredCoinslot
}

func (self *PaymentMethod) Name() string {
	return self.coinslot.Name()
}

func (self *PaymentMethod) PaymentHandler(w http.ResponseWriter, r *http.Request, purchase modelsI.IPurchase) {
	log.Printf("%+v", purchase)
	fmt.Fprintf(w, "Please insert coin")
}

func NewPaymentMethod(api plugin.IPluginApi, c *models.WiredCoinslot) payments.IPaymentMethod {
	return &PaymentMethod{api, c}
}
