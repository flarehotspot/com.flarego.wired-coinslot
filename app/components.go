package app

import (
	"com.flarego.wired-coinslot/app/models"
	"sdk/api/http"
	"sdk/api/plugin"
)

func SetComponents(api sdkplugin.PluginApi, mdl *models.WiredCoinslotModel) {
	api.Http().VueRouter().RegisterPortalRoutes(sdkhttp.VuePortalRoute{
		RouteName:   "insert-coin",
		RoutePath:   "/coinslot/:id/insert-coin",
		Component:   "InsertCoin.vue",
	})
}
