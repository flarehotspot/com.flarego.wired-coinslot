package payment

import "github.com/flarehotspot/sdk/api/payments"

type PaymentMethod struct {
	adminRouteName  string
	portalRouteName string
}

func (p *PaymentMethod) AdminRoute() (name string) {
	return p.adminRouteName
}

func (p *PaymentMethod) PortalRoute() (name string) {
	return p.portalRouteName
}

func (p *PaymentMethod) RequestPayment(params payments.ICheckoutParams) {

}

func NewPaymentMethod() payments.IPaymentMethod {
	return &PaymentMethod{
		adminRouteName:  "coinslots_index",
		portalRouteName: "coinslots_index",
	}
}
