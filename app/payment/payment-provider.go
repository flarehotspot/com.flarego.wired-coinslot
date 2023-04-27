package payment

import (
	"log"

	"github.com/flarehotspot/sdk/api/payments"
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/sdk/api/web/router"
	"github.com/flarehotspot/wired-coinslot/app/models"
	"github.com/flarehotspot/wired-coinslot/app/routes/names"
)

type PaymentProvider struct {
	api   plugin.IPluginApi
	model *models.WiredCoinslotModel
	name  string
	route router.PluginRouteName
}

func (self *PaymentProvider) Name() string {
	return self.name
}

func (self *PaymentProvider) AdminRoute() router.PluginRouteName {
	return self.route
}

func (self *PaymentProvider) PaymentOpts() ([]payments.IPaymentOpt, error) {
	coinslots, err := self.model.All()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	methods := []payments.IPaymentOpt{}
	for _, c := range coinslots {
		wiredCoinslot := NewPaymentOpt(self.api, c)
		methods = append(methods, wiredCoinslot)
	}

	return methods, nil
}

func NewPaymentProvider(api plugin.IPluginApi, mdl *models.WiredCoinslotModel) payments.IPaymentProvider {
	return &PaymentProvider{
		api:   api,
		model: mdl,
		name:  "Wired Coinslots",
		route: names.RouteCoinslotsIndex,
	}
}
