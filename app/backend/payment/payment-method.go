package payment

import (
	"fmt"
	"log"
	"net/http"

	"github.com/flarehotspot/sdk/api/db/models"
	"github.com/flarehotspot/sdk/api/payments"
	"github.com/flarehotspot/sdk/api/plugin"
	// uuid "github.com/flarehotspot/sdk/libs/go.uuid"
)

type PaymentMethod struct {
	api plugin.IPluginApi
}

func (self *PaymentMethod) Name() string {
	return "Coin Alias Here"
}

func (self *PaymentMethod) PaymentHandler(w http.ResponseWriter, r *http.Request, purchase models.IPurchase) {
	log.Printf("%+v", purchase)
	fmt.Fprintf(w, "Please insert coin")
}

func NewPaymentMethod(api plugin.IPluginApi) payments.IPaymentMethod {
	return &PaymentMethod{api}
}
