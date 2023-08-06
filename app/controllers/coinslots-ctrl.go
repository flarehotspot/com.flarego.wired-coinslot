package controllers

import (
	"net/http"

	"github.com/flarehotspot/com.flarego.wired-coinslot/app/models"
	"github.com/flarehotspot/sdk/v1/api"
)

type CoinslotsCtrl struct {
	api   api.IPluginApi
	model *models.WiredCoinslotModel
}

func (ctrl *CoinslotsCtrl) IndexPage(w http.ResponseWriter, r *http.Request) {
	http := ctrl.api.HttpApi()
	http.Respond().AdminView(w, r, "index.html", nil)
}

func NewCoinslotsCtrl(api api.IPluginApi, mdl *models.WiredCoinslotModel) *CoinslotsCtrl {
	return &CoinslotsCtrl{api, mdl}
}
