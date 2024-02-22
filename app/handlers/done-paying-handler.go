package handlers

import (
	"net/http"

	sdkplugin "github.com/flarehotspot/sdk/api/plugin"
)

func DonePayingHandler(api sdkplugin.PluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := api.Http().VueResponse()
		purchase, err := api.Payments().GetPendingPurchase(r)
		if err != nil {
			res.Error(w, err.Error(), 500)
			return
		}
		purchase.Execute(w)
	}
}
