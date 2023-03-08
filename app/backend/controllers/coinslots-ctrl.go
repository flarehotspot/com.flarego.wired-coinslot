package controllers

import (
	"net/http"

	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/wired-coinslot/app/backend/config"
)

type CoinslotsCtrl struct {
	api plugin.IPluginApi
	cfg *config.Config
}

func (t *CoinslotsCtrl) IndexPage(w http.ResponseWriter, r *http.Request) {
	http := t.api.HttpApi()
	cfg, err := t.cfg.Read()
	if err != nil {
		coinslots := []any{}
		http.Respond().AdminView(w, r, "index.html", map[string]any{"coinslots": coinslots})
    return
	}
	coinslots := cfg.Coinslots
	http.Respond().AdminView(w, r, "index.html", map[string]any{"coinslots": coinslots})
}

func NewCoinslotsCtrl(api plugin.IPluginApi, cfg *config.Config) *CoinslotsCtrl {
	return &CoinslotsCtrl{api, cfg}
}
