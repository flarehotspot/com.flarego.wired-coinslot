package app

import (
	"net/http"
	"strconv"

	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	sdkhttp "github.com/flarehotspot/core/sdk/api/http"
	plugin "github.com/flarehotspot/core/sdk/api/plugin"
)

func SetRoutes(api plugin.IPluginApi, mdl *models.WiredCoinslotModel) {

	api.HttpApi().HttpRouter().PluginRouter().Post("/payment-received/{amount}", func(w http.ResponseWriter, r *http.Request) {
		// res := api.HttpApi().VueResponse()
		// vars := api.HttpApi().MuxVars(r)
		// amountstr := vars["amount"]
		// amount, err := strconv.ParseFloat(amountstr, 64)
		// if err != nil {
		//     res.FlashMsg("error", err.Error())
		//     res.Json(w, nil, 500)
		//     return
		// }
	})

	api.HttpApi().VueRouter().RegisterPortalRoutes(sdkhttp.VuePortalRoute{
		RouteName: "insert-coin",
		RoutePath: "/coinslot/:id/insert-coin",
		Component: "InsertCoin.vue",
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			res := api.HttpApi().VueResponse()

			token := r.URL.Query().Get("token")
			purchase, err := api.ModelsApi().Purchase().FindByToken(r.Context(), token)
			if err != nil {
				res.FlashMsg("error", err.Error())
				res.Json(w, nil, 500)
				return
			}

			clnt, err := api.HttpApi().ClientDevice(r)
			if err != nil {
				res.FlashMsg("error", err.Error())
				res.Json(w, nil, 500)
				return
			}

			if purchase.DeviceId() != clnt.Id() {
				res.FlashMsg("error", "This purchase is not for this device.")
				res.Json(w, nil, 500)
				return
			}

			idstr := api.HttpApi().MuxVars(r)["id"]
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

            data := map[string]any{
                "purchase_name": purchase.Name(),
            }
            res.Json(w, data, 200)

		},
		Middlewares: []func(http.Handler) http.Handler{
			api.HttpApi().Middlewares().Device(),
		},
	})

}
