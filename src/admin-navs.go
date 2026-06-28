package src

import (
	"net/http"

	sdkapi "sdk/api"
)

func SetAdminNavs(api sdkapi.IPluginApi) {
	api.Http().Navs().AdminNavsFactory(func(r *http.Request) []sdkapi.AdminNavItemOpt {
		return []sdkapi.AdminNavItemOpt{
			{
			Category:  sdkapi.NavCategoryPayments,
			Label:     api.Translate("label", "Wired Coinslot"),
			RouteName: "admin:wired-coinslots:index",
			Keywords:  []string{"wired", "coinslot", "gpio", "payment", "coin", "relay"},
			Order:     5000,
			Icon:      "<i class='bi bi-coin'></i>",
		},
		}
	})
}
