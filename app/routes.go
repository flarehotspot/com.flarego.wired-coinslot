package app

import (
	"net/http"

	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	plugin "github.com/flarehotspot/core/sdk/api/plugin"
)

func SetRoutes(api plugin.IPluginApi, mdl *models.WiredCoinslotModel) {
	api.HttpApi().HttpRouter().PluginRouter().Post("/payment-received/{amount}", func(w http.ResponseWriter, r *http.Request) {
	})
}
