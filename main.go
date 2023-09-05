//go:build !mono

package main

import (
	"github.com/flarehotspot/com.flarego.wired-coinslot/app"
	"github.com/flarehotspot/core/sdk/api/plugin"
)

func main() {}

func Init(api plugin.IPluginApi) {
	app.Init(api)
}
