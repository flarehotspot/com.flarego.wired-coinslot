package app

import (
	"net/http"

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
	})

}
