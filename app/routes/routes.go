package routes

import (
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/wired-coinslot/app/controllers"
	"github.com/flarehotspot/wired-coinslot/app/models"
	"github.com/flarehotspot/wired-coinslot/app/routes/names"
)

func SetRoutes(api plugin.IPluginApi, mdl *models.WiredCoinslotModel) {
	coinctrl := controllers.NewCoinslotsCtrl(api, mdl)
	router := api.HttpApi().Router()

	router.AdminRouter().Get("/index", coinctrl.IndexPage).Name(names.RouteCoinslotsIndex)
}
