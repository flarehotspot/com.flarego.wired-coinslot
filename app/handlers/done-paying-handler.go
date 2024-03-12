package handlers

import (
	"net/http"

	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	sdkplugin "github.com/flarehotspot/sdk/api/plugin"
)

func DonePayingHandler(api sdkplugin.PluginApi, mdl *models.WiredCoinslotModel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := api.Http().VueResponse()
		clnt, err := api.Http().GetClientDevice(r)
		if err != nil {
			res.Error(w, err.Error(), 500)
			return
		}

		c, err := mdl.FindByClientId(clnt.Id())
		if err != nil {
			res.Error(w, err.Error(), 500)
			return
		}

		purchase, err := api.Payments().GetPendingPurchase(r)
		if err != nil {
			res.Error(w, err.Error(), 500)
			return
		}

		c.SetCurrentDeviceId(0)

		err = c.Update(r.Context())
		if err != nil {
			res.Error(w, err.Error(), 500)
			return
		}

		purchase.Execute(w)
	}
}
