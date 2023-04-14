package navs

import (
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/sdk/api/web/navigation"
	"github.com/flarehotspot/sdk/api/web/navigation/navgen"
	"github.com/flarehotspot/wired-coinslot/app/backend/routes/names"
)

func SetAdminNavs(api plugin.IPluginApi) {
  adminIndex := navgen.NewAdminNav(api, navigation.CategoryPayments, "coinslots_list", names.RouteCoinslotsIndex)
	api.NavApi().NewAdminNav(adminIndex)
}
