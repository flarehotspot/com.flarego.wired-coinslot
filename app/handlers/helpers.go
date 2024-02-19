package handlers

import (
	"net/http"

	sdkhttp "github.com/flarehotspot/core/sdk/api/http"
	sdkpayments "github.com/flarehotspot/core/sdk/api/payments"
)

func ShowPurchase(w http.ResponseWriter, res sdkhttp.VueResponse, purchase sdkpayments.Purchase) {
	state, err := purchase.State()
	if err != nil {
		res.FlashMsg("error", err.Error())
		return
	}
	data := map[string]interface{}{
		"purchase_name":  purchase.Name(),
		"purchase_state": state,
	}
	res.Json(w, data, 200)
}
