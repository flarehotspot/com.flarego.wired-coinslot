package controllers

import (
	"net/http"

	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	plugin "github.com/flarehotspot/core/sdk/api/plugin"
)

type CoinslotsCtrl struct {
	api   plugin.IPluginApi
	model *models.WiredCoinslotModel
}

func (ctrl *CoinslotsCtrl) IndexPage(w http.ResponseWriter, r *http.Request) {
	http := ctrl.api.HttpApi()
	http.HttpResponse().AdminView(w, r, "index.html", nil)
}

func NewCoinslotsCtrl(api plugin.IPluginApi, mdl *models.WiredCoinslotModel) *CoinslotsCtrl {
	return &CoinslotsCtrl{api, mdl}
}
