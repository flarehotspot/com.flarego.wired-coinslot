package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	sdkplugin "github.com/flarehotspot/core/sdk/api/plugin"
)

func PaymentReceivedHandler(api sdkplugin.PluginApi, mdl *models.WiredCoinslotModel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := api.Http().MuxVars(r)
		optname := vars["optname"]
		amountstr := vars["amount"]
		res := api.Http().VueResponse()

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

		if err := purchase.CreatePayment(amount, optname); err != nil {
			log.Println("CreatePayment error:", err)
			res.Error(w, err.Error(), 500)
			return
		}

		res.FlashMsg("success", "Payment received")
		ShowPurchase(w, res, purchase)
	}
}
