package payment

import (
	"net/http"

	"github.com/flarehotspot/sdk/api/payments"
	"github.com/flarehotspot/sdk/api/utils"
)

type PaymentMethod struct {
	utl            utils.IUtils
	adminRouteName string
}

func (p *PaymentMethod) Name() string {
	return "Coin Alias Here"
}

func (p *PaymentMethod) AdminRoute() (name string) {
	return p.adminRouteName
}

func (p *PaymentMethod) RequestPayment(w http.ResponseWriter, r *http.Request, items []*payments.PurchaseItem, amount *payments.UnitAmount, callbackUrl string) {

}

func NewPaymentMethod(utl utils.IUtils) payments.IPaymentMethod {
	return &PaymentMethod{
		utl:            utl,
		adminRouteName: "coinslots_index",
	}
}
