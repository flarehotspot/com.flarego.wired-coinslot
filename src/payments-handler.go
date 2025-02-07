package src

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	sdkapi "sdk/api"
	sdkplugin "sdk/api"

	"com.flarego.wired-coinslot/resources/views"
	sdkutils "github.com/flarehotspot/sdk-utils"
)

func InsertCoinHandler(api sdkplugin.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := api.Http().HttpResponse()

		purchase, err := api.Payments().GetPurchaseRequest(r)
		if err != nil {
			fmt.Println("GetPurchaseRequest error:", err)
			res.FlashMsg(w, r, err.Error(), sdkapi.FlashMsgError)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		clnt, err := api.Http().GetClientDevice(r)
		if err != nil {
			fmt.Println("GetClientDevice error:", err)
			res.FlashMsg(w, r, err.Error(), sdkapi.FlashMsgError)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		clntID := sdkutils.PgUuidToString(clnt.Id())
		coinslotID := api.Http().MuxVars(r)["id"]

		c, err := FindWiredCoinslot(api, coinslotID)
		if err != nil {
			fmt.Println("FindWiredCoinslot error:", err)
			res.FlashMsg(w, r, err.Error(), sdkapi.FlashMsgError)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		if c.DeviceID != nil && *c.DeviceID != clntID {
			fmt.Println("Somebody else is using this coinslot right now.")
			res.FlashMsg(w, r, "Somebody else is using this coinslot right now.", sdkapi.FlashMsgError)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		c.DeviceID = &clntID
		if err := c.Save(); err != nil {
			fmt.Println("WiredCoinslot Save error:", err)
			res.FlashMsg(w, r, err.Error(), sdkapi.FlashMsgError)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		insertCoinPage := views.InsertCoinPage(api, purchase, coinslotID)
		res.PortalView(w, r, sdkplugin.ViewPage{
			Assets: sdkplugin.ViewAssets{
				JsFile: "pages/insert-coin.js",
			},
			PageContent: insertCoinPage,
		})
	}
}

func PaymentReceivedHandler(api sdkplugin.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := api.Http().MuxVars(r)
		idstr := vars["id"]
		amountstr := vars["amount"]
		res := api.Http().HttpResponse()

		amount, err := strconv.ParseFloat(amountstr, 64)
		if err != nil {
			res.Error(w, r, err, 500)
			return
		}

		purchase, err := api.Payments().GetPurchaseRequest(r)
		if err != nil {
			log.Println("GetPendingPurchase error:", err)
			res.Error(w, r, err, 500)
			return
		}

		c, err := FindWiredCoinslot(api, idstr)
		if err != nil {
			res.Error(w, r, err, 500)
			return
		}

		if err := purchase.CreatePayment(amount, c.Name); err != nil {
			log.Println("CreatePayment error:", err)
			res.Error(w, r, err, 500)
			return
		}

		v := views.PaymentReceivedPartial(purchase)
		v.Render(r.Context(), w)
	}
}

func DonePayingHandler(api sdkplugin.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := api.Http().HttpResponse()
		clnt, err := api.Http().GetClientDevice(r)
		if err != nil {
			res.Error(w, r, err, 500)
			return
		}

		c, err := FindWiredCoinslotByDevice(api, clnt.Id())
		if err != nil {
			res.Error(w, r, err, 500)
			return
		}

		purchase, err := api.Payments().GetPurchaseRequest(r)
		if err != nil {
			res.Error(w, r, err, 500)
			return
		}

		c.DeviceID = nil
		if err = c.Save(); err != nil {
			res.Error(w, r, err, 500)
			return
		}

		purchase.Execute(w, r)
	}
}

// func ShowPurchase(w http.ResponseWriter, res sdkhttp.VueResponse, purchase sdkpayments.Purchase) {
// 	state, err := purchase.State()
// 	if err != nil {
// 		res.SetFlashMsg("error", err.Error())
// 		return
// 	}
// 	data := map[string]interface{}{
// 		"purchase_name":  purchase.Name(),
// 		"purchase_state": state,
// 	}
// 	res.Json(w, data, 200)
// }
