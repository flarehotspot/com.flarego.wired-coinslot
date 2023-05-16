package app

import (
	"log"

	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/wired-coinslot/app/models"
	"github.com/flarehotspot/wired-coinslot/app/navs"
	"github.com/flarehotspot/wired-coinslot/app/routes"
)

func Init(api plugin.IPluginApi) {
  mdl, err := models.NewWiredCoinslotModel(api)
  if err != nil {
    log.Println(err)
  }
	routes.SetRoutes(api, mdl)
	navs.SetAdminNavs(api)
}
