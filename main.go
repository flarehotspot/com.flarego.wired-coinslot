//go:build !mono

package main

import (
	"github.com/flarehotspot/com.flarego.wired-coinslot/app"
	"github.com/flarehotspot/sdk/api/plugin"
)

func main() {}

func Init(api plugin.IPluginApi) {
	app.Init(api)
}
