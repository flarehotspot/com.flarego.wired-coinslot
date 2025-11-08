package src

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	sdkapi "sdk/api"
	sdkplugin "sdk/api"

	"com.flarego.wired-coinslot/resources/views"
	"github.com/a-h/templ"
	sdkutils "github.com/flarehotspot/sdk-utils"
)

func InsertCoinHandler(api sdkplugin.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := api.Http().Response()

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

		c, err := LoadWiredCoinslot(api, coinslotID)
		if err != nil {
			fmt.Println("FindWiredCoinslot error:", err)
			res.FlashMsg(w, r, err.Error(), sdkapi.FlashMsgError)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		if !c.CanBeUsedBy(clnt.Id()) {
			fmt.Println("Somebody else is using this coinslot right now.")
			res.FlashMsg(w, r, "Somebody else is using this coinslot right now.", sdkapi.FlashMsgError)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		c.UseBy(clnt.Id())

		ctx := r.Context()

		var insertCoinPage templ.Component

		err = sdkutils.RunInTx(c.api.SqlDB(), ctx, func(tx *sql.Tx) error {
			insertCoinPage = views.InsertCoinPage(tx, ctx, api, purchase, coinslotID)
			return nil
		})
		if err != nil {
			res.FlashMsg(w, r, err.Error(), sdkapi.FlashMsgError)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

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
		res := api.Http().Response()

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

		c, err := LoadWiredCoinslot(api, idstr)
		if err != nil {
			res.Error(w, r, err, 500)
			return
		}

		ctx := r.Context()

		err = sdkutils.RunInTx(c.api.SqlDB(), ctx, func(tx *sql.Tx) error {
			if err := purchase.CreatePayment(tx, ctx, amount, c.GetName()); err != nil {
				return err
			}

			v := views.PaymentReceivedPartial(api, tx, ctx, purchase)
			v.Render(r.Context(), w)
			return nil
		})

		if err != nil {
			api.Logger().Error(err.Error())
			log.Println("CreatePayment error:", err)
			res.Error(w, r, err, 500)
		}
	}
}

func DonePayingHandler(api sdkplugin.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := api.Http().Response()
		clnt, err := api.Http().GetClientDevice(r)
		if err != nil {
			res.Error(w, r, err, 500)
			return
		}

		c, err := FindUsedCoinslot(api, clnt.Id())
		if err != nil {
			res.Error(w, r, err, 500)
			return
		}

		if c == nil {
			res.FlashMsg(w, r, api.Translate("error", "payment_failed"), sdkapi.FlashMsgError)
			res.RedirectToPortal(w, r)
			return
		}

		c.DoneUsing()

		purchase, err := api.Payments().GetPurchaseRequest(r)
		if err != nil {
			res.Error(w, r, err, 500)
			return
		}

		purchase.Execute(w, r)
	}
}
