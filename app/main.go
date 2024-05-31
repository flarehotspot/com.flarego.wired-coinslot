package app

import (
	"log"

	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	plugin "sdk/api/plugin"
)

func Init(api plugin.PluginApi) {
	mdl, err := models.NewWiredCoinslotModel(api)
	if err != nil {
		log.Println(err)
	}

	err = mdl.ResetAll()
	if err != nil {
		log.Println(err)
	}

	SetRoutes(api, mdl)
	SetComponents(api, mdl)

	NewPaymentProvider(api, mdl)
}
