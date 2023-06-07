package navs

import (
	"net/http"

	"github.com/flarehotspot/sdk/api/http/navigation"
	"github.com/flarehotspot/sdk/api/http/navigation/navgen"
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/wired-coinslot/app/routes/names"
)

func SetAdminNavs(api plugin.IPluginApi) {
  adminIndex := navgen.NewAdminNav(api, navigation.CategoryPayments, "coinslots_list", names.RouteCoinslotsIndex)
  api.NavApi().AdminNavsFn(func(r *http.Request) []navigation.IAdminNavItem {
    return []navigation.IAdminNavItem{adminIndex}
  })
}
