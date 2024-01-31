package app

import (
	"net/http"
	"strconv"

	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	plugin "github.com/flarehotspot/core/sdk/api/plugin"
)

func SetRoutes(api plugin.IPluginApi, mdl *models.WiredCoinslotModel) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		vars := api.HttpApi().MuxVars(r)
		optname := vars["optname"]
		token := vars["token"]
		amountstr := vars["amount"]
		res := api.HttpApi().VueResponse()

		// convert to float64
		amount, err := strconv.ParseFloat(amountstr, 64)
		if err != nil {
			res.Error(w, err.Error(), 500)
			return
		}

		if err = api.PaymentsApi().PaymentReceived(r.Context(), token, optname, amount); err != nil {
			res.Error(w, err.Error(), 500)
			return
		}

		res.FlashMsg("success", "Payment received")
		res.Json(w, nil, 200)
	}

	api.HttpApi().HttpRouter().PluginRouter().Post("/payment-received/{optname}/{token}/{amount}", handler).Name("payment:received")
}
