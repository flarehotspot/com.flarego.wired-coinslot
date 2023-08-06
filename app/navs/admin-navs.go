package navs

import (
	"net/http"

	"github.com/flarehotspot/com.flarego.wired-coinslot/app/routes/names"
	"github.com/flarehotspot/sdk/v1/api"
	"github.com/flarehotspot/sdk/v1/api/http/navigation"
	"github.com/flarehotspot/sdk/v1/utils/translate"
)

func SetAdminNavs(API api.IPluginApi) {
	navText := API.Translate(translate.Label, "wired_coinslots")
	adminIndex := navigation.NewAdminNav(navigation.CategoryPayments, navText, API.HttpApi().Router().UrlForRoute(names.RouteCoinslotsIndex))
	API.NavApi().AdminNavsFn(func(r *http.Request) []navigation.IAdminNavItem {
		return []navigation.IAdminNavItem{adminIndex}
	})
}
