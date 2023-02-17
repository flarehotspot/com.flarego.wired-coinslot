package navs

import (
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/sdk/api/web/navigation"
)

func SetAdminMenu(api plugin.IPluginApi) {
	adminIndex := navigation.AdminNav{
		IconPath: api.Helpers().Resource("icon.png"),
		Text:     api.Helpers().Translate("label", "wired_coinslot"),
		Href:     "/",
	}
	api.NavApi().NewAdminNav(adminIndex)
}
