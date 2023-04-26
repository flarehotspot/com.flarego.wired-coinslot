//go:build !mono

package main

import (
	"github.com/flarehotspot/sdk/api/plugin"
	"github.com/flarehotspot/wired-coinslot/app"
)

func main() {}

func Init(api plugin.IPluginApi) {
  app.Init(api)
}
