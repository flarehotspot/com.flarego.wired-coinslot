package navs

import (
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/sdk/api/web/navigation"
	"github.com/flarehotspot/wired-coinslot/app/backend/routes/names"
)

func SetAdminNavs(api plugin.IPluginApi) {
	adminIndex := navigation.AdminNav{
		Text:      "coinslots_list",
		Translate: true,
		RouteName: names.RouteCoinslotsIndex,
	}
	api.NavApi().NewAdminNav(&adminIndex)
}
