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

func (self *PaymentMethod) Name() string {
	return "Coin Alias Here"
}

func (self *PaymentMethod) RequestPayment(w http.ResponseWriter, r *http.Request, params *payments.PaymentRequestParams) {
	fmt.Fprintf(w, "Please insert coin")
}

func NewPaymentMethod(utl utils.IUtils) payments.IPaymentMethod {
	return &PaymentMethod{utl}
}
