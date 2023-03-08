package payment

import (
	"github.com/flarehotspot/sdk/api/payments"
	"github.com/flarehotspot/sdk/api/utils"
)

type PaymentMethod struct {
	utl            utils.IUtils
	adminRouteName string
}

func (p *PaymentMethod) AdminRoute() (name string) {
	return p.adminRouteName
}

func (p *PaymentMethod) PaymentOptions() []payments.IPaymentOption {
	return []payments.IPaymentOption{}
}

func (p *PaymentMethod) RequestPayment(params payments.PurchaseParams) {

}

func NewPaymentMethod() payments.IPaymentMethod {
	return &PaymentMethod{
		adminRouteName: "coinslots_index",
	}
}
