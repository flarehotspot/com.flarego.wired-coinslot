//go:build dev

package src

import (
	"net/http"

	sdkapi "sdk/api"
)

// RegisterMockRoutes mounts the synthetic-coin endpoint. Dev builds only: it lets
// us exercise the full pipeline (pulse counter -> payment session -> SSE) without
// a real coin acceptor wired to the GPIO. In production this is a no-op (see
// mock_prod.go) so the only way to pay is a real coin pulse from the hardware.
func RegisterMockRoutes(api sdkapi.IPluginApi, subrouter sdkapi.IHttpRouterInstance) {
	subrouter.Post("/mock-coin/{id}", MockCoinHandler(api)).Name("payments.mock_coin")
}

// MockCoinURL returns the synthetic-coin URL for the insert-coin page so the dev
// "Mock Insert Coin" button can target it. Only defined here (dev); the prod build
// returns "" so the view renders no button — the single source of dev/prod truth.
func MockCoinURL(api sdkapi.IPluginApi, coinslotID string) string {
	return api.Http().Helpers().UrlForRoute("payments.mock_coin", "id", coinslotID)
}

// MockCoinHandler injects a single synthetic coin pulse into the coinslot's
// counter, simulating the hardware agent emitting a pulse event.
func MockCoinHandler(api sdkapi.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		coinslotID := api.Http().MuxVars(r)["id"]
		if mgr := GetManager(); mgr != nil {
			mgr.InjectPulse(coinslotID)
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
