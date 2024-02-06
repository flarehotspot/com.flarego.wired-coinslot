package app

import (
	"net/http"

	"github.com/flarehotspot/com.flarego.wired-coinslot/app/handlers"
	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	sdkhttp "github.com/flarehotspot/core/sdk/api/http"
	plugin "github.com/flarehotspot/core/sdk/api/plugin"
)

func SetComponents(api plugin.PluginApi, mdl *models.WiredCoinslotModel) {
	insertCoinHandler := handlers.InsertCoinHandler(api, mdl)
	api.Http().VueRouter().RegisterPortalRoutes(sdkhttp.VuePortalRoute{
		RouteName:   "insert-coin",
		RoutePath:   "/coinslot/:id/insert-coin",
		Component:   "InsertCoin.vue",
		HandlerFunc: insertCoinHandler,
		Middlewares: []func(http.Handler) http.Handler{
			api.Http().Middlewares().Device(),
		},
	})

}
