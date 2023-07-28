package controllers

import (
	"net/http"

	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
)

type CoinslotsCtrl struct {
	api   plugin.IPluginApi
	model *models.WiredCoinslotModel
}

func (t *CoinslotsCtrl) IndexPage(w http.ResponseWriter, r *http.Request) {
	http := t.api.HttpApi()
	http.Respond().AdminView(w, r, "index.html", nil, nil)
}

func NewCoinslotsCtrl(api plugin.IPluginApi, mdl *models.WiredCoinslotModel) *CoinslotsCtrl {
	return &CoinslotsCtrl{api, mdl}
}
