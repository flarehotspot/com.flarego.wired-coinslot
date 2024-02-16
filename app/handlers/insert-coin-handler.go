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
			res.FlashMsg("error", err.Error())
			res.Json(w, nil, 500)
			return
		}

		clnt, err := api.Http().GetDevice(r)
		if err != nil {
			res.FlashMsg("error", err.Error())
			res.Json(w, nil, 500)
			return
		}

		idstr := api.Http().MuxVars(r)["id"]
		id, err := strconv.ParseInt(idstr, 10, 64)
		if err != nil {
			res.FlashMsg("error", err.Error())
			res.Json(w, nil, 500)
			return
		}

		c, err := mdl.Find(id)
		if err != nil {
			res.FlashMsg("error", err.Error())
			res.Json(w, nil, 500)
			return
		}

		if c.CurrentDeviceId() > 0 && c.CurrentDeviceId() != clnt.Id() {
			res.FlashMsg("error", "Somebody else is using this coinslot right now.")
			res.Json(w, nil, 500)
			return
		}

		c.SetCurrentDeviceId(clnt.Id())
		err = c.UpdateTx(r.Context())
		if err != nil {
			res.FlashMsg("error", err.Error())
			res.Json(w, nil, 500)
			return
		}

		ShowPurchase(w, res, purchase)
	}
}
