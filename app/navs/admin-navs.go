package navs

import (
	"net/http"

	"github.com/flarehotspot/com.flarego.wired-coinslot/app/routes/names"
	"github.com/flarehotspot/sdk/v1.0.0/api/http/navigation"
	"github.com/flarehotspot/sdk/v1.0.0/api/plugin"
	"github.com/flarehotspot/sdk/v1.0.0/utils/translate"
)

func SetAdminNavs(api plugin.IPluginApi) {
	navText := api.Translate(translate.Label, "wired_coinslots")
	adminIndex := navigation.NewAdminNav(navigation.CategoryPayments, navText, api.HttpApi().Router().UrlForRoute(names.RouteCoinslotsIndex))
	api.NavApi().AdminNavsFn(func(r *http.Request) []navigation.IAdminNavItem {
		return []navigation.IAdminNavItem{adminIndex}
	})
}
