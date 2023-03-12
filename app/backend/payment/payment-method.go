package payment

import (
	"github.com/flarehotspot/sdk/api/payments"
	"github.com/flarehotspot/sdk/api/utils"
	"net/http"
)

type PaymentMethod struct {
	utl utils.IUtils
}

func (p *PaymentMethod) Name() string {
	return "Coin Alias Here"
}

func (p *PaymentMethod) RequestPayment(w http.ResponseWriter, r *http.Request, items []*payments.PurchaseItem, amount *payments.UnitAmount, callbackUrl string) {

}

func NewPaymentMethod(utl utils.IUtils) payments.IPaymentMethod {
	return &PaymentMethod{utl}
}
