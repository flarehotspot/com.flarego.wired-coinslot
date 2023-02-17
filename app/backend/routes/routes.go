package routes

import (
	"log"

	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/wired-coinslot/app/backend/controllers"
)

func SetRoutes(api plugin.IPluginApi) {
  log.Println("wired coinslot set route for /test")
	testctrl := controllers.NewTestCtrl(api)
  http := api.HttpApi()
  http.Router().AdminRouter().Get("/test", testctrl.IndexPage)
}
