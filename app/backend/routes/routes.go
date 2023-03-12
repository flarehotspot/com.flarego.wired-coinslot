package routes

import (
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/wired-coinslot/app/backend/config"
	"github.com/flarehotspot/wired-coinslot/app/backend/controllers"
	"github.com/flarehotspot/wired-coinslot/app/backend/routes/names"
)

func SetRoutes(api plugin.IPluginApi, cfg *config.Config) {
	ctrl := controllers.NewCoinslotsCtrl(api, cfg)
	router := api.HttpApi().Router()
	router.AdminRouter().Get("/index", ctrl.IndexPage).Name(names.RouteCoinslotsIndex)
}
