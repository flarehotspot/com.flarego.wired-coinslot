package src

import (
	"fmt"
	"net/http"

	sdkapi "sdk/api"

	"com.flarego.wired-coinslot/resources/views"
	"github.com/goccy/go-json"
)

// InsertCoinHandler renders the insert-coin page, claims the coinslot for the
// client, starts a payment session and opens the relay so coins are accepted.
func InsertCoinHandler(api sdkapi.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := api.Http().Response()

		purchase, err := api.Payments().GetPurchaseRequest(r)
		if err != nil {
			res.FlashMsg(w, r, err.Error(), sdkapi.FlashMsgError)
			res.RedirectToPortal(w, r)
			return
		}

		clnt, err := api.Http().GetClientDevice(r)
		if err != nil {
			res.FlashMsg(w, r, err.Error(), sdkapi.FlashMsgError)
			res.RedirectToPortal(w, r)
			return
		}

		coinslotID := api.Http().MuxVars(r)["id"]
		c, err := LoadWiredCoinslot(api, coinslotID)
		if err != nil {
			res.FlashMsg(w, r, err.Error(), sdkapi.FlashMsgError)
			res.RedirectToPortal(w, r)
			return
		}

		if !c.CanBeUsedBy(clnt.ID()) {
			res.FlashMsg(w, r, api.Translate("error", "Somebody else is using this coinslot right now."), sdkapi.FlashMsgError)
			res.RedirectToPortal(w, r)
			return
		}
		c.UseBy(clnt.ID())

		if mgr := GetManager(); mgr != nil {
			mgr.Sessions().Begin(coinslotID, clnt.ID(), purchase)
		}

		res.PortalView(w, r, sdkapi.ViewPage{
			Assets: sdkapi.ViewAssets{
				JsFile: "pages/insert-coin.js",
			},
			// mockURL is empty in production builds (see mock_prod.go), so the view
			// renders no synthetic-coin button there.
			PageContent: views.InsertCoinPage(api, purchase, coinslotID, MockCoinURL(api, coinslotID)),
		})
	}
}

// CoinEventsHandler streams live payment updates to the insert-coin page over
// Server-Sent Events. The connection's lifetime gates coin acceptance: leaving
// the page closes the relay (after a short grace) via the session manager.
func CoinEventsHandler(api sdkapi.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		coinslotID := api.Http().MuxVars(r)["id"]
		mgr := GetManager()
		if mgr == nil {
			http.Error(w, "coinslot manager not ready", http.StatusServiceUnavailable)
			return
		}

		ch, unsub, ok := mgr.Sessions().Subscribe(coinslotID)
		if !ok {
			http.Error(w, "no active payment session", http.StatusNotFound)
			return
		}
		defer unsub()

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Send the current state immediately so a reconnecting client is in sync.
		if snap, ok := mgr.Sessions().Snapshot(coinslotID); ok {
			writeSSE(w, flusher, snap)
		}

		for {
			select {
			case <-r.Context().Done():
				return
			case ev, ok := <-ch:
				if !ok {
					return // session ended
				}
				writeSSE(w, flusher, ev)
			}
		}
	}
}

// DonePayingHandler finalizes payment: it ends the session (closes the relay,
// releases the coinslot) and executes the purchase webhook with the accumulated
// total, then redirects the client to the purchase callback.
func DonePayingHandler(api sdkapi.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := api.Http().Response()
		ctx := r.Context()

		clnt, err := api.Http().GetClientDevice(r)
		if err != nil {
			res.Error(w, r, err, http.StatusInternalServerError)
			return
		}

		c, err := FindUsedCoinslot(api, clnt.ID())
		if err != nil {
			res.Error(w, r, err, http.StatusInternalServerError)
			return
		}
		if c == nil {
			res.FlashMsg(w, r, api.Translate("error", "payment_failed"), sdkapi.FlashMsgError)
			res.RedirectToPortal(w, r)
			return
		}

		purchase, err := api.Payments().GetPurchaseRequest(r)
		if err != nil {
			res.Error(w, r, err, http.StatusInternalServerError)
			return
		}

		// Determine the total accumulated payment for the webhook.
		amount := purchase.Price()
		if state, err := purchase.State(ctx); err == nil {
			amount = state.TotalPayment
		}

		// End the session first so the relay is closed even if Execute fails.
		if mgr := GetManager(); mgr != nil {
			mgr.Sessions().End(c.GetID())
		} else {
			c.DoneUsing()
		}

		if err := purchase.Execute(ctx, sdkapi.ExecuteParams{
			Amount:  amount,
			Success: true,
			Message: "Payment completed",
		}); err != nil {
			res.Error(w, r, err, http.StatusInternalServerError)
			return
		}

		purchase.RedirectToCallback(w, r)
	}
}

// =============================================================================
// HELPER FUNCTIONS (internal)
// =============================================================================

func writeSSE(w http.ResponseWriter, flusher http.Flusher, ev CoinEvent) {
	b, err := json.Marshal(ev)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", b)
	flusher.Flush()
}
