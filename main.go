//go:build !mono

package main

import (
	"github.com/flarehotspot/com.flarego.wired-coinslot/app"
	plugin "github.com/flarehotspot/sdk/api/plugin"
)

func main() {}

func Init(api plugin.PluginApi) {
	app.Init(api)
}
