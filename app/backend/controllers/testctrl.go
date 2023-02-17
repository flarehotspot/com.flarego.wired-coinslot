package controllers

import (
	"github.com/flarehotspot/sdk/api/plugin"
	"net/http"
)

type TestCtrl struct {
	api plugin.IPluginApi
}

func (t *TestCtrl) IndexPage(w http.ResponseWriter, r *http.Request) {
	http := t.api.HttpApi()
	http.Respond().PortalView(w, 200, "index.html", map[string]any{"title": "Hello World"})
}

func NewTestCtrl(g plugin.IPluginApi) *TestCtrl {
	return &TestCtrl{g}
}
