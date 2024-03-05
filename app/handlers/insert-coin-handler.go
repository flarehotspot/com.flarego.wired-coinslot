package handlers

import (
	"net/http"
	"strconv"

	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	sdkplugin "github.com/flarehotspot/sdk/api/plugin"
)

func InsertCoinHandler(api sdkplugin.PluginApi, mdl *models.WiredCoinslotModel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := api.Http().VueResponse()

		purchase, err := api.Payments().GetPendingPurchase(r)
		if err != nil {
			res.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		clnt, err := api.Http().GetClientDevice(r)
		if err != nil {
			res.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		idstr := api.Http().MuxVars(r)["id"]
		id, err := strconv.ParseInt(idstr, 10, 64)
		if err != nil {
			res.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		c, err := mdl.Find(id)
		if err != nil {
			res.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if c.CurrentDeviceId() > 0 && c.CurrentDeviceId() != clnt.Id() {
			res.Error(w, "Somebody else is using this coinslot right now.", http.StatusInternalServerError)
			return
		}

		c.SetCurrentDeviceId(clnt.Id())
		err = c.UpdateTx(r.Context())
		if err != nil {
			res.Error(w, err.Error(), 500)
			return
		}

		ShowPurchase(w, res, purchase)
	}
}
