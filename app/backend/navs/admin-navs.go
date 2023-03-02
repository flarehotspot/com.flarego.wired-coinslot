package navs

import (
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/sdk/api/web/navigation"
)

func SetAdminNavs(api plugin.IPluginApi) {
	adminIndex := navigation.AdminNav{
		IconPath: api.Utils().Resource("icon.png"),
		Text:     "coinslots_list",
    Translate: true,
    RouteName: "coinslots_index",
		Href:     "/",
	}
	api.NavApi().NewAdminNav(&adminIndex)
}
