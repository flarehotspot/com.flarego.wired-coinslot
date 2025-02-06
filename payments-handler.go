package main

import (
	"fmt"
	"net/http"

	sdkapi "sdk/api"
	sdkplugin "sdk/api"
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

		coinslotID := api.Http().MuxVars(r)["id"]
		c, err := FindWiredCoinslot(api, coinslotID)
		if err != nil {
			fmt.Println("FindWiredCoinslot error:", err)
			res.FlashMsg(w, r, err.Error(), sdkapi.FlashMsgError)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		if c.DeviceID != nil && *c.DeviceID != clnt.Id() {
			fmt.Println("Somebody else is using this coinslot right now.")
			res.FlashMsg(w, r, "Somebody else is using this coinslot right now.", sdkapi.FlashMsgError)
			res.Redirect(w, r, "portal:index")
			return
		}

		clntID := clnt.Id()
		c.DeviceID = &clntID

		if err := c.Save(); err != nil {
			fmt.Println("WiredCoinslot Save error:", err)
			res.FlashMsg(w, r, err.Error(), sdkapi.FlashMsgError)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		w.Write([]byte(fmt.Sprintf("Show purchase: %+v", purchase)))
		// ShowPurchase(w, res, purchase)
	}
}

func PaymentReceivedHandler(api sdkplugin.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// vars := api.Http().MuxVars(r)
		// idstr := vars["id"]
		// amountstr := vars["amount"]
		// res := api.Http().VueResponse()

		// id, err := strconv.ParseInt(idstr, 10, 64)
		// if err != nil {
		// 	res.Error(w, err.Error(), 500)
		// 	return
		// }

		// amount, err := strconv.ParseFloat(amountstr, 64)
		// if err != nil {
		// 	res.Error(w, err.Error(), 500)
		// 	return
		// }

		// purchase, err := api.Payments().GetPendingPurchase(r)
		// if err != nil {
		// 	log.Println("GetPendingPurchase error:", err)
		// 	res.Error(w, err.Error(), 500)
		// 	return
		// }

		// c, err := mdl.Find(id)
		// if err != nil {
		// 	res.Error(w, err.Error(), 500)
		// 	return
		// }

		// if err := purchase.CreatePayment(amount, c.Name()); err != nil {
		// 	log.Println("CreatePayment error:", err)
		// 	res.Error(w, err.Error(), 500)
		// 	return
		// }

		// res.SetFlashMsg("success", "Payment received")
		// ShowPurchase(w, res, purchase)
	}
}

func DonePayingHandler(api sdkplugin.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// res := api.Http().VueResponse()
		// clnt, err := api.Http().GetClientDevice(r)
		// if err != nil {
		// 	res.Error(w, err.Error(), 500)
		// 	return
		// }

		// c, err := mdl.FindByClientId(clnt.Id())
		// if err != nil {
		// 	res.Error(w, err.Error(), 500)
		// 	return
		// }

		// purchase, err := api.Payments().GetPendingPurchase(r)
		// if err != nil {
		// 	res.Error(w, err.Error(), 500)
		// 	return
		// }

		// c.SetCurrentDeviceId(0)

		// err = c.Update(r.Context())
		// if err != nil {
		// 	res.Error(w, err.Error(), 500)
		// 	return
		// }

		// purchase.Execute(w)
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
