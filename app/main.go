package app

import (
	"log"

	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	"github.com/flarehotspot/com.flarego.wired-coinslot/app/navs"
	"github.com/flarehotspot/com.flarego.wired-coinslot/app/routes"
	"github.com/flarehotspot/sdk/v1/api"
)

func Init(api api.IPluginApi) {
	mdl, err := models.NewWiredCoinslotModel(api)
	if err != nil {
		log.Println(err)
	}
	routes.SetRoutes(api, mdl)
	navs.SetAdminNavs(api)
}
