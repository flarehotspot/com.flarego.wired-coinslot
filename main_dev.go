//go:build dev

package coinslot

import (
	"log"

	"github.com/flarehotspot/sdk/api/config"
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/sdk/api/web/navigation"
	"github.com/flarehotspot/wired-coinslot/app/backend/routes"
)

type SomeCfgType struct {
	Type int `yaml:"type"`
}

type MyCfg struct {
	SomeKey  string       `yaml:"some_key"`
	SomeType *SomeCfgType `yaml:"some_type"`
}

func Init(mgr plugin.IPluginMgr, api plugin.IPluginApi) {
	cfg := &config.PluginConfig{
		"some": "value",
	}
	api.ConfigApi().Write(cfg)
	cfg, err := api.ConfigApi().Read()
	if err != nil {
		log.Println("Error reading config file: ", err)
	}
	log.Printf("Config is: %+v", *cfg)

	nav := navigation.AdminNav{
		IconPath:    "/asdfadf/icon.png",
		Text:        "Settings",
		Href:        "/settings",
		Permissions: []string{"modify-settings"},
	}
	api.NavApi().NewAdminNav(nav)

  // resultPromise := api.PaymentsApi().Checkout(amount: 100)

	// select {
	// case paymentinfo := resultPromise.Done():
		// log.Print("generate session")
	// case err := resultPromise.Error():
		// log.Println("sorry error in payment")
	// }

	routes.SetRoutes(api)
}
