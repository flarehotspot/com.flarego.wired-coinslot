package routes

import (
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/wired-coinslot/app/backend/controllers"
)

func SetRoutes(api plugin.IPluginApi) {
	testctrl := controllers.NewTestCtrl(api)
  router := api.HttpApi().Router()
  router.AdminRouter().Get("/index", testctrl.IndexPage, "coinslots_index")
}
