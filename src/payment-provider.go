package src

import (
	"net/http"

	sdkapi "sdk/api"
)

func NewPaymentProvider(api sdkapi.IPluginApi) *PaymentProvider {
	return &PaymentProvider{
		name: api.Translate("label", "Wired Coinslot"),
		api:  api,
	}
}

type PaymentProvider struct {
	name string
	api  sdkapi.IPluginApi
}

func (self *PaymentProvider) Name() string {
	return self.name
}

func (self *PaymentProvider) OptionsFactory(r *http.Request) []sdkapi.PaymentOption {
	wiredCoinslots, err := GetAllWiredCoinslots(self.api)
	if err != nil {
		_ = self.api.Logger().Error("[wired-coinslot] failed to list coinslots: " + err.Error())
		return nil
	}

	opts := make([]sdkapi.PaymentOption, 0, len(wiredCoinslots))
	for _, c := range wiredCoinslots {
		if c == nil {
			continue
		}
		opts = append(opts, sdkapi.PaymentOption{
			Name:        c.GetName(),
			RouteName:   "payments.insert_coin",
			RouteParams: map[string]string{"id": c.GetID()},
		})
	}
	return opts
}
