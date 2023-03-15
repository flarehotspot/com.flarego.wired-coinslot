package payment

import (
	"fmt"
	"net/http"

	"github.com/flarehotspot/sdk/api/payments"
	"github.com/flarehotspot/sdk/api/utils"
)

type PaymentMethod struct {
	utl utils.IUtils
}

func (p *PaymentMethod) Name() string {
	return "Coin Alias Here"
}

func (p *PaymentMethod) RequestPayment(w http.ResponseWriter, r *http.Request, items []*payments.PaymentRequestItem, amount *payments.UnitAmount, callbackUrl string) {
  fmt.Fprintf(w, "Please insert coin")
}

func NewPaymentMethod(utl utils.IUtils) payments.IPaymentMethod {
	return &PaymentMethod{utl}
}
