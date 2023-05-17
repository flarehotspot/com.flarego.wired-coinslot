package navs

import (
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/sdk/api/http/navigation"
	"github.com/flarehotspot/sdk/api/http/navigation/navgen"
	"github.com/flarehotspot/wired-coinslot/app/routes/names"
)

func SetAdminNavs(api plugin.IPluginApi) {
  adminIndex := navgen.NewAdminNav(api, navigation.CategoryPayments, "coinslots_list", names.RouteCoinslotsIndex)
	api.NavApi().NewAdminNav(adminIndex)
}
