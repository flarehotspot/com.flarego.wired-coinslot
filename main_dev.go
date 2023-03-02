//go:build dev

package coinslot

import (
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/wired-coinslot/app/backend/navs"
	"github.com/flarehotspot/wired-coinslot/app/backend/payment"
	"github.com/flarehotspot/wired-coinslot/app/backend/routes"
)

func Init(api plugin.IPluginApi) {
	routes.SetRoutes(api)
	navs.SetAdminNavs(api)
	payment.PaymentSetup(api)
}
