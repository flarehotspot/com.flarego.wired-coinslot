package src

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	sdkapi "sdk/api"

	"com.flarego.wired-coinslot/resources/views"
	"com.flarego.wired-coinslot/src/gpio"
)

func ListCoinslotsHandler(api sdkapi.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := api.Http().Response()

		coinslots, err := GetAllWiredCoinslots(api)
		if err != nil {
			_ = api.Logger().Error("[wired-coinslot] " + err.Error())
			res.FlashMsg(w, r, api.Translate("error", "Failed to load wired coinslots"), sdkapi.FlashMsgError)
			coinslots = nil
		}

		items := make([]views.CoinslotSettingsData, 0, len(coinslots))
		for _, c := range coinslots {
			if c == nil {
				continue
			}
			items = append(items, views.CoinslotSettingsData{
				ID:                 c.ID,
				Name:               c.Name,
				CoinPin:            c.CoinPin,
				RelayPin:           c.RelayPin,
				RelayActive:        c.RelayActive,
				Pull:               c.Pull,
				Edge:               c.Edge,
				DebounceMs:         c.DebounceMs,
				WindowMs:           c.WindowMs,
				PaymentTimeoutSecs: c.PaymentTimeoutSecs,
				BoardModel:         c.BoardModel,
				Denominations:      denominationsToViewData(c.Denominations),
			})
		}

		deviceModel := ""
		if mgr := GetManager(); mgr != nil {
			deviceModel = mgr.DeviceModel()
		}

		res.AdminView(w, r, sdkapi.ViewPage{
			PageContent: views.CoinslotSettingsPage(api, deviceModel, gpio.BoardModels(), items),
		})
	}
}

func SaveCoinslotSettingsHandler(api sdkapi.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := api.Http().Response()
		redirectURL := api.Http().Helpers().UrlForRoute("admin:wired-coinslots:index")
		id := api.Http().MuxVars(r)["id"]

		if err := r.ParseForm(); err != nil {
			res.FlashMsg(w, r, api.Translate("error", "Failed to parse form"), sdkapi.FlashMsgError)
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}

		c, err := LoadWiredCoinslot(api, id)
		if err != nil {
			res.FlashMsg(w, r, api.Translate("error", "Coinslot not found"), sdkapi.FlashMsgError)
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}

		if name := strings.TrimSpace(r.FormValue("name")); name != "" {
			c.Name = name
		}
		c.CoinPin = atoiDefault(r.FormValue("coin_pin"), c.CoinPin)
		c.RelayPin = atoiDefault(r.FormValue("relay_pin"), c.RelayPin)
		c.RelayActive = atoiDefault(r.FormValue("relay_active"), c.RelayActive)
		c.DebounceMs = atoiDefault(r.FormValue("debounce_ms"), c.DebounceMs)
		c.WindowMs = atoiDefault(r.FormValue("window_ms"), c.WindowMs)
		c.PaymentTimeoutSecs = atoiDefault(r.FormValue("payment_timeout_secs"), c.PaymentTimeoutSecs)
		c.Pull = r.FormValue("pull")
		c.Edge = r.FormValue("edge")
		c.BoardModel = strings.TrimSpace(r.FormValue("board_model"))

		if err := c.Save(); err != nil {
			res.FlashMsg(w, r, err.Error(), sdkapi.FlashMsgError)
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}

		// Rebuild the hardware runtime with the new settings.
		if mgr := GetManager(); mgr != nil {
			mgr.Reload(id)
		}

		res.FlashMsg(w, r, api.Translate("success", "Coinslot settings saved"), sdkapi.FlashMsgSuccess)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	}
}

// AddDenominationHandler appends a new pulses→amount denomination to a coinslot.
func AddDenominationHandler(api sdkapi.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := api.Http().Response()
		redirectURL := api.Http().Helpers().UrlForRoute("admin:wired-coinslots:index")
		id := api.Http().MuxVars(r)["id"]

		c, pulses, amount, ok := parseDenominationForm(api, w, r, redirectURL, id)
		if !ok {
			return
		}

		for _, d := range c.Denominations {
			if d.Pulses == pulses {
				res.FlashMsg(w, r, api.Translate("error", "A denomination with this pulse count already exists"), sdkapi.FlashMsgError)
				http.Redirect(w, r, redirectURL, http.StatusSeeOther)
				return
			}
		}

		c.Denominations = append(c.Denominations, Denomination{Pulses: pulses, Amount: amount})
		sortDenominations(c.Denominations)
		saveAndReload(api, w, r, redirectURL, c, api.Translate("success", "Denomination added successfully"))
	}
}

// UpdateDenominationHandler edits the denomination at the given index.
func UpdateDenominationHandler(api sdkapi.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := api.Http().Response()
		redirectURL := api.Http().Helpers().UrlForRoute("admin:wired-coinslots:index")
		id := api.Http().MuxVars(r)["id"]

		index, err := strconv.Atoi(api.Http().MuxVars(r)["index"])
		if err != nil {
			res.FlashMsg(w, r, api.Translate("error", "Invalid denomination index"), sdkapi.FlashMsgError)
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}

		c, pulses, amount, ok := parseDenominationForm(api, w, r, redirectURL, id)
		if !ok {
			return
		}

		if index < 0 || index >= len(c.Denominations) {
			res.FlashMsg(w, r, api.Translate("error", "Denomination not found"), sdkapi.FlashMsgError)
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}

		for i, d := range c.Denominations {
			if i != index && d.Pulses == pulses {
				res.FlashMsg(w, r, api.Translate("error", "A denomination with this pulse count already exists"), sdkapi.FlashMsgError)
				http.Redirect(w, r, redirectURL, http.StatusSeeOther)
				return
			}
		}

		c.Denominations[index].Pulses = pulses
		c.Denominations[index].Amount = amount
		sortDenominations(c.Denominations)
		saveAndReload(api, w, r, redirectURL, c, api.Translate("success", "Denomination updated successfully"))
	}
}

// DeleteDenominationHandler removes the denomination at the given index.
func DeleteDenominationHandler(api sdkapi.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := api.Http().Response()
		redirectURL := api.Http().Helpers().UrlForRoute("admin:wired-coinslots:index")
		id := api.Http().MuxVars(r)["id"]

		index, err := strconv.Atoi(api.Http().MuxVars(r)["index"])
		if err != nil {
			res.FlashMsg(w, r, api.Translate("error", "Invalid denomination index"), sdkapi.FlashMsgError)
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}

		c, err := LoadWiredCoinslot(api, id)
		if err != nil {
			res.FlashMsg(w, r, api.Translate("error", "Coinslot not found"), sdkapi.FlashMsgError)
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}

		if index < 0 || index >= len(c.Denominations) {
			res.FlashMsg(w, r, api.Translate("error", "Denomination not found"), sdkapi.FlashMsgError)
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}

		if len(c.Denominations) <= 1 {
			res.FlashMsg(w, r, api.Translate("error", "Cannot delete the last denomination"), sdkapi.FlashMsgError)
			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
			return
		}

		c.Denominations = append(c.Denominations[:index], c.Denominations[index+1:]...)
		saveAndReload(api, w, r, redirectURL, c, api.Translate("success", "Denomination deleted successfully"))
	}
}

// =============================================================================
// HELPER FUNCTIONS (internal)
// =============================================================================

func atoiDefault(s string, fallback int) int {
	if v, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return v
	}
	return fallback
}

// parseDenominationForm loads the coinslot and validates the pulses/amount form
// fields shared by the add and update handlers. On any error it writes a flash
// message, redirects, and returns ok=false.
func parseDenominationForm(api sdkapi.IPluginApi, w http.ResponseWriter, r *http.Request, redirectURL, id string) (c *WiredCoinslot, pulses int, amount float64, ok bool) {
	res := api.Http().Response()

	if err := r.ParseForm(); err != nil {
		res.FlashMsg(w, r, api.Translate("error", "Failed to parse form"), sdkapi.FlashMsgError)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return nil, 0, 0, false
	}

	pulses, err := strconv.Atoi(strings.TrimSpace(r.FormValue("pulses")))
	if err != nil || pulses <= 0 {
		res.FlashMsg(w, r, api.Translate("error", "Pulse count must be a positive number"), sdkapi.FlashMsgError)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return nil, 0, 0, false
	}

	amount, err = strconv.ParseFloat(strings.TrimSpace(r.FormValue("amount")), 64)
	if err != nil || amount <= 0 {
		res.FlashMsg(w, r, api.Translate("error", "Amount must be a positive number"), sdkapi.FlashMsgError)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return nil, 0, 0, false
	}

	c, err = LoadWiredCoinslot(api, id)
	if err != nil {
		res.FlashMsg(w, r, api.Translate("error", "Coinslot not found"), sdkapi.FlashMsgError)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return nil, 0, 0, false
	}

	return c, pulses, amount, true
}

// saveAndReload persists the coinslot, rebuilds its hardware runtime so the
// pulse counter picks up the new denominations, and flashes the result.
func saveAndReload(api sdkapi.IPluginApi, w http.ResponseWriter, r *http.Request, redirectURL string, c *WiredCoinslot, successMsg string) {
	res := api.Http().Response()
	if err := c.Save(); err != nil {
		_ = api.Logger().Error("[wired-coinslot] " + err.Error())
		res.FlashMsg(w, r, api.Translate("error", "Failed to save denomination"), sdkapi.FlashMsgError)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}
	if mgr := GetManager(); mgr != nil {
		mgr.Reload(c.ID)
	}
	res.FlashMsg(w, r, successMsg, sdkapi.FlashMsgSuccess)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// sortDenominations orders denominations by ascending pulse count.
func sortDenominations(denoms []Denomination) {
	sort.Slice(denoms, func(i, j int) bool {
		return denoms[i].Pulses < denoms[j].Pulses
	})
}

// denominationsToViewData converts config denominations into indexed view rows.
func denominationsToViewData(denoms []Denomination) []views.DenominationData {
	out := make([]views.DenominationData, len(denoms))
	for i, d := range denoms {
		out[i] = views.DenominationData{Index: i, Pulses: d.Pulses, Amount: d.Amount}
	}
	return out
}
