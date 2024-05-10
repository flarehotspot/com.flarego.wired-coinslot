package app

import (
	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	"github.com/flarehotspot/sdk/api/http"
	"github.com/flarehotspot/sdk/api/plugin"
)

func SetComponents(api sdkplugin.PluginApi, mdl *models.WiredCoinslotModel) {
	api.Http().VueRouter().RegisterPortalRoutes(sdkhttp.VuePortalRoute{
		RouteName:   "insert-coin",
		RoutePath:   "/coinslot/:id/insert-coin",
		Component:   "InsertCoin.vue",
	})
}
