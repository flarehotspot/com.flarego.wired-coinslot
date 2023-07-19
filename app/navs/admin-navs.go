package navs

import (
	"net/http"

	"github.com/flarehotspot/com.flarego.wired-coinslot/app/routes/names"
	"github.com/flarehotspot/sdk/api/http/navigation"
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/sdk/utils/constants"
)

func SetAdminNavs(api plugin.IPluginApi) {
	navText := api.Utils().Translate(cnts.TranslateLabel, "wired_coinslots")
	adminIndex := navigation.NewAdminNav(navigation.CategoryPayments, navText, api.HttpApi().Router().UrlForRoute(names.RouteCoinslotsIndex))
	api.NavApi().AdminNavsFn(func(r *http.Request) []navigation.IAdminNavItem {
		return []navigation.IAdminNavItem{adminIndex}
	})
}
