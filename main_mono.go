//go:build mono

package wiredcoinslot

import (
	"github.com/flarehotspot/sdk/v1.0.0/api/plugin"
	"github.com/flarehotspot/com.flarego.wired-coinslot/app"
)

func Init(api plugin.IPluginApi) {
  app.Init(api)
}
