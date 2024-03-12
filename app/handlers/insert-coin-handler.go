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
			res.SetFlashMsg("error", err.Error())
			res.RedirectToPortal(w)
			return
		}

		clnt, err := api.Http().GetClientDevice(r)
		if err != nil {
			res.SetFlashMsg("error", err.Error())
			res.RedirectToPortal(w)
			return
		}

		idstr := api.Http().MuxVars(r)["id"]
		id, err := strconv.ParseInt(idstr, 10, 64)
		if err != nil {
			res.SetFlashMsg("error", err.Error())
			res.RedirectToPortal(w)
			return
		}

		c, err := mdl.Find(id)
		if err != nil {
			res.SetFlashMsg("error", err.Error())
			res.RedirectToPortal(w)
			return
		}

		if c.CurrentDeviceId() > 0 && c.CurrentDeviceId() != clnt.Id() {
			res.SetFlashMsg("error", "Somebody else is using this coinslot right now.")
			res.RedirectToPortal(w)
			return
		}

		c.SetCurrentDeviceId(clnt.Id())
		err = c.Update(r.Context())
		if err != nil {
			res.SetFlashMsg("error", err.Error())
			res.RedirectToPortal(w)
			return
		}

		ShowPurchase(w, res, purchase)
	}
}
