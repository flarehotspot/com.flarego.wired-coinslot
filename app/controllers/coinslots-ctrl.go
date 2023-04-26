package controllers

import (
	"net/http"

	// "github.com/flarehotspot/sdk/api/payments"
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/wired-coinslot/app/models"
)

type CoinslotsCtrl struct {
	api plugin.IPluginApi
  model *models.WiredCoinslotModel
}

func (t *CoinslotsCtrl) IndexPage(w http.ResponseWriter, r *http.Request) {
	http := t.api.HttpApi()
	// cfg, err := t.cfg.Read()
	// if err != nil {
		// coinslots := []any{}
		// http.Respond().AdminView(w, r, "index.html", map[string]any{"coinslots": coinslots})
    // return
	// }
	// coinslots := cfg.Coinslots

  // var qdata payments.

  // t.api.PaymentsApi().EmitEvent()

	http.Respond().AdminView(w, r, "index.html", map[string]any{"coinslots": []string{}})
}

func NewCoinslotsCtrl(api plugin.IPluginApi, mdl *models.WiredCoinslotModel) *CoinslotsCtrl {
	return &CoinslotsCtrl{api, mdl}
}
