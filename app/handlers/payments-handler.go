package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	sdkhttp "github.com/flarehotspot/sdk/api/http"
	sdkpayments "github.com/flarehotspot/sdk/api/payments"
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

func PaymentReceivedHandler(api sdkplugin.PluginApi, mdl *models.WiredCoinslotModel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := api.Http().MuxVars(r)
		idstr := vars["id"]
		amountstr := vars["amount"]
		res := api.Http().VueResponse()

		id, err := strconv.ParseInt(idstr, 10, 64)
		if err != nil {
			res.Error(w, err.Error(), 500)
			return
		}

		amount, err := strconv.ParseFloat(amountstr, 64)
		if err != nil {
			res.Error(w, err.Error(), 500)
			return
		}

		purchase, err := api.Payments().GetPendingPurchase(r)
		if err != nil {
			log.Println("GetPendingPurchase error:", err)
			res.Error(w, err.Error(), 500)
			return
		}

		c, err := mdl.Find(id)
		if err != nil {
			res.Error(w, err.Error(), 500)
			return
		}

		if err := purchase.CreatePayment(amount, c.Name()); err != nil {
			log.Println("CreatePayment error:", err)
			res.Error(w, err.Error(), 500)
			return
		}

		res.SetFlashMsg("success", "Payment received")
		ShowPurchase(w, res, purchase)
	}
}

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

func ShowPurchase(w http.ResponseWriter, res sdkhttp.VueResponse, purchase sdkpayments.Purchase) {
	state, err := purchase.State()
	if err != nil {
		res.SetFlashMsg("error", err.Error())
		return
	}
	data := map[string]interface{}{
		"purchase_name":  purchase.Name(),
		"purchase_state": state,
	}
	res.Json(w, data, 200)
}
