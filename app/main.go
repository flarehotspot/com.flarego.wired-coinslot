package app

import (
	"log"

	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	plugin "github.com/flarehotspot/core/sdk/api/plugin"
)

func Init(api plugin.IPluginApi) {
	mdl, err := models.NewWiredCoinslotModel(api)
	if err != nil {
		log.Println(err)
	}
	SetRoutes(api, mdl)
}
