package app

import (
	"net/http"
	"strconv"

	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	sdkhttp "github.com/flarehotspot/core/sdk/api/http"
	plugin "github.com/flarehotspot/core/sdk/api/plugin"
)

func SetComponents(api plugin.PluginApi, mdl *models.WiredCoinslotModel) {
	api.Http().VueRouter().RegisterPortalRoutes(sdkhttp.VuePortalRoute{
		RouteName: "insert-coin",
		RoutePath: "/coinslot/:id/insert-coin",
		Component: "InsertCoin.vue",
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
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

			data := map[string]any{
				"purchase_name": purchase.Name(),
			}
			res.Json(w, data, 200)

		},
		Middlewares: []func(http.Handler) http.Handler{
			api.Http().Middlewares().Device(),
		},
	})

}
